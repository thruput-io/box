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

	if client := handlers.ExportResolveClientFromID(*tenant, ""); client != nil {
		t.Fatal("expected nil for empty client ID")
	}

	if client := handlers.ExportResolveClientFromID(*tenant, "not-a-uuid"); client != nil {
		t.Fatal("expected nil for invalid client ID")
	}

	const unknownID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"

	if client := handlers.ExportResolveClientFromID(*tenant, unknownID); client != nil {
		t.Fatal("expected nil for unknown client ID")
	}

	client := handlers.ExportResolveClientFromID(*tenant, parseString(fixture.appID))
	if client == nil {
		t.Fatal("expected client")
	}

	if parseString(client.ClientID()) != parseString(fixture.appID) {
		t.Fatalf("expected client ID %s, got %s", parseString(fixture.appID), parseString(client.ClientID()))
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

	client := handlers.ExportResolveClientFromForm(*tenant, parseString(fixture.appID), "test-secret")
	if client == nil {
		t.Fatal("expected client")
	}

	if parseString(client.ClientID()) != parseString(fixture.appID) {
		t.Fatalf("expected client ID %s, got %s", parseString(fixture.appID), parseString(client.ClientID()))
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

	if got := handlers.ExportResolveClientFromForm(*tenant, parseString(fixture.appID), "wrong"); got != nil {
		t.Fatal("expected nil client for invalid secret")
	}

	const secret = "test-secret" // fixture has test-secret for appID

	got := handlers.ExportResolveClientFromForm(*tenant, parseString(fixture.appID), secret)
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

// --- New Tests & Helpers for domain.SafeTo ---

// TestCustomClaims represents the clean domain struct you want to extract
// without leaking the jwt library's types across your app.
type TestCustomClaims struct {
	Subject string `json:"sub"`
	Tenant  string `json:"tid"`
}

func TestDomain_Parse(t *testing.T) {
	t.Parallel()

	privateKey := generateTestRSAKey(t)
	rawJWT := createTestJWT(t, privateKey)
	accessToken := domain.MustAccessToken(rawJWT)

	rawToken := accessToken.Value()

	var mapClaims jwt.MapClaims

	_, err := jwt.ParseWithClaims(rawToken, &mapClaims, func(_ *jwt.Token) (any, error) {
		return &privateKey.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("jwt.ParseWithClaims: %v", err)
	}

	extractedClaims := &TestCustomClaims{
		Subject: fmt.Sprint(mapClaims["sub"]),
		Tenant:  fmt.Sprint(mapClaims["tid"]),
	}

	if extractedClaims.Subject != "user-123" {
		t.Fatalf("expected subject 'user-123', got %q", extractedClaims.Subject)
	}

	if extractedClaims.Tenant != "tenant-abc" {
		t.Fatalf("expected tenant 'tenant-abc', got %q", extractedClaims.Tenant)
	}
}

func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	const rsaKeySize = 2048

	privateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		t.Fatal(err)
	}

	return privateKey
}

func createTestJWT(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "user-123",
		"tid": "tenant-abc",
	})

	rawJWT, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}

	return rawJWT
}
