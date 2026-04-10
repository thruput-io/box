package domain

import "errors"

const (
	emptyLen    = 0
	emptyString = ""
)

var (
	errNonEmptyStringEmpty   = errors.New("string must not be empty")
	errTenantNameEmpty       = errors.New("tenant name must not be empty")
	errAppNameEmpty          = errors.New("app name must not be empty")
	errIdentifierURIEmpty    = errors.New("identifier URI must not be empty")
	errScopeValueEmpty       = errors.New("scope value must not be empty")
	errRoleValueEmpty        = errors.New("role value must not be empty")
	errGroupNameEmpty        = errors.New("group name must not be empty")
	errUsernameEmpty         = errors.New("username must not be empty")
	errPasswordEmpty         = errors.New("password must not be empty")
	errDisplayNameEmpty      = errors.New("display name must not be empty")
	errEmailEmpty            = errors.New("email must not be empty")
	errRedirectURLEmpty      = errors.New("redirect URL must not be empty")
	errClientSecretEmpty     = errors.New("client secret must not be empty")
	errScopeDescriptionEmpty = errors.New("scope description must not be empty")
	errRoleDescriptionEmpty  = errors.New("role description must not be empty")
	errConfigNoTenants       = errors.New("config must contain at least one tenant")
	errNonceEmpty            = errors.New("nonce must not be empty")
	errCorrelationIDEmpty    = errors.New("correlation ID must not be empty")
	errBaseURLEmpty          = errors.New("base URL must not be empty")
	errBaseURLInvalid        = errors.New("base URL is not a valid URL")
	errIssuerEmpty           = errors.New("issuer must not be empty")
	errTokenVersionEmpty     = errors.New("token version must not be empty")
	errOAuthStateEmpty       = errors.New("OAuth state must not be empty")
	errResponseModeEmpty     = errors.New("response mode must not be empty")
	errResponseTypeEmpty     = errors.New("response type must not be empty")
)
