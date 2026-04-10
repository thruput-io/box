package domain

// Exported sentinel errors for testing only.
var (
	ErrNonEmptyStringEmpty   = errNonEmptyStringEmpty
	ErrTenantNameEmpty       = errTenantNameEmpty
	ErrAppNameEmpty          = errAppNameEmpty
	ErrIdentifierURIEmpty    = errIdentifierURIEmpty
	ErrScopeValueEmpty       = errScopeValueEmpty
	ErrRoleValueEmpty        = errRoleValueEmpty
	ErrGroupNameEmpty        = errGroupNameEmpty
	ErrUsernameEmpty         = errUsernameEmpty
	ErrPasswordEmpty         = errPasswordEmpty
	ErrDisplayNameEmpty      = errDisplayNameEmpty
	ErrEmailEmpty            = errEmailEmpty
	ErrRedirectURLEmpty      = errRedirectURLEmpty
	ErrScopeDescriptionEmpty = errScopeDescriptionEmpty
	ErrRoleDescriptionEmpty  = errRoleDescriptionEmpty
)

// MustParse extracts the underlying string from a domain value type.
// For use in tests and constants only.
func MustParse(v interface{ Value() string }) string {
	return v.Value()
}
