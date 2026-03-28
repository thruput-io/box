package app

import (
	"testing"

	"identity/domain"
)

func mustIdentifierURI(t *testing.T, raw string) domain.IdentifierURI {
	t.Helper()

	v, err := domain.NewIdentifierURI(raw)
	if err != nil {
		t.Fatalf("NewIdentifierURI(%q): %v", raw, err)
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
