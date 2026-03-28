package domain

import (
	"errors"
	"testing"
)

func TestNewNonEmptyString(t *testing.T) {
	t.Run("empty string returns error", func(t *testing.T) {
		_, err := NewNonEmptyString("")
		if err == nil {
			t.Fatalf("expected error")
		}

		if !errors.Is(err, errNonEmptyStringEmpty) {
			t.Fatalf("expected %v, got %v", errNonEmptyStringEmpty, err)
		}
	})

	t.Run("non-empty string succeeds", func(t *testing.T) {
		v, err := NewNonEmptyString("x")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v.String() != "x" {
			t.Fatalf("expected %q, got %q", "x", v.String())
		}
	})
}

func TestNewTenantName_Empty(t *testing.T) {
	_, err := NewTenantName("")
	if err == nil {
		t.Fatalf("expected error")
	}

	if !errors.Is(err, errTenantNameEmpty) {
		t.Fatalf("expected %v, got %v", errTenantNameEmpty, err)
	}
}

func TestNewEmail_Empty(t *testing.T) {
	_, err := NewEmail("")
	if err == nil {
		t.Fatalf("expected error")
	}

	if !errors.Is(err, errEmailEmpty) {
		t.Fatalf("expected %v, got %v", errEmailEmpty, err)
	}
}
