package domain

// ErrorCode is a unique machine-readable identifier for a domain error.
type ErrorCode string

const (
	// ErrCodeTenantNotFound is returned when no matching tenant exists.
	ErrCodeTenantNotFound ErrorCode = "TENANT_NOT_FOUND"

	// ErrCodeAppNotFound is returned when no matching app registration exists.
	ErrCodeAppNotFound ErrorCode = "APP_NOT_FOUND"

	// ErrCodeClientNotFound is returned when no matching client exists.
	ErrCodeClientNotFound ErrorCode = "CLIENT_NOT_FOUND"

	// ErrCodeUserNotFound is returned when no matching user exists.
	ErrCodeUserNotFound ErrorCode = "USER_NOT_FOUND"

	// ErrCodeInvalidCredentials is returned when authentication fails.
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"

	// ErrCodeInvalidRedirectURI is returned when the redirect URI is not permitted.
	ErrCodeInvalidRedirectURI ErrorCode = "INVALID_REDIRECT_URI"

	// ErrCodeInvalidRequest is returned when a required field is missing or malformed.
	ErrCodeInvalidRequest ErrorCode = "INVALID_REQUEST"

	// ErrCodeUnsupportedGrantType is returned when the grant_type is not supported.
	ErrCodeUnsupportedGrantType ErrorCode = "UNSUPPORTED_GRANT_TYPE"

	// ErrCodeInvalidGrant is returned when an auth code or refresh token is invalid.
	ErrCodeInvalidGrant ErrorCode = "INVALID_GRANT"

	// ErrCodeInvalidConfig is returned when the config file fails validation.
	ErrCodeInvalidConfig ErrorCode = "INVALID_CONFIG"
)

// Error is a domain error with a unique code and a human-readable message.
// It has no dependency on transport (HTTP, Kafka, etc.).
type Error struct {
	Code    ErrorCode
	Message string
}

// NewError constructs a domain Error.
func NewError(code ErrorCode, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Error implements the error interface.
func (domainError *Error) Error() string {
	return string(domainError.Code) + ": " + domainError.Message
}
