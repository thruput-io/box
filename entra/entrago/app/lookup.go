package app

import (
	"fmt"
	"slices"

	"identity/domain"
)

const firstTenantIndex = 0

// FindTenant returns the tenant matching tenantID, or the first tenant for "" or "common".
func FindTenant(config *domain.Config, tenantIDStr string) (*domain.Tenant, error) {
	if config == nil {
		return nil, domain.ErrTenantNotFound
	}

	if tenantIDStr == "" || tenantIDStr == "common" {
		return &config.Tenants()[firstTenantIndex], nil
	}

	either := domain.NewTenantID(tenantIDStr)

	id, ok := either.Right()

	if !ok {
		err, _ := either.Left()

		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	return FindTenantByID(config, id)
}

// FindTenantByID returns the tenant matching the given domain.TenantID exactly.
func FindTenantByID(config *domain.Config, tenantID domain.TenantID) (*domain.Tenant, error) {
	if config == nil {
		return nil, domain.ErrTenantNotFound
	}

	for _, tenant := range config.Tenants() {
		if tenant.TenantID() == tenantID {
			return &tenant, nil
		}
	}

	return nil, domain.ErrTenantNotFound
}

// FindClient returns the client matching clientID within a tenant.
func FindClient(tenant domain.Tenant, clientID domain.ClientID) (*domain.Client, error) {
	for _, client := range tenant.Clients() {
		if client.ClientID() == clientID {
			return &client, nil
		}
	}

	return nil, domain.ErrClientNotFound
}

// FindAppRegistration returns the app registration matching clientID within a tenant.
func FindAppRegistration(tenant domain.Tenant, clientID domain.ClientID) (*domain.AppRegistration, error) {
	for _, registration := range tenant.AppRegistrations() {
		if registration.ClientID() == clientID {
			return &registration, nil
		}
	}

	return nil, domain.ErrAppNotFound
}

// FindRedirectURLs returns allowed redirect URLs for a clientID, searching clients then app registrations.
func FindRedirectURLs(tenant domain.Tenant, clientID domain.ClientID) ([]domain.RedirectURL, error) {
	for _, client := range tenant.Clients() {
		if client.ClientID() == clientID {
			return client.RedirectURLs(), nil
		}
	}

	for _, registration := range tenant.AppRegistrations() {
		if registration.ClientID() == clientID {
			return registration.RedirectURLs(), nil
		}
	}

	return []domain.RedirectURL{}, domain.ErrClientNotFound
}

// ValidateRedirectURI checks whether redirectURI is in the allowed list.
func ValidateRedirectURI(redirectURI domain.RedirectURL, allowed []domain.RedirectURL) error {
	if slices.Contains(allowed, redirectURI) {
		return nil
	}

	return domain.ErrInvalidRedirectURI
}

// AuthenticateUser attempts to find a user with matching credentials.
func AuthenticateUser(tenant domain.Tenant, username domain.Username, password domain.Password) (*domain.User, error) {
	for _, user := range tenant.Users() {
		if user.Username() == username && user.Password() == password {
			return &user, nil
		}
	}

	return nil, domain.ErrInvalidCredentials
}

// FindUserByID returns the user with the given ID, if found.
func FindUserByID(tenant domain.Tenant, id domain.UserID) (*domain.User, bool) {
	for _, user := range tenant.Users() {
		if user.ID() == id {
			return &user, true
		}
	}

	return nil, false
}

// ValidateClientSecret checks the client secret using constant-time comparison.
func ValidateClientSecret(client domain.Client, secret *domain.ClientSecret) error {
	if client.Validate(secret).IsLeft() {
		return domain.ErrInvalidCredentials
	}

	return nil
}
