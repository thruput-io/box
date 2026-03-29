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
	RawUser                = rawUser
	RawAppRegistration     = rawAppRegistration
	RawTenant              = rawTenant
	RawGroup               = rawGroup
)

// Exported builder wrapper functions for testing only.

func ExportValidateYAML(yamlData []byte, schemaPath string) error {
	return validateYAML(yamlData, schemaPath)
}

func ExportBuildRedirectURLs(raws []string) ([]RedirectURL, error) {
	return buildRedirectURLs(raws)
}

func ExportBuildUserGroups(username string, rawGroups []string) ([]GroupName, error) {
	return buildUserGroups(username, rawGroups)
}

func ExportBuildScope(raw rawScope) (Scope, error) {
	return buildScope(raw)
}

func ExportBuildRole(raw rawRole) (Role, error) {
	return buildRole(raw)
}

func ExportBuildClient(raw rawClient) (Client, error) {
	return buildClient(raw)
}

func ExportBuildGroupRoleAssignments(raws []rawGroupRoleAssignment) ([]GroupRoleAssignment, error) {
	return buildGroupRoleAssignments(raws)
}

func ExportBuildGroupRoleAssignment(raw rawGroupRoleAssignment) (GroupRoleAssignment, error) {
	return buildGroupRoleAssignment(raw)
}

func ExportBuildUser(raw rawUser) (User, error) {
	return buildUser(raw)
}

func ExportBuildAppRegistration(raw rawAppRegistration) (AppRegistration, error) {
	return buildAppRegistration(raw)
}

func ExportBuildTenant(raw rawTenant) (Tenant, error) {
	return buildTenant(raw)
}

func ExportBuildGroups(raws []rawGroup) ([]Group, error) {
	return buildGroups(raws)
}
