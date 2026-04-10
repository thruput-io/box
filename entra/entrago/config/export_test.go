package config

import "identity/domain"

// Exported for testing only.

func ExportValidateYAML(yamlData []byte, schemaPath string) error {
	return validateYAML(yamlData, schemaPath)
}

func ExportBuildRedirectURLs(raws []string) ([]domain.RedirectURL, error) {
	return buildRedirectURLs(raws)
}

func ExportBuildUserGroups(username string, rawGroups []string) ([]domain.GroupName, error) {
	return buildUserGroups(username, rawGroups)
}

func ExportBuildScope(raw RawScope) (domain.Scope, error) {
	return buildScope(raw)
}

func ExportBuildRole(raw RawRole) (domain.Role, error) {
	return buildRole(raw)
}

func ExportBuildClient(raw RawClient) (*domain.Client, error) {
	return buildClient(raw)
}

func ExportBuildGroupRoleAssignments(raws []RawGroupRoleAssignment) ([]domain.GroupRoleAssignment, error) {
	return buildGroupRoleAssignments(raws)
}

func ExportBuildGroupRoleAssignment(raw RawGroupRoleAssignment) (domain.GroupRoleAssignment, error) {
	return buildGroupRoleAssignment(raw)
}

func ExportBuildUser(raw RawUser) (*domain.User, error) {
	return buildUser(raw)
}

func ExportBuildAppRegistration(raw RawAppRegistration) (*domain.AppRegistration, error) {
	return buildAppRegistration(raw)
}

func ExportBuildTenant(raw RawTenant) (*domain.Tenant, error) {
	return buildTenant(raw)
}
