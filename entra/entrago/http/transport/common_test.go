package transport_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"testing"

	"identity/domain"
)

const (
	testAppURI      = "api://app"
	testCallbackURI = "https://example.com/callback"
	testUsername    = "testuser"
	testPassword    = "testpass"
	exampleCom      = "https://example.com"
	keySize         = 1024
)

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	return key
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

type testHandler struct {
	called bool
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.called = true

	w.WriteHeader(http.StatusOK)
}
