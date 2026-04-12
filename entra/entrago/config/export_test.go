package config

import (
	"github.com/samber/mo"

	"identity/domain"
)

// Exported for testing only.

func ExportValidateYAML(yamlData []byte, schemaPath string) mo.Either[domain.Error, []byte] {
	return validateYAML(yamlData, schemaPath)
}

func ExportBuildRedirectURLs(raws []string) mo.Either[domain.Error, []domain.RedirectURL] {
	return buildRedirectURLs(raws)
}

func ExportBuildUserGroups(username string, rawGroups []string) mo.Either[domain.Error, []domain.GroupName] {
	return buildUserGroups(username, rawGroups)
}

func ExportBuildScope(raw RawScope) mo.Either[domain.Error, domain.Scope] {
	return buildScope(raw)
}

func ExportBuildRole(raw RawRole) mo.Either[domain.Error, domain.Role] {
	return buildRole(raw)
}

func ExportBuildClient(raw RawClient) mo.Either[domain.Error, domain.Client] {
	return buildClient(raw)
}

func ExportBuildGroupRoleAssignments(raws []RawGroupRoleAssignment) mo.Either[domain.Error, []domain.GroupRoleAssignment] {
	return buildGroupRoleAssignments(raws)
}

func ExportBuildGroupRoleAssignment(raw RawGroupRoleAssignment) mo.Either[domain.Error, domain.GroupRoleAssignment] {
	return buildGroupRoleAssignment(raw)
}

func ExportBuildUser(raw RawUser) mo.Either[domain.Error, domain.User] {
	return buildUser(raw)
}

func ExportBuildAppRegistration(raw RawAppRegistration) mo.Either[domain.Error, domain.AppRegistration] {
	return buildAppRegistration(raw)
}

func ExportBuildTenant(raw RawTenant) mo.Either[domain.Error, domain.Tenant] {
	return buildTenant(raw)
}
