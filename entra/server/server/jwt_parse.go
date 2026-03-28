package server

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"

	"identity/app"
	"identity/domain"
)

// authCodeClaims holds the validated claims extracted from an authorization code JWT.
type authCodeClaims struct {
	subject     string
	clientID    string
	redirectURI string
	scope       string
	nonce       string
}

// refreshTokenClaims holds the validated claims extracted from a refresh token JWT.
type refreshTokenClaims struct {
	subject  string
	clientID string
	scope    string
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
		scope:       claimStr(claims, "scope"),
		nonce:       claimStr(claims, "nonce"),
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
		scope:    claimStr(claims, "scope"),
	}, nil
}

func claimStr(claims jwt.MapClaims, key string) string {
	value, ok := claims[key].(string)
	if !ok {
		return emptyValue
	}

	return value
}
