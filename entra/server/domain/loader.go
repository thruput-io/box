package domain

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// rawConfig mirrors the YAML structure exactly for unmarshalling.
// snake_case keys match the config file format.
type rawConfig struct {
	Tenants []rawTenant `yaml:"tenants"`
}

type rawTenant struct {
	TenantID         string               `yaml:"tenantId"`
	Name             string               `yaml:"name"`
	AppRegistrations []rawAppRegistration `yaml:"appRegistrations"`
	Groups           []rawGroup           `yaml:"groups"`
	Users            []rawUser            `yaml:"users"`
	Clients          []rawClient          `yaml:"clients"`
}

type rawAppRegistration struct {
	Name          string     `yaml:"name"`
	ClientID      string     `yaml:"clientId"`
	IdentifierURI string     `yaml:"identifierUri"`
	RedirectURLs  []string   `yaml:"redirectUrls"`
	Scopes        []rawScope `yaml:"scopes"`
	AppRoles      []rawRole  `yaml:"appRoles"`
}

type rawScope struct {
	ID          string `yaml:"id"`
	Value       string `yaml:"value"`
	Description string `yaml:"description"`
}

type rawRole struct {
	ID          string     `yaml:"id"`
	Value       string     `yaml:"value"`
	Description string     `yaml:"description"`
	Scopes      []rawScope `yaml:"scopes"`
}

type rawGroup struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type rawUser struct {
	ID          string   `yaml:"id"`
	Username    string   `yaml:"username"`
	Password    string   `yaml:"password"`
	DisplayName string   `yaml:"displayName"`
	Email       string   `yaml:"email"`
	Groups      []string `yaml:"groups"`
}

type rawClient struct {
	Name                 string                   `yaml:"name"`
	ClientID             string                   `yaml:"clientId"`
	ClientSecret         string                   `yaml:"clientSecret"`
	RedirectURLs         []string                 `yaml:"redirectUrls"`
	GroupRoleAssignments []rawGroupRoleAssignment `yaml:"groupRoleAssignments"`
}

type rawGroupRoleAssignment struct {
	GroupName     string   `yaml:"groupName"`
	Roles         []string `yaml:"roles"`
	ApplicationID string   `yaml:"applicationId"`
}

// LoadConfig reads, validates, and parses Config.yaml into an immutable domain Config.
func LoadConfig(configPath, schemaPath string) (Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read %s: %w", configPath, err)
	}

	if err = validateYAML(data, schemaPath); err != nil {
		return Config{}, fmt.Errorf("config validation failed: %w", err)
	}

	var raw rawConfig

	if err = yaml.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("failed to parse %s: %w", configPath, err)
	}

	return buildConfig(raw)
}

func validateYAML(yamlData []byte, schemaPath string) error {
	var raw any

	if err := yaml.Unmarshal(yamlData, &raw); err != nil {
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
		messages := make([]string, 0, len(result.Errors()))
		for _, desc := range result.Errors() {
			messages = append(messages, desc.String())
		}

		return fmt.Errorf("%w: %s", ErrInvalidConfig, strings.Join(messages, "; "))
	}

	return nil
}

func buildConfig(raw rawConfig) (Config, error) {
	tenants := make([]Tenant, 0, len(raw.Tenants))

	for _, rawTenant := range raw.Tenants {
		tenant, err := buildTenant(rawTenant)
		if err != nil {
			return Config{}, err
		}

		tenants = append(tenants, tenant)
	}

	return NewConfig(tenants)
}

func buildTenant(raw rawTenant) (Tenant, error) {
	tenantID, err := NewTenantID(raw.TenantID)
	if err != nil {
		return Tenant{}, fmt.Errorf("tenant %q: %w", raw.Name, err)
	}

	name, err := NewTenantName(raw.Name)
	if err != nil {
		return Tenant{}, fmt.Errorf("tenant %q: %w", raw.TenantID, err)
	}

	appRegistrations, err := buildAppRegistrations(raw.AppRegistrations)
	if err != nil {
		return Tenant{}, fmt.Errorf("tenant %q: %w", raw.Name, err)
	}

	groups, err := buildGroups(raw.Groups)
	if err != nil {
		return Tenant{}, fmt.Errorf("tenant %q: %w", raw.Name, err)
	}

	users, err := buildUsers(raw.Users)
	if err != nil {
		return Tenant{}, fmt.Errorf("tenant %q: %w", raw.Name, err)
	}

	clients, err := buildClients(raw.Clients)
	if err != nil {
		return Tenant{}, fmt.Errorf("tenant %q: %w", raw.Name, err)
	}

	return NewTenant(tenantID, name, appRegistrations, groups, users, clients)
}

func buildAppRegistrations(raws []rawAppRegistration) ([]AppRegistration, error) {
	result := make([]AppRegistration, 0, len(raws))

	for _, raw := range raws {
		appRegistration, err := buildAppRegistration(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, appRegistration)
	}

	return result, nil
}

func buildAppRegistration(raw rawAppRegistration) (AppRegistration, error) {
	name, err := NewAppName(raw.Name)
	if err != nil {
		return AppRegistration{}, fmt.Errorf("app registration: %w", err)
	}

	clientID, err := NewClientID(raw.ClientID)
	if err != nil {
		return AppRegistration{}, fmt.Errorf("app registration %q: %w", raw.Name, err)
	}

	identifierURI, err := NewIdentifierURI(raw.IdentifierURI)
	if err != nil {
		return AppRegistration{}, fmt.Errorf("app registration %q: %w", raw.Name, err)
	}

	redirectURLs, err := buildRedirectURLs(raw.RedirectURLs)
	if err != nil {
		return AppRegistration{}, fmt.Errorf("app registration %q: %w", raw.Name, err)
	}

	scopes, err := buildScopes(raw.Scopes)
	if err != nil {
		return AppRegistration{}, fmt.Errorf("app registration %q: %w", raw.Name, err)
	}

	roles, err := buildRoles(raw.AppRoles)
	if err != nil {
		return AppRegistration{}, fmt.Errorf("app registration %q: %w", raw.Name, err)
	}

	return NewAppRegistration(name, clientID, identifierURI, redirectURLs, scopes, roles), nil
}

func buildRedirectURLs(raws []string) ([]RedirectURL, error) {
	result := make([]RedirectURL, 0, len(raws))

	for _, raw := range raws {
		redirectURL, err := NewRedirectURL(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, redirectURL)
	}

	return result, nil
}

func buildScopes(raws []rawScope) ([]Scope, error) {
	result := make([]Scope, 0, len(raws))

	for _, raw := range raws {
		scope, err := buildScope(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, scope)
	}

	return result, nil
}

func buildScope(raw rawScope) (Scope, error) {
	scopeID, err := NewScopeID(raw.ID)
	if err != nil {
		return Scope{}, fmt.Errorf("scope %q: %w", raw.Value, err)
	}

	value, err := NewScopeValue(raw.Value)
	if err != nil {
		return Scope{}, fmt.Errorf("scope: %w", err)
	}

	description, err := NewScopeDescription(raw.Description)
	if err != nil {
		return Scope{}, fmt.Errorf("scope %q: %w", raw.Value, err)
	}

	return NewScope(scopeID, value, description), nil
}

func buildRoles(raws []rawRole) ([]Role, error) {
	result := make([]Role, 0, len(raws))

	for _, raw := range raws {
		role, err := buildRole(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, role)
	}

	return result, nil
}

func buildRole(raw rawRole) (Role, error) {
	roleID, err := NewRoleID(raw.ID)
	if err != nil {
		return Role{}, fmt.Errorf("role %q: %w", raw.Value, err)
	}

	value, err := NewRoleValue(raw.Value)
	if err != nil {
		return Role{}, fmt.Errorf("role: %w", err)
	}

	description, err := NewRoleDescription(raw.Description)
	if err != nil {
		return Role{}, fmt.Errorf("role %q: %w", raw.Value, err)
	}

	scopes, err := buildScopes(raw.Scopes)
	if err != nil {
		return Role{}, fmt.Errorf("role %q: %w", raw.Value, err)
	}

	return NewRole(roleID, value, description, scopes), nil
}

func buildGroups(raws []rawGroup) ([]Group, error) {
	result := make([]Group, 0, len(raws))

	for _, raw := range raws {
		groupID, err := NewGroupID(raw.ID)
		if err != nil {
			return nil, fmt.Errorf("group %q: %w", raw.Name, err)
		}

		name, err := NewGroupName(raw.Name)
		if err != nil {
			return nil, fmt.Errorf("group: %w", err)
		}

		result = append(result, NewGroup(groupID, name))
	}

	return result, nil
}

func buildUsers(raws []rawUser) ([]User, error) {
	result := make([]User, 0, len(raws))

	for _, raw := range raws {
		user, err := buildUser(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, user)
	}

	return result, nil
}

func buildUser(raw rawUser) (User, error) {
	userID, err := NewUserID(raw.ID)
	if err != nil {
		return User{}, fmt.Errorf("user %q: %w", raw.Username, err)
	}

	username, err := NewUsername(raw.Username)
	if err != nil {
		return User{}, fmt.Errorf("user: %w", err)
	}

	password, err := NewPassword(raw.Password)
	if err != nil {
		return User{}, fmt.Errorf("user %q: %w", raw.Username, err)
	}

	displayName, err := NewDisplayName(raw.DisplayName)
	if err != nil {
		return User{}, fmt.Errorf("user %q: %w", raw.Username, err)
	}

	email, err := NewEmail(raw.Email)
	if err != nil {
		return User{}, fmt.Errorf("user %q: %w", raw.Username, err)
	}

	groups := make([]GroupName, 0, len(raw.Groups))

	for _, groupName := range raw.Groups {
		name, err := NewGroupName(groupName)
		if err != nil {
			return User{}, fmt.Errorf("user %q group: %w", raw.Username, err)
		}

		groups = append(groups, name)
	}

	return NewUser(userID, username, password, displayName, email, groups), nil
}

func buildClients(raws []rawClient) ([]Client, error) {
	result := make([]Client, 0, len(raws))

	for _, raw := range raws {
		client, err := buildClient(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, client)
	}

	return result, nil
}

func buildClient(raw rawClient) (Client, error) {
	name, err := NewAppName(raw.Name)
	if err != nil {
		return Client{}, fmt.Errorf("client: %w", err)
	}

	clientID, err := NewClientID(raw.ClientID)
	if err != nil {
		return Client{}, fmt.Errorf("client %q: %w", raw.Name, err)
	}

	redirectURLs, err := buildRedirectURLs(raw.RedirectURLs)
	if err != nil {
		return Client{}, fmt.Errorf("client %q: %w", raw.Name, err)
	}

	assignments, err := buildGroupRoleAssignments(raw.GroupRoleAssignments)
	if err != nil {
		return Client{}, fmt.Errorf("client %q: %w", raw.Name, err)
	}

	return NewClient(name, clientID, NewClientSecret(raw.ClientSecret), redirectURLs, assignments), nil
}

func buildGroupRoleAssignments(raws []rawGroupRoleAssignment) ([]GroupRoleAssignment, error) {
	result := make([]GroupRoleAssignment, 0, len(raws))

	for _, raw := range raws {
		assignment, err := buildGroupRoleAssignment(raw)
		if err != nil {
			return nil, err
		}

		result = append(result, assignment)
	}

	return result, nil
}

func buildGroupRoleAssignment(raw rawGroupRoleAssignment) (GroupRoleAssignment, error) {
	groupName, err := NewGroupName(raw.GroupName)
	if err != nil {
		return GroupRoleAssignment{}, fmt.Errorf("group role assignment: %w", err)
	}

	appID, err := uuid.Parse(raw.ApplicationID)
	if err != nil {
		return GroupRoleAssignment{}, fmt.Errorf("group role assignment %q: invalid applicationId: %w", raw.GroupName, err)
	}

	roles := make([]RoleValue, 0, len(raw.Roles))

	for _, roleValue := range raw.Roles {
		role, err := NewRoleValue(roleValue)
		if err != nil {
			return GroupRoleAssignment{}, fmt.Errorf("group role assignment %q: %w", raw.GroupName, err)
		}

		roles = append(roles, role)
	}

	return NewGroupRoleAssignment(groupName, roles, ClientIDFromUUID(appID)), nil
}
