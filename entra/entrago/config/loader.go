// Package config handles the loading and parsing of the mock server configuration.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/samber/mo"
	moeither "github.com/samber/mo/either"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"

	"identity/domain"
)

const (
	emptyLen      = 0
	fmtTenantWrap = "tenant %q: %s"
	fmtAppRegWrap = "app registration %q: %s"
	fmtRoleWrap   = "role %q: %s"
	fmtUserWrap   = "user %q: %s"
	fmtClientWrap = "client %q: %s"
)

func wrap(err domain.Error, format string, args ...any) domain.Error {
	return domain.NewError(err.Code, fmt.Sprintf(format, append(args, err.Message)...))
}

func invalidConfig(message string) domain.Error {
	return domain.NewError(domain.ErrCodeInvalidConfig, message)
}

func invalidConfigf(format string, args ...any) domain.Error {
	return invalidConfig(fmt.Sprintf(format, args...))
}

// RawConfig mirrors the YAML structure exactly for unmarshalling.
// snake_case keys match the config file format.
type RawConfig struct {
	Tenants []RawTenant `yaml:"tenants"`
}

// RawTenant mirrors the YAML structure for a tenant.
type RawTenant struct {
	TenantID         string               `yaml:"tenantId"`
	Name             string               `yaml:"name"`
	AppRegistrations []RawAppRegistration `yaml:"appRegistrations"`
	Groups           []RawGroup           `yaml:"groups"`
	Users            []RawUser            `yaml:"users"`
	Clients          []RawClient          `yaml:"clients"`
}

// RawAppRegistration mirrors the YAML structure for an app registration.
type RawAppRegistration struct {
	Name          string     `yaml:"name"`
	ClientID      string     `yaml:"clientId"`
	IdentifierURI string     `yaml:"identifierUri"`
	RedirectURLs  []string   `yaml:"redirectUrls"`
	Scopes        []RawScope `yaml:"scopes"`
	AppRoles      []RawRole  `yaml:"appRoles"`
}

// RawScope mirrors the YAML structure for a scope.
type RawScope struct {
	ID          string `yaml:"id"`
	Value       string `yaml:"value"`
	Description string `yaml:"description"`
}

// RawRole mirrors the YAML structure for an app role.
type RawRole struct {
	ID          string     `yaml:"id"`
	Value       string     `yaml:"value"`
	Description string     `yaml:"description"`
	Scopes      []RawScope `yaml:"scopes"`
}

// RawGroup mirrors the YAML structure for a group.
type RawGroup struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

// RawUser mirrors the YAML structure for a user.
type RawUser struct {
	ID          string   `yaml:"id"`
	Username    string   `yaml:"username"`
	Password    string   `yaml:"password"`
	DisplayName string   `yaml:"displayName"`
	Email       string   `yaml:"email"`
	Groups      []string `yaml:"groups"`
}

// RawClient mirrors the YAML structure for a client.
type RawClient struct {
	Name                 string                   `yaml:"name"`
	ClientID             string                   `yaml:"clientId"`
	ClientSecret         string                   `yaml:"clientSecret"`
	RedirectURLs         []string                 `yaml:"redirectUrls"`
	GroupRoleAssignments []RawGroupRoleAssignment `yaml:"groupRoleAssignments"`
}

// RawGroupRoleAssignment mirrors the YAML structure for a group role assignment.
type RawGroupRoleAssignment struct {
	GroupName     string   `yaml:"groupName"`
	Roles         []string `yaml:"roles"`
	ApplicationID string   `yaml:"applicationId"`
}

// LoadConfig reads, validates, and parses Config.yaml into an immutable domain Config.

func LoadConfig(configPath, schemaPath string) mo.Either[domain.Error, domain.Config] {
	return moeither.Pipe3(
		readFile(configPath),
		moeither.FlatMapRight(func(data []byte) mo.Either[domain.Error, []byte] {
			return validateYAML(data, schemaPath)
		}),
		moeither.FlatMapRight(unmarshalRawConfig),
		moeither.FlatMapRight(buildConfig),
	)
}

func readFile(path string) mo.Either[domain.Error, []byte] {
	data, err := os.ReadFile(path)
	if err != nil {
		return mo.Left[domain.Error, []byte](invalidConfigf("failed to read %s: %v", path, err))
	}

	return mo.Right[domain.Error](data)
}

func unmarshalRawConfig(data []byte) mo.Either[domain.Error, RawConfig] {
	var raw RawConfig
	err := yaml.Unmarshal(data, &raw)
	if err != nil {
		return mo.Left[domain.Error, RawConfig](invalidConfigf("failed to parse config YAML: %v", err))
	}

	return mo.Right[domain.Error](raw)
}

func validateYAML(yamlData []byte, schemaPath string) mo.Either[domain.Error, []byte] {
	if schemaPath == "" {
		return mo.Right[domain.Error](yamlData)
	}

	var raw any

	if err := yaml.Unmarshal(yamlData, &raw); err != nil {
		return mo.Left[domain.Error, []byte](invalidConfigf("failed to parse YAML for schema validation: %v", err))
	}

	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return mo.Left[domain.Error, []byte](invalidConfigf("failed to read schema %s: %v", schemaPath, err))
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schemaData),
		gojsonschema.NewGoLoader(raw),
	)
	if err != nil {
		return mo.Left[domain.Error, []byte](invalidConfigf("schema validation failed: %v", err))
	}

	if !result.Valid() {
		messages := make([]string, emptyLen, len(result.Errors()))
		for _, desc := range result.Errors() {
			messages = append(messages, desc.String())
		}

		return mo.Left[domain.Error, []byte](
			invalidConfigf("schema violations: %s", strings.Join(messages, "; ")),
		)
	}

	return mo.Right[domain.Error](yamlData)
}

func buildConfig(raw RawConfig) mo.Either[domain.Error, domain.Config] {
	return moeither.Pipe1(
		buildTenants(raw.Tenants),
		moeither.FlatMapRight(domain.NewConfig),
	)
}

func buildTenants(raws []RawTenant) mo.Either[domain.Error, domain.NonEmptyArray[domain.Tenant]] {
	tenants := make([]domain.Tenant, emptyLen, len(raws))

	for _, rawTenant := range raws {
		tenantEither := buildTenant(rawTenant)

		tenant, ok := tenantEither.Right()

		if !ok {
			domErr, _ := tenantEither.Left()

			return mo.Left[domain.Error, domain.NonEmptyArray[domain.Tenant]](domErr)
		}

		tenants = append(tenants, tenant)
	}

	return domain.NewNonEmptyArray(tenants...)
}

func buildTenant(raw RawTenant) mo.Either[domain.Error, domain.Tenant] {
	tenantIDEither := domain.NewTenantID(raw.TenantID)

	tenantID, ok := tenantIDEither.Right()

	if !ok {
		domErr, _ := tenantIDEither.Left()

		return mo.Left[domain.Error, domain.Tenant](wrap(domErr, fmtTenantWrap, raw.Name))
	}

	nameEither := domain.NewTenantName(raw.Name)

	name, ok := nameEither.Right()

	if !ok {
		domErr, _ := nameEither.Left()

		return mo.Left[domain.Error, domain.Tenant](wrap(domErr, fmtTenantWrap, raw.Name))
	}

	appRegistrationsEither := buildAppRegistrations(raw.AppRegistrations)

	appRegistrations, ok := appRegistrationsEither.Right()

	if !ok {
		domErr, _ := appRegistrationsEither.Left()

		return mo.Left[domain.Error, domain.Tenant](wrap(domErr, fmtTenantWrap, raw.Name))
	}

	groupsEither := buildGroups(raw.Groups)

	groups, ok := groupsEither.Right()

	if !ok {
		domErr, _ := groupsEither.Left()

		return mo.Left[domain.Error, domain.Tenant](wrap(domErr, fmtTenantWrap, raw.Name))
	}

	usersEither := buildUsers(raw.Users)

	users, ok := usersEither.Right()

	if !ok {
		domErr, _ := usersEither.Left()

		return mo.Left[domain.Error, domain.Tenant](wrap(domErr, fmtTenantWrap, raw.Name))
	}

	clientsEither := buildClients(raw.Clients)

	clients, ok := clientsEither.Right()

	if !ok {
		domErr, _ := clientsEither.Left()

		return mo.Left[domain.Error, domain.Tenant](wrap(domErr, fmtTenantWrap, raw.Name))
	}

	tenantEither := domain.NewTenant(tenantID, name, appRegistrations, groups, users, clients)

	ten, ok := tenantEither.Right()

	if !ok {
		domErr, _ := tenantEither.Left()

		return mo.Left[domain.Error, domain.Tenant](wrap(domErr, fmtTenantWrap, raw.Name))
	}

	return mo.Right[domain.Error](ten)
}

func buildAppRegistrations(raws []RawAppRegistration) mo.Either[domain.Error, domain.NonEmptyArray[domain.AppRegistration]] {
	result := make([]domain.AppRegistration, emptyLen, len(raws))

	for _, raw := range raws {
		appRegistrationEither := buildAppRegistration(raw)

		appRegistration, ok := appRegistrationEither.Right()

		if !ok {
			domErr, _ := appRegistrationEither.Left()

			return mo.Left[domain.Error, domain.NonEmptyArray[domain.AppRegistration]](domErr)
		}

		result = append(result, appRegistration)
	}

	return domain.NewNonEmptyArray(result...)
}

func buildAppRegistration(raw RawAppRegistration) mo.Either[domain.Error, domain.AppRegistration] {
	nameEither := domain.NewAppName(raw.Name)

	name, ok := nameEither.Right()

	if !ok {
		domErr, _ := nameEither.Left()

		return mo.Left[domain.Error, domain.AppRegistration](wrap(domErr, fmtAppRegWrap, raw.Name))
	}

	clientIDEither := domain.NewClientID(raw.ClientID)

	clientID, ok := clientIDEither.Right()

	if !ok {
		domErr, _ := clientIDEither.Left()

		return mo.Left[domain.Error, domain.AppRegistration](wrap(domErr, fmtAppRegWrap, raw.Name))
	}

	identifierURIEither := domain.NewIdentifierURI(raw.IdentifierURI)

	identifierURI, ok := identifierURIEither.Right()

	if !ok {
		domErr, _ := identifierURIEither.Left()

		return mo.Left[domain.Error, domain.AppRegistration](wrap(domErr, fmtAppRegWrap, raw.Name))
	}

	redirectURLsEither := buildRedirectURLs(raw.RedirectURLs)

	redirectURLs, ok := redirectURLsEither.Right()

	if !ok {
		domErr, _ := redirectURLsEither.Left()

		return mo.Left[domain.Error, domain.AppRegistration](wrap(domErr, fmtAppRegWrap, raw.Name))
	}

	scopesEither := buildScopes(raw.Scopes)

	scopes, ok := scopesEither.Right()

	if !ok {
		domErr, _ := scopesEither.Left()

		return mo.Left[domain.Error, domain.AppRegistration](wrap(domErr, fmtAppRegWrap, raw.Name))
	}

	rolesEither := buildRoles(raw.AppRoles)

	roles, ok := rolesEither.Right()

	if !ok {
		domErr, _ := rolesEither.Left()

		return mo.Left[domain.Error, domain.AppRegistration](wrap(domErr, fmtAppRegWrap, raw.Name))
	}

	return mo.Right[domain.Error](domain.NewAppRegistration(name, clientID, identifierURI, redirectURLs, scopes, roles))
}

func buildRedirectURLs(raws []string) mo.Either[domain.Error, []domain.RedirectURL] {
	result := make([]domain.RedirectURL, emptyLen, len(raws))

	for _, raw := range raws {
		redirectURLEither := domain.NewRedirectURL(raw)

		redirectURL, ok := redirectURLEither.Right()

		if !ok {
			domErr, _ := redirectURLEither.Left()

			return mo.Left[domain.Error, []domain.RedirectURL](
				domain.NewError(domErr.Code, fmt.Sprintf("redirect URL %q: %s", raw, domErr.Message)),
			)
		}

		result = append(result, redirectURL)
	}

	return mo.Right[domain.Error](result)
}

func buildScopes(raws []RawScope) mo.Either[domain.Error, []domain.Scope] {
	result := make([]domain.Scope, emptyLen, len(raws))

	for _, raw := range raws {
		scopeEither := buildScope(raw)

		scope, ok := scopeEither.Right()

		if !ok {
			domErr, _ := scopeEither.Left()

			return mo.Left[domain.Error, []domain.Scope](domErr)
		}

		result = append(result, scope)
	}

	return mo.Right[domain.Error](result)
}

func buildScope(raw RawScope) mo.Either[domain.Error, domain.Scope] {
	scopeIDEither := domain.NewScopeID(raw.ID)

	scopeID, ok := scopeIDEither.Right()

	if !ok {
		domErr, _ := scopeIDEither.Left()

		return mo.Left[domain.Error, domain.Scope](domain.NewError(domErr.Code, fmt.Sprintf("scope %q: %s", raw.Value, domErr.Message)))
	}

	valueEither := domain.NewScopeValue(raw.Value)

	value, ok := valueEither.Right()

	if !ok {
		domErr, _ := valueEither.Left()

		return mo.Left[domain.Error, domain.Scope](domain.NewError(domErr.Code, fmt.Sprintf("scope %q: %s", raw.Value, domErr.Message)))
	}

	descEither := domain.NewScopeDescription(raw.Description)

	description, ok := descEither.Right()

	if !ok {
		domErr, _ := descEither.Left()

		return mo.Left[domain.Error, domain.Scope](domain.NewError(domErr.Code, fmt.Sprintf("scope %q: %s", raw.Value, domErr.Message)))
	}

	return mo.Right[domain.Error](domain.NewScope(scopeID, value, description))
}

func buildRoles(raws []RawRole) mo.Either[domain.Error, []domain.Role] {
	result := make([]domain.Role, emptyLen, len(raws))

	for _, raw := range raws {
		roleEither := buildRole(raw)

		role, ok := roleEither.Right()

		if !ok {
			domErr, _ := roleEither.Left()

			return mo.Left[domain.Error, []domain.Role](domErr)
		}

		result = append(result, role)
	}

	return mo.Right[domain.Error](result)
}

func buildRole(raw RawRole) mo.Either[domain.Error, domain.Role] {
	roleIDEither := domain.NewRoleID(raw.ID)

	roleID, ok := roleIDEither.Right()

	if !ok {
		domErr, _ := roleIDEither.Left()

		return mo.Left[domain.Error, domain.Role](wrap(domErr, fmtRoleWrap, raw.Value))
	}

	valueEither := domain.NewRoleValue(raw.Value)

	value, ok := valueEither.Right()

	if !ok {
		domErr, _ := valueEither.Left()

		return mo.Left[domain.Error, domain.Role](wrap(domErr, fmtRoleWrap, raw.Value))
	}

	descEither := domain.NewRoleDescription(raw.Description)

	description, ok := descEither.Right()

	if !ok {
		domErr, _ := descEither.Left()

		return mo.Left[domain.Error, domain.Role](wrap(domErr, fmtRoleWrap, raw.Value))
	}

	scopesEither := buildScopes(raw.Scopes)

	scopes, ok := scopesEither.Right()

	if !ok {
		domErr, _ := scopesEither.Left()

		return mo.Left[domain.Error, domain.Role](wrap(domErr, fmtRoleWrap, raw.Value))
	}

	return mo.Right[domain.Error](domain.NewRole(roleID, value, description, scopes))
}

func buildGroups(raws []RawGroup) mo.Either[domain.Error, []domain.Group] {
	result := make([]domain.Group, emptyLen, len(raws))

	for _, raw := range raws {
		groupIDEither := domain.NewGroupID(raw.ID)

		groupID, ok := groupIDEither.Right()

		if !ok {
			domErr, _ := groupIDEither.Left()

			return mo.Left[domain.Error, []domain.Group](domain.NewError(domErr.Code, fmt.Sprintf("group %q: %s", raw.Name, domErr.Message)))
		}

		nameEither := domain.NewGroupName(raw.Name)

		name, ok := nameEither.Right()

		if !ok {
			domErr, _ := nameEither.Left()

			return mo.Left[domain.Error, []domain.Group](domain.NewError(domErr.Code, fmt.Sprintf("group %q: %s", raw.Name, domErr.Message)))
		}

		result = append(result, domain.NewGroup(groupID, name))
	}

	return mo.Right[domain.Error](result)
}

func buildUsers(raws []RawUser) mo.Either[domain.Error, domain.NonEmptyArray[domain.User]] {
	result := make([]domain.User, emptyLen, len(raws))

	for _, raw := range raws {
		userEither := buildUser(raw)

		user, ok := userEither.Right()

		if !ok {
			domErr, _ := userEither.Left()

			return mo.Left[domain.Error, domain.NonEmptyArray[domain.User]](domErr)
		}

		result = append(result, user)
	}

	return domain.NewNonEmptyArray(result...)
}

func buildUser(raw RawUser) mo.Either[domain.Error, domain.User] {
	userIDEither := domain.NewUserID(raw.ID)

	userID, ok := userIDEither.Right()

	if !ok {
		domErr, _ := userIDEither.Left()

		return mo.Left[domain.Error, domain.User](wrap(domErr, fmtUserWrap, raw.Username))
	}

	usernameEither := domain.NewUsername(raw.Username)

	username, ok := usernameEither.Right()

	if !ok {
		domErr, _ := usernameEither.Left()

		return mo.Left[domain.Error, domain.User](wrap(domErr, fmtUserWrap, raw.Username))
	}

	passwordEither := domain.NewPassword(raw.Password)

	password, ok := passwordEither.Right()

	if !ok {
		domErr, _ := passwordEither.Left()

		return mo.Left[domain.Error, domain.User](wrap(domErr, fmtUserWrap, raw.Username))
	}

	displayNameEither := domain.NewDisplayName(raw.DisplayName)

	displayName, ok := displayNameEither.Right()

	if !ok {
		domErr, _ := displayNameEither.Left()

		return mo.Left[domain.Error, domain.User](wrap(domErr, fmtUserWrap, raw.Username))
	}

	emailEither := domain.NewEmail(raw.Email)

	email, ok := emailEither.Right()

	if !ok {
		domErr, _ := emailEither.Left()

		return mo.Left[domain.Error, domain.User](wrap(domErr, fmtUserWrap, raw.Username))
	}

	groupsEither := buildUserGroups(raw.Username, raw.Groups)

	groups, ok := groupsEither.Right()

	if !ok {
		domErr, _ := groupsEither.Left()

		return mo.Left[domain.Error, domain.User](wrap(domErr, fmtUserWrap, raw.Username))
	}

	return mo.Right[domain.Error](domain.NewUser(userID, username, password, displayName, email, groups))
}

func buildUserGroups(username string, rawGroups []string) mo.Either[domain.Error, []domain.GroupName] {
	groups := make([]domain.GroupName, emptyLen, len(rawGroups))

	for _, groupName := range rawGroups {
		nameEither := domain.NewGroupName(groupName)

		name, ok := nameEither.Right()

		if !ok {
			domErr, _ := nameEither.Left()

			return mo.Left[domain.Error, []domain.GroupName](
				domain.NewError(domErr.Code, fmt.Sprintf("user %q group %q: %s", username, groupName, domErr.Message)),
			)
		}

		groups = append(groups, name)
	}

	return mo.Right[domain.Error](groups)
}

func buildClients(raws []RawClient) mo.Either[domain.Error, []domain.Client] {
	result := make([]domain.Client, emptyLen, len(raws))

	for _, raw := range raws {
		clientEither := buildClient(raw)

		client, ok := clientEither.Right()

		if !ok {
			domErr, _ := clientEither.Left()

			return mo.Left[domain.Error, []domain.Client](domErr)
		}

		result = append(result, client)
	}

	return mo.Right[domain.Error](result)
}

func buildClient(raw RawClient) mo.Either[domain.Error, domain.Client] {
	nameEither := domain.NewAppName(raw.Name)

	name, ok := nameEither.Right()

	if !ok {
		domErr, _ := nameEither.Left()

		return mo.Left[domain.Error, domain.Client](wrap(domErr, fmtClientWrap, raw.Name))
	}

	clientIDEither := domain.NewClientID(raw.ClientID)

	clientID, ok := clientIDEither.Right()

	if !ok {
		domErr, _ := clientIDEither.Left()

		return mo.Left[domain.Error, domain.Client](wrap(domErr, fmtClientWrap, raw.Name))
	}

	redirectURLsEither := buildRedirectURLs(raw.RedirectURLs)

	redirectURLs, ok := redirectURLsEither.Right()

	if !ok {
		domErr, _ := redirectURLsEither.Left()

		return mo.Left[domain.Error, domain.Client](wrap(domErr, fmtClientWrap, raw.Name))
	}

	assignmentsEither := buildGroupRoleAssignments(raw.GroupRoleAssignments)

	assignments, ok := assignmentsEither.Right()

	if !ok {
		domErr, _ := assignmentsEither.Left()

		return mo.Left[domain.Error, domain.Client](wrap(domErr, fmtClientWrap, raw.Name))
	}

	if raw.ClientSecret == "" {
		return mo.Right[domain.Error](domain.NewClientWithoutSecret(name, clientID, redirectURLs, assignments))
	}

	secretEither := domain.NewClientSecret(raw.ClientSecret)

	clientSecret, ok := secretEither.Right()

	if !ok {
		domErr, _ := secretEither.Left()

		return mo.Left[domain.Error, domain.Client](wrap(domErr, fmtClientWrap, raw.Name))
	}

	return mo.Right[domain.Error](domain.NewClientWithSecret(name, clientID, clientSecret, redirectURLs, assignments))
}

func buildGroupRoleAssignments(raws []RawGroupRoleAssignment) mo.Either[domain.Error, []domain.GroupRoleAssignment] {
	result := make([]domain.GroupRoleAssignment, emptyLen, len(raws))

	for _, raw := range raws {
		assignmentEither := buildGroupRoleAssignment(raw)

		assignment, ok := assignmentEither.Right()

		if !ok {
			domErr, _ := assignmentEither.Left()

			return mo.Left[domain.Error, []domain.GroupRoleAssignment](domErr)
		}

		result = append(result, assignment)
	}

	return mo.Right[domain.Error](result)
}

func buildGroupRoleAssignment(raw RawGroupRoleAssignment) mo.Either[domain.Error, domain.GroupRoleAssignment] {
	groupNameEither := domain.NewGroupName(raw.GroupName)

	groupName, ok := groupNameEither.Right()

	if !ok {
		domErr, _ := groupNameEither.Left()

		return mo.Left[domain.Error, domain.GroupRoleAssignment](
			domain.NewError(domErr.Code, fmt.Sprintf("group role assignment %q: %s", raw.GroupName, domErr.Message)),
		)
	}

	appIDEither := domain.NewClientID(raw.ApplicationID)

	appID, ok := appIDEither.Right()

	if !ok {
		domErr, _ := appIDEither.Left()

		return mo.Left[domain.Error, domain.GroupRoleAssignment](
			domain.NewError(domErr.Code, fmt.Sprintf("group role assignment %q: invalid applicationId: %s", raw.GroupName, domErr.Message)),
		)
	}

	roles := make([]domain.RoleValue, emptyLen, len(raw.Roles))

	for _, roleValue := range raw.Roles {
		roleEither := domain.NewRoleValue(roleValue)

		role, ok := roleEither.Right()

		if !ok {
			domErr, _ := roleEither.Left()

			return mo.Left[domain.Error, domain.GroupRoleAssignment](
				domain.NewError(domErr.Code, fmt.Sprintf("group role assignment %q: %s", raw.GroupName, domErr.Message)),
			)
		}

		roles = append(roles, role)
	}

	return mo.Right[domain.Error](domain.NewGroupRoleAssignment(groupName, roles, appID))
}
