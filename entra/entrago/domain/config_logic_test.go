package domain_test

import (
	"testing"

	"identity/domain"
)

const (
	testTenantID = "11111111-1111-1111-1111-111111111111"
)

func TestAppRegistration_IsAudienceForScope(t *testing.T) {
	t.Parallel()

	clientID := mustClientID(t, "11111111-1111-1111-1111-111111111111")
	idURI := domain.MustIdentifierURI("api://my-app")
	app := domain.NewAppRegistration(
		domain.MustAppName("App"),
		clientID,
		idURI,
		nil, nil, nil,
	)

	tests := []struct {
		scope string
		want  bool
	}{
		{"11111111-1111-1111-1111-111111111111", true},
		{"api://my-app", true},
		{"api://my-app/read", true},
		{"api://my-app/.default", true},
		{"api://other", false},
		{"other-uuid", false},
	}

	for _, tt := range tests {
		if got := app.IsAudienceForScope(mustScopeValue(t, tt.scope)); got != tt.want {
			t.Errorf("IsAudienceForScope(%q) = %v, want %v", tt.scope, got, tt.want)
		}
	}
}

func TestRole_MatchesScope(t *testing.T) {
	t.Parallel()

	scopeID1 := mustScopeID(t, "22222222-2222-2222-2222-222222222222")
	scopeID2 := mustScopeID(t, "33333333-3333-3333-3333-333333333333")
	scopeValue1 := mustScopeValue(t, "read")
	scopeValue2 := mustScopeValue(t, "write")
	scopeDesc1 := mustScopeDescription(t, "Read access")
	scopeDesc2 := mustScopeDescription(t, "Write access")

	scope1 := domain.NewScope(scopeID1, scopeValue1, scopeDesc1)
	scope2 := domain.NewScope(scopeID2, scopeValue2, scopeDesc2)

	role := domain.NewRole(
		mustRoleID(t, "44444444-4444-4444-4444-444444444444"),
		mustRoleValue(t, "Admin"),
		mustRoleDescription(t, "Admin role"),
		[]domain.Scope{scope1, scope2},
	)

	tests := []struct {
		scope string
		want  bool
	}{
		{"read", true},
		{"api://app/read", true},
		{"write", true},
		{"other", false},
	}

	for _, tt := range tests {
		if got := role.MatchesScope(mustScopeValue(t, tt.scope)); got != tt.want {
			t.Errorf("MatchesScope(%q) = %v, want %v", tt.scope, got, tt.want)
		}
	}
}

func TestClient_Validate_Confidential(t *testing.T) {
	t.Parallel()

	clientID := mustClientID(t, "55555555-5555-5555-5555-555555555555")
	secret := mustClientSecret(t, "secret123")
	client := domain.NewClientWithSecret(domain.MustAppName("Conf"), clientID, secret, nil, nil)

	// Correct secret
	validated := client.Validate(&secret).MustRight()
	if validated.ClientID() != clientID {
		t.Errorf("ClientID() = %v, want %v", validated.ClientID(), clientID)
	}

	if validated.ClientSecret() == nil {
		t.Fatal("ClientSecret() = nil")
	}

	if validated.ClientSecret().Value() != secret.Value() {
		t.Error("ClientSecret() does not match expected secret")
	}

	// Wrong secret
	wrong := mustClientSecret(t, "wrong")

	err := client.Validate(&wrong).MustLeft()
	if err.Code != domain.ErrInvalidCredentials.Code {
		t.Errorf("code = %v, want %v", err.Code, domain.ErrInvalidCredentials.Code)
	}

	// Missing secret
	err = client.Validate(nil).MustLeft()
	if err.Code != domain.ErrClientSecretRequired.Code {
		t.Errorf("code = %v, want %v", err.Code, domain.ErrClientSecretRequired.Code)
	}
}

func TestClient_Validate_Public(t *testing.T) {
	t.Parallel()

	clientID := mustClientID(t, "55555555-5555-5555-5555-555555555555")
	secret := mustClientSecret(t, "secret123")
	client := domain.NewClientWithoutSecret(domain.MustAppName("Pub"), clientID, nil, nil)

	// No secret provided
	validated := client.Validate(nil).MustRight()
	if validated.ClientID() != clientID {
		t.Errorf("ClientID() = %v, want %v", validated.ClientID(), clientID)
	}

	if validated.ClientSecret() != nil {
		t.Error("ClientSecret() != nil")
	}

	// Secret provided to public client
	err := client.Validate(&secret).MustLeft()
	if err.Code != domain.ErrPublicClientDoesNotAcceptSecrets.Code {
		t.Errorf("code = %v, want %v", err.Code, domain.ErrPublicClientDoesNotAcceptSecrets.Code)
	}
}

func TestTenant_AsClient(t *testing.T) {
	t.Parallel()

	tenantID := mustTenantID(t, "66666666-6666-6666-6666-666666666666")
	tenantName := mustTenantName(t, "Tenant")
	uID := mustUserID(t, "77777777-7777-7777-7777-777777777777")
	uName := mustUsername(t, "u")
	uPass := mustPassword(t, "p")
	uDisp := domain.MustDisplayName("D")
	uEmail := domain.MustEmail("e")
	user := domain.NewUser(uID, uName, uPass, uDisp, uEmail, nil)

	appReg := domain.NewAppRegistration(
		domain.MustAppName("App"),
		mustClientID(t, "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"),
		domain.MustIdentifierURI("api://app"),
		nil, nil, nil,
	)
	appRegistrations := domain.NewNonEmptyArray(appReg).MustRight()

	users := domain.NewNonEmptyArray(user).MustRight()
	tenant := domain.NewTenant(tenantID, tenantName, appRegistrations, nil, users, nil).MustRight()

	client := tenant.AsClient()

	if client.ClientID().UUID() != tenantID.UUID() {
		t.Errorf("expected client ID %v, got %v", tenantID.UUID(), client.ClientID().UUID())
	}

	if client.ClientSecret() != nil {
		t.Error("expected public client (no secret)")
	}
}

func TestTenantID_AsURL(t *testing.T) {
	t.Parallel()

	id := mustTenantID(t, "88888888-8888-8888-8888-888888888888")
	got := id.AsURL(domain.MustBaseURL("https://entra.test"))
	want := "https://entra.test/88888888-8888-8888-8888-888888888888"

	if got.Value() != want {
		t.Errorf("AsURL() = %q, want %q", got, want)
	}
}
