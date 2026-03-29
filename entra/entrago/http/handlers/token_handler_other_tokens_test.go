package handlers_test

import (
	"testing"

	"identity/domain"
)

func TestDomain_Parse_IDToken(t *testing.T) {
	t.Parallel()

	token := domain.MustIDToken("test-id-token")

	got, err := domain.Parse[string](token, func(s string) (string, error) {
		return "parsed-" + s, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if got != "parsed-test-id-token" {
		t.Errorf("expected parsed-test-id-token, got %q", got)
	}
}

func TestDomain_Parse_RefreshToken(t *testing.T) {
	t.Parallel()

	token := domain.MustRefreshToken("test-refresh-token")

	got, err := domain.Parse[string](token, func(s string) (string, error) {
		return "parsed-" + s, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if got != "parsed-test-refresh-token" {
		t.Errorf("expected parsed-test-refresh-token, got %q", got)
	}
}

func TestDomain_Parse_AuthCode(t *testing.T) {
	t.Parallel()

	token := domain.MustAuthCode("test-auth-code")

	got, err := domain.Parse[string](token, func(s string) (string, error) {
		return "parsed-" + s, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if got != "parsed-test-auth-code" {
		t.Errorf("expected parsed-test-auth-code, got %q", got)
	}
}

func TestDomain_Parse_ClientInfo(t *testing.T) {
	t.Parallel()

	token := domain.MustClientInfo("test-client-info")

	got, err := domain.Parse[string](token, func(s string) (string, error) {
		return "parsed-" + s, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if got != "parsed-test-client-info" {
		t.Errorf("expected parsed-test-client-info, got %q", got)
	}
}
