package domain_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"identity/domain"
)

const emptyInput = ""

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
	scopeValue := mustScopeValue(t, "access")
	scopeDesc := mustScopeDescription(t, "desc")
	scope := domain.NewScope(scopeID, scopeValue, scopeDesc)

	if scope.ID() != scopeID {
		t.Fatal("Scope ID mismatch")
	}
	if scope.Description().String() != "desc" {
		t.Fatal("Scope Description mismatch")
	}
	if scope.Value().String() != "access" {
		t.Fatal("Scope Value mismatch")
	}

	roleID := domain.MustRoleID("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb")
	roleValue := mustRoleValue(t, "Admin")
	roleDesc := mustRoleDescription(t, "desc")
	role := domain.NewRole(roleID, roleValue, roleDesc, []domain.Scope{scope})

	if role.ID() != roleID {
		t.Fatal("Role ID mismatch")
	}
	if role.Description().String() != "desc" {
		t.Fatal("Role Description mismatch")
	}
	if role.Value().String() != "Admin" {
		t.Fatal("Role Value mismatch")
	}
	if len(role.Scopes()) != 1 || role.Scopes()[0].ID() != scopeID {
		t.Fatal("Role Scopes mismatch")
	}

	groupID := domain.MustGroupID("cccccccc-cccc-4ccc-accc-cccccccccccc")
	groupName := mustGroupName(t, "Group")
	group := domain.NewGroup(groupID, groupName)

	if group.ID() != groupID {
		t.Fatal("Group ID mismatch")
	}
	if group.Name().String() != "Group" {
		t.Fatal("Group Name mismatch")
	}
}

func TestIDs_FromUUIDAndAccessors(t *testing.T) {
	t.Parallel()

	value := uuid.New()
	valStr := value.String()

	tenantID := domain.TenantIDFromUUID(value)
	if tenantID.UUID() != value {
		t.Fatal("TenantID UUID mismatch")
	}
	if tenantID.String() != valStr {
		t.Fatal("TenantID String mismatch")
	}

	clientID := domain.ClientIDFromUUID(value)
	if clientID.UUID() != value {
		t.Fatal("ClientID UUID mismatch")
	}
	if clientID.String() != valStr {
		t.Fatal("ClientID String mismatch")
	}

	userID := domain.UserIDFromUUID(value)
	if userID.UUID() != value {
		t.Fatal("UserID UUID mismatch")
	}
	if userID.String() != valStr {
		t.Fatal("UserID String mismatch")
	}

	groupID := domain.GroupIDFromUUID(value)
	if groupID.UUID() != value {
		t.Fatal("GroupID UUID mismatch")
	}
	if groupID.String() != valStr {
		t.Fatal("GroupID String mismatch")
	}

	scopeID := domain.ScopeIDFromUUID(value)
	if scopeID.UUID() != value {
		t.Fatal("ScopeID UUID mismatch")
	}
	if scopeID.String() != valStr {
		t.Fatal("ScopeID String mismatch")
	}

	roleID := domain.RoleIDFromUUID(value)
	if roleID.UUID() != value {
		t.Fatal("RoleID UUID mismatch")
	}
	if roleID.String() != valStr {
		t.Fatal("RoleID String mismatch")
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

func TestDescriptions_String(t *testing.T) {
	t.Parallel()

	sd := mustScopeDescription(t, "scope description")
	if sd.String() != "scope description" {
		t.Fatal("unexpected scope description")
	}

	rd := mustRoleDescription(t, "role description")
	if rd.String() != "role description" {
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
