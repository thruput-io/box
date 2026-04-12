package domain_test

import (
	"testing"

	"identity/domain"
)

func TestMust_Panics(t *testing.T) {
	t.Parallel()

	runPanicTests(t, getStringPanicTests(), "")
	runPanicTests(t, getURLPanicTests(), "://invalid")
}

type panicTest struct {
	name string
	fn   func(string)
}

func getStringPanicTests() []panicTest {
	return []panicTest{
		{name: "MustAppName", fn: func(s string) { domain.MustAppName(s) }},
		{name: "MustIdentifierURI", fn: func(s string) { domain.MustIdentifierURI(s) }},
		{name: "MustDisplayName", fn: func(s string) { domain.MustDisplayName(s) }},
		{name: "MustEmail", fn: func(s string) { domain.MustEmail(s) }},
	}
}

func getURLPanicTests() []panicTest {
	return []panicTest{
		{name: "MustBaseURL", fn: func(s string) { domain.MustBaseURL(s) }},
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
