package domain_test

import (
	"errors"
	"testing"

	"identity/domain"
)

func TestNewNonEmptyString_EmptyStringReturnsError(t *testing.T) {
	t.Parallel()

	_, err := domain.NewNonEmptyString(emptyInput)
	if err == nil {
		t.Fatal(expectedError)
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf(expectedFormat, domain.ErrNonEmptyStringEmpty, err)
	}
}

func TestNewNonEmptyString_NonEmptyStringSucceeds(t *testing.T) {
	t.Parallel()

	const testVal = "x"

	result, err := domain.NewNonEmptyString(testVal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.String() != testVal {
		t.Fatalf("expected %q, got %q", testVal, result.String())
	}
}

func TestNewTenantName_Empty(t *testing.T) {
	t.Parallel()

	_, err := domain.NewTenantName(emptyInput)
	if err == nil {
		t.Fatal(expectedError)
	}

	if !errors.Is(err, domain.ErrTenantNameEmpty) {
		t.Fatalf(expectedFormat, domain.ErrTenantNameEmpty, err)
	}
}

func TestNewEmail_Empty(t *testing.T) {
	t.Parallel()

	_, err := domain.NewEmail(emptyInput)
	if err == nil {
		t.Fatal(expectedError)
	}

	if !errors.Is(err, domain.ErrEmailEmpty) {
		t.Fatalf(expectedFormat, domain.ErrEmailEmpty, err)
	}
}
