package handlers

import (
	"crypto/rsa"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/mo"

	"identity/app"
	"identity/domain"
)

const (
	errMsgInvalidAuthCode     = "invalid or expired authorization code"
	errMsgInvalidRefreshToken = "invalid or expired refresh token"
)

func parseAuthCode(key *rsa.PrivateKey, code string) mo.Either[domain.Error, domain.Claims] {
	rawClaims, err := app.ParseSignedToken(key, code)
	if err != nil {
		return mo.Left[domain.Error, domain.Claims](
			domain.NewError(domain.ErrCodeInvalidGrant, errMsgInvalidAuthCode),
		)
	}

	domainClaims, ok := parseDomainClaims(rawClaims).Right()
	if !ok {
		return mo.Left[domain.Error, domain.Claims](
			domain.NewError(domain.ErrCodeInvalidGrant, errMsgInvalidAuthCode),
		)
	}

	return mo.Right[domain.Error, domain.Claims](domainClaims)
}

func parseRefreshToken(key *rsa.PrivateKey, tokenString string) mo.Either[domain.Error, domain.Claims] {
	claims, err := app.ParseSignedToken(key, tokenString)
	if err != nil {
		return mo.Left[domain.Error, domain.Claims](
			domain.NewError(domain.ErrCodeInvalidGrant, errMsgInvalidRefreshToken),
		)
	}

	domainClaims, ok := parseDomainClaims(claims).Right()
	if !ok {
		return mo.Left[domain.Error, domain.Claims](
			domain.NewError(domain.ErrCodeInvalidGrant, errMsgInvalidRefreshToken),
		)
	}

	return mo.Right[domain.Error, domain.Claims](domainClaims)
}

func parseDomainClaims(rawClaims jwt.MapClaims) mo.Either[domain.Error, domain.Claims] {
	return domain.From(rawClaims)
}

// parseScopeValues splits a space-separated scope string into a slice of ScopeValues.
// Invalid or empty parts are silently skipped since scope strings come from HTTP query params.
func parseScopeValues(raw string) []domain.ScopeValue {
	if raw == emptyValue {
		return nil
	}

	parts := strings.Split(raw, " ")
	result := make([]domain.ScopeValue, 0, len(parts))

	for _, part := range parts {
		if sv, ok := domain.NewScopeValue(part).Right(); ok {
			result = append(result, sv)
		}
	}

	return result
}

// parseOptionalNonce wraps a raw nonce string in mo.Option[Nonce].
func parseOptionalNonce(raw string) mo.Option[domain.Nonce] {
	n, ok := domain.NewNonce(raw).Right()
	if !ok {
		return mo.None[domain.Nonce]()
	}

	return mo.Some(n)
}

// parseOptionalOAuthState wraps a raw state string in mo.Option[OAuthState].
func parseOptionalOAuthState(raw string) mo.Option[domain.OAuthState] {
	s, ok := domain.NewOAuthState(raw).Right()
	if !ok {
		return mo.None[domain.OAuthState]()
	}

	return mo.Some(s)
}

// parseOptionalResponseMode wraps a raw response_mode string in mo.Option[ResponseMode].
func parseOptionalResponseMode(raw string) mo.Option[domain.ResponseMode] {
	rm, ok := domain.NewResponseMode(raw).Right()
	if !ok {
		return mo.None[domain.ResponseMode]()
	}

	return mo.Some(rm)
}

// parseOptionalResponseType wraps a raw response_type string in mo.Option[ResponseType].
func parseOptionalResponseType(raw string) mo.Option[domain.ResponseType] {
	rt, ok := domain.NewResponseType(raw).Right()
	if !ok {
		return mo.None[domain.ResponseType]()
	}

	return mo.Some(rt)
}
