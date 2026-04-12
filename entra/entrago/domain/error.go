package domain

import "errors"

// ErrorCode identifies a domain error using a stable, machine-readable value.
type ErrorCode string

// Error codes returned by domain operations.
const (
	ErrCodeTenantNotFound ErrorCode = "TENANT_NOT_FOUND"

	ErrCodeAppNotFound ErrorCode = "APP_NOT_FOUND"

	ErrCodeClientNotFound ErrorCode = "CLIENT_NOT_FOUND"

	ErrCodeUserNotFound ErrorCode = "USER_NOT_FOUND"

	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"

	ErrCodeInvalidRedirectURI ErrorCode = "INVALID_REDIRECT_URI"

	ErrCodeInvalidRequest ErrorCode = "INVALID_REQUEST"

	ErrCodeUnsupportedGrantType ErrorCode = "UNSUPPORTED_GRANT_TYPE"

	ErrCodeInvalidGrant ErrorCode = "INVALID_GRANT"

	ErrCodeInvalidConfig ErrorCode = "INVALID_CONFIG"

	ErrCodeClientSecretRequired ErrorCode = "CLIENT_SECRET_REQUIRED"

	ErrCodePublicClientDoesNotAcceptSecrets ErrorCode = "PUBLIC_CLIENT_DOES_NOT_ACCEPT_SECRETS"

	ErrCodeNonEmptyStringEmpty ErrorCode = "NON_EMPTY_STRING_EMPTY"
	ErrCodeNonEmptyArrayEmpty  ErrorCode = "NON_EMPTY_ARRAY_EMPTY"
	ErrCodeRedirectURLInvalid  ErrorCode = "REDIRECT_URL_INVALID"
	ErrCodeBaseURLInvalid      ErrorCode = "BASE_URL_INVALID"

	ErrCodeTenantIDInvalid ErrorCode = "TENANT_ID_INVALID"
	ErrCodeClientIDInvalid ErrorCode = "CLIENT_ID_INVALID"
	ErrCodeUserIDInvalid   ErrorCode = "USER_ID_INVALID"
	ErrCodeGroupIDInvalid  ErrorCode = "GROUP_ID_INVALID"
	ErrCodeScopeIDInvalid  ErrorCode = "SCOPE_ID_INVALID"
	ErrCodeRoleIDInvalid   ErrorCode = "ROLE_ID_INVALID"

	ErrCodeClaimValueInvalid  ErrorCode = "CLAIM_VALUE_INVALID"
	ErrCodeClaimTypeMismatch  ErrorCode = "CLAIM_TYPE_MISMATCH"
	ErrCodeClaimShapeInvalid  ErrorCode = "CLAIM_SHAPE_INVALID"
	ErrCodeClaimKeyNotAllowed ErrorCode = "CLAIM_KEY_NOT_ALLOWED"
)

// Sentinel domain errors.
var (
	ErrTenantNotFound = NewError(ErrCodeTenantNotFound, "tenant not found")

	ErrAppNotFound = NewError(ErrCodeAppNotFound, "app registration not found")

	ErrClientNotFound = NewError(ErrCodeClientNotFound, "client not found")

	ErrUserNotFound = NewError(ErrCodeUserNotFound, "user not found")

	ErrInvalidCredentials = NewError(ErrCodeInvalidCredentials, "invalid credentials")

	ErrInvalidRedirectURI = NewError(ErrCodeInvalidRedirectURI, "invalid redirect URI")

	ErrInvalidConfig = NewError(ErrCodeInvalidConfig, "invalid config")

	ErrUnsupportedGrantType = NewError(ErrCodeUnsupportedGrantType, "unsupported grant_type")

	ErrClientSecretRequired = NewError(ErrCodeClientSecretRequired, "client secret required")

	ErrPublicClientDoesNotAcceptSecrets = NewError(
		ErrCodePublicClientDoesNotAcceptSecrets,
		"public client does not accept secrets",
	)

	ErrNonEmptyArrayEmpty = NewError(
		ErrCodeNonEmptyArrayEmpty,
		"non-empty array cannot be empty",
	)

	ErrClaimValueInvalid = NewError(
		ErrCodeClaimValueInvalid,
		"claim value is invalid",
	)

	ErrClaimTypeMismatch = NewError(
		ErrCodeClaimTypeMismatch,
		"claim value type mismatch",
	)

	ErrClaimShapeInvalid = NewError(
		ErrCodeClaimShapeInvalid,
		"claim value shape is invalid",
	)

	ErrClaimKeyNotAllowed = NewError(
		ErrCodeClaimKeyNotAllowed,
		"claim key is not allowlisted",
	)
)

// Error represents a domain error with a stable code and a human-readable message.
type Error struct {
	Code    ErrorCode
	Message string
}

// NewError creates a domain Error.
func NewError(code ErrorCode, message string) Error {
	return Error{Code: code, Message: message}
}

// Error formats the domain error as `<code>: <message>`.
func (domainError Error) Error() string {
	return string(domainError.Code) + ": " + domainError.Message
}

// Is reports whether the target error is a domain Error with the same ErrorCode.
func (domainError Error) Is(target error) bool {
	var targetDomainErr Error
	if !errors.As(target, &targetDomainErr) {
		return false
	}

	return domainError.Code == targetDomainErr.Code
}
