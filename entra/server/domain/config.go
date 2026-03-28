package domain

import "fmt"

// Config is the immutable root domain object loaded once at startup.
// It must contain at least one tenant to be valid.
type Config struct {
	tenants []Tenant
}

// NewConfig constructs a Config, enforcing that at least one tenant is present.
func NewConfig(tenants []Tenant) (Config, error) {
	if len(tenants) == 0 {
		return Config{}, errConfigNoTenants
	}

	return Config{tenants: tenants}, nil
}

// Tenants returns the list of tenants (always non-empty).
func (config Config) Tenants() []Tenant { return config.tenants }

// Tenant is an immutable Azure AD tenant with at least one app registration and one user.
type Tenant struct {
	tenantID         TenantID
	name             TenantName
	appRegistrations []AppRegistration
	groups           []Group
	users            []User
	clients          []Client
}

// NewTenant constructs a Tenant, enforcing required fields and minimum collections.
func NewTenant(
	tenantID TenantID,
	name TenantName,
	appRegistrations []AppRegistration,
	groups []Group,
	users []User,
	clients []Client,
) (Tenant, error) {
	if len(appRegistrations) == 0 {
		return Tenant{}, fmt.Errorf("tenant %q must have at least one app registration", name)
	}

	if len(users) == 0 {
		return Tenant{}, fmt.Errorf("tenant %q must have at least one user", name)
	}

	return Tenant{
		tenantID:         tenantID,
		name:             name,
		appRegistrations: appRegistrations,
		groups:           groups,
		users:            users,
		clients:          clients,
	}, nil
}

// TenantID returns the tenant's unique identifier.
func (tenant Tenant) TenantID() TenantID { return tenant.tenantID }

// Name returns the tenant's display name.
func (tenant Tenant) Name() TenantName { return tenant.name }

// AppRegistrations returns the tenant's app registrations (always non-empty).
func (tenant Tenant) AppRegistrations() []AppRegistration { return tenant.appRegistrations }

// Groups returns the tenant's groups (may be empty).
func (tenant Tenant) Groups() []Group { return tenant.groups }

// Users returns the tenant's users (always non-empty).
func (tenant Tenant) Users() []User { return tenant.users }

// Clients returns the tenant's clients (may be empty).
func (tenant Tenant) Clients() []Client { return tenant.clients }

// AppRegistration is an immutable Azure AD application registration.
type AppRegistration struct {
	name          AppName
	clientID      ClientID
	identifierURI IdentifierURI
	redirectURLs  []RedirectURL
	scopes        []Scope
	appRoles      []Role
}

// NewAppRegistration constructs an AppRegistration, enforcing required fields.
func NewAppRegistration(
	name AppName,
	clientID ClientID,
	identifierURI IdentifierURI,
	redirectURLs []RedirectURL,
	scopes []Scope,
	appRoles []Role,
) AppRegistration {
	return AppRegistration{
		name:          name,
		clientID:      clientID,
		identifierURI: identifierURI,
		redirectURLs:  redirectURLs,
		scopes:        scopes,
		appRoles:      appRoles,
	}
}

// Name returns the app registration's display name.
func (appRegistration AppRegistration) Name() AppName { return appRegistration.name }

// ClientID returns the app registration's client ID.
func (appRegistration AppRegistration) ClientID() ClientID { return appRegistration.clientID }

// IdentifierURI returns the app registration's identifier URI.
func (appRegistration AppRegistration) IdentifierURI() IdentifierURI {
	return appRegistration.identifierURI
}

// RedirectURLs returns the app registration's permitted redirect URLs.
func (appRegistration AppRegistration) RedirectURLs() []RedirectURL {
	return appRegistration.redirectURLs
}

// Scopes returns the app registration's scopes.
func (appRegistration AppRegistration) Scopes() []Scope { return appRegistration.scopes }

// AppRoles returns the app registration's roles.
func (appRegistration AppRegistration) AppRoles() []Role { return appRegistration.appRoles }

// Scope is an immutable OAuth2 scope definition.
type Scope struct {
	id          ScopeID
	value       ScopeValue
	description ScopeDescription
}

// NewScope constructs a Scope with all required fields.
func NewScope(id ScopeID, value ScopeValue, description ScopeDescription) Scope {
	return Scope{id: id, value: value, description: description}
}

// ID returns the scope's unique identifier.
func (scope Scope) ID() ScopeID { return scope.id }

// Value returns the scope's string value.
func (scope Scope) Value() ScopeValue { return scope.value }

// Description returns the scope's human-readable description.
func (scope Scope) Description() ScopeDescription { return scope.description }

// Role is an immutable app role definition.
type Role struct {
	id          RoleID
	value       RoleValue
	description RoleDescription
	scopes      []Scope
}

// NewRole constructs a Role with all required fields.
func NewRole(id RoleID, value RoleValue, description RoleDescription, scopes []Scope) Role {
	return Role{id: id, value: value, description: description, scopes: scopes}
}

// ID returns the role's unique identifier.
func (role Role) ID() RoleID { return role.id }

// Value returns the role's string value.
func (role Role) Value() RoleValue { return role.value }

// Description returns the role's human-readable description.
func (role Role) Description() RoleDescription { return role.description }

// Scopes returns the scopes associated with this role.
func (role Role) Scopes() []Scope { return role.scopes }

// Group is an immutable user group.
type Group struct {
	id   GroupID
	name GroupName
}

// NewGroup constructs a Group with all required fields.
func NewGroup(id GroupID, name GroupName) Group {
	return Group{id: id, name: name}
}

// ID returns the group's unique identifier.
func (group Group) ID() GroupID { return group.id }

// Name returns the group's name.
func (group Group) Name() GroupName { return group.name }

// User is an immutable user account.
// Groups is a list of group names the user belongs to (may be empty).
type User struct {
	id          UserID
	username    Username
	password    Password
	displayName DisplayName
	email       Email
	groups      []GroupName
}

// NewUser constructs a User, enforcing all required fields.
func NewUser(
	id UserID,
	username Username,
	password Password,
	displayName DisplayName,
	email Email,
	groups []GroupName,
) User {
	return User{
		id:          id,
		username:    username,
		password:    password,
		displayName: displayName,
		email:       email,
		groups:      groups,
	}
}

// ID returns the user's unique identifier.
func (user User) ID() UserID { return user.id }

// Username returns the user's login name.
func (user User) Username() Username { return user.username }

// Password returns the user's credential.
func (user User) Password() Password { return user.password }

// DisplayName returns the user's human-readable name.
func (user User) DisplayName() DisplayName { return user.displayName }

// Email returns the user's email address.
func (user User) Email() Email { return user.email }

// Groups returns the names of groups the user belongs to.
func (user User) Groups() []GroupName { return user.groups }

// Client is an immutable OAuth2 confidential or public client.
type Client struct {
	name                 AppName
	clientID             ClientID
	clientSecret         ClientSecret
	redirectURLs         []RedirectURL
	groupRoleAssignments []GroupRoleAssignment
}

// NewClient constructs a Client with all required fields.
func NewClient(
	name AppName,
	clientID ClientID,
	clientSecret ClientSecret,
	redirectURLs []RedirectURL,
	groupRoleAssignments []GroupRoleAssignment,
) Client {
	return Client{
		name:                 name,
		clientID:             clientID,
		clientSecret:         clientSecret,
		redirectURLs:         redirectURLs,
		groupRoleAssignments: groupRoleAssignments,
	}
}

// Name returns the client's display name.
func (client Client) Name() AppName { return client.name }

// ClientID returns the client's unique identifier.
func (client Client) ClientID() ClientID { return client.clientID }

// ClientSecret returns the client's secret (empty for public clients).
func (client Client) ClientSecret() ClientSecret { return client.clientSecret }

// RedirectURLs returns the client's permitted redirect URLs.
func (client Client) RedirectURLs() []RedirectURL { return client.redirectURLs }

// GroupRoleAssignments returns the client's group-to-role mappings.
func (client Client) GroupRoleAssignments() []GroupRoleAssignment {
	return client.groupRoleAssignments
}

// GroupRoleAssignment maps a group to a set of roles for a specific application.
type GroupRoleAssignment struct {
	groupName     GroupName
	roles         []RoleValue
	applicationID ClientID
}

// NewGroupRoleAssignment constructs a GroupRoleAssignment with all required fields.
func NewGroupRoleAssignment(groupName GroupName, roles []RoleValue, applicationID ClientID) GroupRoleAssignment {
	return GroupRoleAssignment{
		groupName:     groupName,
		roles:         roles,
		applicationID: applicationID,
	}
}

// GroupName returns the name of the group.
func (groupRoleAssignment GroupRoleAssignment) GroupName() GroupName {
	return groupRoleAssignment.groupName
}

// Roles returns the role values assigned to the group.
func (groupRoleAssignment GroupRoleAssignment) Roles() []RoleValue { return groupRoleAssignment.roles }

// ApplicationID returns the application this assignment applies to.
func (groupRoleAssignment GroupRoleAssignment) ApplicationID() ClientID {
	return groupRoleAssignment.applicationID
}
