package domain

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type RawValueProvider interface {
	rawCallback(func(string) error) error
}

func Parse[T any](p RawValueProvider, parserFunc func(string) (T, error)) (T, error) {
	var result T
	var err error

	cbErr := p.rawCallback(func(raw string) error {
		result, err = parserFunc(raw)
		return err
	})

	if cbErr != nil {
		var zero T
		return zero, cbErr
	}

	return result, nil
}

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

func (tenantID TenantID) AsClientID() ClientID { return ClientID{tenantID.value} }

func (tenantID TenantID) MarshalJSON() ([]byte, error) {
	return json.Marshal(tenantID.value.String())
}

func (tenantID TenantID) rawCallback(callback func(string) error) error {
	return callback(tenantID.value.String())
}

// ClientID is the unique identifier for an app registration or client.
type ClientID struct {
	value uuid.UUID
}

func AppIDAsClientID(id AppID) ClientID {
	return ClientID(id)
}

type AppID struct {
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

func (clientID ClientID) MarshalJSON() ([]byte, error) {
	return json.Marshal(clientID.value.String())
}

func (clientID ClientID) rawCallback(callback func(string) error) error {
	return callback(clientID.value.String())
}

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

func (userID UserID) MarshalJSON() ([]byte, error) {
	return json.Marshal(userID.value.String())
}

func (userID UserID) rawCallback(callback func(string) error) error {
	return callback(userID.value.String())
}

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

func (groupID GroupID) MarshalJSON() ([]byte, error) {
	return json.Marshal(groupID.value.String())
}

func (groupID GroupID) rawCallback(callback func(string) error) error {
	return callback(groupID.value.String())
}

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

func (scopeID ScopeID) MarshalJSON() ([]byte, error) {
	return json.Marshal(scopeID.value.String())
}

func (scopeID ScopeID) rawCallback(callback func(string) error) error {
	return callback(scopeID.value.String())
}

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

func (roleID RoleID) MarshalJSON() ([]byte, error) {
	return json.Marshal(roleID.value.String())
}

func (roleID RoleID) rawCallback(callback func(string) error) error {
	return callback(roleID.value.String())
}

// TenantName is the display name of a tenant.
type TenantName struct {
	value NonEmptyString
}

func (tenantName TenantName) AsAppName() AppName { return AppName{value: tenantName.value} }

func (tenantName TenantName) MarshalJSON() ([]byte, error) {
	return tenantName.value.MarshalJSON()
}

func (tenantName TenantName) rawCallback(callback func(string) error) error {
	return tenantName.value.rawCallback(callback)
}

// NewTenantName creates a TenantName, returning an error if empty.
func NewTenantName(raw string) (TenantName, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return TenantName{}, errTenantNameEmpty
	}

	return TenantName{value: v}, nil
}

// MustTenantName creates a TenantName, panicking if invalid. For use in tests and constants only.
func MustTenantName(raw string) TenantName {
	name, err := NewTenantName(raw)
	if err != nil {
		panic(err)
	}

	return name
}

// AppName is the display name of an app registration or client.
type AppName struct {
	value NonEmptyString
}

// NewAppName creates an AppName, returning an error if empty.
func NewAppName(raw string) (AppName, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return AppName{}, errAppNameEmpty
	}

	return AppName{value: v}, nil
}

// MustAppName creates an AppName, panicking if invalid. For use in tests and constants only.
func (appName AppName) MarshalJSON() ([]byte, error) {
	return appName.value.MarshalJSON()
}

func (appName AppName) rawCallback(callback func(string) error) error {
	return appName.value.rawCallback(callback)
}

func MustAppName(raw string) AppName {
	name, err := NewAppName(raw)
	if err != nil {
		panic(err)
	}

	return name
}

// IdentifierURI is the application ID URI of an app registration.
type IdentifierURI struct {
	value NonEmptyString
}

// NewIdentifierURI creates an IdentifierURI, returning an error if empty.
func NewIdentifierURI(raw string) (IdentifierURI, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return IdentifierURI{}, errIdentifierURIEmpty
	}

	return IdentifierURI{value: v}, nil
}

func (identifierURI IdentifierURI) MarshalJSON() ([]byte, error) {
	return identifierURI.value.MarshalJSON()
}

func (identifierURI IdentifierURI) rawCallback(callback func(string) error) error {
	return identifierURI.value.rawCallback(callback)
}

func (identifierURI IdentifierURI) Matches(raw string) bool {
	return identifierURI.value.value == raw
}

func (identifierURI IdentifierURI) MatchesPrefix(raw string) bool {
	return strings.HasPrefix(raw, identifierURI.value.value)
}

// MustIdentifierURI creates an IdentifierURI, panicking if invalid. For use in tests and constants only.
func MustIdentifierURI(raw string) IdentifierURI {
	identifierURI, err := NewIdentifierURI(raw)
	if err != nil {
		panic(err)
	}

	return identifierURI
}

// ScopeValue is the string value of a scope (e.g. "read", "api://xxx/.default").
type ScopeValue struct {
	value NonEmptyString
}

// NewScopeValue creates a ScopeValue, returning an error if empty.
func NewScopeValue(raw string) (ScopeValue, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return ScopeValue{}, errScopeValueEmpty
	}

	return ScopeValue{value: v}, nil
}

func (scopeValue ScopeValue) MarshalJSON() ([]byte, error) {
	return scopeValue.value.MarshalJSON()
}

func (scopeValue ScopeValue) rawCallback(callback func(string) error) error {
	return scopeValue.value.rawCallback(callback)
}

func (scopeValue ScopeValue) Matches(raw string) bool {
	v := scopeValue.value.value
	return v == raw || strings.HasSuffix(raw, "/"+v)
}

// MustScopeValue creates a ScopeValue, panicking if invalid. For use in tests and constants only.
func MustScopeValue(raw string) ScopeValue {
	v, err := NewScopeValue(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// RoleValue is the string value of a role (e.g. "Admin", "Reader").
type RoleValue struct {
	value NonEmptyString
}

// NewRoleValue creates a RoleValue, returning an error if empty.
func NewRoleValue(raw string) (RoleValue, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return RoleValue{}, errRoleValueEmpty
	}

	return RoleValue{value: v}, nil
}

func (roleValue RoleValue) MarshalJSON() ([]byte, error) {
	return roleValue.value.MarshalJSON()
}

func (roleValue RoleValue) rawCallback(callback func(string) error) error {
	return roleValue.value.rawCallback(callback)
}

func (roleValue RoleValue) Matches(raw string) bool {
	return roleValue.value.value == raw
}

// MustRoleValue creates a RoleValue, panicking if invalid. For use in tests and constants only.
func MustRoleValue(raw string) RoleValue {
	v, err := NewRoleValue(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// GroupName is the name of a group.
type GroupName struct {
	value NonEmptyString
}

// NewGroupName creates a GroupName, returning an error if empty.
func NewGroupName(raw string) (GroupName, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return GroupName{}, errGroupNameEmpty
	}

	return GroupName{value: v}, nil
}

func (groupName GroupName) MarshalJSON() ([]byte, error) {
	return groupName.value.MarshalJSON()
}

func (groupName GroupName) rawCallback(callback func(string) error) error {
	return groupName.value.rawCallback(callback)
}

func (groupName GroupName) Matches(raw string) bool {
	return groupName.value.value == raw
}

// MustGroupName creates a GroupName, panicking if invalid. For use in tests and constants only.
func MustGroupName(raw string) GroupName {
	v, err := NewGroupName(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Username is the login name of a user.
type Username struct {
	value NonEmptyString
}

// NewUsername creates a Username, returning an error if empty.
func NewUsername(raw string) (Username, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return Username{}, errUsernameEmpty
	}

	return Username{value: v}, nil
}

func (username Username) MarshalJSON() ([]byte, error) {
	return username.value.MarshalJSON()
}

func (username Username) rawCallback(callback func(string) error) error {
	return username.value.rawCallback(callback)
}

// MustUsername creates a Username, panicking if invalid. For use in tests and constants only.
func MustUsername(raw string) Username {
	username, err := NewUsername(raw)
	if err != nil {
		panic(err)
	}

	return username
}

// Password is the credential of a user.
type Password struct {
	value NonEmptyString
}

// NewPassword creates a Password, returning an error if empty.
func NewPassword(raw string) (Password, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return Password{}, errPasswordEmpty
	}

	return Password{value: v}, nil
}

func (password Password) MarshalJSON() ([]byte, error) {
	return password.value.MarshalJSON()
}

func (password Password) rawCallback(callback func(string) error) error {
	return password.value.rawCallback(callback)
}

// MustPassword creates a Password, panicking if invalid. For use in tests and constants only.
func MustPassword(raw string) Password {
	password, err := NewPassword(raw)
	if err != nil {
		panic(err)
	}

	return password
}

// DisplayName is the human-readable name of a user.
type DisplayName struct {
	value NonEmptyString
}

// NewDisplayName creates a DisplayName, returning an error if empty.
func NewDisplayName(raw string) (DisplayName, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return DisplayName{}, errDisplayNameEmpty
	}

	return DisplayName{value: v}, nil
}

func (displayName DisplayName) MarshalJSON() ([]byte, error) {
	return displayName.value.MarshalJSON()
}

func (displayName DisplayName) rawCallback(callback func(string) error) error {
	return displayName.value.rawCallback(callback)
}

// MustDisplayName creates a DisplayName, panicking if invalid. For use in tests and constants only.
func MustDisplayName(raw string) DisplayName {
	displayName, err := NewDisplayName(raw)
	if err != nil {
		panic(err)
	}

	return displayName
}

// Email is the email address of a user.
type Email struct {
	value NonEmptyString
}

// NewEmail creates an Email, returning an error if empty.
func NewEmail(raw string) (Email, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return Email{}, errEmailEmpty
	}

	return Email{value: v}, nil
}

func (email Email) MarshalJSON() ([]byte, error) {
	return email.value.MarshalJSON()
}

func (email Email) rawCallback(callback func(string) error) error {
	return email.value.rawCallback(callback)
}

// MustEmail creates an Email, panicking if invalid. For use in tests and constants only.
func MustEmail(raw string) Email {
	email, err := NewEmail(raw)
	if err != nil {
		panic(err)
	}

	return email
}

// RedirectURL is a permitted OAuth2 redirect URI.
type RedirectURL struct {
	value NonEmptyString
}

// NewRedirectURL creates a RedirectURL, returning an error if empty.
func NewRedirectURL(raw string) (RedirectURL, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return RedirectURL{}, errRedirectURLEmpty
	}

	return RedirectURL{value: v}, nil
}

func (redirectURL RedirectURL) MarshalJSON() ([]byte, error) {
	return redirectURL.value.MarshalJSON()
}

func (redirectURL RedirectURL) rawCallback(callback func(string) error) error {
	return redirectURL.value.rawCallback(callback)
}

// MustRedirectURL creates a RedirectURL, panicking if invalid. For use in tests and constants only.
func MustRedirectURL(raw string) RedirectURL {
	redirectURL, err := NewRedirectURL(raw)
	if err != nil {
		panic(err)
	}

	return redirectURL
}

// ClientSecret is the secret credential of a confidential client.
type ClientSecret struct {
	value NonEmptyString
}

// NewClientSecret creates a ClientSecret, returning an error if empty.
func NewClientSecret(raw string) (ClientSecret, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return ClientSecret{}, errClientSecretEmpty
	}

	return ClientSecret{value: v}, nil
}

func (clientSecret ClientSecret) MarshalJSON() ([]byte, error) {
	return clientSecret.value.MarshalJSON()
}

func (clientSecret ClientSecret) rawCallback(callback func(string) error) error {
	return clientSecret.value.rawCallback(callback)
}

// MustClientSecret creates a ClientSecret, panicking if invalid. For use in tests and constants only.
func MustClientSecret(raw string) ClientSecret {
	clientSecret, err := NewClientSecret(raw)
	if err != nil {
		panic(err)
	}

	return clientSecret
}

// Match reports whether this client secret matches the raw secret provided,
// using constant-time comparison.
func (clientSecret ClientSecret) Match(other ClientSecret) bool {
	expected := []byte(clientSecret.value.value)
	provided := []byte(other.value.value)

	if len(expected) != len(provided) {
		return false
	}

	return subtle.ConstantTimeCompare(expected, provided) == 1
}

// ScopeDescription is the human-readable description of a scope.
type ScopeDescription struct {
	value NonEmptyString
}

// NewScopeDescription creates a ScopeDescription, returning an error if empty.
func NewScopeDescription(raw string) (ScopeDescription, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return ScopeDescription{}, errScopeDescriptionEmpty
	}

	return ScopeDescription{value: v}, nil
}

func (scopeDescription ScopeDescription) MarshalJSON() ([]byte, error) {
	return scopeDescription.value.MarshalJSON()
}

func (scopeDescription ScopeDescription) rawCallback(callback func(string) error) error {
	return scopeDescription.value.rawCallback(callback)
}

// MustScopeDescription creates a ScopeDescription, panicking if invalid. For use in tests and constants only.
func MustScopeDescription(raw string) ScopeDescription {
	v, err := NewScopeDescription(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// RoleDescription is the human-readable description of a role.
type RoleDescription struct {
	value NonEmptyString
}

// NewRoleDescription creates a RoleDescription, returning an error if empty.
func NewRoleDescription(raw string) (RoleDescription, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return RoleDescription{}, errRoleDescriptionEmpty
	}

	return RoleDescription{value: v}, nil
}

func (roleDescription RoleDescription) MarshalJSON() ([]byte, error) {
	return roleDescription.value.MarshalJSON()
}

func (roleDescription RoleDescription) rawCallback(callback func(string) error) error {
	return roleDescription.value.rawCallback(callback)
}

// MustRoleDescription creates a RoleDescription, panicking if invalid. For use in tests and constants only.
func MustRoleDescription(raw string) RoleDescription {
	v, err := NewRoleDescription(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// AccessToken is an issued OAuth2 access token.
type AccessToken struct {
	value NonEmptyString
}

func NewAccessToken(raw string) (AccessToken, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return AccessToken{}, err
	}

	return AccessToken{value: v}, nil
}

func MustAccessToken(raw string) AccessToken {
	v, err := NewAccessToken(raw)
	if err != nil {
		panic(err)
	}

	return v
}

func (at AccessToken) AsByteArray() []byte { return ([]byte)(at.value.value + "\n") }
func (at AccessToken) rawCallback(callback func(string) error) error {
	return callback(at.value.value)
}

// ParseTokenAdapter is a standalone function that handles the generics.
func ParseTokenAdapter[T any](at AccessToken, parserFunc func(string) (T, error)) (T, error) {
	var zero T
	var result T
	var err error

	// We use the struct's callback to get the string, and execute the parser!
	cbErr := at.rawCallback(func(rawToken string) error {
		result, err = parserFunc(rawToken)
		return err
	})

	if cbErr != nil {
		return zero, cbErr
	}

	return result, nil
}

// TokenType is the type of an issued token (e.g. "Bearer").
type TokenType struct {
	value NonEmptyString
}

func NewTokenType(raw string) (TokenType, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return TokenType{}, err
	}

	return TokenType{value: v}, nil
}

func MustTokenType(raw string) TokenType {
	v, err := NewTokenType(raw)
	if err != nil {
		panic(err)
	}

	return v
}

func (tt TokenType) MarshalJSON() ([]byte, error) { return tt.value.MarshalJSON() }

func (tt TokenType) rawCallback(callback func(string) error) error {
	return tt.value.rawCallback(callback)
}

// IDToken is an issued OIDC ID token.
type IDToken struct {
	value NonEmptyString
}

func (rt RefreshToken) MarshalJSON() ([]byte, error) { return rt.value.MarshalJSON() }

func (rt RefreshToken) rawCallback(callback func(string) error) error {
	return rt.value.rawCallback(callback)
}

func NewIDToken(raw string) (IDToken, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return IDToken{}, err
	}

	return IDToken{value: v}, nil
}

func MustIDToken(raw string) IDToken {
	v, err := NewIDToken(raw)
	if err != nil {
		panic(err)
	}

	return v
}

func (it IDToken) MarshalJSON() ([]byte, error) { return it.value.MarshalJSON() }

func (it IDToken) rawCallback(callback func(string) error) error {
	return it.value.rawCallback(callback)
}

// RefreshToken is an issued OAuth2 refresh token.
type RefreshToken struct {
	value NonEmptyString
}

func NewRefreshToken(raw string) (RefreshToken, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return RefreshToken{}, err
	}

	return RefreshToken{value: v}, nil
}

func MustRefreshToken(raw string) RefreshToken {
	v, err := NewRefreshToken(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// ClientInfo is the base64-encoded MSAL client information.
type ClientInfo struct {
	value NonEmptyString
}

func NewClientInfo(raw string) (ClientInfo, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return ClientInfo{}, err
	}

	return ClientInfo{value: v}, nil
}

func MustClientInfo(raw string) ClientInfo {
	v, err := NewClientInfo(raw)
	if err != nil {
		panic(err)
	}

	return v
}

func (ci ClientInfo) MarshalJSON() ([]byte, error) { return ci.value.MarshalJSON() }

func (ci ClientInfo) rawCallback(callback func(string) error) error {
	return ci.value.rawCallback(callback)
}

// AuthCode is an issued OAuth2 authorization code.
type AuthCode struct {
	value NonEmptyString
}

func NewAuthCode(raw string) (AuthCode, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return AuthCode{}, err
	}

	return AuthCode{value: v}, nil
}

func MustAuthCode(raw string) AuthCode {
	v, err := NewAuthCode(raw)
	if err != nil {
		panic(err)
	}

	return v
}

func (ac AuthCode) MarshalJSON() ([]byte, error) { return ac.value.MarshalJSON() }

func (ac AuthCode) rawCallback(callback func(string) error) error {
	return ac.value.rawCallback(callback)
}

func JoinRoleValues(roles []RoleValue, sep string) string {
	ss := make([]string, len(roles))
	for i, r := range roles {
		ss[i] = r.value.value
	}

	return strings.Join(ss, sep)
}
