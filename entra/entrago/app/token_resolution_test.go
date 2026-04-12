package app_test

import (
	"testing"

	"identity/app"
	"identity/domain"
)

func TestResolveAudienceForTest(t *testing.T) {
	t.Parallel()

	clientID := mustClientID(t, testClientID)
	redirectURL := mustRedirectURL(t, testCallback)

	registration := domain.NewAppRegistration(
		mustAppName(t, "App"),
		clientID,
		mustIdentifierURI(t, testApp),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)

	user := domain.NewUser(
		mustUserID(t, testUserID),
		mustUsername(t, testUser),
		mustPassword(t, testPass),
		mustDisplayName(t, "User"),
		mustEmail(t, "user@example.com"),
		nil,
	)

	tenant := domain.NewTenant(
		mustTenantID(t, testTenantID),
		mustTenantName(t, "Tenant"),
		domain.NewNonEmptyArray(registration).MustRight(),
		nil,
		domain.NewNonEmptyArray(user).MustRight(),
		nil,
	).MustRight()

	gotAud, gotApps := app.ExportResolveAudienceForTest(&tenant, []domain.ScopeValue{
		mustScopeValue(t, "openid"),
		mustScopeValue(t, "api://app/access"),
	})
	if !gotAud.Matches(testApp) {
		t.Fatalf("expected audience %q, got %q", testApp, gotAud)
	}

	if !gotApps[clientID] {
		t.Fatalf("expected target apps to include %s", clientID.UUID().String())
	}

	gotAud, gotApps = app.ExportResolveAudienceForTest(&tenant, []domain.ScopeValue{
		mustScopeValue(t, "openid"),
	})
	if !gotAud.Matches("api://default") {
		t.Fatalf("expected default audience %q, got %q", "api://default", gotAud)
	}

	const expectedLen = 0

	if len(gotApps) != expectedLen {
		t.Fatal("expected no target apps for oidc-only scope")
	}
}

func TestResolveRolesForTest_AssignmentsAndScopeMatchedRoles(t *testing.T) {
	t.Parallel()

	appClientID := mustClientID(t, testClientID)
	redirectURL := mustRedirectURL(t, testCallback)

	const (
		scopeID = "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"
		roleID  = "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb"
	)

	scope := domain.NewScope(
		mustScopeID(t, scopeID),
		mustScopeValue(t, "access"),
		mustScopeDescription(t, "desc"),
	)
	roleFromScope := domain.NewRole(
		mustRoleID(t, roleID),
		mustRoleValue(t, "RoleFromScope"),
		mustRoleDescription(t, "desc"),
		[]domain.Scope{scope},
	)

	registration := domain.NewAppRegistration(
		mustAppName(t, "App"),
		appClientID,
		mustIdentifierURI(t, testApp),
		[]domain.RedirectURL{redirectURL},
		[]domain.Scope{scope},
		[]domain.Role{roleFromScope},
	)

	user, client := setupUserAndClientForRoles(t, appClientID, redirectURL)

	tenant := domain.NewTenant(
		mustTenantID(t, testTenantID),
		mustTenantName(t, "Tenant"),
		domain.NewNonEmptyArray(registration).MustRight(),
		nil,
		domain.NewNonEmptyArray(user).MustRight(),
		[]domain.Client{client},
	).MustRight()

	verifyResolvedRoles(t, &tenant, &client, &user, appClientID)
}

func setupUserAndClientForRoles(
	t *testing.T,
	appClientID domain.ClientID,
	redirectURL domain.RedirectURL,
) (domain.User, domain.Client) {
	t.Helper()

	user := domain.NewUser(
		mustUserID(t, testUserID),
		mustUsername(t, testUser),
		mustPassword(t, testPass),
		mustDisplayName(t, "User"),
		mustEmail(t, "user@example.com"),
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
		mustClientID(t, "00000000-0000-0000-0000-000000000000"),
	)

	client := domain.NewClientWithSecret(
		mustAppName(t, "Client"),
		appClientID,
		mustClientSecret(t, testSecret),
		[]domain.RedirectURL{redirectURL},
		[]domain.GroupRoleAssignment{assignmentMatching, assignmentWrongGroup, assignmentNilApp},
	)

	return user, client
}

func verifyResolvedRoles(
	t *testing.T,
	tenant *domain.Tenant,
	client *domain.Client,
	user *domain.User,
	appClientID domain.ClientID,
) {
	t.Helper()

	targetApps := map[domain.ClientID]bool{appClientID: true}
	requested := []domain.ScopeValue{mustScopeValue(t, testApp+"/access")}

	roles := app.ExportResolveRolesForTest(tenant, client, user, targetApps, requested)
	roleSet := make(map[string]bool)

	for _, r := range roles {
		roleStr := r.Value()
		roleSet[roleStr] = true
	}

	if !roleSet["RoleFromScope"] {
		t.Fatal("expected RoleFromScope")
	}

	if !roleSet["RoleFromAssignment"] {
		t.Fatal("expected RoleFromAssignment")
	}

	if !roleSet["RoleFromNilApp"] {
		t.Fatal("expected RoleFromNilApp")
	}

	if roleSet["ShouldNotAppear"] {
		t.Fatal("did not expect role from non-member group")
	}
}
