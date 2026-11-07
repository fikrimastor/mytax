package policy

import (
	"strconv"
	"strings"
	"testing"
)

func TestDraftPackageServesYA2025Draft(t *testing.T) {
	pkg, err := DraftPackage(2025)
	if err != nil {
		t.Fatalf("DraftPackage(2025): %v", err)
	}
	if pkg.Version != "ya2025-draft-001" || pkg.Status != Draft || pkg.AssessmentYear != 2025 {
		t.Fatalf("got version=%q status=%q year=%d", pkg.Version, pkg.Status, pkg.AssessmentYear)
	}
	if len(pkg.Rules) == 0 {
		t.Fatal("draft package has no rules")
	}
}

func TestDraftPackageFailsClosedForUnsupportedYears(t *testing.T) {
	for _, year := range []int{2024, 2026, 0, -1} {
		t.Run(strconv.Itoa(year), func(t *testing.T) {
			pkg, err := DraftPackage(year)
			if err == nil {
				t.Fatalf("DraftPackage(%d) returned a package: %+v", year, pkg)
			}
			if !strings.Contains(err.Error(), "no draft rule package for assessment year") {
				t.Fatalf("unexpected error: %v", err)
			}
			if pkg.Version != "" || len(pkg.Rules) != 0 {
				t.Fatalf("error result leaked a package: %+v", pkg)
			}
		})
	}
}
