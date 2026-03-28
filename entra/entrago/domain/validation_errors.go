package domain

import "errors"

const (
	emptyLen      = 0
	emptyString   = ""
	fmtTenantWrap = "tenant %q: %w"
	fmtAppRegWrap = "app registration %q: %w"
	fmtRoleWrap   = "role %q: %w"
	fmtUserWrap   = "user %q: %w"
	fmtClientWrap = "client %q: %w"
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
	errScopeDescriptionEmpty = errors.New("scope description must not be empty")
	errRoleDescriptionEmpty  = errors.New("role description must not be empty")
	errConfigNoTenants       = errors.New("config must contain at least one tenant")
)
