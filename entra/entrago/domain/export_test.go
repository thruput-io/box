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

// Exported for testing only.

// MockProvider is for testing domain.Parse.
type MockProvider struct {
	Val string
	Err error
}

func (m MockProvider) rawCallback(cb func(string) error) error {
	if m.Err != nil {
		return m.Err
	}

	return cb(m.Val)
}
