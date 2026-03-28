package app

import (
	"crypto/rsa"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"identity/domain"
)

const (
	accessTokenDuration  = 24 * time.Hour
	refreshTokenDuration = 90 * 24 * time.Hour
	authCodeDuration     = 5 * time.Minute
	graphClientID        = "00000003-0000-0000-c000-000000000000"
	graphAudience        = "https://graph.microsoft.com"
)

// RawTokenRequest holds the parsed-but-not-yet-validated inputs from an HTTP token request.
type RawTokenRequest struct {
	GrantType     string
	ClientID      string
	ClientSecret  string
	Username      string
	Password      string
	Scope         string
	Code          string
	RedirectURI   string
	RefreshToken  string
	Nonce         string
	TenantID      string
	IsV2          bool
	CorrelationID string
	BaseURL       string
}

// TokenResponse holds the issued tokens and metadata.
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

// AuthCodeClaims holds the validated claims extracted from an authorization code.
type AuthCodeClaims struct {
	Subject     string
	ClientID    string
	RedirectURI string
	Scope       string
	Nonce       string
}

// RefreshTokenClaims holds the validated claims extracted from a refresh token.
type RefreshTokenClaims struct {
	Subject  string
	ClientID string
	Scope    string
}

// IssueToken resolves tenant/client/user from config and issues a signed access token.
func IssueToken(config domain.Config, key *rsa.PrivateKey, raw RawTokenRequest) (TokenResponse, error) {
	tenant, err := FindTenant(config, raw.TenantID)
	if err != nil {
		return TokenResponse{}, err
	}

	switch raw.GrantType {
	case "password", "test":
		return issuePasswordGrant(tenant, key, raw)
	case "client_credentials":
		return issueClientCredentials(tenant, key, raw)
	case "authorization_code":
		return issueAuthorizationCode(tenant, key, raw)
	case "refresh_token":
		return issueRefreshToken(tenant, key, raw)
	default:
		return TokenResponse{}, domain.ErrUnsupportedGrantType
	}
}

// IssueAuthCode issues a short-lived signed JWT used as an authorization code.
func IssueAuthCode(key *rsa.PrivateKey, user domain.User, clientID domain.ClientID, redirectURI, scope, tenantID, nonce string) string {
	claims := jwt.MapClaims{
		"sub":          user.ID().String(),
		"client_id":    clientID.String(),
		"redirect_uri": redirectURI,
		"scope":        scope,
		"tenant":       tenantID,
		"nonce":        nonce,
		"exp":          time.Now().Add(authCodeDuration).Unix(),
	}

	return SignClaims(key, claims)
}

// ParseAuthCode validates and extracts claims from a signed authorization code JWT.
func ParseAuthCode(key *rsa.PrivateKey, code string) (AuthCodeClaims, error) {
	claims, err := ParseSignedToken(key, code)
	if err != nil {
		return AuthCodeClaims{}, err
	}

	return AuthCodeClaims{
		Subject:     stringClaim(claims, "sub"),
		ClientID:    stringClaim(claims, "client_id"),
		RedirectURI: stringClaim(claims, "redirect_uri"),
		Scope:       stringClaim(claims, "scope"),
		Nonce:       stringClaim(claims, "nonce"),
	}, nil
}

// ParseRefreshToken validates and extracts claims from a signed refresh token JWT.
func ParseRefreshToken(key *rsa.PrivateKey, tokenString string) (RefreshTokenClaims, error) {
	claims, err := ParseSignedToken(key, tokenString)
	if err != nil {
		return RefreshTokenClaims{}, err
	}

	return RefreshTokenClaims{
		Subject:  stringClaim(claims, "sub"),
		ClientID: stringClaim(claims, "client_id"),
		Scope:    stringClaim(claims, "scope"),
	}, nil
}

func issuePasswordGrant(tenant domain.Tenant, key *rsa.PrivateKey, raw RawTokenRequest) (TokenResponse, error) {
	user, err := AuthenticateUser(tenant, raw.Username, raw.Password)
	if err != nil {
		return TokenResponse{}, err
	}

	var activeClient *domain.Client

	if raw.ClientID != "" {
		clientID, parseErr := parseClientIDString(raw.ClientID)
		if parseErr == nil {
			client, findErr := FindClient(tenant, clientID)
			if findErr == nil {
				if !client.ClientSecret().IsEmpty() && raw.ClientSecret != "" {
					validateErr := ValidateClientSecret(client, raw.ClientSecret)
					if validateErr != nil {
						return TokenResponse{}, validateErr
					}
				}

				activeClient = &client
			}
		}
	}

	return buildTokenResponse(tenant, key, raw, &user, activeClient), nil
}

func issueClientCredentials(tenant domain.Tenant, key *rsa.PrivateKey, raw RawTokenRequest) (TokenResponse, error) {
	clientID, err := parseClientIDString(raw.ClientID)
	if err != nil {
		return TokenResponse{}, domain.ErrClientNotFound
	}

	client, err := FindClient(tenant, clientID)
	if err != nil {
		return TokenResponse{}, err
	}

	if err = ValidateClientSecret(client, raw.ClientSecret); err != nil {
		return TokenResponse{}, err
	}

	return buildTokenResponse(tenant, key, raw, nil, &client), nil
}

func issueAuthorizationCode(tenant domain.Tenant, key *rsa.PrivateKey, raw RawTokenRequest) (TokenResponse, error) {
	codeClaims, err := ParseAuthCode(key, raw.Code)
	if err != nil {
		return TokenResponse{}, domain.ErrInvalidCredentials
	}

	if raw.RedirectURI != "" && raw.RedirectURI != codeClaims.RedirectURI {
		return TokenResponse{}, domain.ErrInvalidRedirectURI
	}

	effectiveScope := raw.Scope
	if effectiveScope == "" {
		effectiveScope = codeClaims.Scope
	}

	effectiveClientID := raw.ClientID
	if effectiveClientID == "" {
		effectiveClientID = codeClaims.ClientID
	}

	enriched := raw
	enriched.Scope = effectiveScope
	enriched.ClientID = effectiveClientID
	enriched.Nonce = codeClaims.Nonce

	var activeUser *domain.User

	if user, found := FindUserByID(tenant, codeClaims.Subject); found {
		activeUser = &user
	}

	var activeClient *domain.Client

	if clientID, parseErr := parseClientIDString(effectiveClientID); parseErr == nil {
		if client, findErr := FindClient(tenant, clientID); findErr == nil {
			activeClient = &client
		}
	}

	return buildTokenResponse(tenant, key, enriched, activeUser, activeClient), nil
}

func issueRefreshToken(tenant domain.Tenant, key *rsa.PrivateKey, raw RawTokenRequest) (TokenResponse, error) {
	refreshClaims, err := ParseRefreshToken(key, raw.RefreshToken)
	if err != nil {
		return TokenResponse{}, domain.ErrInvalidCredentials
	}

	effectiveScope := raw.Scope
	if effectiveScope == "" {
		effectiveScope = refreshClaims.Scope
	}

	effectiveClientID := raw.ClientID
	if effectiveClientID == "" {
		effectiveClientID = refreshClaims.ClientID
	}

	enriched := raw
	enriched.Scope = effectiveScope
	enriched.ClientID = effectiveClientID

	var activeUser *domain.User

	if user, found := FindUserByID(tenant, refreshClaims.Subject); found {
		activeUser = &user
	}

	var activeClient *domain.Client

	if clientID, parseErr := parseClientIDString(effectiveClientID); parseErr == nil {
		if client, findErr := FindClient(tenant, clientID); findErr == nil {
			activeClient = &client
		}
	}

	return buildTokenResponse(tenant, key, enriched, activeUser, activeClient), nil
}

func buildTokenResponse(tenant domain.Tenant, key *rsa.PrivateKey, raw RawTokenRequest, user *domain.User, client *domain.Client) TokenResponse {
	issuer, version := buildIssuer(raw.BaseURL, tenant.TenantID().String(), raw.IsV2)
	tenantID := tenant.TenantID().String()

	subject := resolveSubject(raw, user, client)
	displayName, email := resolveUserInfo(user)

	targetAudience, targetAppIDs := resolveAudience(tenant, raw.Scope)
	roles := resolveRoles(tenant, client, user, targetAppIDs, strings.Split(raw.Scope, " "))

	now := time.Now()

	claims := buildAccessClaims(issuer, subject, targetAudience, tenantID, version, raw, user, client, email, displayName, roles, now)

	accessToken := SignClaims(key, claims)

	response := TokenResponse{
		AccessToken:   accessToken,
		TokenType:     "Bearer",
		ExpiresIn:     3600,
		Scope:         raw.Scope,
		IDToken:       nil,
		RefreshToken:  nil,
		ClientInfo:    nil,
		CorrelationID: raw.CorrelationID,
	}

	requestedScopes := strings.Split(raw.Scope, " ")

	if containsScope(requestedScopes, "openid") && user != nil {
		idToken := SignClaims(key, buildIDClaims(issuer, subject, raw.ClientID, tenantID, version, raw.Nonce, displayName, email, now))
		response.IDToken = &idToken
	}

	if containsScope(requestedScopes, "offline_access") || raw.GrantType == "refresh_token" {
		refreshToken := SignClaims(key, buildRefreshClaims(issuer, subject, raw.ClientID, tenantID, raw.Scope, now))
		response.RefreshToken = &refreshToken
	}

	if user != nil {
		clientInfo := BuildClientInfo(user.ID().String(), tenantID)
		response.ClientInfo = &clientInfo
	}

	return response
}

func buildAccessClaims(issuer, subject, audience, tenantID, version string, raw RawTokenRequest, user *domain.User, client *domain.Client, email, displayName string, roles []string, now time.Time) jwt.MapClaims {
	claims := jwt.MapClaims{
		"iss":   issuer,
		"sub":   subject,
		"aud":   audience,
		"exp":   now.Add(accessTokenDuration).Unix(),
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"jti":   uuid.New().String(),
		"tid":   tenantID,
		"ver":   version,
		"oid":   subject,
		"roles": roles,
	}

	if client != nil {
		claims["azp"] = client.ClientID().String()
		claims["azpacls"] = "0"
		claims["appid"] = client.ClientID().String()
	}

	if user != nil {
		claims["name"] = displayName
		claims["preferred_username"] = email
		claims["email"] = email
		claims["unique_name"] = email

		if audience == graphClientID || audience == graphAudience {
			claims["scp"] = strings.Join(roles, " ")
		} else {
			claims["scp"] = raw.Scope
		}
	}

	if raw.Nonce != "" {
		claims["nonce"] = raw.Nonce
	}

	return claims
}

func buildIDClaims(issuer, subject, clientID, tenantID, version, nonce, displayName, email string, now time.Time) jwt.MapClaims {
	claims := jwt.MapClaims{
		"iss":                issuer,
		"sub":                subject,
		"aud":                clientID,
		"exp":                now.Add(accessTokenDuration).Unix(),
		"iat":                now.Unix(),
		"nbf":                now.Unix(),
		"tid":                tenantID,
		"ver":                version,
		"oid":                subject,
		"name":               displayName,
		"preferred_username": email,
		"email":              email,
	}

	if nonce != "" {
		claims["nonce"] = nonce
	}

	return claims
}

func buildRefreshClaims(issuer, subject, clientID, tenantID, scope string, now time.Time) jwt.MapClaims {
	return jwt.MapClaims{
		"iss":       issuer,
		"sub":       subject,
		"aud":       issuer,
		"exp":       now.Add(refreshTokenDuration).Unix(),
		"iat":       now.Unix(),
		"client_id": clientID,
		"scope":     scope,
		"tid":       tenantID,
		"typ":       "Refresh",
	}
}

func resolveSubject(raw RawTokenRequest, user *domain.User, client *domain.Client) string {
	if user != nil {
		return user.ID().String()
	}

	if client != nil {
		return client.ClientID().String()
	}

	return raw.ClientID
}

func resolveUserInfo(user *domain.User) (displayName, email string) {
	if user != nil {
		return user.DisplayName().String(), user.Email().String()
	}

	return "Mock User", "user@example.com"
}

func buildIssuer(baseURL, tenantID string, isV2 bool) (string, string) {
	if isV2 {
		return baseURL + "/" + tenantID + "/v2.0", "2.0"
	}

	return baseURL + "/" + tenantID, "1.0"
}

func containsScope(scopes []string, target string) bool {
	return slices.Contains(scopes, target)
}

func resolveAudience(tenant domain.Tenant, scope string) (string, map[string]bool) {
	targetAudience := "api://default"
	targetAppIDs := make(map[string]bool)

	for scopePart := range strings.SplitSeq(scope, " ") {
		if isOIDCScope(scopePart) {
			continue
		}

		for _, registration := range tenant.AppRegistrations() {
			identifierURI := registration.IdentifierURI().String()
			clientIDStr := registration.ClientID().String()

			if clientIDStr == scopePart ||
				identifierURI == scopePart ||
				strings.HasPrefix(scopePart, identifierURI+"/") ||
				scopePart == identifierURI+"/.default" {
				targetAppIDs[clientIDStr] = true
				targetAudience = identifierURI
			}

			for _, role := range registration.AppRoles() {
				for _, roleScope := range role.Scopes() {
					roleScopeValue := roleScope.Value().String()

					if roleScopeValue == scopePart || strings.HasSuffix(scopePart, "/"+roleScopeValue) {
						targetAppIDs[clientIDStr] = true
						targetAudience = identifierURI
					}
				}
			}
		}
	}

	return targetAudience, targetAppIDs
}

func resolveRoles(tenant domain.Tenant, client *domain.Client, user *domain.User, targetAppIDs map[string]bool, requestedScopes []string) []string {
	resolved := make(map[string]bool)

	userGroups := make(map[string]bool)

	if user != nil {
		for _, groupName := range user.Groups() {
			userGroups[groupName.String()] = true
		}
	}

	if client != nil {
		for _, assignment := range client.GroupRoleAssignments() {
			if user != nil && !userGroups[assignment.GroupName().String()] {
				continue
			}

			appIDStr := assignment.ApplicationID().String()
			if appIDStr != uuid.Nil.String() && !targetAppIDs[appIDStr] {
				continue
			}

			for _, roleValue := range assignment.Roles() {
				resolved[roleValue.String()] = true
			}
		}
	}

	for _, registration := range tenant.AppRegistrations() {
		if !targetAppIDs[registration.ClientID().String()] {
			continue
		}

		for _, role := range registration.AppRoles() {
			for _, requestedScope := range requestedScopes {
				if isOIDCScope(requestedScope) {
					continue
				}

				for _, roleScope := range role.Scopes() {
					if roleScope.Value().String() == requestedScope || strings.HasSuffix(requestedScope, "/"+roleScope.Value().String()) {
						resolved[role.Value().String()] = true
					}
				}
			}
		}
	}

	roles := make([]string, 0, len(resolved))
	for role := range resolved {
		roles = append(roles, role)
	}

	return roles
}

func isOIDCScope(scope string) bool {
	switch scope {
	case "openid", "profile", "offline_access", "email":
		return true
	default:
		return false
	}
}

func stringClaim(claims jwt.MapClaims, key string) string {
	value, _ := claims[key].(string)

	return value
}

func parseClientIDString(raw string) (domain.ClientID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return domain.ClientID{}, domain.ErrClientNotFound
	}

	return domain.ClientIDFromUUID(parsed), nil
}

// ResolveAudienceForTest exposes resolveAudience for testing.
func ResolveAudienceForTest(tenant domain.Tenant, scope string) (string, map[string]bool) {
	return resolveAudience(tenant, scope)
}

// ResolveRolesForTest exposes resolveRoles for testing.
func ResolveRolesForTest(tenant domain.Tenant, client *domain.Client, user *domain.User, targetAppIDs map[string]bool, requestedScopes []string) []string {
	return resolveRoles(tenant, client, user, targetAppIDs, requestedScopes)
}
