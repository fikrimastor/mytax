// Package mcpserver exposes the mytax draft relief catalog over the Model Context
// Protocol. It is a read-only adapter over internal/policy and holds no tax logic of
// its own: it never calculates, approves or evaluates a claim.
package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fikrimastor/mytax/services/api/internal/policy"
)

// Notice is attached to every successful response so a caller can never mistake the
// data for an approved or filing-ready result.
const Notice = "Developer preview. Draft rules, not reviewed by a tax professional and not approved by LHDN. Nothing here is an entitlement, an approved deduction or a filing-ready figure."

const (
	listToolName = "list_tax_reliefs"
	ruleToolName = "get_tax_relief_rule"

	errCodeRuleUnavailable = "RULE_UNAVAILABLE"
	errCodeUnknownRule     = "UNKNOWN_RELIEF_RULE"

	registryYear = 2025
)

// RuleSummary is the compact form of a draft rule returned by list_tax_reliefs.
type RuleSummary struct {
	ID              string `json:"id"`
	Group           string `json:"group,omitempty"`
	ExclusiveFamily string `json:"exclusive_family,omitempty"`
	LimitSen        int64  `json:"limit_sen"`
}

// ListOutput is the list_tax_reliefs response.
type ListOutput struct {
	AssessmentYear  int                    `json:"assessment_year"`
	PackageVersion  string                 `json:"package_version"`
	PackageStatus   policy.Status          `json:"package_status"`
	FilingReady     bool                   `json:"filing_ready"`
	Notice          string                 `json:"notice"`
	Groups          []policy.Group         `json:"groups"`
	Rules           []RuleSummary          `json:"rules"`
	RegistryEntries []policy.RegistryEntry `json:"registry_entries" jsonschema:"descriptive inventory of the official relief list; entries are not evaluable rules"`
}

// RuleOutput is the get_tax_relief_rule response.
type RuleOutput struct {
	AssessmentYear int           `json:"assessment_year"`
	PackageVersion string        `json:"package_version"`
	PackageStatus  policy.Status `json:"package_status"`
	FilingReady    bool          `json:"filing_ready"`
	Notice         string        `json:"notice"`
	Rule           policy.Rule   `json:"rule"`
	Group          *policy.Group `json:"group,omitempty" jsonschema:"shared cap this rule counts against, if any"`
}

type listInput struct {
	AssessmentYear int `json:"assessment_year" jsonschema:"Year of Assessment, e.g. 2025"`
}

type ruleInput struct {
	AssessmentYear int    `json:"assessment_year" jsonschema:"Year of Assessment, e.g. 2025"`
	RuleID         string `json:"rule_id" jsonschema:"Rule ID exactly as returned by list_tax_reliefs, e.g. LIFESTYLE_DEVICE"`
}

// New builds the read-only mytax MCP server. version is reported to clients as the
// server implementation version.
func New(version string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "mytax", Version: version}, &mcp.ServerOptions{
		Instructions: Notice + " Amounts are integer sen. Use rule IDs exactly as listed; an unsupported assessment year or unknown rule is an error, never a guess.",
	})
	mcp.AddTool(s, &mcp.Tool{
		Name: listToolName,
		Description: "Lists the draft relief rules the mytax engine can evaluate, their caps and shared groups, " +
			"plus the descriptive registry of reliefs not yet modelled. Nothing is approved.",
		Annotations: readOnlyAnnotations(),
	}, listReliefs)
	mcp.AddTool(s, &mcp.Tool{
		Name: ruleToolName,
		Description: "Returns one draft rule: its cap, shared group cap, mutually exclusive family, " +
			"the facts that must be verified, and its sources.",
		Annotations: readOnlyAnnotations(),
	}, getReliefRule)
	return s
}

func readOnlyAnnotations() *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		ReadOnlyHint:   true,
		IdempotentHint: true,
		OpenWorldHint:  new(bool), // false: only the local catalog is read
	}
}

func listReliefs(_ context.Context, _ *mcp.CallToolRequest, in listInput) (*mcp.CallToolResult, ListOutput, error) {
	pkg, err := policy.DraftPackage(in.AssessmentYear)
	if err != nil {
		return nil, ListOutput{}, ruleUnavailableError(in.AssessmentYear)
	}

	rules := make([]RuleSummary, 0, len(pkg.Rules))
	for _, r := range pkg.Rules {
		rules = append(rules, RuleSummary{ID: r.ID, Group: r.Group, ExclusiveFamily: r.ExclusiveFamily, LimitSen: r.LimitSen})
	}

	return nil, ListOutput{
		AssessmentYear:  pkg.AssessmentYear,
		PackageVersion:  pkg.Version,
		PackageStatus:   pkg.Status,
		FilingReady:     false,
		Notice:          Notice,
		Groups:          append([]policy.Group{}, pkg.Groups...),
		Rules:           rules,
		RegistryEntries: registryEntries(pkg.AssessmentYear),
	}, nil
}

func getReliefRule(_ context.Context, _ *mcp.CallToolRequest, in ruleInput) (*mcp.CallToolResult, RuleOutput, error) {
	pkg, err := policy.DraftPackage(in.AssessmentYear)
	if err != nil {
		return nil, RuleOutput{}, ruleUnavailableError(in.AssessmentYear)
	}

	rule, ok := findRule(pkg, in.RuleID)
	if !ok {
		return nil, RuleOutput{}, fmt.Errorf("%s: %q is not in package %s", errCodeUnknownRule, in.RuleID, pkg.Version)
	}

	out := RuleOutput{
		AssessmentYear: pkg.AssessmentYear,
		PackageVersion: pkg.Version,
		PackageStatus:  pkg.Status,
		FilingReady:    false,
		Notice:         Notice,
		Rule:           copyRule(rule),
		Group:          findGroup(pkg, rule.Group),
	}
	return nil, out, nil
}

func ruleUnavailableError(year int) error {
	return fmt.Errorf("%s: no rule package for assessment year %d", errCodeRuleUnavailable, year)
}

// registryEntries returns the descriptive relief inventory for the package year, or an
// empty non-nil slice when the year has no registry.
func registryEntries(year int) []policy.RegistryEntry {
	if year != registryYear {
		return []policy.RegistryEntry{}
	}
	return append([]policy.RegistryEntry{}, policy.ReliefRegistry2025()...)
}

func findRule(pkg policy.Package, id string) (policy.Rule, bool) {
	for _, r := range pkg.Rules {
		if r.ID == id {
			return r, true
		}
	}
	return policy.Rule{}, false
}

// findGroup returns the package group with the given ID, or nil when the rule belongs to
// no shared group.
func findGroup(pkg policy.Package, id string) *policy.Group {
	if id == "" {
		return nil
	}
	for _, g := range pkg.Groups {
		if g.ID == id {
			return &policy.Group{ID: g.ID, LimitSen: g.LimitSen}
		}
	}
	return nil
}

// copyRule deep-copies the slices so a response never aliases catalog state and never
// contains a nil slice (a null array would fail the output schema).
func copyRule(r policy.Rule) policy.Rule {
	r.RequiredFacts = append([]string{}, r.RequiredFacts...)
	r.Sources = append([]policy.Source{}, r.Sources...)
	return r
}
