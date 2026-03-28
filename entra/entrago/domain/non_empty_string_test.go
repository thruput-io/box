package domain_test

import (
	"errors"
	"testing"

	"identity/domain"
)

func TestNewNonEmptyString_EmptyStringReturnsError(t *testing.T) {
	t.Parallel()

	_, err := domain.NewNonEmptyString("")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, domain.ErrNonEmptyStringEmpty) {
		t.Fatalf("expected %v, got %v", domain.ErrNonEmptyStringEmpty, err)
	}
}

func TestNewNonEmptyString_NonEmptyStringSucceeds(t *testing.T) {
	t.Parallel()

	result, err := domain.NewNonEmptyString("x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.String() != "x" {
		t.Fatalf("expected %q, got %q", "x", result.String())
	}
}

func TestNewTenantName_Empty(t *testing.T) {
	t.Parallel()

	_, err := domain.NewTenantName("")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, domain.ErrTenantNameEmpty) {
		t.Fatalf("expected %v, got %v", domain.ErrTenantNameEmpty, err)
	}
}

func TestNewEmail_Empty(t *testing.T) {
	t.Parallel()

	_, err := domain.NewEmail("")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, domain.ErrEmailEmpty) {
		t.Fatalf("expected %v, got %v", domain.ErrEmailEmpty, err)
	}
}
