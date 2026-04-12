package domain

import (
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/mo"
	moeither "github.com/samber/mo/either"
)

// TenantID is the unique identifier for a tenant.
type TenantID struct {
	value uuid.UUID
}

// NewTenantID creates a TenantID from a UUID string, returning an error if invalid.

func NewTenantID(raw string) mo.Either[Error, TenantID] {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return mo.Left[Error, TenantID](errTenantIDInvalid)
	}

	return mo.Right[Error](TenantID{value: parsed})
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
func NewClientID(raw string) mo.Either[Error, ClientID] {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return mo.Left[Error, ClientID](errClientIDInvalid)
	}

	return mo.Right[Error](ClientID{value: parsed})
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
func NewUserID(raw string) mo.Either[Error, UserID] {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return mo.Left[Error, UserID](errUserIDInvalid)
	}

	return mo.Right[Error](UserID{value: parsed})
}

// UserIDFromUUID wraps an already-parsed uuid.UUID.

// UUID returns the underlying uuid.UUID value.

// Value returns the UUID string representation.
func (userID UserID) Value() string {
	return userID.value.String()
}

// GroupID is the unique identifier for a group.
type GroupID struct {
	value uuid.UUID
}

// NewGroupID creates a GroupID from a UUID string, returning an error if invalid.
func NewGroupID(raw string) mo.Either[Error, GroupID] {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return mo.Left[Error, GroupID](errGroupIDInvalid)
	}

	return mo.Right[Error](GroupID{value: parsed})
}

// GroupIDFromUUID wraps an already-parsed uuid.UUID.

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
func NewScopeID(raw string) mo.Either[Error, ScopeID] {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return mo.Left[Error, ScopeID](errScopeIDInvalid)
	}

	return mo.Right[Error](ScopeID{value: parsed})
}

// ScopeIDFromUUID wraps an already-parsed uuid.UUID.

// UUID returns the underlying uuid.UUID value.

// Value returns the UUID string representation.

// RoleID is the unique identifier for a role.
type RoleID struct {
	value uuid.UUID
}

// NewRoleID creates a RoleID from a UUID string, returning an error if invalid.
func NewRoleID(raw string) mo.Either[Error, RoleID] {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return mo.Left[Error, RoleID](errRoleIDInvalid)
	}

	return mo.Right[Error](RoleID{value: parsed})
}

// RoleIDFromUUID wraps an already-parsed uuid.UUID.

// UUID returns the underlying uuid.UUID value.

// Value returns the UUID string representation.

// TenantName is the display name of a tenant.
type TenantName struct {
	value NonEmptyString
}

// NewTenantName creates a TenantName, returning an error if empty.
func NewTenantName(raw string) mo.Either[Error, TenantName] {
	return moeither.MapRight[Error, NonEmptyString, TenantName](func(nes NonEmptyString) TenantName {
		return TenantName{value: nes}
	})(NewNonEmptyString(raw))
}

// MustTenantName creates a TenantName, panicking if invalid. For use in tests and constants only.

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
func NewAppName(raw string) mo.Either[Error, AppName] {
	return moeither.MapRight[Error, NonEmptyString, AppName](func(nes NonEmptyString) AppName {
		return AppName{value: nes}
	})(NewNonEmptyString(raw))
}

// MustAppName creates an AppName, panicking if invalid. For use in tests and constants only.
func MustAppName(raw string) AppName {
	return NewAppName(raw).MustRight()
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
func NewIdentifierURI(raw string) mo.Either[Error, IdentifierURI] {
	return moeither.MapRight[Error, NonEmptyString, IdentifierURI](func(nes NonEmptyString) IdentifierURI {
		return IdentifierURI{value: nes}
	})(NewNonEmptyString(raw))
}

// MustIdentifierURI creates an IdentifierURI, panicking if invalid. For use in tests and constants only.
func MustIdentifierURI(raw string) IdentifierURI {
	return NewIdentifierURI(raw).MustRight()
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
func NewScopeValue(raw string) mo.Either[Error, ScopeValue] {
	return moeither.MapRight[Error, NonEmptyString, ScopeValue](func(nes NonEmptyString) ScopeValue {
		return ScopeValue{value: nes}
	})(NewNonEmptyString(raw))
}

// MustScopeValue creates a ScopeValue, panicking if invalid. For use in tests and constants only.

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
func NewRoleValue(raw string) mo.Either[Error, RoleValue] {
	return moeither.MapRight[Error, NonEmptyString, RoleValue](func(nes NonEmptyString) RoleValue {
		return RoleValue{value: nes}
	})(NewNonEmptyString(raw))
}

// MustRoleValue creates a RoleValue, panicking if invalid. For use in tests and constants only.

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
func NewGroupName(raw string) mo.Either[Error, GroupName] {
	return moeither.MapRight[Error, NonEmptyString, GroupName](func(nes NonEmptyString) GroupName {
		return GroupName{value: nes}
	})(NewNonEmptyString(raw))
}

// MustGroupName creates a GroupName, panicking if invalid. For use in tests and constants only.

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
func NewUsername(raw string) mo.Either[Error, Username] {
	return moeither.MapRight[Error, NonEmptyString, Username](func(nes NonEmptyString) Username {
		return Username{value: nes}
	})(NewNonEmptyString(raw))
}

// MustUsername creates a Username, panicking if invalid. For use in tests and constants only.

// Value returns the underlying string value.
func (username Username) Value() string {
	return username.value.Value()
}

// Password is the credential of a user.
type Password struct {
	value NonEmptyString
}

// NewPassword creates a Password from a raw string, returning an error if empty.
func NewPassword(raw string) mo.Either[Error, Password] {
	return moeither.MapRight[Error, NonEmptyString, Password](func(nes NonEmptyString) Password {
		return Password{value: nes}
	})(NewNonEmptyString(raw))
}

// MustPassword creates a Password, panicking if invalid. For use in tests and constants only.

// Value returns the underlying string value.
func (password Password) Value() string {
	return password.value.Value()
}

// DisplayName is the human-readable name of a user.
type DisplayName struct {
	value NonEmptyString
}

// NewDisplayName creates a DisplayName from a raw string, returning an error if empty.
func NewDisplayName(raw string) mo.Either[Error, DisplayName] {
	return moeither.MapRight[Error, NonEmptyString, DisplayName](func(nes NonEmptyString) DisplayName {
		return DisplayName{value: nes}
	})(NewNonEmptyString(raw))
}

// MustDisplayName creates a DisplayName, panicking if invalid. For use in tests and constants only.
func MustDisplayName(raw string) DisplayName {
	return NewDisplayName(raw).MustRight()
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
func NewEmail(raw string) mo.Either[Error, Email] {
	return moeither.MapRight[Error, NonEmptyString, Email](func(nes NonEmptyString) Email {
		return Email{value: nes}
	})(NewNonEmptyString(raw))
}

// MustEmail creates an Email, panicking if invalid. For use in tests and constants only.
func MustEmail(raw string) Email {
	return NewEmail(raw).MustRight()
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
func NewRedirectURL(raw string) mo.Either[Error, RedirectURL] {
	return moeither.FlatMapRight[Error, NonEmptyString, RedirectURL](func(nes NonEmptyString) mo.Either[Error, RedirectURL] {
		v, err := url.Parse(nes.Value())
		if err != nil || v == nil {
			return mo.Left[Error, RedirectURL](errRedirectURLInvalid)
		}

		return mo.Right[Error](RedirectURL{value: *v})
	})(NewNonEmptyString(raw))
}

// MustRedirectURL creates a RedirectURL, panicking if invalid. For use in tests and constants only.

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
func NewClientSecret(raw string) mo.Either[Error, ClientSecret] {
	return moeither.MapRight[Error, NonEmptyString, ClientSecret](func(nes NonEmptyString) ClientSecret {
		return ClientSecret{value: nes}
	})(NewNonEmptyString(raw))
}

// MustClientSecret creates a ClientSecret, panicking if invalid. For use in tests and constants only.

// Value returns the underlying string value.
func (clientSecret ClientSecret) Value() string {
	return clientSecret.value.Value()
}

// ScopeDescription is the human-readable description of a scope.
type ScopeDescription struct {
	value NonEmptyString
}

// NewScopeDescription creates a ScopeDescription from a raw string, returning an error if empty.
func NewScopeDescription(raw string) mo.Either[Error, ScopeDescription] {
	return moeither.MapRight[Error, NonEmptyString, ScopeDescription](func(nes NonEmptyString) ScopeDescription {
		return ScopeDescription{value: nes}
	})(NewNonEmptyString(raw))
}

// MustScopeDescription creates a ScopeDescription, panicking if invalid. For use in tests and constants only.

// Value returns the underlying string value.
func (scopeDescription ScopeDescription) Value() string {
	return scopeDescription.value.Value()
}

// RoleDescription is the human-readable description of a role.
type RoleDescription struct {
	value NonEmptyString
}

// NewRoleDescription creates a RoleDescription from a raw string, returning an error if empty.
func NewRoleDescription(raw string) mo.Either[Error, RoleDescription] {
	return moeither.MapRight[Error, NonEmptyString, RoleDescription](func(nes NonEmptyString) RoleDescription {
		return RoleDescription{value: nes}
	})(NewNonEmptyString(raw))
}

// MustRoleDescription creates a RoleDescription, panicking if invalid. For use in tests and constants only.

// Value returns the underlying string value.
func (roleDescription RoleDescription) Value() string {
	return roleDescription.value.Value()
}

// Nonce is an OAuth2 nonce parameter used to prevent replay attacks.
type Nonce struct {
	value NonEmptyString
}

// NewNonce creates a Nonce from a raw string, returning an error if empty.
func NewNonce(raw string) mo.Either[Error, Nonce] {
	return moeither.MapRight[Error, NonEmptyString, Nonce](func(nes NonEmptyString) Nonce {
		return Nonce{value: nes}
	})(NewNonEmptyString(raw))
}

// MustNonce creates a Nonce, panicking if invalid. For use in tests and constants only.

// Value returns the underlying string value.
func (n Nonce) Value() string { return n.value.value }

// CorrelationID is an opaque request correlation identifier.
type CorrelationID struct {
	value NonEmptyString
}

// NewCorrelationID creates a CorrelationID from a raw string, returning an error if empty.
func NewCorrelationID(raw string) mo.Either[Error, CorrelationID] {
	return moeither.MapRight[Error, NonEmptyString, CorrelationID](func(nes NonEmptyString) CorrelationID {
		return CorrelationID{value: nes}
	})(NewNonEmptyString(raw))
}

// MustCorrelationID creates a CorrelationID, panicking if invalid. For use in tests and constants only.
func MustCorrelationID(raw string) CorrelationID {
	return NewCorrelationID(raw).MustRight()
}

// Value returns the underlying string value.
func (c CorrelationID) Value() string { return c.value.value }

// BaseURL is the validated scheme+host origin of the identity server (e.g. https://login.microsoftonline.com).
type BaseURL struct {
	value url.URL
}

// NewBaseURL creates a BaseURL from a raw string, returning an error if empty or invalid.
func NewBaseURL(raw string) mo.Either[Error, BaseURL] {
	return moeither.FlatMapRight[Error, NonEmptyString, BaseURL](func(nes NonEmptyString) mo.Either[Error, BaseURL] {
		parsed, err := url.ParseRequestURI(nes.Value())
		if err != nil || parsed.Host == emptyString {
			return mo.Left[Error, BaseURL](errBaseURLInvalid)
		}

		return mo.Right[Error](BaseURL{value: *parsed})
	})(NewNonEmptyString(raw))
}

// MustBaseURL creates a BaseURL, panicking if invalid. For use in tests and constants only.
func MustBaseURL(raw string) BaseURL {
	return NewBaseURL(raw).MustRight()
}

// Value returns the underlying string value.
func (b BaseURL) Value() string { return b.value.String() }

// Issuer is a validated JWT issuer URI (scheme + host + optional path).
type Issuer struct {
	value NonEmptyString
}

// NewIssuer creates an Issuer from a raw string, returning an error if empty.
func NewIssuer(raw string) mo.Either[Error, Issuer] {
	return moeither.MapRight[Error, NonEmptyString, Issuer](func(nes NonEmptyString) Issuer {
		return Issuer{value: nes}
	})(NewNonEmptyString(raw))
}

// MustIssuer creates an Issuer, panicking if invalid. For use in tests and constants only.
func MustIssuer(raw string) Issuer {
	return NewIssuer(raw).MustRight()
}

// Value returns the underlying string value.
func (i Issuer) Value() string { return i.value.value }

// Subject is a JWT subject claim value — the unique identifier of the principal.
type Subject struct {
	value NonEmptyString
}

// Value returns the underlying string value.
func (s Subject) Value() string { return s.value.value }

// TokenVersion is the version string embedded in JWT claims (e.g. "1.0" or "2.0").
type TokenVersion struct {
	value NonEmptyString
}

// NewTokenVersion creates a TokenVersion from a raw string, returning an error if empty.
func NewTokenVersion(raw string) mo.Either[Error, TokenVersion] {
	return moeither.MapRight[Error, NonEmptyString, TokenVersion](func(nes NonEmptyString) TokenVersion {
		return TokenVersion{value: nes}
	})(NewNonEmptyString(raw))
}

// MustTokenVersion creates a TokenVersion, panicking if invalid. For use in tests and constants only.
func MustTokenVersion(raw string) TokenVersion {
	return NewTokenVersion(raw).MustRight()
}

// Value returns the underlying string value.
func (tv TokenVersion) Value() string { return tv.value.value }

// OAuthState is the opaque state parameter passed through the OAuth2 authorization flow.
type OAuthState struct {
	value NonEmptyString
}

// NewOAuthState creates an OAuthState from a raw string, returning an error if empty.
func NewOAuthState(raw string) mo.Either[Error, OAuthState] {
	return moeither.MapRight[Error, NonEmptyString, OAuthState](func(nes NonEmptyString) OAuthState {
		return OAuthState{value: nes}
	})(NewNonEmptyString(raw))
}

// Value returns the underlying string value.
func (o OAuthState) Value() string { return o.value.value }

// ResponseMode is the OAuth2 response_mode parameter value (e.g. "query", "fragment", "form_post").
type ResponseMode struct {
	value NonEmptyString
}

// NewResponseMode creates a ResponseMode from a raw string, returning an error if empty.
func NewResponseMode(raw string) mo.Either[Error, ResponseMode] {
	return moeither.MapRight[Error, NonEmptyString, ResponseMode](func(nes NonEmptyString) ResponseMode {
		return ResponseMode{value: nes}
	})(NewNonEmptyString(raw))
}

// Value returns the underlying string value.
func (rm ResponseMode) Value() string { return rm.value.value }

// ResponseType is the OAuth2 response_type parameter value (e.g. "code", "id_token").
type ResponseType struct {
	value NonEmptyString
}

// NewResponseType creates a ResponseType from a raw string, returning an error if empty.
func NewResponseType(raw string) mo.Either[Error, ResponseType] {
	return moeither.MapRight[Error, NonEmptyString, ResponseType](func(nes NonEmptyString) ResponseType {
		return ResponseType{value: nes}
	})(NewNonEmptyString(raw))
}

// Value returns the underlying string value.
func (rt ResponseType) Value() string { return rt.value.value }

type claimKey struct {
	key         string
	valueKind   claimValueKind
	allowsArray bool
}

var (
	subClaim   = claimKey{key: "sub", valueKind: claimValueKindString, allowsArray: false}
	nonceClaim = claimKey{key: "nonce", valueKind: claimValueKindString, allowsArray: false}

	expClaim = claimKey{key: "exp", valueKind: claimValueKindInt64, allowsArray: false}
	iatClaim = claimKey{key: "iat", valueKind: claimValueKindInt64, allowsArray: false}
	nbfClaim = claimKey{key: "nbf", valueKind: claimValueKindInt64, allowsArray: false}
	utiClaim = claimKey{key: "uti", valueKind: claimValueKindString, allowsArray: false}

	tidClaim = claimKey{key: "tid", valueKind: claimValueKindString, allowsArray: false}
	verClaim = claimKey{key: "ver", valueKind: claimValueKindString, allowsArray: false}
	oidClaim = claimKey{key: "oid", valueKind: claimValueKindString, allowsArray: false}

	rolesClaim  = claimKey{key: "roles", valueKind: claimValueKindString, allowsArray: true}
	groupsClaim = claimKey{key: "groups", valueKind: claimValueKindString, allowsArray: true}
	scpClaim    = claimKey{key: "scp", valueKind: claimValueKindString, allowsArray: false}

	azpClaim    = claimKey{key: "azp", valueKind: claimValueKindString, allowsArray: false}
	azpacrClaim = claimKey{key: "azpacr", valueKind: claimValueKindString, allowsArray: false}
	appidClaim  = claimKey{key: "appid", valueKind: claimValueKindString, allowsArray: false}

	nameClaim              = claimKey{key: "name", valueKind: claimValueKindString, allowsArray: false}
	preferredUsernameClaim = claimKey{key: "preferred_username", valueKind: claimValueKindString, allowsArray: false}
	emailClaim             = claimKey{key: "email", valueKind: claimValueKindString, allowsArray: false}
	uniqueNameClaim        = claimKey{key: "unique_name", valueKind: claimValueKindString, allowsArray: false}

	issClaim = claimKey{key: "iss", valueKind: claimValueKindString, allowsArray: false}
	audClaim = claimKey{key: "aud", valueKind: claimValueKindString, allowsArray: true}
	typClaim = claimKey{key: "typ", valueKind: claimValueKindString, allowsArray: false}

	aioClaim    = claimKey{key: "aio", valueKind: claimValueKindString, allowsArray: false}
	rhClaim     = claimKey{key: "rh", valueKind: claimValueKindString, allowsArray: false}
	sidClaim    = claimKey{key: "sid", valueKind: claimValueKindString, allowsArray: false}
	xmsFtdClaim = claimKey{key: "xms_ftd", valueKind: claimValueKindString, allowsArray: false}

	jtiClaim     = claimKey{key: "jti", valueKind: claimValueKindString, allowsArray: false}
	azpaclsClaim = claimKey{key: "azpacls", valueKind: claimValueKindString, allowsArray: false}
)

var claimKeyMap = map[string]claimKey{
	subClaim.key:   subClaim,
	nonceClaim.key: nonceClaim,

	expClaim.key: expClaim,
	iatClaim.key: iatClaim,
	nbfClaim.key: nbfClaim,
	utiClaim.key: utiClaim,

	tidClaim.key: tidClaim,
	verClaim.key: verClaim,
	oidClaim.key: oidClaim,

	rolesClaim.key:  rolesClaim,
	groupsClaim.key: groupsClaim,
	scpClaim.key:    scpClaim,

	azpClaim.key:    azpClaim,
	azpacrClaim.key: azpacrClaim,
	appidClaim.key:  appidClaim,

	nameClaim.key:              nameClaim,
	preferredUsernameClaim.key: preferredUsernameClaim,
	emailClaim.key:             emailClaim,
	uniqueNameClaim.key:        uniqueNameClaim,

	issClaim.key: issClaim,
	audClaim.key: audClaim,
	typClaim.key: typClaim,

	aioClaim.key:    aioClaim,
	rhClaim.key:     rhClaim,
	sidClaim.key:    sidClaim,
	xmsFtdClaim.key: xmsFtdClaim,

	jtiClaim.key:     jtiClaim,
	azpaclsClaim.key: azpaclsClaim,
}

func ValidClaim(rawKey string) mo.Option[ValidClaimKey] {
	ck, ok := claimKeyMap[rawKey]
	if !ok {
		return mo.None[ValidClaimKey]()
	}

	return mo.Some(ValidClaimKey{key: ck})
}

func AsString(uuid uuid.UUID) NonEmptyString {
	return NonEmptyString{value: uuid.String()}
}

func (tenantID TenantID) AsClaim() Claim {
	return newStringClaim(tidClaim, AsString(tenantID.value))
}

func (clientID ClientID) AsClaim() Claim {
	return newStringClaim(azpClaim, AsString(clientID.value))
}

func (nonce Nonce) AsClaim() Claim {
	return newStringClaim(nonceClaim, nonce.value)
}

func (issuer Issuer) AsClaim() Claim {
	return newStringClaim(issClaim, issuer.value)
}

func (aud IdentifierURI) AsClaim() Claim {
	return newStringClaim(audClaim, aud.value)
}

func (subject Subject) AsClaim() Claim {
	return newStringClaim(subClaim, subject.value)
}

func (version TokenVersion) AsClaim() Claim {
	return newStringClaim(verClaim, version.value)
}

func (displayName DisplayName) AsClaim() Claim {
	return newStringClaim(nameClaim, displayName.value)
}

func (email Email) AsClaim() Claim {
	return newStringClaim(emailClaim, email.value)
}

func (email Email) AsPreferredUsernameClaim() Claim {
	return newStringClaim(preferredUsernameClaim, email.value)
}

func (email Email) AsUniqueNameClaim() Claim {
	return newStringClaim(uniqueNameClaim, email.value)
}

func (subject Subject) AsOidClaim() Claim {
	return newStringClaim(oidClaim, subject.value)
}

func (clientID ClientID) AsAppidClaim() Claim {
	return newStringClaim(appidClaim, AsString(clientID.value))
}

func (clientID ClientID) AsAudClaim() Claim {
	return newStringClaim(audClaim, AsString(clientID.value))
}

func (issuer Issuer) AsAudClaim() Claim {
	return newStringClaim(audClaim, issuer.value)
}

// ClaimComparable implementations — one per domain type used as a claim value.
// Each is placed here, alongside its type definition.

func (v TenantID) claimComparable()         {}
func (v TenantID) toClaimValue() claimValue { return newClaimStringValue(AsString(v.value)) }

func (v ClientID) claimComparable()         {}
func (v ClientID) toClaimValue() claimValue { return newClaimStringValue(AsString(v.value)) }

func (v GroupID) claimComparable()         {}
func (v GroupID) toClaimValue() claimValue { return newClaimStringValue(AsString(v.value)) }

func (v Nonce) claimComparable()         {}
func (v Nonce) toClaimValue() claimValue { return newClaimStringValue(v.value) }

func (v Issuer) claimComparable()         {}
func (v Issuer) toClaimValue() claimValue { return newClaimStringValue(v.value) }

func (v IdentifierURI) claimComparable()         {}
func (v IdentifierURI) toClaimValue() claimValue { return newClaimStringValue(v.value) }

func (v Subject) claimComparable()         {}
func (v Subject) toClaimValue() claimValue { return newClaimStringValue(v.value) }

func (v TokenVersion) claimComparable()         {}
func (v TokenVersion) toClaimValue() claimValue { return newClaimStringValue(v.value) }

func (v DisplayName) claimComparable()         {}
func (v DisplayName) toClaimValue() claimValue { return newClaimStringValue(v.value) }

func (v Email) claimComparable()         {}
func (v Email) toClaimValue() claimValue { return newClaimStringValue(v.value) }

func (v ScopeValue) claimComparable()         {}
func (v ScopeValue) toClaimValue() claimValue { return newClaimStringValue(v.value) }

func (v RoleValue) claimComparable()         {}
func (v RoleValue) toClaimValue() claimValue { return newClaimStringValue(v.value) }

func (v NonEmptyString) claimComparable()         {}
func (v NonEmptyString) toClaimValue() claimValue { return newClaimStringValue(v) }
