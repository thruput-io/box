package handlers_test

import (
	"testing"

	"identity/domain"
)

func TestDomain_Parse_IDToken(t *testing.T) {
	t.Parallel()

	token := domain.MustIDToken("test-id-token")

	if got := token.Value(); got != "test-id-token" {
		t.Errorf("expected test-id-token, got %q", got)
	}
}

func TestDomain_Parse_RefreshToken(t *testing.T) {
	t.Parallel()

	token := domain.MustRefreshToken("test-refresh-token")

	if got := token.Value(); got != "test-refresh-token" {
		t.Errorf("expected test-refresh-token, got %q", got)
	}
}

func TestDomain_Parse_AuthCode(t *testing.T) {
	t.Parallel()

	token := domain.MustAuthCode("test-auth-code")

	if got := token.Value(); got != "test-auth-code" {
		t.Errorf("expected test-auth-code, got %q", got)
	}
}

func TestDomain_Parse_ClientInfo(t *testing.T) {
	t.Parallel()

	token := domain.MustClientInfo("test-client-info")

	if got := token.Value(); got != "test-client-info" {
		t.Errorf("expected test-client-info, got %q", got)
	}
}
