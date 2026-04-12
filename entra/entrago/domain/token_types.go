package domain

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/samber/mo"
	moeither "github.com/samber/mo/either"
)

// AccessToken is an issued OAuth2 access token.
type AccessToken struct {
	value NonEmptyString
}

// NewAccessToken creates an AccessToken from a raw string, returning an error if empty.
func NewAccessToken(raw string) mo.Either[Error, AccessToken] {
	return moeither.MapRight[Error, NonEmptyString, AccessToken](func(nes NonEmptyString) AccessToken {
		return AccessToken{value: nes}
	})(NewNonEmptyString(raw))
}

// MustAccessToken creates an AccessToken from a raw string, panicking if invalid.
func MustAccessToken(raw string) AccessToken {
	return NewAccessToken(raw).MustRight()
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
func NewTokenType(raw string) mo.Either[Error, TokenType] {
	return moeither.MapRight[Error, NonEmptyString, TokenType](func(nes NonEmptyString) TokenType {
		return TokenType{value: nes}
	})(NewNonEmptyString(raw))
}

// MustTokenType creates a TokenType from a raw string, panicking if invalid.
func MustTokenType(raw string) TokenType {
	return NewTokenType(raw).MustRight()
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
func NewIDToken(raw string) mo.Either[Error, IDToken] {
	return moeither.MapRight[Error, NonEmptyString, IDToken](func(nes NonEmptyString) IDToken {
		return IDToken{value: nes}
	})(NewNonEmptyString(raw))
}

// MustIDToken creates an IDToken from a raw string, panicking if invalid.
func MustIDToken(raw string) IDToken {
	return NewIDToken(raw).MustRight()
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
func NewRefreshToken(raw string) mo.Either[Error, RefreshToken] {
	return moeither.MapRight[Error, NonEmptyString, RefreshToken](func(nes NonEmptyString) RefreshToken {
		return RefreshToken{value: nes}
	})(NewNonEmptyString(raw))
}

// MustRefreshToken creates a RefreshToken from a raw string, panicking if invalid.
func MustRefreshToken(raw string) RefreshToken {
	return NewRefreshToken(raw).MustRight()
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
func NewClientInfo(raw string) mo.Either[Error, ClientInfo] {
	return moeither.MapRight[Error, NonEmptyString, ClientInfo](func(nes NonEmptyString) ClientInfo {
		return ClientInfo{value: nes}
	})(NewNonEmptyString(raw))
}

// MustClientInfo creates a ClientInfo from a raw string, panicking if invalid.
func MustClientInfo(raw string) ClientInfo {
	return NewClientInfo(raw).MustRight()
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
func NewAuthCode(raw string) mo.Either[Error, AuthCode] {
	return moeither.MapRight[Error, NonEmptyString, AuthCode](func(nes NonEmptyString) AuthCode {
		return AuthCode{value: nes}
	})(NewNonEmptyString(raw))
}

// MustAuthCode creates an AuthCode from a raw string, panicking if invalid.
func MustAuthCode(raw string) AuthCode {
	return NewAuthCode(raw).MustRight()
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
	return Subject{value: NonEmptyString{value: userID.value.String()}}
}

// AsSubject returns the ClientID as a Subject claim value.
func (clientID ClientID) AsSubject() Subject {
	return Subject{value: NonEmptyString{value: clientID.value.String()}}
}
