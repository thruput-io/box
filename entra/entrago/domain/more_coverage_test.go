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

	if scope.Description().RawString() != testDesc {
		t.Fatal("Scope Description mismatch")
	}

	if scope.Value().RawString() != testAccess {
		t.Fatal("Scope Value mismatch")
	}
}

func assertRoleFields(t *testing.T, role domain.Role, expectedID domain.RoleID, expectedScopeID domain.ScopeID) {
	t.Helper()

	if role.ID() != expectedID {
		t.Fatal("Role ID mismatch")
	}

	if role.Description().RawString() != testDesc {
		t.Fatal("Role Description mismatch")
	}

	if role.Value().RawString() != testAdmin {
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

	if group.Name().RawString() != testGroup {
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

func TestClientSecret_IsEmpty(t *testing.T) {
	t.Parallel()

	if !domain.NewClientSecret(emptyInput).IsEmpty() {
		t.Fatal("expected empty secret to be IsEmpty")
	}

	if domain.NewClientSecret("secret").IsEmpty() {
		t.Fatal("expected non-empty secret to not be IsEmpty")
	}
}

func TestDescriptions_RawString(t *testing.T) {
	t.Parallel()

	sd := mustScopeDescription(t, "scope description")
	if sd.RawString() != "scope description" {
		t.Fatal("unexpected scope description")
	}

	rd := mustRoleDescription(t, "role description")
	if rd.RawString() != "role description" {
		t.Fatal("unexpected role description")
	}
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
