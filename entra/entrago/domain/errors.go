package domain

import "errors"

// ErrTenantNotFound is returned when a tenant cannot be located by ID.
var ErrTenantNotFound = errors.New("tenant not found")

// ErrAppNotFound is returned when an app registration cannot be located by client ID.
var ErrAppNotFound = errors.New("app registration not found")

// ErrClientNotFound is returned when a client cannot be located by client ID.
var ErrClientNotFound = errors.New("client not found")

// ErrUserNotFound is returned when a user cannot be located by username or ID.
var ErrUserNotFound = errors.New("user not found")

// ErrInvalidCredentials is returned when username/password or client secret does not match.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrInvalidRedirectURI is returned when the redirect URI is not in the allowed list.
var ErrInvalidRedirectURI = errors.New("invalid redirect URI")

// ErrInvalidConfig is returned when the config file fails schema validation.
var ErrInvalidConfig = errors.New("invalid config")

// ErrUnsupportedGrantType is returned when the grant_type is not supported.
var ErrUnsupportedGrantType = errors.New("unsupported grant_type")

// ErrTenantNoAppRegistrations is returned when a tenant has no app registrations.
var ErrTenantNoAppRegistrations = errors.New("tenant must have at least one app registration")

// ErrTenantNoUsers is returned when a tenant has no users.
var ErrTenantNoUsers = errors.New("tenant must have at least one user")

// ErrClientSecretRequired is returned when a client secret is needed but not provided.
var ErrClientSecretRequired = errors.New("client secret required")

// ErrPublicClientDoesNotAcceptSecrets is returned when a secret is provided for a public client.
var ErrPublicClientDoesNotAcceptSecrets = errors.New("public client does not accept secrets")
