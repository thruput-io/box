package handlers_test

import (
	"testing"

	"identity/domain"
	"identity/http/handlers"
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

	user, client := setupUserAndClientForHelpers(t, appID, redirectURL)

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

	verifyAuthHelpers(t, user, client, tenant, appID)
}

func setupUserAndClientForHelpers(
	t *testing.T,
	appID domain.ClientID,
	redirectURL domain.RedirectURL,
) (domain.User, domain.Client) {
	t.Helper()

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

	client := domain.NewClientWithoutSecret(
		domain.MustAppName("Client"),
		appID,
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

	return user, client
}

func verifyAuthHelpers(
	t *testing.T,
	user domain.User,
	client domain.Client,
	tenant domain.Tenant,
	appID domain.ClientID,
) {
	t.Helper()

	appRoles := handlers.ExportCollectAssignmentRoles(user, client)

	const expectedRolesLen = 1

	if got := appRoles[appID]; len(got) != expectedRolesLen || got[0] != "Role1" {
		t.Fatalf("unexpected appRoles: %#v", appRoles)
	}

	if got := handlers.ExportResolveAppName(tenant, appID); got != "App" {
		t.Fatalf("expected App, got %q", got)
	}

	if got := handlers.ExportResolveAppName(tenant, domain.ClientID{}); got != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("expected fallback app id, got %q", got)
	}

	roles := handlers.ExportResolveDisplayRoles(user, client, tenant)
	if len(roles) != expectedRolesLen {
		t.Fatalf("expected 1 role display, got %d", len(roles))
	}
}
