package domain_test

import (
	"testing"

	"identity/domain"
)

func TestMust_Panics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "MustTenantID",
			fn:   func() { domain.MustTenantID("invalid") },
		},
		{
			name: "MustClientID",
			fn:   func() { domain.MustClientID("invalid") },
		},
		{
			name: "MustUserID",
			fn:   func() { domain.MustUserID("invalid") },
		},
		{
			name: "MustGroupID",
			fn:   func() { domain.MustGroupID("invalid") },
		},
		{
			name: "MustScopeID",
			fn:   func() { domain.MustScopeID("invalid") },
		},
		{
			name: "MustRoleID",
			fn:   func() { domain.MustRoleID("invalid") },
		},
		{
			name: "MustTenantName",
			fn:   func() { domain.MustTenantName("") },
		},
		{
			name: "MustAppName",
			fn:   func() { domain.MustAppName("") },
		},
		{
			name: "MustUsername",
			fn:   func() { domain.MustUsername("") },
		},
		{
			name: "MustPassword",
			fn:   func() { domain.MustPassword("") },
		},
		{
			name: "MustDisplayName",
			fn:   func() { domain.MustDisplayName("") },
		},
		{
			name: "MustEmail",
			fn:   func() { domain.MustEmail("") },
		},
		{
			name: "MustNonEmptyString",
			fn:   func() { domain.MustNonEmptyString("") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("%s: expected panic", tt.name)
				}
			}()

			tt.fn()
		})
	}
}
