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

func mustTenantID(t *testing.T, raw string) domain.TenantID {
	t.Helper()

	return domain.NewTenantID(raw).MustRight()
}

func mustTenantName(t *testing.T, raw string) domain.TenantName {
	t.Helper()

	return domain.NewTenantName(raw).MustRight()
}

func mustClientID(t *testing.T, raw string) domain.ClientID {
	t.Helper()

	return domain.NewClientID(raw).MustRight()
}

func mustClientSecret(t *testing.T, raw string) domain.ClientSecret {
	t.Helper()

	return domain.NewClientSecret(raw).MustRight()
}

func mustUserID(t *testing.T, raw string) domain.UserID {
	t.Helper()

	return domain.NewUserID(raw).MustRight()
}

func mustUsername(t *testing.T, raw string) domain.Username {
	t.Helper()

	return domain.NewUsername(raw).MustRight()
}

func mustPassword(t *testing.T, raw string) domain.Password {
	t.Helper()

	return domain.NewPassword(raw).MustRight()
}

func mustAppName(t *testing.T, raw string) domain.AppName {
	t.Helper()

	return domain.NewAppName(raw).MustRight()
}

func mustDisplayName(t *testing.T, raw string) domain.DisplayName {
	t.Helper()

	return domain.NewDisplayName(raw).MustRight()
}

func mustEmail(t *testing.T, raw string) domain.Email {
	t.Helper()

	return domain.NewEmail(raw).MustRight()
}

func mustNonce(t *testing.T, raw string) domain.Nonce {
	t.Helper()

	return domain.NewNonce(raw).MustRight()
}

func mustCorrelationID(t *testing.T, raw string) domain.CorrelationID {
	t.Helper()

	return domain.NewCorrelationID(raw).MustRight()
}

func mustScopeID(t *testing.T, raw string) domain.ScopeID {
	t.Helper()

	return domain.NewScopeID(raw).MustRight()
}

func mustRoleID(t *testing.T, raw string) domain.RoleID {
	t.Helper()

	return domain.NewRoleID(raw).MustRight()
}

func mustRedirectURL(t *testing.T, raw string) domain.RedirectURL {
	t.Helper()

	return domain.NewRedirectURL(raw).MustRight()
}

func mustIdentifierURI(t *testing.T, raw string) domain.IdentifierURI {
	t.Helper()

	return domain.NewIdentifierURI(raw).MustRight()
}

func mustGroupName(t *testing.T, raw string) domain.GroupName {
	t.Helper()

	return domain.NewGroupName(raw).MustRight()
}

func mustScopeValue(t *testing.T, raw string) domain.ScopeValue {
	t.Helper()

	return domain.NewScopeValue(raw).MustRight()
}

func mustScopeDescription(t *testing.T, raw string) domain.ScopeDescription {
	t.Helper()

	return domain.NewScopeDescription(raw).MustRight()
}

func mustRoleValue(t *testing.T, raw string) domain.RoleValue {
	t.Helper()

	return domain.NewRoleValue(raw).MustRight()
}

func mustRoleDescription(t *testing.T, raw string) domain.RoleDescription {
	t.Helper()

	return domain.NewRoleDescription(raw).MustRight()
}

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, testKeySize)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	return key
}

func parseString(v interface{ Value() string }) string {
	return v.Value()
}
