package app

import (
	"slices"

	"identity/domain"
)

// FindTenant returns the tenant matching tenantID, or the first tenant for "" or "common".
func FindTenant(config domain.Config, tenantIDStr string) (domain.Tenant, error) {
	if tenantIDStr == "" || tenantIDStr == "common" {
		return config.Tenants()[0], nil
	}

	id, err := domain.NewTenantID(tenantIDStr)
	if err != nil {
		return domain.Tenant{}, err
	}

	return FindTenantByID(config, id)
}

// FindTenantByID returns the tenant matching the given domain.TenantID exactly.
func FindTenantByID(config domain.Config, tenantID domain.TenantID) (domain.Tenant, error) {
	for _, tenant := range config.Tenants() {
		if tenant.TenantID() == tenantID {
			return tenant, nil
		}
	}

	return domain.Tenant{}, domain.ErrTenantNotFound
}

// FindClient returns the client matching clientID within a tenant.
func FindClient(tenant domain.Tenant, clientID domain.ClientID) (domain.Client, error) {
	for _, client := range tenant.Clients() {
		if client.ClientID() == clientID {
			return client, nil
		}
	}

	return nil, domain.ErrClientNotFound
}

// FindAppRegistration returns the app registration matching clientID within a tenant.
func FindAppRegistration(tenant domain.Tenant, clientID domain.ClientID) (domain.AppRegistration, error) {
	for _, registration := range tenant.AppRegistrations() {
		if registration.ClientID() == clientID {
			return registration, nil
		}
	}

	return domain.AppRegistration{}, domain.ErrAppNotFound
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

	return nil, domain.ErrClientNotFound
}

// ValidateRedirectURI checks whether redirectURI is in the allowed list.
func ValidateRedirectURI(redirectURI domain.RedirectURL, allowed []domain.RedirectURL) error {
	if slices.Contains(allowed, redirectURI) {
		return nil
	}

	return domain.ErrInvalidRedirectURI
}

func AuthenticateUser(tenant domain.Tenant, username domain.Username, password domain.Password) (domain.User, error) {
	for _, user := range tenant.Users() {
		if user.Username() == username && user.Password() == password {
			return user, nil
		}
	}

	return domain.User{}, domain.ErrInvalidCredentials
}

func FindUserByID(tenant domain.Tenant, id domain.UserID) (domain.User, bool) {
	for _, user := range tenant.Users() {
		if user.ID() == id {
			return user, true
		}
	}

	return domain.User{}, false
}

// ValidateClientSecret checks the client secret using constant-time comparison.
func ValidateClientSecret(client domain.Client, secret *domain.ClientSecret) error {
	err := client.Validate(secret)
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	return nil
}
