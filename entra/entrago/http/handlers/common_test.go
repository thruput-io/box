package handlers_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
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
	firstIndex      = 0

	// Mock utility paths for unit tests.
	pathMockUtils     = "/mock-utils/"
	pathMockUtilsSign = "/mock-utils/sign"

	// Common test server URL.
	testServerURL = "http://entra.test"
)

func mustGroupName(t *testing.T, raw string) domain.GroupName {
	t.Helper()

	return domain.NewGroupName(raw).MustRight()
}

func mustRoleValue(t *testing.T, raw string) domain.RoleValue {
	t.Helper()

	return domain.NewRoleValue(raw).MustRight()
}

func mustIdentifierURI(t *testing.T, raw string) domain.IdentifierURI {
	t.Helper()

	return domain.NewIdentifierURI(raw).MustRight()
}

func mustRedirectURL(t *testing.T, raw string) domain.RedirectURL {
	t.Helper()

	return domain.NewRedirectURL(raw).MustRight()
}

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

func newTestRequest(t *testing.T, method, url, body string) *http.Request {
	t.Helper()

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("newTestRequest: %v", err)
	}

	return req
}

func parseString(v interface{ Value() string }) string {
	return v.Value()
}

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	const rsaBits = 2048

	key, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	return key
}
