package main

import (
	"testing"

	"identity/domain"
)

func mustTenantID(t *testing.T, raw string) domain.TenantID {
	t.Helper()

	return domain.NewTenantID(raw).MustRight()
}

func mustTenantName(t *testing.T, raw string) domain.TenantName {
	t.Helper()

	return domain.NewTenantName(raw).MustRight()
}

func mustClientID(t *testing.T, raw string) domain.ClientID {
	t.Helper()

	return domain.NewClientID(raw).MustRight()
}

func mustUserID(t *testing.T, raw string) domain.UserID {
	t.Helper()

	return domain.NewUserID(raw).MustRight()
}

func mustUsername(t *testing.T, raw string) domain.Username {
	t.Helper()

	return domain.NewUsername(raw).MustRight()
}

func mustPassword(t *testing.T, raw string) domain.Password {
	t.Helper()

	return domain.NewPassword(raw).MustRight()
}

func mustAppName(t *testing.T, raw string) domain.AppName {
	t.Helper()

	return domain.NewAppName(raw).MustRight()
}

func mustIdentifierURI(t *testing.T, raw string) domain.IdentifierURI {
	t.Helper()

	return domain.NewIdentifierURI(raw).MustRight()
}

func mustRedirectURL(t *testing.T, raw string) domain.RedirectURL {
	t.Helper()

	return domain.NewRedirectURL(raw).MustRight()
}

func mustDisplayName(t *testing.T, raw string) domain.DisplayName {
	t.Helper()

	return domain.NewDisplayName(raw).MustRight()
}

func mustEmail(t *testing.T, raw string) domain.Email {
	t.Helper()

	return domain.NewEmail(raw).MustRight()
}

func mustClientSecret(t *testing.T, raw string) domain.ClientSecret {
	t.Helper()

	return domain.NewClientSecret(raw).MustRight()
}
