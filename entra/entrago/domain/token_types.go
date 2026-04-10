package domain

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// AccessToken is an issued OAuth2 access token.
type AccessToken struct {
	value NonEmptyString
}

// NewAccessToken creates an AccessToken from a raw string, returning an error if empty.
func NewAccessToken(raw string) (AccessToken, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return AccessToken{}, err
	}

	return AccessToken{value: v}, nil
}

// MustAccessToken creates an AccessToken from a raw string, panicking if invalid.
func MustAccessToken(raw string) AccessToken {
	v, err := NewAccessToken(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// AsByteArray returns the token as a byte slice with a trailing newline.
func (at AccessToken) AsByteArray() []byte { return ([]byte)(at.value.value + "\n") }

// MarshalJSON serializes AccessToken as its raw token string.
func (at AccessToken) MarshalJSON() ([]byte, error) {
	encoded, err := json.Marshal(at.value.value)
	if err != nil {
		return nil, fmt.Errorf("marshal access token: %w", err)
	}

	return encoded, nil
}

// Value returns the raw token string.
func (at AccessToken) Value() string {
	return at.value.value
}

// TokenType is the type of an issued token (e.g. "Bearer").
type TokenType struct {
	value NonEmptyString
}

// NewTokenType creates a TokenType from a raw string, returning an error if empty.
func NewTokenType(raw string) (TokenType, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return TokenType{}, err
	}

	return TokenType{value: v}, nil
}

// MustTokenType creates a TokenType from a raw string, panicking if invalid.
func MustTokenType(raw string) TokenType {
	v, err := NewTokenType(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// MarshalJSON serializes TokenType as its raw string value.
func (tt TokenType) MarshalJSON() ([]byte, error) {
	encoded, err := json.Marshal(tt.value.value)
	if err != nil {
		return nil, fmt.Errorf("marshal token type: %w", err)
	}

	return encoded, nil
}

// IDToken is an issued OIDC ID token.
type IDToken struct {
	value NonEmptyString
}

// NewIDToken creates an IDToken from a raw string, returning an error if empty.
func NewIDToken(raw string) (IDToken, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return IDToken{}, err
	}

	return IDToken{value: v}, nil
}

// MustIDToken creates an IDToken from a raw string, panicking if invalid.
func MustIDToken(raw string) IDToken {
	v, err := NewIDToken(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the raw token string.
func (it IDToken) Value() string {
	return it.value.value
}

// MarshalJSON serializes IDToken as its raw token string.
func (it IDToken) MarshalJSON() ([]byte, error) {
	encoded, err := json.Marshal(it.value.value)
	if err != nil {
		return nil, fmt.Errorf("marshal id token: %w", err)
	}

	return encoded, nil
}

// RefreshToken is an issued OAuth2 refresh token.
type RefreshToken struct {
	value NonEmptyString
}

// NewRefreshToken creates a RefreshToken from a raw string, returning an error if empty.
func NewRefreshToken(raw string) (RefreshToken, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return RefreshToken{}, err
	}

	return RefreshToken{value: v}, nil
}

// MustRefreshToken creates a RefreshToken from a raw string, panicking if invalid.
func MustRefreshToken(raw string) RefreshToken {
	v, err := NewRefreshToken(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// MarshalJSON serializes RefreshToken as its raw token string.
func (rt RefreshToken) MarshalJSON() ([]byte, error) {
	encoded, err := json.Marshal(rt.value.value)
	if err != nil {
		return nil, fmt.Errorf("marshal refresh token: %w", err)
	}

	return encoded, nil
}

// Value returns the underlying string value.
func (rt RefreshToken) Value() string {
	return rt.value.Value()
}

// ClientInfo is the base64-encoded MSAL client information.
type ClientInfo struct {
	value NonEmptyString
}

// NewClientInfo creates a ClientInfo from a raw string, returning an error if empty.
func NewClientInfo(raw string) (ClientInfo, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return ClientInfo{}, err
	}

	return ClientInfo{value: v}, nil
}

// MustClientInfo creates a ClientInfo from a raw string, panicking if invalid.
func MustClientInfo(raw string) ClientInfo {
	v, err := NewClientInfo(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// MarshalJSON serializes ClientInfo as its raw string value.
func (ci ClientInfo) MarshalJSON() ([]byte, error) {
	encoded, err := json.Marshal(ci.value.value)
	if err != nil {
		return nil, fmt.Errorf("marshal client info: %w", err)
	}

	return encoded, nil
}

// Value returns the underlying string value.
func (ci ClientInfo) Value() string {
	return ci.value.Value()
}

// AuthCode is an issued OAuth2 authorization code.
type AuthCode struct {
	value NonEmptyString
}

// NewAuthCode creates an AuthCode from a raw string, returning an error if empty.
func NewAuthCode(raw string) (AuthCode, error) {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		return AuthCode{}, err
	}

	return AuthCode{value: v}, nil
}

// MustAuthCode creates an AuthCode from a raw string, panicking if invalid.
func MustAuthCode(raw string) AuthCode {
	v, err := NewAuthCode(raw)
	if err != nil {
		panic(err)
	}

	return v
}

// Value returns the underlying string value.
func (ac AuthCode) Value() string {
	return ac.value.Value()
}

// AddTo sets the authorization code in the provided url.Values under the "code" key.
func (ac AuthCode) AddTo(values url.Values) {
	values.Add("code", ac.value.value)
}

// JoinRoleValues joins multiple role values into a single string using the provided separator.
func JoinRoleValues(roles []RoleValue, sep string) string {
	ss := make([]string, len(roles))
	for i, r := range roles {
		ss[i] = r.value.value
	}

	return strings.Join(ss, sep)
}

// JoinScopeValues joins multiple scope values into a space-separated string.
func JoinScopeValues(scopes []ScopeValue) string {
	ss := make([]string, len(scopes))
	for i, s := range scopes {
		ss[i] = s.value.value
	}

	return strings.Join(ss, " ")
}

// AsSubject returns the UserID as a Subject claim value.
func (userID UserID) AsSubject() Subject {
	return Subject{value: MustNonEmptyString(userID.value.String())}
}

// AsSubject returns the ClientID as a Subject claim value.
func (clientID ClientID) AsSubject() Subject {
	return Subject{value: MustNonEmptyString(clientID.value.String())}
}
