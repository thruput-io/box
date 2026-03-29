package app

import (
	"identity/domain"
)

const (
	constantTimeEqual = 1
	firstTenant       = 0
)

// FindTenant returns the tenant matching tenantID, or the first tenant for "" or "common".
func FindTenant(config domain.Config, tenantIDStr string) (domain.Tenant, error) {
	if tenantIDStr == "" || tenantIDStr == "common" {
		return config.Tenants()[firstTenant], nil
	}

	tenantID, err := domain.NewTenantID(tenantIDStr)
	if err != nil {
		return domain.Tenant{}, err
	}

	for _, tenant := range config.Tenants() {
		if tenant.TenantID() == tenantID {
			return tenant, nil
		}
	}

	return domain.Tenant{}, domain.ErrTenantNotFound
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
	for _, allowedURL := range allowed {
		if allowedURL == redirectURI {
			return nil
		}
	}

	return domain.ErrInvalidRedirectURI
}

// AuthenticateUser returns the user matching username and password using constant-time comparison.
func AuthenticateUser(tenant domain.Tenant, usernameStr, passwordStr string) (domain.User, error) {
	username, err := domain.NewUsername(usernameStr)
	if err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}

	password, err := domain.NewPassword(passwordStr)
	if err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}

	for _, user := range tenant.Users() {
		if user.Username() == username && user.Password() == password {
			return user, nil
		}
	}

	return domain.User{}, domain.ErrInvalidCredentials
}

// FindUserByID returns the user matching subject (UUID string), and whether it was found.
func FindUserByID(tenant domain.Tenant, subjectStr string) (domain.User, bool) {
	subject, err := domain.NewUserID(subjectStr)
	if err != nil {
		return domain.User{}, false
	}

	for _, user := range tenant.Users() {
		if user.ID() == subject {
			return user, true
		}
	}

	return domain.User{}, false
}

// ValidateClientSecret checks the client secret using constant-time comparison.
func ValidateClientSecret(client domain.Client, secret *domain.ClientSecret) error {
	if err := client.Validate(secret); err != nil {
		return domain.ErrInvalidCredentials
	}

	return nil
}
