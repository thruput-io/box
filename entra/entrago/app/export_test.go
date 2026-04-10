package app

import (
	"crypto/rsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/mo"

	"identity/domain"
)

// ExportFindTenant is for testing FindTenant from app_test.
func ExportFindTenant(config *domain.Config, tenantID string) (*domain.Tenant, error) {
	return FindTenant(config, tenantID)
}

// ExportFindTenantByID is for testing FindTenantByID from app_test.
func ExportFindTenantByID(config *domain.Config, tenantID domain.TenantID) (*domain.Tenant, error) {
	return FindTenantByID(config, tenantID)
}

// ExportFindClient is for testing FindClient from app_test.
func ExportFindClient(tenant *domain.Tenant, clientID domain.ClientID) (*domain.Client, error) {
	return FindClient(*tenant, clientID)
}

// ExportFindAppRegistration is for testing FindAppRegistration from app_test.
func ExportFindAppRegistration(tenant *domain.Tenant, clientID domain.ClientID) (*domain.AppRegistration, error) {
	return FindAppRegistration(*tenant, clientID)
}

// ExportFindRedirectURLs is for testing FindRedirectURLs from app_test.
func ExportFindRedirectURLs(tenant *domain.Tenant, clientID domain.ClientID) ([]domain.RedirectURL, error) {
	return FindRedirectURLs(*tenant, clientID)
}

// ExportValidateRedirectURI is for testing ValidateRedirectURI from app_test.
func ExportValidateRedirectURI(redirectURIStr string, allowed []domain.RedirectURL) error {
	redirectURI, err := domain.NewRedirectURL(redirectURIStr)
	if err != nil {
		return fmt.Errorf("invalid redirect URI: %w", err)
	}

	return ValidateRedirectURI(redirectURI, allowed)
}

// ExportAuthenticateUser is for testing AuthenticateUser from app_test.
func ExportAuthenticateUser(tenant *domain.Tenant, usernameStr, passwordStr string) (*domain.User, error) {
	username, _ := domain.NewUsername(usernameStr)
	password, _ := domain.NewPassword(passwordStr)

	return AuthenticateUser(*tenant, username, password)
}

func ExportFindUserByID(tenant *domain.Tenant, id domain.UserID) (*domain.User, bool) {
	return FindUserByID(*tenant, id)
}

// ExportValidateClientSecret is for testing ValidateClientSecret from app_test.
func ExportValidateClientSecret(client *domain.Client, secretStr string) error {
	var secret *domain.ClientSecret

	if secretStr != "" {
		s, err := domain.NewClientSecret(secretStr)
		if err != nil {
			return fmt.Errorf("invalid client secret: %w", err)
		}

		secret = &s
	}

	return ValidateClientSecret(*client, secret)
}

// ExportIssueAuthCode is for testing IssueAuthCode from app_test.
func ExportIssueAuthCode(
	key *rsa.PrivateKey,
	user *domain.User,
	clientID domain.ClientID,
	redirectURI domain.RedirectURL,
	scope []domain.ScopeValue,
	tenantID domain.TenantID,
	nonce mo.Option[domain.Nonce],
) domain.AuthCode {
	return IssueAuthCode(key, *user, clientID, redirectURI, scope, tenantID, nonce)
}

// ExportParseSignedToken is for testing ParseSignedToken from app_test.
func ExportParseSignedToken(key *rsa.PrivateKey, tokenString string) (jwt.MapClaims, error) {
	return ParseSignedToken(key, tokenString)
}

// ExportResolveAudienceForTest is for testing ResolveAudience from app_test.
func ExportResolveAudienceForTest(
	tenant *domain.Tenant,
	scope []domain.ScopeValue,
) (domain.IdentifierURI, map[domain.ClientID]bool) {
	return ResolveAudienceForTest(tenant, scope)
}

// ExportResolveRolesForTest is for testing ResolveRoles from app_test.
func ExportResolveRolesForTest(
	tenant *domain.Tenant,
	client *domain.Client,
	user *domain.User,
	targetAppIDs map[domain.ClientID]bool,
	requestedScopes []domain.ScopeValue,
) []domain.RoleValue {
	return ResolveRolesForTest(tenant, client, user, targetAppIDs, requestedScopes)
}

// ExportSignClaims is for testing SignClaims from app_test.
func ExportSignClaims(key *rsa.PrivateKey, claims jwt.MapClaims) string {
	return SignClaims(key, claims)
}

// IsOIDCScopeForTest is for testing isOIDCScope from app_test.
func IsOIDCScopeForTest(scope domain.ScopeValue) bool {
	return isOIDCScope(scope)
}

// ClaimSub exports claimSub constant for testing.
const ClaimSub = claimSub

// ClaimClientID exports claimClientID constant for testing.
const ClaimClientID = claimClientID

// ClaimRedirectURI exports claimRedirectURI constant for testing.
const ClaimRedirectURI = claimRedirectURI

// ClaimScope exports claimScope constant for testing.
const ClaimScope = claimScope

// ClaimTenant exports claimTenant constant for testing.
const ClaimTenant = claimTenant

// ClaimNonce exports claimNonce constant for testing.
const ClaimNonce = claimNonce

// ClaimExp exports claimExp constant for testing.
const ClaimExp = claimExp
