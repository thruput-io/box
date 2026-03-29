package app_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"identity/domain"
)

const (
	testTenantID = "11111111-1111-4111-8111-111111111111"
	testClientID = "22222222-2222-4222-8222-222222222222"
	testUserID   = "33333333-3333-4333-8333-333333333333"
	testUser     = "user"
	testPass     = "pass"
	testWrong    = "wrong"
	testCallback = "https://example.com/callback"
	testSecret   = "secret"
	testNonce    = "nonce"
	testCorr     = "corr"
	testScope    = "openid offline_access api://app/access"
	testKeySize  = 1024
	testApp      = "api://app"
)

func mustRedirectURL(t *testing.T, raw string) domain.RedirectURL {
	t.Helper()

	url, err := domain.NewRedirectURL(raw)
	if err != nil {
		t.Fatalf("NewRedirectURL(%q): %v", raw, err)
	}

	return url
}

func mustIdentifierURI(t *testing.T, raw string) domain.IdentifierURI {
	t.Helper()

	v, err := domain.NewIdentifierURI(raw)
	if err != nil {
		t.Fatalf("NewIdentifierURI(%q): %v", raw, err)
	}

	return v
}

func mustGroupName(t *testing.T, raw string) domain.GroupName {
	t.Helper()

	v, err := domain.NewGroupName(raw)
	if err != nil {
		t.Fatalf("NewGroupName(%q): %v", raw, err)
	}

	return v
}

func mustScopeValue(t *testing.T, raw string) domain.ScopeValue {
	t.Helper()

	v, err := domain.NewScopeValue(raw)
	if err != nil {
		t.Fatalf("NewScopeValue(%q): %v", raw, err)
	}

	return v
}

func mustScopeDescription(t *testing.T, raw string) domain.ScopeDescription {
	t.Helper()

	v, err := domain.NewScopeDescription(raw)
	if err != nil {
		t.Fatalf("NewScopeDescription(%q): %v", raw, err)
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

func mustRoleDescription(t *testing.T, raw string) domain.RoleDescription {
	t.Helper()

	v, err := domain.NewRoleDescription(raw)
	if err != nil {
		t.Fatalf("NewRoleDescription(%q): %v", raw, err)
	}

	return v
}

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, testKeySize)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	return key
}

func parseString(p domain.RawValueProvider) string {
	s, _ := domain.Parse[string](p, func(v string) (string, error) { return v, nil })
	return s
}
