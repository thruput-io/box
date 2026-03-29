package domain

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
	Tenant        Tenant
	User          *User
	Client        Client
	Scope         string
	Nonce         string
	IsV2          bool
	BaseURL       string
	CorrelationID string
}

// TokenResponse holds the issued tokens returned by IssueToken.
type TokenResponse struct {
	AccessToken   string
	TokenType     string
	ExpiresIn     int
	Scope         string
	IDToken       *string
	RefreshToken  *string
	ClientInfo    *string
	CorrelationID string
}
