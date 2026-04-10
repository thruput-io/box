package domain

import "github.com/samber/mo"

// GrantType is a validated OAuth2 grant type.
type GrantType string

const (
	// GrantPassword is the Resource Owner Password Credentials grant.
	GrantPassword GrantType = "password"

	// GrantClientCredentials is the Client Credentials grant.
	GrantClientCredentials GrantType = "client_credentials"

	// GrantAuthorizationCode is the Authorization Code grant.
	GrantAuthorizationCode GrantType = "authorization_code"

	// GrantRefreshToken is the Refresh Token grant.
	GrantRefreshToken GrantType = "refresh_token"

	// GrantTest is a dev-only grant for issuing test tokens directly.
	GrantTest GrantType = "test"
)

// TokenInput holds all pre-resolved, pre-validated inputs for token issuance.
// By the time this struct is constructed, all domain lookups and authentication
// have already succeeded. IssueToken receives this and cannot fail.
type TokenInput struct {
	Grant         GrantType
	Tenant        *Tenant
	User          *User
	Client        *Client
	Scope         []ScopeValue
	Nonce         mo.Option[Nonce]
	IsV2          bool
	BaseURL       BaseURL
	CorrelationID mo.Option[CorrelationID]
}

type (
	// RequestedClaim is a list of claim values.
	RequestedClaim []NonEmptyString
	// RequestedClaims is a list of requested claims.
	RequestedClaims []RequestedClaim
	// TestTokenInput is a container for testing token issuance with specific claims.
	TestTokenInput struct {
		RequestedClaims RequestedClaims
		TokenInput      TokenInput
	}
)

// TokenResponse holds the issued tokens returned by IssueToken.
type TokenResponse struct {
	AccessToken   AccessToken
	TokenType     TokenType
	ExpiresIn     int
	Scope         []ScopeValue
	IDToken       mo.Option[IDToken]
	RefreshToken  mo.Option[RefreshToken]
	ClientInfo    mo.Option[ClientInfo]
	CorrelationID mo.Option[CorrelationID]
}
