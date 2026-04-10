package domain_test

import (
	"testing"

	"identity/domain"
)

func TestMust_Panics(t *testing.T) {
	t.Parallel()

	runPanicTests(t, getIDPanicTests(), "invalid")
	runPanicTests(t, getStringPanicTests(), "")
}

type panicTest struct {
	name string
	fn   func(string)
}

func getIDPanicTests() []panicTest {
	return []panicTest{
		{name: "MustTenantID", fn: func(s string) { domain.MustTenantID(s) }},
		{name: "MustClientID", fn: func(s string) { domain.MustClientID(s) }},
		{name: "MustUserID", fn: func(s string) { domain.MustUserID(s) }},
		{name: "MustGroupID", fn: func(s string) { domain.MustGroupID(s) }},
		{name: "MustScopeID", fn: func(s string) { domain.MustScopeID(s) }},
		{name: "MustRoleID", fn: func(s string) { domain.MustRoleID(s) }},
	}
}

func getStringPanicTests() []panicTest {
	return []panicTest{
		{name: "MustTenantName", fn: func(s string) { domain.MustTenantName(s) }},
		{name: "MustAppName", fn: func(s string) { domain.MustAppName(s) }},
		{name: "MustUsername", fn: func(s string) { domain.MustUsername(s) }},
		{name: "MustPassword", fn: func(s string) { domain.MustPassword(s) }},
		{name: "MustDisplayName", fn: func(s string) { domain.MustDisplayName(s) }},
		{name: "MustEmail", fn: func(s string) { domain.MustEmail(s) }},
		{name: "MustNonEmptyString", fn: func(s string) { domain.MustNonEmptyString(s) }},
	}
}

func runPanicTests(t *testing.T, tests []panicTest, input string) {
	t.Helper()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertPanic(t, tc.fn, tc.name, input)
		})
	}
}

func assertPanic(t *testing.T, panicFunc func(string), name, input string) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected panic", name)
		}
	}()

	panicFunc(input)
}
