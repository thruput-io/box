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
