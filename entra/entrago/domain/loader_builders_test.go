package domain_test

import (
	"testing"

	"identity/domain"
)

const expectedOneAssignment = 1

func TestBuildScope_Success(t *testing.T) {
	t.Parallel()

	scope, err := domain.BuildScope(domain.RawScope{
		ID:          "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa",
		Value:       "access",
		Description: "desc",
	})
	if err != nil {
		t.Fatalf("BuildScope: %v", err)
	}

	_ = scope.ID()
	_ = scope.Description()
}

func TestBuildRole_Success(t *testing.T) {
	t.Parallel()

	role, err := domain.BuildRole(domain.RawRole{
		ID:          "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb",
		Value:       "Admin",
		Description: "desc",
		Scopes: []domain.RawScope{{
			ID:          "cccccccc-cccc-4ccc-accc-cccccccccccc",
			Value:       "scope",
			Description: "desc",
		}},
	})
	if err != nil {
		t.Fatalf("BuildRole: %v", err)
	}

	_ = role.ID()
	_ = role.Description()
}

func TestBuildClient_WithGroupRoleAssignment(t *testing.T) {
	t.Parallel()

	client, err := domain.BuildClient(domain.RawClient{
		Name:         "Client",
		ClientID:     "22222222-2222-4222-8222-222222222222",
		ClientSecret: "secret",
		RedirectURLs: []string{"https://example.com/callback"},
		GroupRoleAssignments: []domain.RawGroupRoleAssignment{{
			GroupName:     "GroupA",
			Roles:         []string{"RoleA"},
			ApplicationID: "22222222-2222-4222-8222-222222222222",
		}},
	})
	if err != nil {
		t.Fatalf("BuildClient: %v", err)
	}

	if len(client.GroupRoleAssignments()) != expectedOneAssignment {
		t.Fatal("expected 1 assignment")
	}
}

func TestBuildGroupRoleAssignments_Success(t *testing.T) {
	t.Parallel()

	assignments, err := domain.BuildGroupRoleAssignments([]domain.RawGroupRoleAssignment{{
		GroupName:     "GroupA",
		Roles:         []string{"RoleA"},
		ApplicationID: "22222222-2222-4222-8222-222222222222",
	}})
	if err != nil {
		t.Fatalf("BuildGroupRoleAssignments: %v", err)
	}

	if len(assignments) != expectedOneAssignment {
		t.Fatal("expected 1 assignment")
	}
}

func TestBuildGroupRoleAssignment_Success(t *testing.T) {
	t.Parallel()

	assignment, err := domain.BuildGroupRoleAssignment(domain.RawGroupRoleAssignment{
		GroupName:     "GroupA",
		Roles:         []string{"RoleA"},
		ApplicationID: "22222222-2222-4222-8222-222222222222",
	})
	if err != nil {
		t.Fatalf("BuildGroupRoleAssignment: %v", err)
	}

	if assignment.GroupName().String() != "GroupA" {
		t.Fatal("unexpected group name")
	}
}
