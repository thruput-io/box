package domain_test

import (
	"errors"
	"testing"

	"identity/domain"
)

func TestError_Is_matchesByCode(t *testing.T) {
	t.Parallel()

	custom := domain.NewError(domain.ErrCodeInvalidCredentials, "invalid username or password")

	if !errors.Is(custom, domain.ErrInvalidCredentials) {
		t.Fatal("expected errors.Is(custom, ErrInvalidCredentials) to be true")
	}

	if !domain.ErrInvalidCredentials.Is(custom) {
		t.Fatal("expected ErrInvalidCredentials.Is(custom) to be true")
	}

	if errors.Is(custom, domain.ErrTenantNotFound) {
		t.Fatal("expected errors.Is(custom, ErrTenantNotFound) to be false")
	}
}
