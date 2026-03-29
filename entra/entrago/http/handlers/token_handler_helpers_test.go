package handlers_test

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"

	"identity/app"
	"identity/domain"
	"identity/http/handlers"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenHelpers_ResolveClientFromID(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(fixture.application.Config, fixture.tenantID)
	if err != nil {
		const fmtErr = "FindTenantByID: %v"

		t.Fatalf(fmtErr, err)
	}

	if client := handlers.ExportResolveClientFromID(tenant, ""); client != nil {
		t.Fatal("expected nil for empty client ID")
	}

	if client := handlers.ExportResolveClientFromID(tenant, "not-a-uuid"); client != nil {
		t.Fatal("expected nil for invalid client ID")
	}

	const unknownID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"

	if client := handlers.ExportResolveClientFromID(tenant, unknownID); client != nil {
		t.Fatal("expected nil for unknown client ID")
	}

	client := handlers.ExportResolveClientFromID(tenant, fixture.appID.UUID().String())
	if client == nil {
		t.Fatal("expected client")
	}

	if client.ClientID().UUID().String() != fixture.appID.UUID().String() {
		t.Fatalf("expected client ID %s, got %s", fixture.appID.UUID().String(), client.ClientID().UUID().String())
	}
}

func TestTokenHelpers_ResolveClientFromForm_WithoutSecretClient(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(fixture.application.Config, fixture.tenantID)
	if err != nil {
		const fmtErr = "FindTenantByID: %v"

		t.Fatalf(fmtErr, err)
	}

	client := handlers.ExportResolveClientFromForm(tenant, fixture.appID.UUID().String(), "test-secret")
	if client == nil {
		t.Fatal("expected client")
	}

	if client.ClientID().UUID().String() != fixture.appID.UUID().String() {
		t.Fatalf("expected client ID %s, got %s", fixture.appID.UUID().String(), client.ClientID().UUID().String())
	}
}

func TestTokenHelpers_ResolveClientFromForm_WithSecretValidation(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(fixture.application.Config, fixture.tenantID)
	if err != nil {
		const fmtErr = "FindTenantByID: %v"

		t.Fatalf(fmtErr, err)
	}

	if got := handlers.ExportResolveClientFromForm(tenant, fixture.appID.UUID().String(), "wrong"); got != nil {
		t.Fatal("expected nil client for invalid secret")
	}

	const secret = "test-secret" // fixture has test-secret for appID

	got := handlers.ExportResolveClientFromForm(tenant, fixture.appID.UUID().String(), secret)
	if got == nil {
		t.Fatal("expected client for valid secret")
	}
}

func TestTokenHelpers_FirstOf(t *testing.T) {
	t.Parallel()

	const fallback = "fallback"

	if got := handlers.ExportFirstOf("primary", fallback); got != "primary" {
		t.Fatalf("expected primary, got %q", got)
	}

	const empty = ""

	if got := handlers.ExportFirstOf(empty, fallback); got != fallback {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestTokenHelpers_CorrelationID(t *testing.T) {
	t.Parallel()

	const emptyBody = ""

	request := newTestRequest(t, "POST", "https://example.test/token", emptyBody)
	request.Header.Set("Client-Request-Id", "cid-123")

	if got := handlers.ExportCorrelationID(request); got != "cid-123" {
		t.Fatalf("expected header correlation ID, got %q", got)
	}

	request = newTestRequest(t, "POST", "https://example.test/token", emptyBody)

	generated := handlers.ExportCorrelationID(request)
	if generated == "" {
		t.Fatal("expected generated correlation ID")
	}
}

// --- New Tests & Helpers for domain.Parse ---

// TestCustomClaims represents the clean domain struct you want to extract
// without leaking the jwt library's types across your app.
type TestCustomClaims struct {
	Subject string `json:"sub"`
	Tenant  string `json:"tid"`
}

func TestDomain_Parse(t *testing.T) {
	t.Parallel()

	// 1. Generate a dummy RSA key for signing the test token
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Create a valid mock JWT to feed into your system
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "user-123",
		"tid": "tenant-abc",
	})
	rawJWT, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	// 3. Wrap it in your domain.AccessToken
	accessToken := domain.MustAccessToken(rawJWT)

	// 4. PARSE IT!
	// We explicitly tell Parse we expect it to extract a *TestCustomClaims pointer.
	extractedClaims, err := domain.ParseTokenAdapter[*TestCustomClaims](accessToken, func(rawToken string) (*TestCustomClaims, error) {
		// messy framework lifting is strictly safely contained in this closure callback
		var mapClaims jwt.MapClaims
		_, err := jwt.ParseWithClaims(rawToken, &mapClaims, func(t *jwt.Token) (any, error) {
			return &privateKey.PublicKey, nil
		})
		if err != nil {
			return nil, err
		}

		// Map the framework results to your clean struct
		return &TestCustomClaims{
			Subject: fmt.Sprint(mapClaims["sub"]),
			Tenant:  fmt.Sprint(mapClaims["tid"]),
		}, nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	// 5. Assert the mapping was successful
	if extractedClaims.Subject != "user-123" {
		t.Fatalf("expected subject 'user-123', got %q", extractedClaims.Subject)
	}
	if extractedClaims.Tenant != "tenant-abc" {
		t.Fatalf("expected tenant 'tenant-abc', got %q", extractedClaims.Tenant)
	}
}

func TestDomain_Parse_OtherTokens(t *testing.T) {
	t.Parallel()

	t.Run("IDToken", func(t *testing.T) {
		token := domain.MustIDToken("test-id-token")
		got, err := domain.ParseIdTokenAdapter[string](token, func(s string) (string, error) {
			return "parsed-" + s, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if got != "parsed-test-id-token" {
			t.Errorf("expected parsed-test-id-token, got %q", got)
		}
	})

	t.Run("RefreshToken", func(t *testing.T) {
		token := domain.MustRefreshToken("test-refresh-token")
		got, err := domain.ParseIdTokenAdapter[string](token, func(s string) (string, error) {
			return "parsed-" + s, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if got != "parsed-test-refresh-token" {
			t.Errorf("expected parsed-test-refresh-token, got %q", got)
		}
	})

	t.Run("AuthCode", func(t *testing.T) {
		token := domain.MustAuthCode("test-auth-code")
		got, err := domain.Parse[string](token.RawString(), func(s string) (string, error) {
			return "parsed-" + s, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if got != "parsed-test-auth-code" {
			t.Errorf("expected parsed-test-auth-code, got %q", got)
		}
	})

	t.Run("ClientInfo", func(t *testing.T) {
		token := domain.MustClientInfo("test-client-info")
		got, err := domain.Parse[string](token.RawString(), func(s string) (string, error) {
			return "parsed-" + s, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if got != "parsed-test-client-info" {
			t.Errorf("expected parsed-test-client-info, got %q", got)
		}
	})
}
