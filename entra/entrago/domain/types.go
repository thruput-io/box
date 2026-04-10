package domain

import (
	"crypto/subtle"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

const (
	comparisonMatch = 1
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

// AsClientID returns the tenant ID as a ClientID.
func (tenantID TenantID) AsClientID() ClientID { return ClientID(tenantID) }

// Value returns the UUID string representation.
func (tenantID TenantID) Value() string {
	return tenantID.value.String()
}

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

// Value returns the UUID string representation.
func (clientID ClientID) Value() string {
	return clientID.value.String()
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

// Value returns the UUID string representation.
func (userID UserID) Value() string {
	return userID.value.String()
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

// Value returns the UUID string representation.
func (groupID GroupID) Value() string {
	return groupID.value.String()
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

// Value returns the UUID string representation.
func (scopeID ScopeID) Value() string {
	return scopeID.value.String()
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

// Value returns the UUID string representation.
func (roleID RoleID) Value() string {
	return roleID.value.String()
}

// TenantName is the display name of a tenant.
type TenantName struct {
	value NonEmptyString
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

// AsAppName returns the tenant name as an AppName.
func (tenantName TenantName) AsAppName() AppName { return AppName(tenantName) }

// Value returns the underlying string value.
func (tenantName TenantName) Value() string {
	return tenantName.value.Value()
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
func MustAppName(raw string) AppName {
	name, err := NewAppName(raw)
	if err != nil {
		panic(err)
	}

	return name
}

// Value returns the underlying string value.
func (appName AppName) Value() string {
	return appName.value.Value()
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

// MustIdentifierURI creates an IdentifierURI, panicking if invalid. For use in tests and constants only.
func MustIdentifierURI(raw string) IdentifierURI {
	identifierURI, err := NewIdentifierURI(raw)
	if err != nil {
		panic(err)
	}

	return identifierURI
}

// Matches reports whether the identifier URI exactly matches the given raw string.
func (identifierURI IdentifierURI) Matches(raw string) bool {
	return identifierURI.value.value == raw
}

// MatchesPrefix reports whether the given raw string starts with the identifier URI.
func (identifierURI IdentifierURI) MatchesPrefix(raw string) bool {
	return strings.HasPrefix(raw, identifierURI.value.value)
}

// Value returns the underlying string value.
func (identifierURI IdentifierURI) Value() string {
	return identifierURI.value.Value()
}

// ScopeValue is the string value of a scope (e.g. "read", "api://xxx/.default").
type ScopeValue struct {
	value NonEmptyString
}

// NewScopeValue creates a ScopeValue from a raw string, returning an error if empty.
func NewScopeValue(raw string) (ScopeValue, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return ScopeValue{}, errScopeValueEmpty
	}

	return ScopeValue{value: v}, nil
}

// MustScopeValue creates a ScopeValue, panicking if invalid. For use in tests and constants only.
func MustScopeValue(raw string) ScopeValue {
	v, err := NewScopeValue(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Matches reports whether the scope value matches the given raw string.
func (scopeValue ScopeValue) Matches(raw string) bool {
	v := scopeValue.value.value

	return v == raw || strings.HasSuffix(raw, "/"+v)
}

// Value returns the underlying string value.
func (scopeValue ScopeValue) Value() string {
	return scopeValue.value.Value()
}

// RoleValue is the string value of a role (e.g. "Admin", "Reader").
type RoleValue struct {
	value NonEmptyString
}

// NewRoleValue creates a RoleValue from a raw string, returning an error if empty.
func NewRoleValue(raw string) (RoleValue, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return RoleValue{}, errRoleValueEmpty
	}

	return RoleValue{value: v}, nil
}

// MustRoleValue creates a RoleValue, panicking if invalid. For use in tests and constants only.
func MustRoleValue(raw string) RoleValue {
	v, err := NewRoleValue(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Matches reports whether the role value matches the given raw string.
func (roleValue RoleValue) Matches(raw string) bool {
	return roleValue.value.value == raw
}

// Value returns the underlying string value.
func (roleValue RoleValue) Value() string {
	return roleValue.value.Value()
}

// GroupName is the name of a group.
type GroupName struct {
	value NonEmptyString
}

// NewGroupName creates a GroupName from a raw string, returning an error if empty.
func NewGroupName(raw string) (GroupName, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return GroupName{}, errGroupNameEmpty
	}

	return GroupName{value: v}, nil
}

// MustGroupName creates a GroupName, panicking if invalid. For use in tests and constants only.
func MustGroupName(raw string) GroupName {
	v, err := NewGroupName(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Matches reports whether the group name matches the given raw string.
func (groupName GroupName) Matches(raw string) bool {
	return groupName.value.value == raw
}

// Value returns the underlying string value.
func (groupName GroupName) Value() string {
	return groupName.value.Value()
}

// Username is the login name of a user.
type Username struct {
	value NonEmptyString
}

// NewUsername creates a Username from a raw string, returning an error if empty.
func NewUsername(raw string) (Username, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return Username{}, errUsernameEmpty
	}

	return Username{value: v}, nil
}

// MustUsername creates a Username, panicking if invalid. For use in tests and constants only.
func MustUsername(raw string) Username {
	username, err := NewUsername(raw)
	if err != nil {
		panic(err)
	}

	return username
}

// Value returns the underlying string value.
func (username Username) Value() string {
	return username.value.Value()
}

// Password is the credential of a user.
type Password struct {
	value NonEmptyString
}

// NewPassword creates a Password from a raw string, returning an error if empty.
func NewPassword(raw string) (Password, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return Password{}, errPasswordEmpty
	}

	return Password{value: v}, nil
}

// MustPassword creates a Password, panicking if invalid. For use in tests and constants only.
func MustPassword(raw string) Password {
	password, err := NewPassword(raw)
	if err != nil {
		panic(err)
	}

	return password
}

// Value returns the underlying string value.
func (password Password) Value() string {
	return password.value.Value()
}

// DisplayName is the human-readable name of a user.
type DisplayName struct {
	value NonEmptyString
}

// NewDisplayName creates a DisplayName from a raw string, returning an error if empty.
func NewDisplayName(raw string) (DisplayName, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return DisplayName{}, errDisplayNameEmpty
	}

	return DisplayName{value: v}, nil
}

// MustDisplayName creates a DisplayName, panicking if invalid. For use in tests and constants only.
func MustDisplayName(raw string) DisplayName {
	displayName, err := NewDisplayName(raw)
	if err != nil {
		panic(err)
	}

	return displayName
}

// Value returns the underlying string value.
func (displayName DisplayName) Value() string {
	return displayName.value.Value()
}

// Email is the email address of a user.
type Email struct {
	value NonEmptyString
}

// NewEmail creates an Email from a raw string, returning an error if empty.
func NewEmail(raw string) (Email, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return Email{}, errEmailEmpty
	}

	return Email{value: v}, nil
}

// MustEmail creates an Email, panicking if invalid. For use in tests and constants only.
func MustEmail(raw string) Email {
	email, err := NewEmail(raw)
	if err != nil {
		panic(err)
	}

	return email
}

// Value returns the underlying string value.
func (email Email) Value() string {
	return email.value.Value()
}

// RedirectURL is a permitted OAuth2 redirect URI.
type RedirectURL struct {
	value url.URL
}

// NewRedirectURL creates a RedirectURL from a raw string, returning an error if empty or invalid.
func NewRedirectURL(raw string) (RedirectURL, error) {
	if raw == emptyString {
		return RedirectURL{}, errRedirectURLEmpty
	}

	v, err := url.Parse(raw)
	if err != nil {
		return RedirectURL{}, errRedirectURLEmpty
	}

	return RedirectURL{*v}, nil
}

// MustRedirectURL creates a RedirectURL, panicking if invalid. For use in tests and constants only.
func MustRedirectURL(raw string) RedirectURL {
	redirectURL, err := NewRedirectURL(raw)
	if err != nil {
		panic(err)
	}

	return redirectURL
}

// AsURL returns the underlying url.URL value.
func (redirectURL RedirectURL) AsURL() url.URL {
	return redirectURL.value
}

// Value returns the URL string representation.
func (redirectURL RedirectURL) Value() string {
	return redirectURL.value.String()
}

// ClientSecret is the secret credential of a confidential client.
type ClientSecret struct {
	value NonEmptyString
}

// NewClientSecret creates a ClientSecret from a raw string, returning an error if empty.
func NewClientSecret(raw string) (ClientSecret, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return ClientSecret{}, errClientSecretEmpty
	}

	return ClientSecret{value: v}, nil
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

	return subtle.ConstantTimeCompare(expected, provided) == comparisonMatch
}

// Value returns the underlying string value.
func (clientSecret ClientSecret) Value() string {
	return clientSecret.value.Value()
}

// ScopeDescription is the human-readable description of a scope.
type ScopeDescription struct {
	value NonEmptyString
}

// NewScopeDescription creates a ScopeDescription from a raw string, returning an error if empty.
func NewScopeDescription(raw string) (ScopeDescription, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return ScopeDescription{}, errScopeDescriptionEmpty
	}

	return ScopeDescription{value: v}, nil
}

// MustScopeDescription creates a ScopeDescription, panicking if invalid. For use in tests and constants only.
func MustScopeDescription(raw string) ScopeDescription {
	v, err := NewScopeDescription(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (scopeDescription ScopeDescription) Value() string {
	return scopeDescription.value.Value()
}

// RoleDescription is the human-readable description of a role.
type RoleDescription struct {
	value NonEmptyString
}

// NewRoleDescription creates a RoleDescription from a raw string, returning an error if empty.
func NewRoleDescription(raw string) (RoleDescription, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return RoleDescription{}, errRoleDescriptionEmpty
	}

	return RoleDescription{value: v}, nil
}

// MustRoleDescription creates a RoleDescription, panicking if invalid. For use in tests and constants only.
func MustRoleDescription(raw string) RoleDescription {
	v, err := NewRoleDescription(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (roleDescription RoleDescription) Value() string {
	return roleDescription.value.Value()
}

// Nonce is an OAuth2 nonce parameter used to prevent replay attacks.
type Nonce struct {
	value NonEmptyString
}

// NewNonce creates a Nonce from a raw string, returning an error if empty.
func NewNonce(raw string) (Nonce, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return Nonce{}, errNonceEmpty
	}

	return Nonce{value: v}, nil
}

// MustNonce creates a Nonce, panicking if invalid. For use in tests and constants only.
func MustNonce(raw string) Nonce {
	v, err := NewNonce(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (n Nonce) Value() string { return n.value.value }

// CorrelationID is an opaque request correlation identifier.
type CorrelationID struct {
	value NonEmptyString
}

// NewCorrelationID creates a CorrelationID from a raw string, returning an error if empty.
func NewCorrelationID(raw string) (CorrelationID, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return CorrelationID{}, errCorrelationIDEmpty
	}

	return CorrelationID{value: v}, nil
}

// MustCorrelationID creates a CorrelationID, panicking if invalid. For use in tests and constants only.
func MustCorrelationID(raw string) CorrelationID {
	v, err := NewCorrelationID(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (c CorrelationID) Value() string { return c.value.value }

// BaseURL is the validated scheme+host origin of the identity server (e.g. https://login.microsoftonline.com).
type BaseURL struct {
	value url.URL
}

// NewBaseURL creates a BaseURL from a raw string, returning an error if empty or invalid.
func NewBaseURL(raw string) (BaseURL, error) {
	if raw == emptyString {
		return BaseURL{}, errBaseURLEmpty
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == emptyString {
		return BaseURL{}, errBaseURLInvalid
	}

	return BaseURL{value: *parsed}, nil
}

// MustBaseURL creates a BaseURL, panicking if invalid. For use in tests and constants only.
func MustBaseURL(raw string) BaseURL {
	v, err := NewBaseURL(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (b BaseURL) Value() string { return b.value.String() }

// Issuer is a validated JWT issuer URI (scheme + host + optional path).
type Issuer struct {
	value NonEmptyString
}

// NewIssuer creates an Issuer from a raw string, returning an error if empty.
func NewIssuer(raw string) (Issuer, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return Issuer{}, errIssuerEmpty
	}

	return Issuer{value: v}, nil
}

// MustIssuer creates an Issuer, panicking if invalid. For use in tests and constants only.
func MustIssuer(raw string) Issuer {
	v, err := NewIssuer(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (i Issuer) Value() string { return i.value.value }

// Subject is a JWT subject claim value — the unique identifier of the principal.
type Subject struct {
	value NonEmptyString
}

// NewSubject creates a Subject from a raw string, returning an error if empty.
func NewSubject(raw string) (Subject, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return Subject{}, errSubjectEmpty
	}

	return Subject{value: v}, nil
}

// MustSubject creates a Subject, panicking if invalid. For use in tests and constants only.
func MustSubject(raw string) Subject {
	v, err := NewSubject(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (s Subject) Value() string { return s.value.value }

// TokenVersion is the version string embedded in JWT claims (e.g. "1.0" or "2.0").
type TokenVersion struct {
	value NonEmptyString
}

// NewTokenVersion creates a TokenVersion from a raw string, returning an error if empty.
func NewTokenVersion(raw string) (TokenVersion, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return TokenVersion{}, errTokenVersionEmpty
	}

	return TokenVersion{value: v}, nil
}

// MustTokenVersion creates a TokenVersion, panicking if invalid. For use in tests and constants only.
func MustTokenVersion(raw string) TokenVersion {
	v, err := NewTokenVersion(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (tv TokenVersion) Value() string { return tv.value.value }

// OAuthState is the opaque state parameter passed through the OAuth2 authorization flow.
type OAuthState struct {
	value NonEmptyString
}

// NewOAuthState creates an OAuthState from a raw string, returning an error if empty.
func NewOAuthState(raw string) (OAuthState, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return OAuthState{}, errOAuthStateEmpty
	}

	return OAuthState{value: v}, nil
}

// MustOAuthState creates an OAuthState, panicking if invalid. For use in tests and constants only.
func MustOAuthState(raw string) OAuthState {
	v, err := NewOAuthState(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (o OAuthState) Value() string { return o.value.value }

// ResponseMode is the OAuth2 response_mode parameter value (e.g. "query", "fragment", "form_post").
type ResponseMode struct {
	value NonEmptyString
}

// NewResponseMode creates a ResponseMode from a raw string, returning an error if empty.
func NewResponseMode(raw string) (ResponseMode, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return ResponseMode{}, errResponseModeEmpty
	}

	return ResponseMode{value: v}, nil
}

// MustResponseMode creates a ResponseMode, panicking if invalid. For use in tests and constants only.
func MustResponseMode(raw string) ResponseMode {
	v, err := NewResponseMode(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (rm ResponseMode) Value() string { return rm.value.value }

// ResponseType is the OAuth2 response_type parameter value (e.g. "code", "id_token").
type ResponseType struct {
	value NonEmptyString
}

// NewResponseType creates a ResponseType from a raw string, returning an error if empty.
func NewResponseType(raw string) (ResponseType, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return ResponseType{}, errResponseTypeEmpty
	}

	return ResponseType{value: v}, nil
}

// MustResponseType creates a ResponseType, panicking if invalid. For use in tests and constants only.
func MustResponseType(raw string) ResponseType {
	v, err := NewResponseType(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (rt ResponseType) Value() string { return rt.value.value }
