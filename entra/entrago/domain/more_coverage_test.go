package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func mustScopeValue(t *testing.T, raw string) ScopeValue {
	t.Helper()

	v, err := NewScopeValue(raw)
	if err != nil {
		t.Fatalf("NewScopeValue(%q): %v", raw, err)
	}

	return v
}

func mustScopeDescription(t *testing.T, raw string) ScopeDescription {
	t.Helper()

	v, err := NewScopeDescription(raw)
	if err != nil {
		t.Fatalf("NewScopeDescription(%q): %v", raw, err)
	}

	return v
}

func mustRoleValue(t *testing.T, raw string) RoleValue {
	t.Helper()

	v, err := NewRoleValue(raw)
	if err != nil {
		t.Fatalf("NewRoleValue(%q): %v", raw, err)
	}

	return v
}

func mustRoleDescription(t *testing.T, raw string) RoleDescription {
	t.Helper()

	v, err := NewRoleDescription(raw)
	if err != nil {
		t.Fatalf("NewRoleDescription(%q): %v", raw, err)
	}

	return v
}

func mustGroupName(t *testing.T, raw string) GroupName {
	t.Helper()

	v, err := NewGroupName(raw)
	if err != nil {
		t.Fatalf("NewGroupName(%q): %v", raw, err)
	}

	return v
}

func TestDomainError_ErrorString(t *testing.T) {
	t.Parallel()

	err := NewError(ErrCodeTenantNotFound, "no such tenant")
	if got := err.Error(); got != "TENANT_NOT_FOUND: no such tenant" {
		t.Fatalf("unexpected error string: %q", got)
	}
}

func TestConfig_ScopeRoleGroupGetters(t *testing.T) {
	t.Parallel()

	scope := NewScope(
		MustScopeID("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"),
		mustScopeValue(t, "access"),
		mustScopeDescription(t, "desc"),
	)
	_ = scope.ID()
	_ = scope.Description()

	role := NewRole(
		MustRoleID("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb"),
		mustRoleValue(t, "Admin"),
		mustRoleDescription(t, "desc"),
		[]Scope{scope},
	)
	_ = role.ID()
	_ = role.Description()

	group := NewGroup(
		MustGroupID("cccccccc-cccc-4ccc-accc-cccccccccccc"),
		mustGroupName(t, "Group"),
	)
	_ = group.ID()
	_ = group.Name()
}

func TestNonEmptyString_MustPanicsOnEmpty(t *testing.T) {
	t.Parallel()

	_ = MustNonEmptyString("x")

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	_ = MustNonEmptyString("")
}

func TestIDs_FromUUIDAndAccessors(t *testing.T) {
	t.Parallel()

	value := uuid.New()

	tenantID := TenantIDFromUUID(value)
	if tenantID.UUID() != value {
		t.Fatalf("TenantID UUID mismatch")
	}

	_ = tenantID.String()

	clientID := ClientIDFromUUID(value)
	if clientID.UUID() != value {
		t.Fatalf("ClientID UUID mismatch")
	}

	_ = clientID.String()

	userID := UserIDFromUUID(value)
	if userID.UUID() != value {
		t.Fatalf("UserID UUID mismatch")
	}

	_ = userID.String()

	groupID := GroupIDFromUUID(value)
	if groupID.UUID() != value {
		t.Fatalf("GroupID UUID mismatch")
	}

	_ = groupID.String()

	scopeID := ScopeIDFromUUID(value)
	if scopeID.UUID() != value {
		t.Fatalf("ScopeID UUID mismatch")
	}

	_ = scopeID.String()

	roleID := RoleIDFromUUID(value)
	if roleID.UUID() != value {
		t.Fatalf("RoleID UUID mismatch")
	}

	_ = roleID.String()
}

func TestClientSecret_IsEmpty(t *testing.T) {
	t.Parallel()

	if !NewClientSecret("").IsEmpty() {
		t.Fatalf("expected empty secret to be IsEmpty")
	}

	if NewClientSecret("secret").IsEmpty() {
		t.Fatalf("expected non-empty secret to not be IsEmpty")
	}
}

func TestDescriptions_String(t *testing.T) {
	t.Parallel()

	sd := mustScopeDescription(t, "scope description")
	if sd.String() != "scope description" {
		t.Fatalf("unexpected scope description")
	}

	rd := mustRoleDescription(t, "role description")
	if rd.String() != "role description" {
		t.Fatalf("unexpected role description")
	}
}

func TestStringWrappers_EmptyErrors(t *testing.T) {
	t.Parallel()

	if _, err := NewGroupID(""); err == nil {
		t.Fatalf("expected group id error")
	}

	if _, err := NewTenantName(""); !errors.Is(err, errTenantNameEmpty) {
		t.Fatalf("expected errTenantNameEmpty, got %v", err)
	}

	if _, err := NewAppName(""); !errors.Is(err, errAppNameEmpty) {
		t.Fatalf("expected errAppNameEmpty, got %v", err)
	}

	if _, err := NewIdentifierURI(""); !errors.Is(err, errIdentifierURIEmpty) {
		t.Fatalf("expected errIdentifierURIEmpty, got %v", err)
	}

	if _, err := NewScopeValue(""); !errors.Is(err, errScopeValueEmpty) {
		t.Fatalf("expected errScopeValueEmpty, got %v", err)
	}

	if _, err := NewRoleValue(""); !errors.Is(err, errRoleValueEmpty) {
		t.Fatalf("expected errRoleValueEmpty, got %v", err)
	}

	if _, err := NewGroupName(""); !errors.Is(err, errGroupNameEmpty) {
		t.Fatalf("expected errGroupNameEmpty, got %v", err)
	}

	if _, err := NewUsername(""); !errors.Is(err, errUsernameEmpty) {
		t.Fatalf("expected errUsernameEmpty, got %v", err)
	}

	if _, err := NewPassword(""); !errors.Is(err, errPasswordEmpty) {
		t.Fatalf("expected errPasswordEmpty, got %v", err)
	}

	if _, err := NewDisplayName(""); !errors.Is(err, errDisplayNameEmpty) {
		t.Fatalf("expected errDisplayNameEmpty, got %v", err)
	}

	if _, err := NewEmail(""); !errors.Is(err, errEmailEmpty) {
		t.Fatalf("expected errEmailEmpty, got %v", err)
	}

	if _, err := NewRedirectURL(""); !errors.Is(err, errRedirectURLEmpty) {
		t.Fatalf("expected errRedirectURLEmpty, got %v", err)
	}

	if _, err := NewScopeDescription(""); !errors.Is(err, errScopeDescriptionEmpty) {
		t.Fatalf("expected errScopeDescriptionEmpty, got %v", err)
	}

	if _, err := NewRoleDescription(""); !errors.Is(err, errRoleDescriptionEmpty) {
		t.Fatalf("expected errRoleDescriptionEmpty, got %v", err)
	}
}
