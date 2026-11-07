package mcpserver

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fikrimastor/mytax/services/api/internal/policy"
)

// newSession starts an in-memory MCP session against a fresh server. The server must
// connect first; the client initializes the session during its own connect.
func newSession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ss, err := New("test").Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		ss.Close()
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		cs.Close()
		ss.Close()
	})
	return cs
}

func callList(t *testing.T, cs *mcp.ClientSession, year int) (*mcp.CallToolResult, ListOutput) {
	t.Helper()
	return call[ListOutput](t, cs, listToolName, map[string]any{"assessment_year": year})
}

func callRule(t *testing.T, cs *mcp.ClientSession, year int, ruleID string) (*mcp.CallToolResult, RuleOutput) {
	t.Helper()
	return call[RuleOutput](t, cs, ruleToolName, map[string]any{"assessment_year": year, "rule_id": ruleID})
}

// call invokes a tool and decodes StructuredContent into out.
func call[T any](t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) (*mcp.CallToolResult, T) {
	t.Helper()
	var out T
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: protocol error: %v", name, err)
	}
	if res.IsError {
		return res, out
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("%s: marshal structured content: %v", name, err)
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: unmarshal structured content %s: %v", name, raw, err)
	}
	return res, out
}

// errorText returns the first text content of a tool error result.
func errorText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatal("tool error result has no content")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("tool error content is %T, want *mcp.TextContent", res.Content[0])
	}
	return tc.Text
}

func TestListToolsExposesExactlyTheReadOnlyTools(t *testing.T) {
	cs := newSession(t)
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	got := make([]string, 0, len(res.Tools))
	for _, tool := range res.Tools {
		got = append(got, tool.Name)
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Errorf("tool %q is not marked read-only", tool.Name)
		}
		if tool.Annotations != nil && !tool.Annotations.IdempotentHint {
			t.Errorf("tool %q is not marked idempotent", tool.Name)
		}
		if tool.Annotations != nil && (tool.Annotations.OpenWorldHint == nil || *tool.Annotations.OpenWorldHint) {
			t.Errorf("tool %q is not marked closed-world", tool.Name)
		}
		if tool.Description == "" {
			t.Errorf("tool %q has no description", tool.Name)
		}
	}

	want := map[string]bool{listToolName: true, ruleToolName: true}
	if len(got) != len(want) {
		t.Fatalf("tools = %v, want exactly %d", got, len(want))
	}
	for _, name := range got {
		if !want[name] {
			t.Fatalf("unexpected tool %q (all: %v)", name, got)
		}
	}
}

func TestListReliefsServesTheDraftPackage(t *testing.T) {
	cs := newSession(t)
	pkg := policy.YA2025Draft()

	res, out := callList(t, cs, 2025)
	if res.IsError {
		t.Fatalf("list_tax_reliefs 2025 errored: %s", errorText(t, res))
	}
	if out.PackageStatus != policy.Draft || out.PackageVersion != pkg.Version || out.AssessmentYear != 2025 {
		t.Fatalf("package = (%q, %q, %d), want (DRAFT, %q, 2025)", out.PackageStatus, out.PackageVersion, out.AssessmentYear, pkg.Version)
	}
	if out.FilingReady {
		t.Fatal("draft package reported as filing ready")
	}
	if out.Notice != Notice {
		t.Fatalf("notice = %q, want %q", out.Notice, Notice)
	}

	if len(out.Rules) != len(pkg.Rules) {
		t.Fatalf("got %d rules, want %d", len(out.Rules), len(pkg.Rules))
	}
	for i, r := range pkg.Rules {
		if out.Rules[i].ID != r.ID || out.Rules[i].LimitSen != r.LimitSen || out.Rules[i].Group != r.Group || out.Rules[i].ExclusiveFamily != r.ExclusiveFamily {
			t.Fatalf("rule %d = %+v, want %+v", i, out.Rules[i], r)
		}
	}

	if !equalGroups(out.Groups, pkg.Groups) {
		t.Fatalf("groups = %+v, want %+v", out.Groups, pkg.Groups)
	}

	registry := policy.ReliefRegistry2025()
	if len(out.RegistryEntries) != len(registry) {
		t.Fatalf("got %d registry entries, want %d", len(out.RegistryEntries), len(registry))
	}
}

func TestListReliefsRejectsUnsupportedYears(t *testing.T) {
	cs := newSession(t)
	for _, year := range []int{2024, 2026} {
		t.Run(strconv.Itoa(year), func(t *testing.T) {
			res, _ := callList(t, cs, year)
			if !res.IsError {
				t.Fatalf("list_tax_reliefs %d succeeded: %+v", year, res)
			}
			if text := errorText(t, res); !strings.HasPrefix(text, "RULE_UNAVAILABLE") {
				t.Fatalf("error = %q, want RULE_UNAVAILABLE prefix", text)
			}
		})
	}
}

func TestGetReliefRuleReturnsCapGroupAndSources(t *testing.T) {
	cs := newSession(t)
	res, out := callRule(t, cs, 2025, "LIFESTYLE_DEVICE")
	if res.IsError {
		t.Fatalf("get_tax_relief_rule errored: %s", errorText(t, res))
	}
	if out.Rule.ID != "LIFESTYLE_DEVICE" || out.Rule.LimitSen != 250000 {
		t.Fatalf("rule = %+v, want LIFESTYLE_DEVICE at 250000 sen", out.Rule)
	}
	if out.Group == nil || out.Group.ID != "LIFESTYLE" || out.Group.LimitSen != 250000 {
		t.Fatalf("group = %+v, want LIFESTYLE at 250000 sen", out.Group)
	}
	if out.FilingReady {
		t.Fatal("draft rule reported as filing ready")
	}
	if out.Notice != Notice {
		t.Fatalf("notice = %q, want %q", out.Notice, Notice)
	}
	if len(out.Rule.Sources) == 0 {
		t.Fatal("rule has no sources")
	}
	for _, s := range out.Rule.Sources {
		if !strings.HasPrefix(s.URL, "https://www.hasil.gov.my/") {
			t.Fatalf("source URL is not an official HASiL URL: %q", s.URL)
		}
	}
	if len(out.Rule.RequiredFacts) == 0 {
		t.Fatal("rule has no required facts")
	}
}

func TestGetReliefRuleWithoutSharedGroup(t *testing.T) {
	cs := newSession(t)
	res, out := callRule(t, cs, 2025, "FIRST_HOME_LOW")
	if res.IsError {
		t.Fatalf("get_tax_relief_rule errored: %s", errorText(t, res))
	}
	if out.Group != nil {
		t.Fatalf("group = %+v, want absent", out.Group)
	}
	if out.Rule.ExclusiveFamily != "FIRST_HOME" {
		t.Fatalf("exclusive_family = %q, want FIRST_HOME", out.Rule.ExclusiveFamily)
	}
}

func TestGetReliefRuleRejectsUnknownRuleAndYear(t *testing.T) {
	cs := newSession(t)
	res, _ := callRule(t, cs, 2025, "LIFESTYLE")
	if !res.IsError {
		t.Fatalf("unknown rule succeeded: %+v", res)
	}
	if text := errorText(t, res); !strings.HasPrefix(text, "UNKNOWN_RELIEF_RULE") {
		t.Fatalf("error = %q, want UNKNOWN_RELIEF_RULE prefix", text)
	}

	res, _ = callRule(t, cs, 2026, "LIFESTYLE_DEVICE")
	if !res.IsError {
		t.Fatalf("unsupported year succeeded: %+v", res)
	}
	if text := errorText(t, res); !strings.HasPrefix(text, "RULE_UNAVAILABLE") {
		t.Fatalf("error = %q, want RULE_UNAVAILABLE prefix", text)
	}
}

func equalGroups(a, b []policy.Group) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
