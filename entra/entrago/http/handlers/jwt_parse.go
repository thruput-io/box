package handlers

import (
	"crypto/rsa"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/mo"

	"identity/app"
	"identity/domain"
)

// authCodeClaims holds the validated claims extracted from an authorization code JWT.
type authCodeClaims struct {
	subject     string
	clientID    string
	redirectURI string
	scope       []domain.ScopeValue
	nonce       mo.Option[domain.Nonce]
}

// refreshTokenClaims holds the validated claims extracted from a refresh token JWT.
type refreshTokenClaims struct {
	subject  string
	clientID string
	scope    []domain.ScopeValue
}

func parseAuthCode(key *rsa.PrivateKey, code string) (authCodeClaims, *domain.Error) {
	claims, err := app.ParseSignedToken(key, code)
	if err != nil {
		return authCodeClaims{}, domain.NewError(domain.ErrCodeInvalidGrant, "invalid or expired authorization code")
	}

	return authCodeClaims{
		subject:     claimStr(claims, "sub"),
		clientID:    claimStr(claims, "client_id"),
		redirectURI: claimStr(claims, "redirect_uri"),
		scope:       parseScopeValues(claimStr(claims, "scope")),
		nonce:       parseOptionalNonce(claimStr(claims, "nonce")),
	}, nil
}

func parseRefreshToken(key *rsa.PrivateKey, tokenString string) (refreshTokenClaims, *domain.Error) {
	claims, err := app.ParseSignedToken(key, tokenString)
	if err != nil {
		return refreshTokenClaims{}, domain.NewError(domain.ErrCodeInvalidGrant, "invalid or expired refresh token")
	}

	return refreshTokenClaims{
		subject:  claimStr(claims, "sub"),
		clientID: claimStr(claims, "client_id"),
		scope:    parseScopeValues(claimStr(claims, "scope")),
	}, nil
}

func claimStr(claims jwt.MapClaims, key string) string {
	value, ok := claims[key].(string)
	if !ok {
		return emptyValue
	}

	return value
}

// parseScopeValues splits a space-separated scope string into validated ScopeValue slice.
// Invalid or empty tokens are silently dropped — scope parsing is best-effort at the JWT boundary.
func parseScopeValues(raw string) []domain.ScopeValue {
	parts := strings.Fields(raw)
	values := make([]domain.ScopeValue, emptySliceSize, len(parts))

	for _, part := range parts {
		sv, err := domain.NewScopeValue(part)
		if err != nil {
			continue
		}

		values = append(values, sv)
	}

	return values
}

// parseOptionalNonce wraps a raw nonce string in mo.Option[Nonce].
// Returns None if the raw string is empty or invalid.
func parseOptionalNonce(raw string) mo.Option[domain.Nonce] {
	n, err := domain.NewNonce(raw)
	if err != nil {
		return mo.None[domain.Nonce]()
	}

	return mo.Some(n)
}

// parseOptionalOAuthState wraps a raw state string in mo.Option[OAuthState].
func parseOptionalOAuthState(raw string) mo.Option[domain.OAuthState] {
	s, err := domain.NewOAuthState(raw)
	if err != nil {
		return mo.None[domain.OAuthState]()
	}

	return mo.Some(s)
}

// parseOptionalResponseMode wraps a raw response_mode string in mo.Option[ResponseMode].
func parseOptionalResponseMode(raw string) mo.Option[domain.ResponseMode] {
	rm, err := domain.NewResponseMode(raw)
	if err != nil {
		return mo.None[domain.ResponseMode]()
	}

	return mo.Some(rm)
}

// parseOptionalResponseType wraps a raw response_type string in mo.Option[ResponseType].
func parseOptionalResponseType(raw string) mo.Option[domain.ResponseType] {
	rt, err := domain.NewResponseType(raw)
	if err != nil {
		return mo.None[domain.ResponseType]()
	}

	return mo.Some(rt)
}
