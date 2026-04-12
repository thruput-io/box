package domain_test

import (
	"testing"

	"identity/domain"
)

func mustClientID(t *testing.T, raw string) domain.ClientID {
	t.Helper()

	return domain.NewClientID(raw).MustRight()
}

func mustClientSecret(t *testing.T, raw string) domain.ClientSecret {
	t.Helper()

	return domain.NewClientSecret(raw).MustRight()
}

func mustTenantID(t *testing.T, raw string) domain.TenantID {
	t.Helper()

	return domain.NewTenantID(raw).MustRight()
}

func mustTenantName(t *testing.T, raw string) domain.TenantName {
	t.Helper()

	return domain.NewTenantName(raw).MustRight()
}

func mustUserID(t *testing.T, raw string) domain.UserID {
	t.Helper()

	return domain.NewUserID(raw).MustRight()
}

func mustUsername(t *testing.T, raw string) domain.Username {
	t.Helper()

	return domain.NewUsername(raw).MustRight()
}

func mustPassword(t *testing.T, raw string) domain.Password {
	t.Helper()

	return domain.NewPassword(raw).MustRight()
}

func mustScopeID(t *testing.T, raw string) domain.ScopeID {
	t.Helper()

	return domain.NewScopeID(raw).MustRight()
}

func mustScopeValue(t *testing.T, raw string) domain.ScopeValue {
	t.Helper()

	return domain.NewScopeValue(raw).MustRight()
}

func mustScopeDescription(t *testing.T, raw string) domain.ScopeDescription {
	t.Helper()

	return domain.NewScopeDescription(raw).MustRight()
}

func mustRoleID(t *testing.T, raw string) domain.RoleID {
	t.Helper()

	return domain.NewRoleID(raw).MustRight()
}

func mustRoleValue(t *testing.T, raw string) domain.RoleValue {
	t.Helper()

	return domain.NewRoleValue(raw).MustRight()
}

func mustRoleDescription(t *testing.T, raw string) domain.RoleDescription {
	t.Helper()

	return domain.NewRoleDescription(raw).MustRight()
}

func mustGroupID(t *testing.T, raw string) domain.GroupID {
	t.Helper()

	return domain.NewGroupID(raw).MustRight()
}

func mustGroupName(t *testing.T, raw string) domain.GroupName {
	t.Helper()

	return domain.NewGroupName(raw).MustRight()
}
