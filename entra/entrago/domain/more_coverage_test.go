package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"identity/domain"
)

func mustScopeValue(t *testing.T, raw string) domain.ScopeValue {
	t.Helper()

	v, err := domain.NewScopeValue(raw)
	if err != nil {
		t.Fatalf("NewScopeValue(%q): %v", raw, err)
	}

	return v
}

func mustScopeDescription(t *testing.T, raw string) domain.ScopeDescription {
	t.Helper()

	v, err := domain.NewScopeDescription(raw)
	if err != nil {
		t.Fatalf("NewScopeDescription(%q): %v", raw, err)
	}

	return v
}

func mustRoleValue(t *testing.T, raw string) domain.RoleValue {
	t.Helper()

	v, err := domain.NewRoleValue(raw)
	if err != nil {
		t.Fatalf("NewRoleValue(%q): %v", raw, err)
	}

	return v
}

func mustRoleDescription(t *testing.T, raw string) domain.RoleDescription {
	t.Helper()

	v, err := domain.NewRoleDescription(raw)
	if err != nil {
		t.Fatalf("NewRoleDescription(%q): %v", raw, err)
	}

	return v
}

func mustGroupName(t *testing.T, raw string) domain.GroupName {
	t.Helper()

	v, err := domain.NewGroupName(raw)
	if err != nil {
		t.Fatalf("NewGroupName(%q): %v", raw, err)
	}

	return v
}

func TestDomainError_ErrorString(t *testing.T) {
	t.Parallel()

	err := domain.NewError(domain.ErrCodeTenantNotFound, "no such tenant")
	if got := err.Error(); got != "TENANT_NOT_FOUND: no such tenant" {
		t.Fatalf("unexpected error string: %q", got)
	}
}

func TestConfig_ScopeRoleGroupGetters(t *testing.T) {
	t.Parallel()

	scopeID := domain.MustScopeID("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	scopeValue := mustScopeValue(t, testAccess)
	scopeDesc := mustScopeDescription(t, testDesc)
	scope := domain.NewScope(scopeID, scopeValue, scopeDesc)

	assertScopeFields(t, scope, scopeID)

	roleID := domain.MustRoleID("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	roleValue := mustRoleValue(t, testAdmin)
	roleDesc := mustRoleDescription(t, testDesc)
	role := domain.NewRole(roleID, roleValue, roleDesc, []domain.Scope{scope})

	assertRoleFields(t, role, roleID, scopeID)

	groupID := domain.MustGroupID("cccccccc-cccc-4ccc-accc-cccccccccccc")
	groupName := mustGroupName(t, testGroup)
	group := domain.NewGroup(groupID, groupName)

	assertGroupFields(t, group, groupID)
}

func assertScopeFields(t *testing.T, scope domain.Scope, expectedID domain.ScopeID) {
	t.Helper()

	if scope.ID() != expectedID {
		t.Fatal("Scope ID mismatch")
	}

	if parseString(scope.Description()) != testDesc {
		t.Fatal("Scope Description mismatch")
	}

	if parseString(scope.Value()) != testAccess {
		t.Fatal("Scope Value mismatch")
	}
}

func assertRoleFields(t *testing.T, role domain.Role, expectedID domain.RoleID, expectedScopeID domain.ScopeID) {
	t.Helper()

	if role.ID() != expectedID {
		t.Fatal("Role ID mismatch")
	}

	if parseString(role.Description()) != testDesc {
		t.Fatal("Role Description mismatch")
	}

	if parseString(role.Value()) != testAdmin {
		t.Fatal("Role Value mismatch")
	}

	if len(role.Scopes()) != expectedScopes || role.Scopes()[0].ID() != expectedScopeID {
		t.Fatal("Role Scopes mismatch")
	}
}

func assertGroupFields(t *testing.T, group domain.Group, expectedID domain.GroupID) {
	t.Helper()

	if group.ID() != expectedID {
		t.Fatal("Group ID mismatch")
	}

	if parseString(group.Name()) != testGroup {
		t.Fatal("Group Name mismatch")
	}
}

func TestIDs_FromUUIDAndAccessors(t *testing.T) {
	t.Parallel()

	value := uuid.New()
	valStr := value.String()

	assertIDAccessors(t, value, valStr)
}

func assertIDAccessors(t *testing.T, value uuid.UUID, valStr string) {
	t.Helper()

	assertCoreIDs(t, value, valStr)
	assertResourceIDs(t, value, valStr)
}

func assertCoreIDs(t *testing.T, value uuid.UUID, valStr string) {
	t.Helper()

	if id := domain.TenantIDFromUUID(value); id.UUID() != value || id.UUID().String() != valStr {
		t.Fatal("TenantID mismatch")
	}

	if id := domain.ClientIDFromUUID(value); id.UUID() != value || id.UUID().String() != valStr {
		t.Fatal("ClientID mismatch")
	}

	if id := domain.UserIDFromUUID(value); id.UUID() != value || id.UUID().String() != valStr {
		t.Fatal("UserID mismatch")
	}
}

func assertResourceIDs(t *testing.T, value uuid.UUID, valStr string) {
	t.Helper()

	if id := domain.GroupIDFromUUID(value); id.UUID() != value || id.UUID().String() != valStr {
		t.Fatal("GroupID mismatch")
	}

	if id := domain.ScopeIDFromUUID(value); id.UUID() != value || id.UUID().String() != valStr {
		t.Fatal("ScopeID mismatch")
	}

	if id := domain.RoleIDFromUUID(value); id.UUID() != value || id.UUID().String() != valStr {
		t.Fatal("RoleID mismatch")
	}
}

func TestDescriptions_Parse(t *testing.T) {
	t.Parallel()

	sd := mustScopeDescription(t, "scope description")
	if parseString(sd) != "scope description" {
		t.Fatal("unexpected scope description")
	}

	rd := mustRoleDescription(t, "role description")
	if parseString(rd) != "role description" {
		t.Fatal("unexpected role description")
	}
}

func parseString(v interface{ Value() string }) string {
	return v.Value()
}

func TestStringWrappers_EmptyErrors_IDsAndNames(t *testing.T) {
	t.Parallel()

	var err error

	_, err = domain.NewGroupID(emptyInput)
	if err == nil {
		t.Fatal("expected group id error")
	}

	_, err = domain.NewTenantName(emptyInput)
	if !errors.Is(err, domain.ErrTenantNameEmpty) {
		t.Fatalf("expected ErrTenantNameEmpty, got %v", err)
	}

	_, err = domain.NewAppName(emptyInput)
	if !errors.Is(err, domain.ErrAppNameEmpty) {
		t.Fatalf("expected ErrAppNameEmpty, got %v", err)
	}

	_, err = domain.NewIdentifierURI(emptyInput)
	if !errors.Is(err, domain.ErrIdentifierURIEmpty) {
		t.Fatalf("expected ErrIdentifierURIEmpty, got %v", err)
	}

	_, err = domain.NewScopeValue(emptyInput)
	if !errors.Is(err, domain.ErrScopeValueEmpty) {
		t.Fatalf("expected ErrScopeValueEmpty, got %v", err)
	}

	_, err = domain.NewRoleValue(emptyInput)
	if !errors.Is(err, domain.ErrRoleValueEmpty) {
		t.Fatalf("expected ErrRoleValueEmpty, got %v", err)
	}

	_, err = domain.NewGroupName(emptyInput)
	if !errors.Is(err, domain.ErrGroupNameEmpty) {
		t.Fatalf("expected ErrGroupNameEmpty, got %v", err)
	}
}

func TestGroupID_Value(t *testing.T) {
	t.Parallel()

	id := domain.MustGroupID(testGroupIDUUID)
	got := id.Value()

	if got != testGroupIDUUID {
		t.Errorf("GroupID.Value() = %q, want %q", got, testGroupIDUUID)
	}
}

func TestScopeID_Value(t *testing.T) {
	t.Parallel()

	id := domain.MustScopeID(testScopeIDUUID)
	got := id.Value()

	if got != testScopeIDUUID {
		t.Errorf("ScopeID.Value() = %q, want %q", got, testScopeIDUUID)
	}
}

func TestRoleID_Value(t *testing.T) {
	t.Parallel()

	id := domain.MustRoleID(testRoleIDUUID)
	got := id.Value()

	if got != testRoleIDUUID {
		t.Errorf("RoleID.Value() = %q, want %q", got, testRoleIDUUID)
	}
}

func TestIdentifierURI_MatchesPrefix(t *testing.T) {
	t.Parallel()

	uri := domain.MustIdentifierURI("api://myapp")

	if !uri.MatchesPrefix("api://myapp/.default") {
		t.Error("expected MatchesPrefix to return true for prefix match")
	}

	if uri.MatchesPrefix("api://other") {
		t.Error("expected MatchesPrefix to return false for non-prefix")
	}
}

func TestRoleValue_Matches(t *testing.T) {
	t.Parallel()

	roleValue := domain.MustRoleValue(testAdmin)

	if !roleValue.Matches(testAdmin) {
		t.Error("expected Matches to return true for exact match")
	}

	if roleValue.Matches("User") {
		t.Error("expected Matches to return false for non-match")
	}
}

func TestGroupName_Matches(t *testing.T) {
	t.Parallel()

	groupName := domain.MustGroupName("engineers")

	if !groupName.Matches("engineers") {
		t.Error("expected Matches to return true for exact match")
	}

	if groupName.Matches("managers") {
		t.Error("expected Matches to return false for non-match")
	}
}

func TestJoinRoleValues(t *testing.T) {
	t.Parallel()

	roles := []domain.RoleValue{domain.MustRoleValue(testAdmin), domain.MustRoleValue("User")}
	got := domain.JoinRoleValues(roles, " ")

	if got != "Admin User" {
		t.Errorf("JoinRoleValues() = %q, want %q", got, "Admin User")
	}
}

func TestIDToken_MarshalJSON(t *testing.T) {
	t.Parallel()

	token := domain.MustIDToken("some.jwt.token")

	data, err := token.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if string(data) != `"some.jwt.token"` {
		t.Errorf("MarshalJSON() = %s, want %q", data, "some.jwt.token")
	}
}

func TestStringWrappers_EmptyErrors_CredentialsAndURIs(t *testing.T) {
	t.Parallel()

	var err error

	_, err = domain.NewUsername(emptyInput)
	if !errors.Is(err, domain.ErrUsernameEmpty) {
		t.Fatalf("expected ErrUsernameEmpty, got %v", err)
	}

	_, err = domain.NewPassword(emptyInput)
	if !errors.Is(err, domain.ErrPasswordEmpty) {
		t.Fatalf("expected ErrPasswordEmpty, got %v", err)
	}

	_, err = domain.NewDisplayName(emptyInput)
	if !errors.Is(err, domain.ErrDisplayNameEmpty) {
		t.Fatalf("expected ErrDisplayNameEmpty, got %v", err)
	}

	_, err = domain.NewEmail(emptyInput)
	if !errors.Is(err, domain.ErrEmailEmpty) {
		t.Fatalf("expected ErrEmailEmpty, got %v", err)
	}

	_, err = domain.NewRedirectURL(emptyInput)
	if !errors.Is(err, domain.ErrRedirectURLEmpty) {
		t.Fatalf("expected ErrRedirectURLEmpty, got %v", err)
	}

	_, err = domain.NewScopeDescription(emptyInput)
	if !errors.Is(err, domain.ErrScopeDescriptionEmpty) {
		t.Fatalf("expected ErrScopeDescriptionEmpty, got %v", err)
	}

	_, err = domain.NewRoleDescription(emptyInput)
	if !errors.Is(err, domain.ErrRoleDescriptionEmpty) {
		t.Fatalf("expected ErrRoleDescriptionEmpty, got %v", err)
	}
}
