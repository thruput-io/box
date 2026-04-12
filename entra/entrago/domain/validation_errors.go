package domain

const (
	emptyLen    = 0
	emptyString = ""
)

var (
	errNonEmptyStringEmpty = NewError(ErrCodeNonEmptyStringEmpty, "string must not be empty")
	errRedirectURLInvalid  = NewError(ErrCodeRedirectURLInvalid, "redirect URL is not a valid URL")
	errBaseURLInvalid      = NewError(ErrCodeBaseURLInvalid, "base URL is not a valid URL")

	errTenantIDInvalid = NewError(ErrCodeTenantIDInvalid, "invalid tenant ID")
	errClientIDInvalid = NewError(ErrCodeClientIDInvalid, "invalid client ID")
	errUserIDInvalid   = NewError(ErrCodeUserIDInvalid, "invalid user ID")
	errGroupIDInvalid  = NewError(ErrCodeGroupIDInvalid, "invalid group ID")
	errScopeIDInvalid  = NewError(ErrCodeScopeIDInvalid, "invalid scope ID")
	errRoleIDInvalid   = NewError(ErrCodeRoleIDInvalid, "invalid role ID")
)
