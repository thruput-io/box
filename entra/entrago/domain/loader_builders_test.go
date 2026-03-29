package domain_test

import (
	"testing"

	"identity/domain"
)

func TestBuildScope_Success(t *testing.T) {
	t.Parallel()

	scope, err := domain.ExportBuildScope(domain.RawScope{
		ID:          "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa",
		Value:       testAccess,
		Description: testDesc,
	})
	if err != nil {
		t.Fatalf("BuildScope: %v", err)
	}

	if scope.ID() != domain.MustScopeID("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa") {
		t.Fatal("Scope ID mismatch")
	}

	if scope.Description() != domain.MustScopeDescription(testDesc) {
		t.Fatal("Scope Description mismatch")
	}

	if scope.Value() != domain.MustScopeValue(testAccess) {
		t.Fatal("Scope Value mismatch")
	}
}

func TestBuildRole_Success(t *testing.T) {
	t.Parallel()

	role, err := domain.ExportBuildRole(domain.RawRole{
		ID:          "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb",
		Value:       testAdmin,
		Description: testDesc,
		Scopes: []domain.RawScope{{
			ID:          "cccccccc-cccc-4ccc-accc-cccccccccccc",
			Value:       "scope",
			Description: testDesc,
		}},
	})
	if err != nil {
		t.Fatalf("BuildRole: %v", err)
	}

	if role.ID() != domain.MustRoleID("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb") {
		t.Fatal("Role ID mismatch")
	}

	if role.Description() != domain.MustRoleDescription(testDesc) {
		t.Fatal("Role Description mismatch")
	}

	if role.Value() != domain.MustRoleValue(testAdmin) {
		t.Fatal("Role Value mismatch")
	}

	if len(role.Scopes()) != expectedOneAssignment || role.Scopes()[0].Value() != domain.MustScopeValue("scope") {
		t.Fatal("Role Scopes mismatch")
	}
}

func TestBuildClient_WithGroupRoleAssignment(t *testing.T) {
	t.Parallel()

	clientSecret := domain.MustClientSecret("secret")
	client, err := domain.ExportBuildClient(domain.RawClient{
		Name:         "Client",
		ClientID:     testClientID,
		ClientSecret: "secret",
		RedirectURLs: []string{testCallback},
		GroupRoleAssignments: []domain.RawGroupRoleAssignment{{
			GroupName:     "GroupA",
			Roles:         []string{"RoleA"},
			ApplicationID: testClientID,
		}},
	})
	if err != nil {
		t.Fatalf("BuildClient: %v", err)
	}

	if client.Name() != domain.MustAppName("Client") {
		t.Fatal("Client name mismatch")
	}

	if client.ClientID() != domain.MustClientID(testClientID) {
		t.Fatal("Client ID mismatch")
	}

	// For secrets, we use Match since it's a secret.
	if err := client.Validate(&clientSecret); err != nil {
		t.Fatal("Client secret mismatch")
	}

	if len(client.RedirectURLs()) != expectedOneAssignment || client.RedirectURLs()[0] != domain.MustRedirectURL(testCallback) {
		t.Fatal("Client RedirectURLs mismatch")
	}

	if len(client.GroupRoleAssignments()) != expectedOneAssignment {
		t.Fatal("expected 1 assignment")
	}
}

func TestBuildGroupRoleAssignments_Success(t *testing.T) {
	t.Parallel()

	assignments, err := domain.ExportBuildGroupRoleAssignments([]domain.RawGroupRoleAssignment{{
		GroupName:     "GroupA",
		Roles:         []string{"RoleA"},
		ApplicationID: testClientID,
	}})
	if err != nil {
		t.Fatalf("BuildGroupRoleAssignments: %v", err)
	}

	if len(assignments) != expectedOneAssignment {
		t.Fatal("expected 1 assignment")
	}
}

func TestBuildGroupRoleAssignment_Success(t *testing.T) {
	t.Parallel()

	assignment, err := domain.ExportBuildGroupRoleAssignment(domain.RawGroupRoleAssignment{
		GroupName:     "GroupA",
		Roles:         []string{"RoleA"},
		ApplicationID: testClientID,
	})
	if err != nil {
		t.Fatalf("BuildGroupRoleAssignment: %v", err)
	}

	if assignment.GroupName() != domain.MustGroupName("GroupA") {
		t.Fatal("unexpected group name")
	}

	if len(assignment.Roles()) != expectedOneAssignment || assignment.Roles()[0] != domain.MustRoleValue("RoleA") {
		t.Fatal("unexpected roles")
	}

	if assignment.ApplicationID() != domain.MustClientID(testClientID) {
		t.Fatal("unexpected application id")
	}
}

func TestBuildUser_Success(t *testing.T) {
	t.Parallel()

	user, err := domain.ExportBuildUser(domain.RawUser{
		ID:          "33333333-3333-4333-8333-333333333333",
		Username:    "user",
		Password:    "pass",
		DisplayName: "User",
		Email:       "user@example.com",
		Groups:      []string{"GroupA"},
	})
	if err != nil {
		t.Fatalf("BuildUser: %v", err)
	}

	assertUserFields(t, user)
}

func assertUserFields(t *testing.T, user domain.User) {
	t.Helper()

	if user.ID() != domain.MustUserID("33333333-3333-4333-8333-333333333333") {
		t.Fatal("User ID mismatch")
	}

	if user.Username() != domain.MustUsername("user") {
		t.Fatal("User Username mismatch")
	}

	if user.Password() != domain.MustPassword("pass") {
		t.Fatal("User Password mismatch")
	}

	if user.DisplayName() != domain.MustDisplayName("User") {
		t.Fatal("User DisplayName mismatch")
	}

	if user.Email() != domain.MustEmail("user@example.com") {
		t.Fatal("User Email mismatch")
	}

	if len(user.Groups()) != expectedOneAssignment || user.Groups()[0] != domain.MustGroupName("GroupA") {
		t.Fatal("User Groups mismatch")
	}
}

func TestBuildAppRegistration_Success(t *testing.T) {
	t.Parallel()

	app, err := domain.ExportBuildAppRegistration(domain.RawAppRegistration{
		Name:          "App",
		ClientID:      testClientID,
		IdentifierURI: "api://app",
		RedirectURLs:  []string{testCallback},
		Scopes: []domain.RawScope{{
			ID:          "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa",
			Value:       testAccess,
			Description: testDesc,
		}},
		AppRoles: []domain.RawRole{{
			ID:          "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb",
			Value:       testAdmin,
			Description: testDesc,
			Scopes:      []domain.RawScope{},
		}},
	})
	if err != nil {
		t.Fatalf("BuildAppRegistration: %v", err)
	}

	assertAppFields(t, app)
}

func assertAppFields(t *testing.T, app domain.AppRegistration) {
	t.Helper()

	if app.Name() != domain.MustAppName("App") {
		t.Fatal("App Name mismatch")
	}

	if app.ClientID() != domain.MustClientID(testClientID) {
		t.Fatal("App ClientID mismatch")
	}

	if app.IdentifierURI() != domain.MustIdentifierURI("api://app") {
		t.Fatal("App IdentifierURI mismatch")
	}

	assertAppCollections(t, app)
}

func assertAppCollections(t *testing.T, app domain.AppRegistration) {
	t.Helper()

	if len(app.RedirectURLs()) != expectedOneAssignment || app.RedirectURLs()[0] != domain.MustRedirectURL(testCallback) {
		t.Fatal("App RedirectURLs mismatch")
	}

	if len(app.Scopes()) != expectedOneAssignment || app.Scopes()[0].Value() != domain.MustScopeValue(testAccess) {
		t.Fatal("App Scopes mismatch")
	}

	if len(app.AppRoles()) != expectedOneAssignment || app.AppRoles()[0].Value() != domain.MustRoleValue(testAdmin) {
		t.Fatal("App AppRoles mismatch")
	}
}

func TestBuildTenant_Success(t *testing.T) {
	t.Parallel()

	tenant, err := domain.ExportBuildTenant(domain.RawTenant{
		TenantID: "11111111-1111-4111-8111-111111111111",
		Name:     "Tenant",
		AppRegistrations: []domain.RawAppRegistration{{
			Name: "App", ClientID: testClientID, IdentifierURI: "api://app",
			RedirectURLs: []string{testCallback}, Scopes: []domain.RawScope{}, AppRoles: []domain.RawRole{},
		}},
		Groups: []domain.RawGroup{{
			ID: "cccccccc-cccc-4ccc-accc-cccccccccccc", Name: "Group",
		}},
		Users: []domain.RawUser{{
			ID: "33333333-3333-4333-8333-333333333333", Username: "user", Password: "pass",
			DisplayName: "User", Email: "user@example.com", Groups: []string{},
		}},
		Clients: []domain.RawClient{{
			Name: "Client", ClientID: "44444444-4444-4444-4444-444444444444", ClientSecret: "secret",
			RedirectURLs: []string{},
			GroupRoleAssignments: []domain.RawGroupRoleAssignment{{
				GroupName: "Group", Roles: []string{"Role"}, ApplicationID: testClientID,
			}},
		}},
	})
	if err != nil {
		t.Fatalf("BuildTenant: %v", err)
	}

	assertTenantFields(t, tenant)
}

func assertTenantFields(t *testing.T, tenant domain.Tenant) {
	t.Helper()

	if tenant.TenantID() != domain.MustTenantID("11111111-1111-4111-8111-111111111111") {
		t.Fatal("Tenant ID mismatch")
	}

	if tenant.Name() != domain.MustTenantName("Tenant") {
		t.Fatal("Tenant Name mismatch")
	}

	if len(tenant.AppRegistrations()) != expectedOneAssignment {
		t.Fatal("Tenant AppRegistrations mismatch")
	}

	if len(tenant.Groups()) != expectedOneAssignment {
		t.Fatal("Tenant Groups mismatch")
	}

	if len(tenant.Users()) != expectedOneAssignment {
		t.Fatal("Tenant Users mismatch")
	}

	if len(tenant.Clients()) != expectedOneAssignment {
		t.Fatal("Tenant Clients mismatch")
	}
}
