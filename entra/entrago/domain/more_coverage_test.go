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

	scope := domain.NewScope(
		domain.MustScopeID("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"),
		mustScopeValue(t, "access"),
		mustScopeDescription(t, "desc"),
	)
	_ = scope.ID()
	_ = scope.Description()

	role := domain.NewRole(
		domain.MustRoleID("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb"),
		mustRoleValue(t, "Admin"),
		mustRoleDescription(t, "desc"),
		[]domain.Scope{scope},
	)
	_ = role.ID()
	_ = role.Description()

	group := domain.NewGroup(
		domain.MustGroupID("cccccccc-cccc-4ccc-accc-cccccccccccc"),
		mustGroupName(t, "Group"),
	)
	_ = group.ID()
	_ = group.Name()
}

func TestNonEmptyString_MustPanicsOnEmpty(t *testing.T) {
	t.Parallel()

	_ = domain.MustNonEmptyString("x")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()

	_ = domain.MustNonEmptyString(emptyInput)
}

func TestIDs_FromUUIDAndAccessors(t *testing.T) {
	t.Parallel()

	value := uuid.New()

	tenantID := domain.TenantIDFromUUID(value)
	if tenantID.UUID() != value {
		t.Fatal("TenantID UUID mismatch")
	}

	_ = tenantID.String()

	clientID := domain.ClientIDFromUUID(value)
	if clientID.UUID() != value {
		t.Fatal("ClientID UUID mismatch")
	}

	_ = clientID.String()

	userID := domain.UserIDFromUUID(value)
	if userID.UUID() != value {
		t.Fatal("UserID UUID mismatch")
	}

	_ = userID.String()

	groupID := domain.GroupIDFromUUID(value)
	if groupID.UUID() != value {
		t.Fatal("GroupID UUID mismatch")
	}

	_ = groupID.String()

	scopeID := domain.ScopeIDFromUUID(value)
	if scopeID.UUID() != value {
		t.Fatal("ScopeID UUID mismatch")
	}

	_ = scopeID.String()

	roleID := domain.RoleIDFromUUID(value)
	if roleID.UUID() != value {
		t.Fatal("RoleID UUID mismatch")
	}

	_ = roleID.String()
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
