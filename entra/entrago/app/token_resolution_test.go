package app

import (
	"testing"

	"identity/domain"
)

func TestResolveAudienceForTest(t *testing.T) {
	t.Parallel()

	clientID := domain.MustClientID("22222222-2222-4222-8222-222222222222")
	redirectURL := mustRedirectURL(t, "https://example.com/callback")

	registration := domain.NewAppRegistration(
		domain.MustAppName("App"),
		clientID,
		mustIdentifierURI(t, "api://app"),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)

	user := domain.NewUser(
		domain.MustUserID("33333333-3333-4333-8333-333333333333"),
		domain.MustUsername("user"),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		nil,
	)

	tenant, err := domain.NewTenant(
		domain.MustTenantID("11111111-1111-4111-8111-111111111111"),
		domain.MustTenantName("Tenant"),
		[]domain.AppRegistration{registration},
		nil,
		[]domain.User{user},
		nil,
	)
	if err != nil {
		t.Fatalf("NewTenant: %v", err)
	}

	gotAud, gotApps := ResolveAudienceForTest(tenant, "openid api://app/access")
	if gotAud != "api://app" {
		t.Fatalf("expected audience %q, got %q", "api://app", gotAud)
	}

	if !gotApps[clientID.String()] {
		t.Fatalf("expected target apps to include %s", clientID)
	}

	gotAud, gotApps = ResolveAudienceForTest(tenant, "openid")
	if gotAud != "api://default" {
		t.Fatalf("expected default audience %q, got %q", "api://default", gotAud)
	}

	if len(gotApps) != 0 {
		t.Fatalf("expected no target apps for oidc-only scope")
	}
}

func TestResolveRolesForTest_AssignmentsAndScopeMatchedRoles(t *testing.T) {
	t.Parallel()

	appClientID := domain.MustClientID("22222222-2222-4222-8222-222222222222")
	redirectURL := mustRedirectURL(t, "https://example.com/callback")

	scope := domain.NewScope(
		domain.MustScopeID("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"),
		mustScopeValue(t, "access"),
		mustScopeDescription(t, "desc"),
	)
	roleFromScope := domain.NewRole(
		domain.MustRoleID("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb"),
		mustRoleValue(t, "RoleFromScope"),
		mustRoleDescription(t, "desc"),
		[]domain.Scope{scope},
	)

	registration := domain.NewAppRegistration(
		domain.MustAppName("App"),
		appClientID,
		mustIdentifierURI(t, "api://app"),
		[]domain.RedirectURL{redirectURL},
		[]domain.Scope{scope},
		[]domain.Role{roleFromScope},
	)

	user := domain.NewUser(
		domain.MustUserID("33333333-3333-4333-8333-333333333333"),
		domain.MustUsername("user"),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		[]domain.GroupName{mustGroupName(t, "GroupA")},
	)

	assignmentMatching := domain.NewGroupRoleAssignment(
		mustGroupName(t, "GroupA"),
		[]domain.RoleValue{mustRoleValue(t, "RoleFromAssignment")},
		appClientID,
	)
	assignmentWrongGroup := domain.NewGroupRoleAssignment(
		mustGroupName(t, "Other"),
		[]domain.RoleValue{mustRoleValue(t, "ShouldNotAppear")},
		appClientID,
	)
	assignmentNilApp := domain.NewGroupRoleAssignment(
		mustGroupName(t, "GroupA"),
		[]domain.RoleValue{mustRoleValue(t, "RoleFromNilApp")},
		domain.MustClientID("00000000-0000-0000-0000-000000000000"),
	)

	client := domain.NewClient(
		domain.MustAppName("Client"),
		appClientID,
		domain.NewClientSecret("secret"),
		[]domain.RedirectURL{redirectURL},
		[]domain.GroupRoleAssignment{assignmentMatching, assignmentWrongGroup, assignmentNilApp},
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

	targetApps := map[string]bool{appClientID.String(): true}
	requested := []string{"api://app/access"}

	roles := ResolveRolesForTest(tenant, &client, &user, targetApps, requested)

	roleSet := make(map[string]bool)

	for _, r := range roles {
		roleSet[r] = true
	}

	if !roleSet["RoleFromScope"] {
		t.Fatalf("expected RoleFromScope")
	}

	if !roleSet["RoleFromAssignment"] {
		t.Fatalf("expected RoleFromAssignment")
	}

	if !roleSet["RoleFromNilApp"] {
		t.Fatalf("expected RoleFromNilApp")
	}

	if roleSet["ShouldNotAppear"] {
		t.Fatalf("did not expect role from non-member group")
	}
}
