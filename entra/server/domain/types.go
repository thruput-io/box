package domain

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// TenantID is the unique identifier for a tenant.
type TenantID struct {
	value uuid.UUID
}

// NewTenantID creates a TenantID from a UUID string, returning an error if invalid.
func NewTenantID(raw string) (TenantID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return TenantID{}, fmt.Errorf("invalid tenant ID %q: %w", raw, err)
	}

	return TenantID{value: parsed}, nil
}

// MustTenantID creates a TenantID, panicking if invalid. For use in tests and constants only.
func MustTenantID(raw string) TenantID {
	tenantID, err := NewTenantID(raw)
	if err != nil {
		panic(err)
	}

	return tenantID
}

// TenantIDFromUUID wraps an already-parsed uuid.UUID.
func TenantIDFromUUID(value uuid.UUID) TenantID {
	return TenantID{value: value}
}

// UUID returns the underlying uuid.UUID value.
func (tenantID TenantID) UUID() uuid.UUID { return tenantID.value }

// String returns the canonical UUID string representation.
func (tenantID TenantID) String() string { return tenantID.value.String() }

// ClientID is the unique identifier for an app registration or client.
type ClientID struct {
	value uuid.UUID
}

// NewClientID creates a ClientID from a UUID string, returning an error if invalid.
func NewClientID(raw string) (ClientID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return ClientID{}, fmt.Errorf("invalid client ID %q: %w", raw, err)
	}

	return ClientID{value: parsed}, nil
}

// MustClientID creates a ClientID, panicking if invalid. For use in tests and constants only.
func MustClientID(raw string) ClientID {
	clientID, err := NewClientID(raw)
	if err != nil {
		panic(err)
	}

	return clientID
}

// ClientIDFromUUID wraps an already-parsed uuid.UUID.
func ClientIDFromUUID(value uuid.UUID) ClientID {
	return ClientID{value: value}
}

// UUID returns the underlying uuid.UUID value.
func (clientID ClientID) UUID() uuid.UUID { return clientID.value }

// String returns the canonical UUID string representation.
func (clientID ClientID) String() string { return clientID.value.String() }

// UserID is the unique identifier for a user.
type UserID struct {
	value uuid.UUID
}

// NewUserID creates a UserID from a UUID string, returning an error if invalid.
func NewUserID(raw string) (UserID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return UserID{}, fmt.Errorf("invalid user ID %q: %w", raw, err)
	}

	return UserID{value: parsed}, nil
}

// MustUserID creates a UserID, panicking if invalid. For use in tests and constants only.
func MustUserID(raw string) UserID {
	userID, err := NewUserID(raw)
	if err != nil {
		panic(err)
	}

	return userID
}

// UserIDFromUUID wraps an already-parsed uuid.UUID.
func UserIDFromUUID(value uuid.UUID) UserID {
	return UserID{value: value}
}

// UUID returns the underlying uuid.UUID value.
func (userID UserID) UUID() uuid.UUID { return userID.value }

// String returns the canonical UUID string representation.
func (userID UserID) String() string { return userID.value.String() }

// GroupID is the unique identifier for a group.
type GroupID struct {
	value uuid.UUID
}

// NewGroupID creates a GroupID from a UUID string, returning an error if invalid.
func NewGroupID(raw string) (GroupID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return GroupID{}, fmt.Errorf("invalid group ID %q: %w", raw, err)
	}

	return GroupID{value: parsed}, nil
}

// MustGroupID creates a GroupID, panicking if invalid. For use in tests and constants only.
func MustGroupID(raw string) GroupID {
	groupID, err := NewGroupID(raw)
	if err != nil {
		panic(err)
	}

	return groupID
}

// GroupIDFromUUID wraps an already-parsed uuid.UUID.
func GroupIDFromUUID(value uuid.UUID) GroupID {
	return GroupID{value: value}
}

// UUID returns the underlying uuid.UUID value.
func (groupID GroupID) UUID() uuid.UUID { return groupID.value }

// String returns the canonical UUID string representation.
func (groupID GroupID) String() string { return groupID.value.String() }

// ScopeID is the unique identifier for a scope.
type ScopeID struct {
	value uuid.UUID
}

// NewScopeID creates a ScopeID from a UUID string, returning an error if invalid.
func NewScopeID(raw string) (ScopeID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return ScopeID{}, fmt.Errorf("invalid scope ID %q: %w", raw, err)
	}

	return ScopeID{value: parsed}, nil
}

// MustScopeID creates a ScopeID, panicking if invalid. For use in tests and constants only.
func MustScopeID(raw string) ScopeID {
	scopeID, err := NewScopeID(raw)
	if err != nil {
		panic(err)
	}

	return scopeID
}

// ScopeIDFromUUID wraps an already-parsed uuid.UUID.
func ScopeIDFromUUID(value uuid.UUID) ScopeID {
	return ScopeID{value: value}
}

// UUID returns the underlying uuid.UUID value.
func (scopeID ScopeID) UUID() uuid.UUID { return scopeID.value }

// String returns the canonical UUID string representation.
func (scopeID ScopeID) String() string { return scopeID.value.String() }

// RoleID is the unique identifier for a role.
type RoleID struct {
	value uuid.UUID
}

// NewRoleID creates a RoleID from a UUID string, returning an error if invalid.
func NewRoleID(raw string) (RoleID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return RoleID{}, fmt.Errorf("invalid role ID %q: %w", raw, err)
	}

	return RoleID{value: parsed}, nil
}

// MustRoleID creates a RoleID, panicking if invalid. For use in tests and constants only.
func MustRoleID(raw string) RoleID {
	roleID, err := NewRoleID(raw)
	if err != nil {
		panic(err)
	}

	return roleID
}

// RoleIDFromUUID wraps an already-parsed uuid.UUID.
func RoleIDFromUUID(value uuid.UUID) RoleID {
	return RoleID{value: value}
}

// UUID returns the underlying uuid.UUID value.
func (roleID RoleID) UUID() uuid.UUID { return roleID.value }

// String returns the canonical UUID string representation.
func (roleID RoleID) String() string { return roleID.value.String() }

// TenantName is the display name of a tenant.
type TenantName struct {
	value string
}

// NewTenantName creates a TenantName, returning an error if empty.
func NewTenantName(raw string) (TenantName, error) {
	if raw == "" {
		return TenantName{}, errors.New("tenant name must not be empty")
	}

	return TenantName{value: raw}, nil
}

// String returns the tenant name value.
func (tenantName TenantName) String() string { return tenantName.value }

// AppName is the display name of an app registration or client.
type AppName struct {
	value string
}

// NewAppName creates an AppName, returning an error if empty.
func NewAppName(raw string) (AppName, error) {
	if raw == "" {
		return AppName{}, errors.New("app name must not be empty")
	}

	return AppName{value: raw}, nil
}

// String returns the app name value.
func (appName AppName) String() string { return appName.value }

// IdentifierURI is the application ID URI of an app registration.
type IdentifierURI struct {
	value string
}

// NewIdentifierURI creates an IdentifierURI, returning an error if empty.
func NewIdentifierURI(raw string) (IdentifierURI, error) {
	if raw == "" {
		return IdentifierURI{}, errors.New("identifier URI must not be empty")
	}

	return IdentifierURI{value: raw}, nil
}

// String returns the identifier URI value.
func (identifierURI IdentifierURI) String() string { return identifierURI.value }

// ScopeValue is the string value of a scope (e.g. "read", "api://xxx/.default").
type ScopeValue struct {
	value string
}

// NewScopeValue creates a ScopeValue, returning an error if empty.
func NewScopeValue(raw string) (ScopeValue, error) {
	if raw == "" {
		return ScopeValue{}, errors.New("scope value must not be empty")
	}

	return ScopeValue{value: raw}, nil
}

// String returns the scope value.
func (scopeValue ScopeValue) String() string { return scopeValue.value }

// RoleValue is the string value of a role (e.g. "Admin", "Reader").
type RoleValue struct {
	value string
}

// NewRoleValue creates a RoleValue, returning an error if empty.
func NewRoleValue(raw string) (RoleValue, error) {
	if raw == "" {
		return RoleValue{}, errors.New("role value must not be empty")
	}

	return RoleValue{value: raw}, nil
}

// String returns the role value.
func (roleValue RoleValue) String() string { return roleValue.value }

// GroupName is the name of a group.
type GroupName struct {
	value string
}

// NewGroupName creates a GroupName, returning an error if empty.
func NewGroupName(raw string) (GroupName, error) {
	if raw == "" {
		return GroupName{}, errors.New("group name must not be empty")
	}

	return GroupName{value: raw}, nil
}

// String returns the group name value.
func (groupName GroupName) String() string { return groupName.value }

// Username is the login name of a user.
type Username struct {
	value string
}

// NewUsername creates a Username, returning an error if empty.
func NewUsername(raw string) (Username, error) {
	if raw == "" {
		return Username{}, errors.New("username must not be empty")
	}

	return Username{value: raw}, nil
}

// String returns the username value.
func (username Username) String() string { return username.value }

// Password is the credential of a user.
type Password struct {
	value string
}

// NewPassword creates a Password, returning an error if empty.
func NewPassword(raw string) (Password, error) {
	if raw == "" {
		return Password{}, errors.New("password must not be empty")
	}

	return Password{value: raw}, nil
}

// String returns the password value.
func (password Password) String() string { return password.value }

// DisplayName is the human-readable name of a user.
type DisplayName struct {
	value string
}

// NewDisplayName creates a DisplayName, returning an error if empty.
func NewDisplayName(raw string) (DisplayName, error) {
	if raw == "" {
		return DisplayName{}, errors.New("display name must not be empty")
	}

	return DisplayName{value: raw}, nil
}

// String returns the display name value.
func (displayName DisplayName) String() string { return displayName.value }

// Email is the email address of a user.
type Email struct {
	value string
}

// NewEmail creates an Email, returning an error if empty.
func NewEmail(raw string) (Email, error) {
	if raw == "" {
		return Email{}, errors.New("email must not be empty")
	}

	return Email{value: raw}, nil
}

// String returns the email value.
func (email Email) String() string { return email.value }

// RedirectURL is a permitted OAuth2 redirect URI.
type RedirectURL struct {
	value string
}

// NewRedirectURL creates a RedirectURL, returning an error if empty.
func NewRedirectURL(raw string) (RedirectURL, error) {
	if raw == "" {
		return RedirectURL{}, errors.New("redirect URL must not be empty")
	}

	return RedirectURL{value: raw}, nil
}

// String returns the redirect URL value.
func (redirectURL RedirectURL) String() string { return redirectURL.value }

// ClientSecret is the secret credential of a confidential client.
// Optional — public clients have no secret.
type ClientSecret struct {
	value string
}

// NewClientSecret creates a ClientSecret from a raw string.
// An empty string is valid and represents a public client.
func NewClientSecret(raw string) ClientSecret {
	return ClientSecret{value: raw}
}

// String returns the client secret value.
func (clientSecret ClientSecret) String() string { return clientSecret.value }

// IsEmpty reports whether this is a public client (no secret).
func (clientSecret ClientSecret) IsEmpty() bool { return clientSecret.value == "" }

// ScopeDescription is the human-readable description of a scope.
type ScopeDescription struct {
	value string
}

// NewScopeDescription creates a ScopeDescription, returning an error if empty.
func NewScopeDescription(raw string) (ScopeDescription, error) {
	if raw == "" {
		return ScopeDescription{}, errors.New("scope description must not be empty")
	}

	return ScopeDescription{value: raw}, nil
}

// String returns the scope description value.
func (scopeDescription ScopeDescription) String() string { return scopeDescription.value }

// RoleDescription is the human-readable description of a role.
type RoleDescription struct {
	value string
}

// NewRoleDescription creates a RoleDescription, returning an error if empty.
func NewRoleDescription(raw string) (RoleDescription, error) {
	if raw == "" {
		return RoleDescription{}, errors.New("role description must not be empty")
	}

	return RoleDescription{value: raw}, nil
}

// String returns the role description value.
func (roleDescription RoleDescription) String() string { return roleDescription.value }
