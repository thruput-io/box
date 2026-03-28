package handlers

import (
	"testing"

	"identity/domain"
)

func TestAuthHelpers_CollectAssignmentRolesAndResolveAppName(t *testing.T) {
	t.Parallel()

	appID := domain.MustClientID("22222222-2222-4222-8222-222222222222")

	redirectURL, err := domain.NewRedirectURL("https://example.com/callback")
	if err != nil {
		t.Fatalf("NewRedirectURL: %v", err)
	}

	registration := domain.NewAppRegistration(
		domain.MustAppName("App"),
		appID,
		mustIdentifierURI(t, "api://app"),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)

	assignmentMatching := domain.NewGroupRoleAssignment(
		mustGroupName(t, "GroupA"),
		[]domain.RoleValue{mustRoleValue(t, "Role1")},
		appID,
	)
	assignmentWrongGroup := domain.NewGroupRoleAssignment(
		mustGroupName(t, "Other"),
		[]domain.RoleValue{mustRoleValue(t, "Nope")},
		appID,
	)

	client := domain.NewClient(
		domain.MustAppName("Client"),
		appID,
		domain.NewClientSecret(""),
		[]domain.RedirectURL{redirectURL},
		[]domain.GroupRoleAssignment{assignmentMatching, assignmentWrongGroup},
	)

	user := domain.NewUser(
		domain.MustUserID("33333333-3333-4333-8333-333333333333"),
		domain.MustUsername("user"),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		[]domain.GroupName{mustGroupName(t, "GroupA")},
	)

	tenant, err := domain.NewTenant(
		domain.MustTenantID("11111111-1111-4111-8111-111111111111"),
		domain.MustTenantName("Tenant"),
		[]domain.AppRegistration{registration},
		nil,
		[]domain.User{user},
		[]domain.Client{client},
	)
	if err != nil {
		t.Fatalf("NewTenant: %v", err)
	}

	appRoles := collectAssignmentRoles(user, client)
	if got := appRoles[appID.String()]; len(got) != 1 || got[0] != "Role1" {
		t.Fatalf("unexpected appRoles: %#v", appRoles)
	}

	if got := resolveAppName(tenant, appID.String()); got != "App" {
		t.Fatalf("expected App, got %q", got)
	}

	if got := resolveAppName(tenant, "nope"); got != "nope" {
		t.Fatalf("expected fallback app id, got %q", got)
	}

	roles := resolveDisplayRoles(user, &client, tenant)
	if len(roles) != 1 {
		t.Fatalf("expected 1 role display, got %d", len(roles))
	}
}
