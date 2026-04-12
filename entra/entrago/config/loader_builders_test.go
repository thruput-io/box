package config_test

import (
	"testing"

	"identity/config"
	"identity/domain"
)

func TestBuildScope_Success(t *testing.T) {
	t.Parallel()

	scope := config.ExportBuildScope(config.RawScope{
		ID:          "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa",
		Value:       testAccess,
		Description: testDesc,
	}).MustRight()

	if scope.Description() != domain.NewScopeDescription(testDesc).MustRight() {
		t.Fatal("Scope Description mismatch")
	}

	if scope.Value() != domain.NewScopeValue(testAccess).MustRight() {
		t.Fatal("Scope Value mismatch")
	}
}

func TestBuildRole_Success(t *testing.T) {
	t.Parallel()

	role := config.ExportBuildRole(config.RawRole{
		ID:          "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb",
		Value:       testAdmin,
		Description: testDesc,
		Scopes: []config.RawScope{{
			ID:          "cccccccc-cccc-4ccc-accc-cccccccccccc",
			Value:       "scope",
			Description: testDesc,
		}},
	}).MustRight()

	if role.Description() != domain.NewRoleDescription(testDesc).MustRight() {
		t.Fatal("Role Description mismatch")
	}

	if role.Value() != domain.NewRoleValue(testAdmin).MustRight() {
		t.Fatal("Role Value mismatch")
	}

	if len(role.Scopes()) != expectedOneAssignment || role.Scopes()[0].Value() != domain.NewScopeValue("scope").MustRight() {
		t.Fatal("Role Scopes mismatch")
	}
}

func TestBuildClient_WithGroupRoleAssignment(t *testing.T) {
	t.Parallel()

	clientSecret := domain.NewClientSecret("secret").MustRight()

	client := config.ExportBuildClient(config.RawClient{
		Name:         "Client",
		ClientID:     testClientID,
		ClientSecret: "secret",
		RedirectURLs: []string{testCallback},
		GroupRoleAssignments: []config.RawGroupRoleAssignment{{
			GroupName:     "GroupA",
			Roles:         []string{"RoleA"},
			ApplicationID: testClientID,
		}},
	}).MustRight()

	if client.Name() != domain.NewAppName("Client").MustRight() {
		t.Fatal("Client name mismatch")
	}

	if client.ClientID() != domain.NewClientID(testClientID).MustRight() {
		t.Fatal("Client ID mismatch")
	}

	// Client secrets are mock values in this repo; validation is strict equality.
	validated := client.Validate(&clientSecret).MustRight()
	if validated.ClientSecret() == nil || validated.ClientSecret().Value() != clientSecret.Value() {
		t.Fatal("Client secret mismatch")
	}

	if len(client.RedirectURLs()) != expectedOneAssignment ||
		client.RedirectURLs()[firstIndex] != domain.NewRedirectURL(testCallback).MustRight() {
		t.Fatal("Client RedirectURLs mismatch")
	}

	if len(client.GroupRoleAssignments()) != expectedOneAssignment {
		t.Fatal("expected 1 assignment")
	}
}

func TestBuildGroupRoleAssignments_Success(t *testing.T) {
	t.Parallel()

	assignments := config.ExportBuildGroupRoleAssignments([]config.RawGroupRoleAssignment{{
		GroupName:     "GroupA",
		Roles:         []string{"RoleA"},
		ApplicationID: testClientID,
	}}).MustRight()

	if len(assignments) != expectedOneAssignment {
		t.Fatal("expected 1 assignment")
	}
}

func TestBuildGroupRoleAssignment_Success(t *testing.T) {
	t.Parallel()

	assignment := config.ExportBuildGroupRoleAssignment(config.RawGroupRoleAssignment{
		GroupName:     "GroupA",
		Roles:         []string{"RoleA"},
		ApplicationID: testClientID,
	}).MustRight()

	if assignment.GroupName() != domain.NewGroupName("GroupA").MustRight() {
		t.Fatal("unexpected group name")
	}

	if len(assignment.Roles()) != expectedOneAssignment ||
		assignment.Roles()[firstIndex] != domain.NewRoleValue("RoleA").MustRight() {
		t.Fatal("unexpected roles")
	}

	if assignment.ApplicationID() != domain.NewClientID(testClientID).MustRight() {
		t.Fatal("unexpected application id")
	}
}

func TestBuildUser_Success(t *testing.T) {
	t.Parallel()

	user := config.ExportBuildUser(config.RawUser{
		ID:          "33333333-3333-4333-8333-333333333333",
		Username:    "user",
		Password:    "pass",
		DisplayName: "User",
		Email:       "user@example.com",
		Groups:      []string{"GroupA"},
	}).MustRight()

	assertUserFields(t, user)
}

func assertUserFields(t *testing.T, user domain.User) {
	t.Helper()

	if user.ID() != domain.NewUserID("33333333-3333-4333-8333-333333333333").MustRight() {
		t.Fatal("User ID mismatch")
	}

	if user.Username() != domain.NewUsername("user").MustRight() {
		t.Fatal("User Username mismatch")
	}

	if user.Password() != domain.NewPassword("pass").MustRight() {
		t.Fatal("User Password mismatch")
	}

	if user.DisplayName() != domain.NewDisplayName("User").MustRight() {
		t.Fatal("User DisplayName mismatch")
	}

	if user.Email() != domain.NewEmail("user@example.com").MustRight() {
		t.Fatal("User Email mismatch")
	}

	if len(user.Groups()) != expectedOneAssignment || user.Groups()[firstIndex] != domain.NewGroupName("GroupA").MustRight() {
		t.Fatal("User Groups mismatch")
	}
}

func TestBuildAppRegistration_Success(t *testing.T) {
	t.Parallel()

	app := config.ExportBuildAppRegistration(config.RawAppRegistration{
		Name:          "App",
		ClientID:      testClientID,
		IdentifierURI: "api://app",
		RedirectURLs:  []string{testCallback},
		Scopes: []config.RawScope{{
			ID:          "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa",
			Value:       testAccess,
			Description: testDesc,
		}},
		AppRoles: []config.RawRole{{
			ID:          "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb",
			Value:       testAdmin,
			Description: testDesc,
			Scopes:      []config.RawScope{},
		}},
	}).MustRight()

	assertAppFields(t, app)
}

func assertAppFields(t *testing.T, app domain.AppRegistration) {
	t.Helper()

	if app.Name() != domain.NewAppName("App").MustRight() {
		t.Fatal("App Name mismatch")
	}

	if app.ClientID() != domain.NewClientID(testClientID).MustRight() {
		t.Fatal("App ClientID mismatch")
	}

	if app.IdentifierURI() != domain.NewIdentifierURI("api://app").MustRight() {
		t.Fatal("App IdentifierURI mismatch")
	}

	assertAppCollections(t, app)
}

func assertAppCollections(t *testing.T, app domain.AppRegistration) {
	t.Helper()

	if len(app.RedirectURLs()) != expectedOneAssignment ||
		app.RedirectURLs()[firstIndex] != domain.NewRedirectURL(testCallback).MustRight() {
		t.Fatal("App RedirectURLs mismatch")
	}

	if len(app.Scopes()) != expectedOneAssignment ||
		app.Scopes()[firstIndex].Value() != domain.NewScopeValue(testAccess).MustRight() {
		t.Fatal("App Scopes mismatch")
	}

	if len(app.AppRoles()) != expectedOneAssignment ||
		app.AppRoles()[firstIndex].Value() != domain.NewRoleValue(testAdmin).MustRight() {
		t.Fatal("App AppRoles mismatch")
	}
}

func TestBuildTenant_Success(t *testing.T) {
	t.Parallel()

	tenant := config.ExportBuildTenant(config.RawTenant{
		TenantID: "11111111-1111-4111-8111-111111111111",
		Name:     "Tenant",
		AppRegistrations: []config.RawAppRegistration{{
			Name: "App", ClientID: testClientID, IdentifierURI: "api://app",
			RedirectURLs: []string{testCallback}, Scopes: []config.RawScope{}, AppRoles: []config.RawRole{},
		}},
		Groups: []config.RawGroup{{
			ID: "cccccccc-cccc-4ccc-accc-cccccccccccc", Name: "Group",
		}},
		Users: []config.RawUser{{
			ID: "33333333-3333-4333-8333-333333333333", Username: "user", Password: "pass",
			DisplayName: "User", Email: "user@example.com", Groups: []string{},
		}},
		Clients: []config.RawClient{{
			Name: "Client", ClientID: "44444444-4444-4444-4444-444444444444", ClientSecret: "secret",
			RedirectURLs: []string{},
			GroupRoleAssignments: []config.RawGroupRoleAssignment{{
				GroupName: "Group", Roles: []string{"Role"}, ApplicationID: testClientID,
			}},
		}},
	}).MustRight()

	assertTenantFields(t, tenant)
}

func assertTenantFields(t *testing.T, tenant domain.Tenant) {
	t.Helper()

	if tenant.TenantID() != domain.NewTenantID("11111111-1111-4111-8111-111111111111").MustRight() {
		t.Fatal("Tenant ID mismatch")
	}

	if tenant.Name() != domain.NewTenantName("Tenant").MustRight() {
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
