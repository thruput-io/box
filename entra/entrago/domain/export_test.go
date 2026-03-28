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

// Exported raw types for testing only.
type (
	RawScope               = rawScope
	RawRole                = rawRole
	RawClient              = rawClient
	RawGroupRoleAssignment = rawGroupRoleAssignment
)

// Exported builder wrapper functions for testing only.

func ValidateYAML(yamlData []byte, schemaPath string) error {
	return validateYAML(yamlData, schemaPath)
}

func BuildRedirectURLs(raws []string) ([]RedirectURL, error) {
	return buildRedirectURLs(raws)
}

func BuildUserGroups(username string, rawGroups []string) ([]GroupName, error) {
	return buildUserGroups(username, rawGroups)
}

func BuildScope(raw rawScope) (Scope, error) {
	return buildScope(raw)
}

func BuildRole(raw rawRole) (Role, error) {
	return buildRole(raw)
}

func BuildClient(raw rawClient) (Client, error) {
	return buildClient(raw)
}

func BuildGroupRoleAssignments(raws []rawGroupRoleAssignment) ([]GroupRoleAssignment, error) {
	return buildGroupRoleAssignments(raws)
}

func BuildGroupRoleAssignment(raw rawGroupRoleAssignment) (GroupRoleAssignment, error) {
	return buildGroupRoleAssignment(raw)
}
