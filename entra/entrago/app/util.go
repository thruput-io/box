package app

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"identity/domain"
)

var errInvalidClaimsType = errors.New("invalid claims type in JWT")

// MaxBodyBytes is the maximum allowed request body size.
const MaxBodyBytes = 1024 * 1024

// MaxParamLength is the maximum allowed length for any single parameter value.
const MaxParamLength = 2048

// JWKSKey holds the RSA public key components for the JWKS endpoint.
type JWKSKey struct {
	N string
	E string
}

// PublicKey returns the JWKS key components for the given RSA private key.
func PublicKey(key *rsa.PrivateKey) JWKSKey {
	return JWKSKey{
		N: base64URLEncode(key.N.Bytes()),
		E: base64URLEncode(big.NewInt(int64(key.E)).Bytes()),
	}
}

// BuildClientInfo encodes uid+utid as a base64url JSON blob for MSAL.
func BuildClientInfo(userID domain.UserID, tenantID domain.TenantID) domain.ClientInfo {
	info := map[string]any{
		"uid":  userID.Value(),
		"utid": tenantID.Value(),
	}

	data, err := json.Marshal(info)
	if err != nil {
		panic("failed to marshal client_info: " + err.Error())
	}

	return domain.MustClientInfo(base64URLEncode(data))
}

func roleValuesToStrings(vals []domain.RoleValue) []string {
	res := make([]string, len(vals))
	for i, v := range vals {
		res[i] = v.Value()
	}

	return res
}

// SignClaims signs a domain.Claims as a JWT with RS256 and kid=1.
func SignClaims(key *rsa.PrivateKey, claims domain.Claims) string {
	return SignMapClaims(key, claims.ToJWT())
}

// SignMapClaims signs a raw jwt.MapClaims as a JWT with RS256 and kid=1.
// Use this only at boundaries where raw claim maps are unavoidable (e.g. test token overrides).
func SignMapClaims(key *rsa.PrivateKey, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "1"

	signed, err := token.SignedString(key)
	if err != nil {
		panic("failed to sign JWT: " + err.Error())
	}

	return signed
}

// ParseSignedToken validates a JWT signed by key and returns its claims.
func ParseSignedToken(key *rsa.PrivateKey, tokenString string) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(tokenString, func(_ *jwt.Token) (any, error) {
		return key.Public(), nil
	})
	if err != nil || !parsed.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errInvalidClaimsType
	}

	return claims, nil
}

func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}
