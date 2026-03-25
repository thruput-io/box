package main

import (
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
)

func TestResolveAudience(t *testing.T) {
	u1 := uuid.MustParse("11111111-1111-4111-a111-111111111111")
	u2 := uuid.MustParse("22222222-2222-4222-a222-222222222222")
	tenant := &Tenant{
		AppRegistrations: []AppRegistration{
			{
				Name:          "App1",
				ClientID:      u1,
				IdentifierUri: "api://app1",
			},
			{
				Name:          "App2",
				ClientID:      u2,
				IdentifierUri: "api://app2",
				AppRoles: []Role{
					{
						Value: "Role1",
						Scopes: []Scope{
							{Value: "scope1"},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name           string
		scope          string
		wantAudience   string
		wantAppIDs     map[string]bool
	}{
		{
			name:         "Empty scope",
			scope:        "",
			wantAudience: "api://default",
			wantAppIDs:   map[string]bool{},
		},
		{
			name:         "Exact ClientID match",
			scope:        u1.String(),
			wantAudience: "api://app1",
			wantAppIDs:   map[string]bool{u1.String(): true},
		},
		{
			name:         "Exact IdentifierUri match",
			scope:        "api://app1",
			wantAudience: "api://app1",
			wantAppIDs:   map[string]bool{u1.String(): true},
		},
		{
			name:         ".default suffix",
			scope:        "api://app1/.default",
			wantAudience: "api://app1",
			wantAppIDs:   map[string]bool{u1.String(): true},
		},
		{
			name:         "Scope match",
			scope:        "scope1",
			wantAudience: "api://app2",
			wantAppIDs:   map[string]bool{u2.String(): true},
		},
		{
			name:         "Scope with URI prefix",
			scope:        "api://app2/scope1",
			wantAudience: "api://app2",
			wantAppIDs:   map[string]bool{u2.String(): true},
		},
		{
			name:         "Multiple scopes",
			scope:        "openid profile api://app1/.default",
			wantAudience: "api://app1",
			wantAppIDs:   map[string]bool{u1.String(): true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAud, gotAppIDs := resolveAudience(tenant, tt.scope)
			if gotAud != tt.wantAudience {
				t.Errorf("resolveAudience() gotAud = %v, want %v", gotAud, tt.wantAudience)
			}
			if !reflect.DeepEqual(gotAppIDs, tt.wantAppIDs) {
				t.Errorf("resolveAudience() gotAppIDs = %v, want %v", gotAppIDs, tt.wantAppIDs)
			}
		})
	}
}

func TestResolveRoles(t *testing.T) {
	api1 := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	tenant := &Tenant{
		AppRegistrations: []AppRegistration{
			{
				ClientID:      api1,
				IdentifierUri: "api://api-1",
				AppRoles: []Role{
					{
						Value: "Admin",
						Scopes: []Scope{
							{Value: "admin-scope"},
						},
					},
					{
						Value: "Reader",
						Scopes: []Scope{
							{Value: "read-scope"},
						},
					},
				},
			},
		},
	}

	client := &Client{
		GroupRoleAssignments: []GroupRoleAssignment{
			{
				GroupName:     "group-1",
				Roles:         []string{"Reader"},
				ApplicationID: api1,
			},
			{
				GroupName:     "group-2",
				Roles:         []string{"Admin"},
				ApplicationID: api1,
			},
		},
	}

	user := &User{
		Groups: []string{"group-1"},
	}

	tests := []struct {
		name            string
		client          *Client
		user            *User
		targetAppIDs    map[string]bool
		requestedScopes []string
		wantRoles       []string
	}{
		{
			name:            "User role from group",
			client:          client,
			user:            user,
			targetAppIDs:    map[string]bool{api1.String(): true},
			requestedScopes: []string{"api://api-1/.default"},
			wantRoles:       []string{"Reader"},
		},
		{
			name:            "Client credentials (no user)",
			client:          client,
			user:            nil,
			targetAppIDs:    map[string]bool{api1.String(): true},
			requestedScopes: []string{"api://api-1/.default"},
			wantRoles:       []string{"Admin", "Reader"},
		},
		{
			name:            "Scope to Role mapping",
			client:          nil,
			user:            nil,
			targetAppIDs:    map[string]bool{api1.String(): true},
			requestedScopes: []string{"admin-scope"},
			wantRoles:       []string{"Admin"},
		},
		{
			name:            "Filter by targetAppIDs",
			client:          client,
			user:            nil,
			targetAppIDs:    map[string]bool{"other-api": true},
			requestedScopes: []string{"api://api-1/.default"},
			wantRoles:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveRoles(tenant, tt.client, tt.user, tt.targetAppIDs, tt.requestedScopes)
			sort.Strings(got)
			sort.Strings(tt.wantRoles)
			if !reflect.DeepEqual(got, tt.wantRoles) {
				t.Errorf("resolveRoles() = %v, want %v", got, tt.wantRoles)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "Valid config",
			yaml: `
tenants:
  - tenant_id: "b5a920d6-7d3c-44fe-baad-4ffed6b8774d"
    name: "Default Tenant"
    app_registrations:
      - name: "App1"
        client_id: "33333333-3333-4333-a333-333333333333"
        identifier_uri: "api://app1"
        scopes: []
        app_roles: []
    groups: []
    users: []
    clients: []
`,
			wantErr: false,
		},
		{
			name: "Invalid UUID",
			yaml: `
tenants:
  - tenant_id: "not-a-uuid"
    name: "Default Tenant"
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
