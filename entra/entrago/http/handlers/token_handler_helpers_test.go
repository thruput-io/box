package handlers_test

import (
	"testing"

	"identity/app"
	"identity/http/handlers"
)

func TestTokenHelpers_ResolveClientFromID(t *testing.T) {
	t.Parallel()

	fixture := mustAppForTestTokenHandler(t)

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

	fixture := mustAppForTestTokenHandler(t)

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

	fixture := mustAppForTestTokenHandler(t)

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
