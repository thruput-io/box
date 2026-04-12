package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"identity/domain"
)

func TestDomainError_ErrorString(t *testing.T) {
	t.Parallel()

	err := domain.NewError(domain.ErrCodeTenantNotFound, "no such tenant")
	if got := err.Error(); got != "TENANT_NOT_FOUND: no such tenant" {
		t.Fatalf("unexpected error string: %q", got)
	}
}

func TestConfig_ScopeRoleGroupGetters(t *testing.T) {
	t.Parallel()

	scopeID := mustScopeID(t, "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	scopeValue := mustScopeValue(t, testAccess)
	scopeDesc := mustScopeDescription(t, testDesc)
	scope := domain.NewScope(scopeID, scopeValue, scopeDesc)

	assertScopeFields(t, scope)

	roleID := mustRoleID(t, "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	roleValue := mustRoleValue(t, testAdmin)
	roleDesc := mustRoleDescription(t, testDesc)
	role := domain.NewRole(roleID, roleValue, roleDesc, []domain.Scope{scope})

	assertRoleFields(t, role)

	groupID := mustGroupID(t, "cccccccc-cccc-4ccc-accc-cccccccccccc")
	groupName := mustGroupName(t, testGroup)
	group := domain.NewGroup(groupID, groupName)

	assertGroupFields(t, group, groupID)
}

func assertScopeFields(t *testing.T, scope domain.Scope) {
	t.Helper()

	if parseString(scope.Description()) != testDesc {
		t.Fatal("Scope Description mismatch")
	}

	if parseString(scope.Value()) != testAccess {
		t.Fatal("Scope Value mismatch")
	}
}

func assertRoleFields(t *testing.T, role domain.Role) {
	t.Helper()

	if parseString(role.Description()) != testDesc {
		t.Fatal("Role Description mismatch")
	}

	if parseString(role.Value()) != testAdmin {
		t.Fatal("Role Value mismatch")
	}

	if len(role.Scopes()) != expectedScopes || parseString(role.Scopes()[0].Value()) != testAccess {
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

	tenantID := domain.NewTenantID(valStr).MustRight()
	if tenantID.UUID() != value || tenantID.Value() != valStr {
		t.Fatal("TenantID mismatch")
	}

	clientID := domain.NewClientID(valStr).MustRight()
	if clientID.UUID() != value || clientID.Value() != valStr {
		t.Fatal("ClientID mismatch")
	}

	groupID := domain.NewGroupID(valStr).MustRight()
	if groupID.UUID() != value || groupID.Value() != valStr {
		t.Fatal("GroupID mismatch")
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

	if _, ok := domain.NewGroupID(emptyInput).Left(); !ok {
		t.Fatal("expected group id error")
	}

	err, ok := domain.NewTenantName(emptyInput).Left()
	if !ok {
		t.Fatal("expected tenant name error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}

	err, ok = domain.NewAppName(emptyInput).Left()
	if !ok {
		t.Fatal("expected app name error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}

	err, ok = domain.NewIdentifierURI(emptyInput).Left()
	if !ok {
		t.Fatal("expected identifier URI error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}

	err, ok = domain.NewScopeValue(emptyInput).Left()
	if !ok {
		t.Fatal("expected scope value error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}

	err, ok = domain.NewRoleValue(emptyInput).Left()
	if !ok {
		t.Fatal("expected role value error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}

	err, ok = domain.NewGroupName(emptyInput).Left()
	if !ok {
		t.Fatal("expected group name error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}
}

func TestGroupID_Value(t *testing.T) {
	t.Parallel()

	id := mustGroupID(t, testGroupIDUUID)
	got := id.Value()

	if got != testGroupIDUUID {
		t.Errorf("GroupID.Value() = %q, want %q", got, testGroupIDUUID)
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

	roleValue := mustRoleValue(t, testAdmin)

	if !roleValue.Matches(testAdmin) {
		t.Error("expected Matches to return true for exact match")
	}

	if roleValue.Matches("User") {
		t.Error("expected Matches to return false for non-match")
	}
}

func TestGroupName_Matches(t *testing.T) {
	t.Parallel()

	groupName := mustGroupName(t, "engineers")

	if !groupName.Matches("engineers") {
		t.Error("expected Matches to return true for exact match")
	}

	if groupName.Matches("managers") {
		t.Error("expected Matches to return false for non-match")
	}
}

func TestJoinRoleValues(t *testing.T) {
	t.Parallel()

	roles := []domain.RoleValue{mustRoleValue(t, testAdmin), mustRoleValue(t, "User")}
	got := domain.JoinRoleValues(roles, " ")

	if got != "Admin User" {
		t.Errorf("JoinRoleValues() = %q, want %q", got, "Admin User")
	}
}

func TestIDToken_MarshalJSON(t *testing.T) {
	t.Parallel()

	token := domain.NewIDToken("some.jwt.token").MustRight()

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

	err, ok := domain.NewUsername(emptyInput).Left()
	if !ok {
		t.Fatal("expected username error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}

	err, ok = domain.NewPassword(emptyInput).Left()
	if !ok {
		t.Fatal("expected password error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}

	err, ok = domain.NewDisplayName(emptyInput).Left()
	if !ok {
		t.Fatal("expected display name error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}

	err, ok = domain.NewEmail(emptyInput).Left()
	if !ok {
		t.Fatal("expected email error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}

	err, ok = domain.NewRedirectURL(emptyInput).Left()
	if !ok {
		t.Fatal("expected redirect URL error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}

	err, ok = domain.NewScopeDescription(emptyInput).Left()
	if !ok {
		t.Fatal("expected scope description error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}

	err, ok = domain.NewRoleDescription(emptyInput).Left()
	if !ok {
		t.Fatal("expected role description error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected ErrNonEmptyStringEmpty, got %v", err)
	}
}
