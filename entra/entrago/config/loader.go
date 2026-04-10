// Package config handles the loading and parsing of the mock server configuration.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"

	"identity/domain"
)

const (
	emptyLen      = 0
	fmtTenantWrap = "tenant %q: %w"
	fmtAppRegWrap = "app registration %q: %w"
	fmtRoleWrap   = "role %q: %w"
	fmtUserWrap   = "user %q: %w"
	fmtClientWrap = "client %q: %w"
)

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
func LoadConfig(configPath, schemaPath string) (*domain.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", configPath, err)
	}

	err = validateYAML(data, schemaPath)
	if err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	var raw RawConfig

	err = yaml.Unmarshal(data, &raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", configPath, err)
	}

	return buildConfig(raw)
}

func validateYAML(yamlData []byte, schemaPath string) error {
	if schemaPath == "" {
		return nil
	}

	var raw any

	err := yaml.Unmarshal(yamlData, &raw)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schemaData),
		gojsonschema.NewGoLoader(raw),
	)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		messages := make([]string, emptyLen, len(result.Errors()))
		for _, desc := range result.Errors() {
			messages = append(messages, desc.String())
		}

		return fmt.Errorf("%w: %s", domain.ErrInvalidConfig, strings.Join(messages, "; "))
	}

	return nil
}

func buildConfig(raw RawConfig) (*domain.Config, error) {
	tenants := make([]domain.Tenant, emptyLen, len(raw.Tenants))

	for _, rawTenant := range raw.Tenants {
		tenant, err := buildTenant(rawTenant)
		if err != nil {
			return nil, err
		}

		tenants = append(tenants, *tenant)
	}

	cfg, err := domain.NewConfig(tenants)
	if err != nil {
		return nil, fmt.Errorf("NewConfig: %w", err)
	}

	return &cfg, nil
}

func buildTenant(raw RawTenant) (*domain.Tenant, error) {
	tenantID, err := domain.NewTenantID(raw.TenantID)
	if err != nil {
		return nil, fmt.Errorf(fmtTenantWrap, raw.Name, err)
	}

	name, err := domain.NewTenantName(raw.Name)
	if err != nil {
		return nil, fmt.Errorf(fmtTenantWrap, raw.Name, err)
	}

	appRegistrations, err := buildAppRegistrations(raw.AppRegistrations)
	if err != nil {
		return nil, fmt.Errorf(fmtTenantWrap, raw.Name, err)
	}

	groups, err := buildGroups(raw.Groups)
	if err != nil {
		return nil, fmt.Errorf(fmtTenantWrap, raw.Name, err)
	}

	users, err := buildUsers(raw.Users)
	if err != nil {
		return nil, fmt.Errorf(fmtTenantWrap, raw.Name, err)
	}

	clients, err := buildClients(raw.Clients)
	if err != nil {
		return nil, fmt.Errorf(fmtTenantWrap, raw.Name, err)
	}

	ten, err := domain.NewTenant(tenantID, name, appRegistrations, groups, users, clients)
	if err != nil {
		return nil, fmt.Errorf(fmtTenantWrap, raw.Name, err)
	}

	return &ten, nil
}

func buildAppRegistrations(raws []RawAppRegistration) ([]domain.AppRegistration, error) {
	result := make([]domain.AppRegistration, emptyLen, len(raws))

	for _, raw := range raws {
		appRegistration, err := buildAppRegistration(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, *appRegistration)
	}

	return result, nil
}

func buildAppRegistration(raw RawAppRegistration) (*domain.AppRegistration, error) {
	name, err := domain.NewAppName(raw.Name)
	if err != nil {
		return nil, fmt.Errorf("app registration: %w", err)
	}

	clientID, err := domain.NewClientID(raw.ClientID)
	if err != nil {
		return nil, fmt.Errorf(fmtAppRegWrap, raw.Name, err)
	}

	identifierURI, err := domain.NewIdentifierURI(raw.IdentifierURI)
	if err != nil {
		return nil, fmt.Errorf(fmtAppRegWrap, raw.Name, err)
	}

	redirectURLs, err := buildRedirectURLs(raw.RedirectURLs)
	if err != nil {
		return nil, fmt.Errorf(fmtAppRegWrap, raw.Name, err)
	}

	scopes, err := buildScopes(raw.Scopes)
	if err != nil {
		return nil, fmt.Errorf(fmtAppRegWrap, raw.Name, err)
	}

	roles, err := buildRoles(raw.AppRoles)
	if err != nil {
		return nil, fmt.Errorf(fmtAppRegWrap, raw.Name, err)
	}

	reg := domain.NewAppRegistration(name, clientID, identifierURI, redirectURLs, scopes, roles)

	return &reg, nil
}

func buildRedirectURLs(raws []string) ([]domain.RedirectURL, error) {
	result := make([]domain.RedirectURL, emptyLen, len(raws))

	for _, raw := range raws {
		redirectURL, err := domain.NewRedirectURL(raw)
		if err != nil {
			return nil, fmt.Errorf("NewRedirectURL: %w", err)
		}

		result = append(result, redirectURL)
	}

	return result, nil
}

func buildScopes(raws []RawScope) ([]domain.Scope, error) {
	result := make([]domain.Scope, emptyLen, len(raws))

	for _, raw := range raws {
		scope, err := buildScope(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, scope)
	}

	return result, nil
}

func buildScope(raw RawScope) (domain.Scope, error) {
	scopeID, err := domain.NewScopeID(raw.ID)
	if err != nil {
		return domain.Scope{}, fmt.Errorf("scope %q: %w", raw.Value, err)
	}

	value, err := domain.NewScopeValue(raw.Value)
	if err != nil {
		return domain.Scope{}, fmt.Errorf("scope: %w", err)
	}

	description, err := domain.NewScopeDescription(raw.Description)
	if err != nil {
		return domain.Scope{}, fmt.Errorf("scope %q: %w", raw.Value, err)
	}

	return domain.NewScope(scopeID, value, description), nil
}

func buildRoles(raws []RawRole) ([]domain.Role, error) {
	result := make([]domain.Role, emptyLen, len(raws))

	for _, raw := range raws {
		role, err := buildRole(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, role)
	}

	return result, nil
}

func buildRole(raw RawRole) (domain.Role, error) {
	roleID, err := domain.NewRoleID(raw.ID)
	if err != nil {
		return domain.Role{}, fmt.Errorf(fmtRoleWrap, raw.Value, err)
	}

	value, err := domain.NewRoleValue(raw.Value)
	if err != nil {
		return domain.Role{}, fmt.Errorf("role: %w", err)
	}

	description, err := domain.NewRoleDescription(raw.Description)
	if err != nil {
		return domain.Role{}, fmt.Errorf(fmtRoleWrap, raw.Value, err)
	}

	scopes, err := buildScopes(raw.Scopes)
	if err != nil {
		return domain.Role{}, fmt.Errorf(fmtRoleWrap, raw.Value, err)
	}

	return domain.NewRole(roleID, value, description, scopes), nil
}

func buildGroups(raws []RawGroup) ([]domain.Group, error) {
	result := make([]domain.Group, emptyLen, len(raws))

	for _, raw := range raws {
		groupID, err := domain.NewGroupID(raw.ID)
		if err != nil {
			return nil, fmt.Errorf("group %q: %w", raw.Name, err)
		}

		name, err := domain.NewGroupName(raw.Name)
		if err != nil {
			return nil, fmt.Errorf("group: %w", err)
		}

		result = append(result, domain.NewGroup(groupID, name))
	}

	return result, nil
}

func buildUsers(raws []RawUser) ([]domain.User, error) {
	result := make([]domain.User, emptyLen, len(raws))

	for _, raw := range raws {
		user, err := buildUser(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, *user)
	}

	return result, nil
}

func buildUser(raw RawUser) (*domain.User, error) {
	userID, err := domain.NewUserID(raw.ID)
	if err != nil {
		return nil, fmt.Errorf(fmtUserWrap, raw.Username, err)
	}

	username, err := domain.NewUsername(raw.Username)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	password, err := domain.NewPassword(raw.Password)
	if err != nil {
		return nil, fmt.Errorf(fmtUserWrap, raw.Username, err)
	}

	displayName, err := domain.NewDisplayName(raw.DisplayName)
	if err != nil {
		return nil, fmt.Errorf(fmtUserWrap, raw.Username, err)
	}

	email, err := domain.NewEmail(raw.Email)
	if err != nil {
		return nil, fmt.Errorf(fmtUserWrap, raw.Username, err)
	}

	groups, err := buildUserGroups(raw.Username, raw.Groups)
	if err != nil {
		return nil, err
	}

	u := domain.NewUser(userID, username, password, displayName, email, groups)

	return &u, nil
}

func buildUserGroups(username string, rawGroups []string) ([]domain.GroupName, error) {
	groups := make([]domain.GroupName, emptyLen, len(rawGroups))

	for _, groupName := range rawGroups {
		name, err := domain.NewGroupName(groupName)
		if err != nil {
			return nil, fmt.Errorf("user %q group: %w", username, err)
		}

		groups = append(groups, name)
	}

	return groups, nil
}

func buildClients(raws []RawClient) ([]domain.Client, error) {
	result := make([]domain.Client, emptyLen, len(raws))

	for _, raw := range raws {
		client, err := buildClient(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, *client)
	}

	return result, nil
}

func buildClient(raw RawClient) (*domain.Client, error) {
	name, err := domain.NewAppName(raw.Name)
	if err != nil {
		return nil, fmt.Errorf("client: %w", err)
	}

	clientID, err := domain.NewClientID(raw.ClientID)
	if err != nil {
		return nil, fmt.Errorf(fmtClientWrap, raw.Name, err)
	}

	redirectURLs, err := buildRedirectURLs(raw.RedirectURLs)
	if err != nil {
		return nil, fmt.Errorf(fmtClientWrap, raw.Name, err)
	}

	assignments, err := buildGroupRoleAssignments(raw.GroupRoleAssignments)
	if err != nil {
		return nil, fmt.Errorf(fmtClientWrap, raw.Name, err)
	}

	if raw.ClientSecret == "" {
		c := domain.NewClientWithoutSecret(name, clientID, redirectURLs, assignments)

		return &c, nil
	}

	clientSecret, err := domain.NewClientSecret(raw.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf(fmtClientWrap, raw.Name, err)
	}

	c := domain.NewClientWithSecret(name, clientID, clientSecret, redirectURLs, assignments)

	return &c, nil
}

func buildGroupRoleAssignments(raws []RawGroupRoleAssignment) ([]domain.GroupRoleAssignment, error) {
	result := make([]domain.GroupRoleAssignment, emptyLen, len(raws))

	for _, raw := range raws {
		assignment, err := buildGroupRoleAssignment(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, assignment)
	}

	return result, nil
}

func buildGroupRoleAssignment(raw RawGroupRoleAssignment) (domain.GroupRoleAssignment, error) {
	groupName, err := domain.NewGroupName(raw.GroupName)
	if err != nil {
		return domain.GroupRoleAssignment{}, fmt.Errorf("group role assignment: %w", err)
	}

	appID, err := uuid.Parse(raw.ApplicationID)
	if err != nil {
		return domain.GroupRoleAssignment{}, fmt.Errorf(
			"group role assignment %q: invalid applicationId: %w", raw.GroupName, err,
		)
	}

	roles := make([]domain.RoleValue, emptyLen, len(raw.Roles))

	for _, roleValue := range raw.Roles {
		role, err := domain.NewRoleValue(roleValue)
		if err != nil {
			return domain.GroupRoleAssignment{}, fmt.Errorf("group role assignment %q: %w", raw.GroupName, err)
		}

		roles = append(roles, role)
	}

	return domain.NewGroupRoleAssignment(groupName, roles, domain.ClientIDFromUUID(appID)), nil
}
