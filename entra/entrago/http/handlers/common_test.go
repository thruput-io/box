package handlers_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"identity/domain"
)

const (
	testAppURI      = "api://app"
	testCallbackURI = "https://example.com/callback"
	fmtStatus       = "expected %d, got %d"
	pathConfig      = "/config/"
	pathCsharp      = "/csharp"
)

func mustGroupName(t *testing.T, raw string) domain.GroupName {
	t.Helper()

	v, err := domain.NewGroupName(raw)
	if err != nil {
		t.Fatalf("NewGroupName(%q): %v", raw, err)
	}

	return v
}

func mustRoleValue(t *testing.T, raw string) domain.RoleValue {
	t.Helper()

	v, err := domain.NewRoleValue(raw)
	if err != nil {
		t.Fatalf("NewRoleValue(%q): %v", raw, err)
	}

	return v
}

func mustIdentifierURI(t *testing.T, raw string) domain.IdentifierURI {
	t.Helper()

	v, err := domain.NewIdentifierURI(raw)
	if err != nil {
		t.Fatalf("NewIdentifierURI(%q): %v", raw, err)
	}

	return v
}

func newTestRequest(t *testing.T, method, url, body string) *http.Request {
	t.Helper()

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("newTestRequest: %v", err)
	}

	return req
}

func parseString(p domain.RawValueProvider) string {
	s, _ := domain.Parse[string](p, func(v string) (string, error) { return v, nil })

	return s
}
