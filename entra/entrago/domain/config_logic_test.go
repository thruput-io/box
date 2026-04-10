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

	clientID := domain.MustClientID("11111111-1111-1111-1111-111111111111")
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
		if got := app.IsAudienceForScope(domain.MustScopeValue(tt.scope)); got != tt.want {
			t.Errorf("IsAudienceForScope(%q) = %v, want %v", tt.scope, got, tt.want)
		}
	}
}

func TestRole_MatchesScope(t *testing.T) {
	t.Parallel()

	scopeID1 := domain.MustScopeID("22222222-2222-2222-2222-222222222222")
	scopeID2 := domain.MustScopeID("33333333-3333-3333-3333-333333333333")
	scopeValue1 := domain.MustScopeValue("read")
	scopeValue2 := domain.MustScopeValue("write")
	scopeDesc1 := domain.MustScopeDescription("Read access")
	scopeDesc2 := domain.MustScopeDescription("Write access")

	scope1 := domain.NewScope(scopeID1, scopeValue1, scopeDesc1)
	scope2 := domain.NewScope(scopeID2, scopeValue2, scopeDesc2)

	role := domain.NewRole(
		domain.MustRoleID("44444444-4444-4444-4444-444444444444"),
		domain.MustRoleValue("Admin"),
		domain.MustRoleDescription("Admin role"),
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
		if got := role.MatchesScope(domain.MustScopeValue(tt.scope)); got != tt.want {
			t.Errorf("MatchesScope(%q) = %v, want %v", tt.scope, got, tt.want)
		}
	}
}

func TestClient_Validate_Confidential(t *testing.T) {
	t.Parallel()

	clientID := domain.MustClientID("55555555-5555-5555-5555-555555555555")
	secret := domain.MustClientSecret("secret123")
	client := domain.NewClientWithSecret(domain.MustAppName("Conf"), clientID, secret, nil, nil)

	// Correct secret
	err := client.Validate(&secret)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Wrong secret
	wrong := domain.MustClientSecret("wrong")

	err = client.Validate(&wrong)
	if err == nil {
		t.Error("expected error for wrong secret")
	}

	// Missing secret
	err = client.Validate(nil)
	if err == nil {
		t.Error("expected error for missing secret")
	}
}

func TestClient_Validate_Public(t *testing.T) {
	t.Parallel()

	clientID := domain.MustClientID("55555555-5555-5555-5555-555555555555")
	secret := domain.MustClientSecret("secret123")
	client := domain.NewClientWithoutSecret(domain.MustAppName("Pub"), clientID, nil, nil)

	// No secret provided
	err := client.Validate(nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Secret provided to public client
	err = client.Validate(&secret)
	if err == nil {
		t.Error("expected error when secret provided to public client")
	}
}

func TestTenant_AsClient(t *testing.T) {
	t.Parallel()

	tenantID := domain.MustTenantID("66666666-6666-6666-6666-666666666666")
	tenantName := domain.MustTenantName("Tenant")
	uID := domain.MustUserID("77777777-7777-7777-7777-777777777777")
	uName := domain.MustUsername("u")
	uPass := domain.MustPassword("p")
	uDisp := domain.MustDisplayName("D")
	uEmail := domain.MustEmail("e")
	user := domain.NewUser(uID, uName, uPass, uDisp, uEmail, nil)

	appReg := domain.NewAppRegistration(
		domain.MustAppName("App"),
		domain.MustClientID("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"),
		domain.MustIdentifierURI("api://app"),
		nil, nil, nil,
	)

	tenant, err := domain.NewTenant(
		tenantID, tenantName, []domain.AppRegistration{appReg},
		nil, []domain.User{user}, nil,
	)
	if err != nil {
		t.Fatalf("NewTenant failed: %v", err)
	}

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

	id := domain.MustTenantID("88888888-8888-8888-8888-888888888888")
	got := id.AsURL(domain.MustBaseURL("https://entra.test"))
	want := "https://entra.test/88888888-8888-8888-8888-888888888888"

	if got.Value() != want {
		t.Errorf("AsURL() = %q, want %q", got, want)
	}
}

func TestNewTenant_Errors(t *testing.T) {
	t.Parallel()

	tenantID := domain.MustTenantID(testTenantID)
	tenantName := domain.MustTenantName("T")

	// No users
	appReg := domain.NewAppRegistration(
		domain.MustAppName("A"),
		domain.MustClientID("22222222-2222-2222-2222-222222222222"),
		domain.MustIdentifierURI("api://a"),
		nil, nil, nil,
	)

	_, err := domain.NewTenant(tenantID, tenantName, []domain.AppRegistration{appReg}, nil, nil, nil)
	if err == nil {
		t.Error("expected error for no users")
	}

	// No app registrations
	uID := domain.MustUserID("33333333-3333-3333-3333-333333333333")
	uName := domain.MustUsername("u")
	uPass := domain.MustPassword("p")
	uDisp := domain.MustDisplayName("D")
	uEmail := domain.MustEmail("e")
	user := domain.NewUser(uID, uName, uPass, uDisp, uEmail, nil)

	_, err = domain.NewTenant(tenantID, tenantName, nil, nil, []domain.User{user}, nil)
	if err == nil {
		t.Error("expected error for no app registrations")
	}
}
