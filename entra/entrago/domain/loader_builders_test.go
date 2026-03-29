package domain_test

import (
	"testing"

	"identity/domain"
)

const expectedOneAssignment = 1

func TestBuildScope_Success(t *testing.T) {
	t.Parallel()

	scope, err := domain.BuildScope(domain.RawScope{
		ID:          "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa",
		Value:       "access",
		Description: "desc",
	})
	if err != nil {
		t.Fatalf("BuildScope: %v", err)
	}

	if scope.ID().String() != "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa" {
		t.Fatal("Scope ID mismatch")
	}
	if scope.Description().String() != "desc" {
		t.Fatal("Scope Description mismatch")
	}
	if scope.Value().String() != "access" {
		t.Fatal("Scope Value mismatch")
	}
}

func TestBuildRole_Success(t *testing.T) {
	t.Parallel()

	role, err := domain.BuildRole(domain.RawRole{
		ID:          "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb",
		Value:       "Admin",
		Description: "desc",
		Scopes: []domain.RawScope{{
			ID:          "cccccccc-cccc-4ccc-accc-cccccccccccc",
			Value:       "scope",
			Description: "desc",
		}},
	})
	if err != nil {
		t.Fatalf("BuildRole: %v", err)
	}

	if role.ID().String() != "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb" {
		t.Fatal("Role ID mismatch")
	}
	if role.Description().String() != "desc" {
		t.Fatal("Role Description mismatch")
	}
	if role.Value().String() != "Admin" {
		t.Fatal("Role Value mismatch")
	}
	if len(role.Scopes()) != 1 || role.Scopes()[0].Value().String() != "scope" {
		t.Fatal("Role Scopes mismatch")
	}
}

func TestBuildClient_WithGroupRoleAssignment(t *testing.T) {
	t.Parallel()

	client, err := domain.BuildClient(domain.RawClient{
		Name:         "Client",
		ClientID:     "22222222-2222-4222-8222-222222222222",
		ClientSecret: "secret",
		RedirectURLs: []string{"https://example.com/callback"},
		GroupRoleAssignments: []domain.RawGroupRoleAssignment{{
			GroupName:     "GroupA",
			Roles:         []string{"RoleA"},
			ApplicationID: "22222222-2222-4222-8222-222222222222",
		}},
	})
	if err != nil {
		t.Fatalf("BuildClient: %v", err)
	}

	if client.Name().String() != "Client" {
		t.Fatal("Client name mismatch")
	}
	if client.ClientID().String() != "22222222-2222-4222-8222-222222222222" {
		t.Fatal("Client ID mismatch")
	}
	if client.ClientSecret().String() != "secret" {
		t.Fatal("Client secret mismatch")
	}
	if len(client.RedirectURLs()) != 1 || client.RedirectURLs()[0].String() != "https://example.com/callback" {
		t.Fatal("Client RedirectURLs mismatch")
	}
	if len(client.GroupRoleAssignments()) != expectedOneAssignment {
		t.Fatal("expected 1 assignment")
	}
}

func TestBuildGroupRoleAssignments_Success(t *testing.T) {
	t.Parallel()

	assignments, err := domain.BuildGroupRoleAssignments([]domain.RawGroupRoleAssignment{{
		GroupName:     "GroupA",
		Roles:         []string{"RoleA"},
		ApplicationID: "22222222-2222-4222-8222-222222222222",
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

	assignment, err := domain.BuildGroupRoleAssignment(domain.RawGroupRoleAssignment{
		GroupName:     "GroupA",
		Roles:         []string{"RoleA"},
		ApplicationID: "22222222-2222-4222-8222-222222222222",
	})
	if err != nil {
		t.Fatalf("BuildGroupRoleAssignment: %v", err)
	}

	if assignment.GroupName().String() != "GroupA" {
		t.Fatal("unexpected group name")
	}
	if len(assignment.Roles()) != 1 || assignment.Roles()[0].String() != "RoleA" {
		t.Fatal("unexpected roles")
	}
	if assignment.ApplicationID().String() != "22222222-2222-4222-8222-222222222222" {
		t.Fatal("unexpected application id")
	}
}

func TestBuildUser_Success(t *testing.T) {
	t.Parallel()

	user, err := domain.BuildUser(domain.RawUser{
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

	if user.ID().String() != "33333333-3333-4333-8333-333333333333" {
		t.Fatal("User ID mismatch")
	}
	if user.Username().String() != "user" {
		t.Fatal("User Username mismatch")
	}
	if user.Password().String() != "pass" {
		t.Fatal("User Password mismatch")
	}
	if user.DisplayName().String() != "User" {
		t.Fatal("User DisplayName mismatch")
	}
	if user.Email().String() != "user@example.com" {
		t.Fatal("User Email mismatch")
	}
	if len(user.Groups()) != 1 || user.Groups()[0].String() != "GroupA" {
		t.Fatal("User Groups mismatch")
	}
}

func TestBuildAppRegistration_Success(t *testing.T) {
	t.Parallel()

	app, err := domain.BuildAppRegistration(domain.RawAppRegistration{
		Name:          "App",
		ClientID:      "22222222-2222-4222-8222-222222222222",
		IdentifierURI: "api://app",
		RedirectURLs:  []string{"https://example.com/callback"},
		Scopes: []domain.RawScope{{
			ID:          "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa",
			Value:       "access",
			Description: "desc",
		}},
		AppRoles: []domain.RawRole{{
			ID:          "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb",
			Value:       "Admin",
			Description: "desc",
			Scopes:      []domain.RawScope{},
		}},
	})
	if err != nil {
		t.Fatalf("BuildAppRegistration: %v", err)
	}

	if app.Name().String() != "App" {
		t.Fatal("App Name mismatch")
	}
	if app.ClientID().String() != "22222222-2222-4222-8222-222222222222" {
		t.Fatal("App ClientID mismatch")
	}
	if app.IdentifierURI().String() != "api://app" {
		t.Fatal("App IdentifierURI mismatch")
	}
	if len(app.RedirectURLs()) != 1 || app.RedirectURLs()[0].String() != "https://example.com/callback" {
		t.Fatal("App RedirectURLs mismatch")
	}
	if len(app.Scopes()) != 1 || app.Scopes()[0].Value().String() != "access" {
		t.Fatal("App Scopes mismatch")
	}
	if len(app.AppRoles()) != 1 || app.AppRoles()[0].Value().String() != "Admin" {
		t.Fatal("App AppRoles mismatch")
	}
}

func TestBuildTenant_Success(t *testing.T) {
	t.Parallel()

	tenant, err := domain.BuildTenant(domain.RawTenant{
		TenantID: "11111111-1111-4111-8111-111111111111",
		Name:     "Tenant",
		AppRegistrations: []domain.RawAppRegistration{{
			Name:          "App",
			ClientID:      "22222222-2222-4222-8222-222222222222",
			IdentifierURI: "api://app",
			RedirectURLs:  []string{"https://example.com/callback"},
			Scopes:        []domain.RawScope{},
			AppRoles:      []domain.RawRole{},
		}},
		Groups: []domain.RawGroup{{
			ID:   "cccccccc-cccc-4ccc-accc-cccccccccccc",
			Name: "Group",
		}},
		Users: []domain.RawUser{{
			ID:          "33333333-3333-4333-8333-333333333333",
			Username:    "user",
			Password:    "pass",
			DisplayName: "User",
			Email:       "user@example.com",
			Groups:      []string{},
		}},
		Clients: []domain.RawClient{{
			Name:         "Client",
			ClientID:     "44444444-4444-4444-4444-444444444444",
			ClientSecret: "secret",
			RedirectURLs: []string{},
			GroupRoleAssignments: []domain.RawGroupRoleAssignment{{
				GroupName:     "Group",
				Roles:         []string{"Role"},
				ApplicationID: "22222222-2222-4222-8222-222222222222",
			}},
		}},
	})
	if err != nil {
		t.Fatalf("BuildTenant: %v", err)
	}

	if tenant.TenantID().String() != "11111111-1111-4111-8111-111111111111" {
		t.Fatal("Tenant ID mismatch")
	}
	if tenant.Name().String() != "Tenant" {
		t.Fatal("Tenant Name mismatch")
	}
	if len(tenant.AppRegistrations()) != 1 {
		t.Fatal("Tenant AppRegistrations mismatch")
	}
	if len(tenant.Groups()) != 1 {
		t.Fatal("Tenant Groups mismatch")
	}
	if len(tenant.Users()) != 1 {
		t.Fatal("Tenant Users mismatch")
	}
	if len(tenant.Clients()) != 1 {
		t.Fatal("Tenant Clients mismatch")
	}
}
