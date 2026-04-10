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

	v, err := domain.NewIdentifierURI(raw)
	if err != nil {
		t.Fatalf("NewIdentifierURI(%q): %v", raw, err)
	}

	return v
}

func mustRedirectURL(t *testing.T, raw string) domain.RedirectURL {
	t.Helper()

	v, err := domain.NewRedirectURL(raw)
	if err != nil {
		t.Fatalf("NewRedirectURL(%q): %v", raw, err)
	}

	return v
}

type testHandler struct {
	called bool
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.called = true

	w.WriteHeader(http.StatusOK)
}
