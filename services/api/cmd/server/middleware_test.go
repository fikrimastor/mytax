package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fikrimastor/mytax/services/api/internal/policy"
	"github.com/labstack/echo/v4"
)

func bodyLimitServer() *echo.Echo {
	e := echo.New()
	useMiddleware(e, "")
	e.POST("/x", func(c echo.Context) error {
		var in struct {
			Claims []policy.Claim `json:"claims"`
		}
		if err := c.Bind(&in); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
	return e
}

func post(e *echo.Echo, body []byte) int {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	e.ServeHTTP(rec, req)
	return rec.Code
}

// A request at policy's 1,000-claim limit, every claim carrying the largest fact set in the
// catalog, must still fit under the body limit.
func TestBodyLimitAdmitsAFullClaimList(t *testing.T) {
	facts := map[string]bool{}
	for _, r := range policy.YA2025Draft().Rules {
		if len(r.RequiredFacts) > len(facts) {
			facts = map[string]bool{}
			for _, f := range r.RequiredFacts {
				facts[f] = true
			}
		}
	}
	claims := make([]policy.Claim, 1000)
	for i := range claims {
		claims[i] = policy.Claim{ID: fmt.Sprintf("claim-%04d", i), RuleID: "MEDICAL_LEARNING_DISABILITY", AmountSen: 600000,
			ExpenseKey: fmt.Sprintf("receipt-%04d-0123456789abcdef", i), EvidenceReference: fmt.Sprintf("evidence-%04d-0123456789abcdef", i), Facts: facts}
	}
	body, _ := json.Marshal(map[string]any{"assessment_year": 2025, "claims": claims})
	if code := post(bodyLimitServer(), body); code != http.StatusNoContent {
		t.Fatalf("full claim list (%d bytes) got %d", len(body), code)
	}
}

func TestBodyLimitRejectsOversizedBody(t *testing.T) {
	body := append([]byte(`{"claims":[],"pad":"`), bytes.Repeat([]byte("a"), 1<<20)...)
	body = append(body, `"}`...)
	if code := post(bodyLimitServer(), body); code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body got %d, want 413", code)
	}
}
