package handlers

import (
	"testing"

	"identity/app"
	"identity/domain"
)

func TestTokenHelpers_ResolveClientFromID(t *testing.T) {
	t.Parallel()

	application, tenantID, clientID, _ := mustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(application.Config, tenantID)
	if err != nil {
		t.Fatalf("FindTenantByID: %v", err)
	}

	if client := resolveClientFromID(tenant, ""); client != nil {
		t.Fatalf("expected nil for empty client ID")
	}

	if client := resolveClientFromID(tenant, "not-a-uuid"); client != nil {
		t.Fatalf("expected nil for invalid client ID")
	}

	if client := resolveClientFromID(tenant, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"); client != nil {
		t.Fatalf("expected nil for unknown client ID")
	}

	client := resolveClientFromID(tenant, clientID.String())
	if client == nil {
		t.Fatalf("expected client")
	}

	if client.ClientID().String() != clientID.String() {
		t.Fatalf("expected client ID %s, got %s", clientID.String(), client.ClientID().String())
	}
}

func TestTokenHelpers_ResolveClientFromForm_WithoutSecretClient(t *testing.T) {
	t.Parallel()

	application, tenantID, clientID, _ := mustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(application.Config, tenantID)
	if err != nil {
		t.Fatalf("FindTenantByID: %v", err)
	}

	client := resolveClientFromForm(tenant, clientID.String(), "")
	if client == nil {
		t.Fatalf("expected client")
	}

	if client.ClientID().String() != clientID.String() {
		t.Fatalf("expected client ID %s, got %s", clientID.String(), client.ClientID().String())
	}
}

func TestTokenHelpers_ResolveClientFromForm_WithSecretValidation(t *testing.T) {
	t.Parallel()

	tenantID := domain.MustTenantID("11111111-1111-4111-8111-111111111111")
	clientID := domain.MustClientID("44444444-4444-4444-8444-444444444444")

	redirectURL, err := domain.NewRedirectURL("https://example.com/callback")
	if err != nil {
		t.Fatalf("NewRedirectURL: %v", err)
	}

	identifierURI, err := domain.NewIdentifierURI("api://secret-client")
	if err != nil {
		t.Fatalf("NewIdentifierURI: %v", err)
	}

	user := domain.NewUser(
		domain.MustUserID("33333333-3333-4333-8333-333333333333"),
		domain.MustUsername("user"),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		nil,
	)

	registration := domain.NewAppRegistration(
		domain.MustAppName("App"),
		clientID,
		identifierURI,
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)

	client := domain.NewClient(
		domain.MustAppName("SecretClient"),
		clientID,
		domain.NewClientSecret("secret"),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	tenant, err := domain.NewTenant(
		tenantID,
		domain.MustTenantName("Tenant"),
		[]domain.AppRegistration{registration},
		nil,
		[]domain.User{user},
		[]domain.Client{client},
	)
	if err != nil {
		t.Fatalf("NewTenant: %v", err)
	}

	if got := resolveClientFromForm(tenant, clientID.String(), "wrong"); got != nil {
		t.Fatalf("expected nil client for invalid secret")
	}

	got := resolveClientFromForm(tenant, clientID.String(), "secret")
	if got == nil {
		t.Fatalf("expected client for valid secret")
	}
}

func TestTokenHelpers_FirstOf(t *testing.T) {
	t.Parallel()

	if got := firstOf("primary", "fallback"); got != "primary" {
		t.Fatalf("expected primary, got %q", got)
	}

	if got := firstOf("", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestTokenHelpers_CorrelationID(t *testing.T) {
	t.Parallel()

	request := newTestRequest(t, "POST", "https://example.test/token", "")
	request.Header.Set("Client-Request-Id", "cid-123")

	if got := correlationID(request); got != "cid-123" {
		t.Fatalf("expected header correlation ID, got %q", got)
	}

	request = newTestRequest(t, "POST", "https://example.test/token", "")

	generated := correlationID(request)
	if generated == "" {
		t.Fatalf("expected generated correlation ID")
	}
}
