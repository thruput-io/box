package domain_test

import (
	"errors"
	"testing"

	"identity/domain"
)

func TestNewNonEmptyString_EmptyStringReturnsError(t *testing.T) {
	t.Parallel()

	err, ok := domain.NewNonEmptyString(emptyInput).Left()
	if !ok {
		t.Fatal(expectedError)
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf(expectedFormat, domain.ErrNonEmptyStringEmpty, err)
	}
}
