package domain

import "testing"

func TestLoader_BuildScopeRoleClientAndAssignments(t *testing.T) {
	t.Parallel()

	scope, err := buildScope(rawScope{
		ID:          "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa",
		Value:       "access",
		Description: "desc",
	})
	if err != nil {
		t.Fatalf("buildScope: %v", err)
	}

	_ = scope.ID()
	_ = scope.Description()

	role, err := buildRole(rawRole{
		ID:          "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb",
		Value:       "Admin",
		Description: "desc",
		Scopes: []rawScope{{
			ID:          "cccccccc-cccc-4ccc-accc-cccccccccccc",
			Value:       "scope",
			Description: "desc",
		}},
	})
	if err != nil {
		t.Fatalf("buildRole: %v", err)
	}

	_ = role.ID()
	_ = role.Description()

	client, err := buildClient(rawClient{
		Name:         "Client",
		ClientID:     "22222222-2222-4222-8222-222222222222",
		ClientSecret: "secret",
		RedirectURLs: []string{"https://example.com/callback"},
		GroupRoleAssignments: []rawGroupRoleAssignment{{
			GroupName:     "GroupA",
			Roles:         []string{"RoleA"},
			ApplicationID: "22222222-2222-4222-8222-222222222222",
		}},
	})
	if err != nil {
		t.Fatalf("buildClient: %v", err)
	}

	if len(client.GroupRoleAssignments()) != 1 {
		t.Fatalf("expected 1 assignment")
	}

	assignments, err := buildGroupRoleAssignments([]rawGroupRoleAssignment{{
		GroupName:     "GroupA",
		Roles:         []string{"RoleA"},
		ApplicationID: "22222222-2222-4222-8222-222222222222",
	}})
	if err != nil {
		t.Fatalf("buildGroupRoleAssignments: %v", err)
	}

	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment")
	}

	assignment, err := buildGroupRoleAssignment(rawGroupRoleAssignment{
		GroupName:     "GroupA",
		Roles:         []string{"RoleA"},
		ApplicationID: "22222222-2222-4222-8222-222222222222",
	})
	if err != nil {
		t.Fatalf("buildGroupRoleAssignment: %v", err)
	}

	if assignment.GroupName().String() != "GroupA" {
		t.Fatalf("unexpected group name")
	}
}
