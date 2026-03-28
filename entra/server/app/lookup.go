package app

import (
	"crypto/subtle"

	"identity/domain"
)

// FindTenant returns the tenant matching tenantID, or the first tenant for "" or "common".
func FindTenant(config domain.Config, tenantID string) (domain.Tenant, error) {
	if tenantID == "" || tenantID == "common" {
		return config.Tenants()[0], nil
	}

	for _, tenant := range config.Tenants() {
		if tenant.TenantID().String() == tenantID {
			return tenant, nil
		}
	}

	return domain.Tenant{}, domain.ErrTenantNotFound
}

// FindTenantByID returns the tenant matching the given domain.TenantID exactly.
func FindTenantByID(config domain.Config, tenantID domain.TenantID) (domain.Tenant, error) {
	for _, tenant := range config.Tenants() {
		if tenant.TenantID().String() == tenantID.String() {
			return tenant, nil
		}
	}

	return domain.Tenant{}, domain.ErrTenantNotFound
}

// FindClient returns the client matching clientID within a tenant.
func FindClient(tenant domain.Tenant, clientID domain.ClientID) (domain.Client, error) {
	for _, client := range tenant.Clients() {
		if client.ClientID().String() == clientID.String() {
			return client, nil
		}
	}

	return domain.Client{}, domain.ErrClientNotFound
}

// FindAppRegistration returns the app registration matching clientID within a tenant.
func FindAppRegistration(tenant domain.Tenant, clientID domain.ClientID) (domain.AppRegistration, error) {
	for _, registration := range tenant.AppRegistrations() {
		if registration.ClientID().String() == clientID.String() {
			return registration, nil
		}
	}

	return domain.AppRegistration{}, domain.ErrAppNotFound
}

// FindRedirectURLs returns allowed redirect URLs for a clientID, searching clients then app registrations.
func FindRedirectURLs(tenant domain.Tenant, clientID domain.ClientID) ([]domain.RedirectURL, error) {
	for _, client := range tenant.Clients() {
		if client.ClientID().String() == clientID.String() {
			return client.RedirectURLs(), nil
		}
	}

	for _, registration := range tenant.AppRegistrations() {
		if registration.ClientID().String() == clientID.String() {
			return registration.RedirectURLs(), nil
		}
	}

	return nil, domain.ErrClientNotFound
}

// ValidateRedirectURI checks whether redirectURI is in the allowed list.
func ValidateRedirectURI(redirectURI string, allowed []domain.RedirectURL) error {
	for _, allowedURL := range allowed {
		if allowedURL.String() == redirectURI {
			return nil
		}
	}

	return domain.ErrInvalidRedirectURI
}

// AuthenticateUser returns the user matching username and password using constant-time comparison.
func AuthenticateUser(tenant domain.Tenant, username, password string) (domain.User, error) {
	for _, user := range tenant.Users() {
		usernameMatch := subtle.ConstantTimeCompare([]byte(user.Username().String()), []byte(username)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(user.Password().String()), []byte(password)) == 1

		if usernameMatch && passwordMatch {
			return user, nil
		}
	}

	return domain.User{}, domain.ErrInvalidCredentials
}

// FindUserByID returns the user matching subject (UUID string), and whether it was found.
func FindUserByID(tenant domain.Tenant, subject string) (domain.User, bool) {
	for _, user := range tenant.Users() {
		if user.ID().String() == subject {
			return user, true
		}
	}

	return domain.User{}, false
}

// ValidateClientSecret checks the client secret using constant-time comparison.
func ValidateClientSecret(client domain.Client, secret string) error {
	if subtle.ConstantTimeCompare([]byte(client.ClientSecret().String()), []byte(secret)) != 1 {
		return domain.ErrInvalidCredentials
	}

	return nil
}
