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
	accessTokenExpiry    = 3600
	scopeSeparator       = " "
	graphClientID        = "00000003-0000-0000-c000-000000000000"
	graphAudience        = "https://graph.microsoft.com"
	emptyString          = ""
	emptySliceLen        = 0
	defaultDisplayName   = "Mock User"
	defaultEmail         = "user@example.com"
)

// tokenClaimsBase groups the common fields used to build JWT claim sets.
type tokenClaimsBase struct {
	issuer   string
	subject  any
	audience domain.IdentifierURI
	tenantID domain.TenantID
	version  string
	now      time.Time
}

// IssueToken issues a signed access token from pre-validated domain inputs.
// This function is pure: inputs are fully validated before calling it, so it cannot fail.
func IssueToken(key *rsa.PrivateKey, input domain.TokenInput) domain.TokenResponse {
	issuer, version := buildIssuerForInput(input)
	tenantID := input.Tenant.TenantID()
	subject := resolveSubject(input)
	displayName, email := resolveUserInfo(input.User)
	targetAudience, targetAppIDs := resolveAudience(input.Tenant, input.Scope)

	requestedScopes := strings.Split(input.Scope, scopeSeparator)
	roles := resolveRoles(
		input.Tenant, input.Client, input.User,
		targetAppIDs, requestedScopes,
	)

	now := time.Now()
	base := tokenClaimsBase{
		issuer:   issuer,
		subject:  subject,
		audience: targetAudience,
		tenantID: tenantID,
		version:  version,
		now:      now,
	}

	accessToken := domain.MustAccessToken(SignClaims(key, buildAccessClaims(base, input, email, displayName, roles)))

	response := domain.TokenResponse{
		AccessToken:   accessToken,
		TokenType:     domain.MustTokenType("Bearer"),
		ExpiresIn:     accessTokenExpiry,
		Scope:         input.Scope,
		IDToken:       nil,
		RefreshToken:  nil,
		ClientInfo:    nil,
		CorrelationID: input.CorrelationID,
	}

	if slices.Contains(requestedScopes, "openid") && input.User != nil {
		idToken := domain.MustIDToken(SignClaims(key, buildIDClaims(base, clientID(input), input.Nonce, displayName, email)))
		response.IDToken = &idToken
	}

	if slices.Contains(requestedScopes, "offline_access") || input.Grant == domain.GrantRefreshToken {
		refreshToken := domain.MustRefreshToken(SignClaims(
			key,
			buildRefreshClaims(issuer, subject, clientID(input), tenantID, input.Scope, now),
		))
		response.RefreshToken = &refreshToken
	}

	if input.User != nil {
		clientInfo := BuildClientInfo(input.User.ID(), input.Tenant.TenantID())
		response.ClientInfo = &clientInfo
	}

	return response
}

// IssueAuthCode issues a short-lived signed JWT used as an authorization code.
func IssueAuthCode(
	key *rsa.PrivateKey,
	user domain.User,
	clientID domain.ClientID,
	redirectURI domain.RedirectURL,
	scope string,
	tenantID domain.TenantID,
	nonce string,
) domain.AuthCode {
	claims := jwt.MapClaims{
		claimSub:         user.ID(),
		claimClientID:    clientID,
		claimRedirectURI: redirectURI,
		claimScope:       scope,
		claimTenant:      tenantID,
		claimNonce:       nonce,
		claimExp:         time.Now().Add(authCodeDuration).Unix(),
	}

	return domain.MustAuthCode(SignClaims(key, claims))
}

func buildAccessClaims(
	base tokenClaimsBase,
	input domain.TokenInput,
	email domain.Email,
	displayName domain.DisplayName,
	roles []domain.RoleValue,
) jwt.MapClaims {
	claims := jwt.MapClaims{
		claimIss:   base.issuer,
		claimSub:   base.subject,
		claimAud:   base.audience,
		claimExp:   base.now.Add(accessTokenDuration).Unix(),
		claimIat:   base.now.Unix(),
		claimNbf:   base.now.Unix(),
		claimJti:   uuid.New().String(),
		claimTid:   base.tenantID,
		claimVer:   base.version,
		claimOid:   base.subject,
		claimRoles: roles,
	}

	if input.Client != nil {
		claims[claimAzp] = input.Client.ClientID()
		claims[claimAzpacls] = claimAzpacls0
		claims[claimAppid] = input.Client.ClientID()
	}

	if input.User != nil {
		claims[claimName] = displayName
		claims[claimPreferredUsername] = email
		claims[claimEmail] = email
		claims[claimUniqueName] = email

		if base.audience.Matches(graphClientID) || base.audience.Matches(graphAudience) {
			claims[claimScp] = domain.JoinRoleValues(roles, scopeSeparator)
		} else {
			claims[claimScp] = input.Scope
		}
	}

	if input.Nonce != emptyString {
		claims[claimNonce] = input.Nonce
	}

	return claims
}

func buildIDClaims(base tokenClaimsBase, clientID domain.ClientID, nonce string, displayName domain.DisplayName, email domain.Email) jwt.MapClaims {
	claims := jwt.MapClaims{
		claimIss:               base.issuer,
		claimSub:               base.subject,
		claimAud:               clientID,
		claimExp:               base.now.Add(accessTokenDuration).Unix(),
		claimIat:               base.now.Unix(),
		claimNbf:               base.now.Unix(),
		claimTid:               base.tenantID,
		claimVer:               base.version,
		claimOid:               base.subject,
		claimName:              displayName,
		claimPreferredUsername: email,
		claimEmail:             email,
	}

	if nonce != emptyString {
		claims[claimNonce] = nonce
	}

	return claims
}

func buildRefreshClaims(issuer string, subject any, clientID domain.ClientID, tenantID domain.TenantID, scope string, now time.Time) jwt.MapClaims {
	return jwt.MapClaims{
		claimIss:      issuer,
		claimSub:      subject,
		claimAud:      issuer,
		claimExp:      now.Add(refreshTokenDuration).Unix(),
		claimIat:      now.Unix(),
		claimClientID: clientID,
		claimScope:    scope,
		claimTid:      tenantID,
		claimTyp:      claimRefreshTyp,
	}
}

func resolveSubject(input domain.TokenInput) any {
	if input.User != nil {
		return input.User.ID()
	}

	if input.Client != nil {
		return input.Client.ClientID()
	}

	return emptyString
}

func clientID(input domain.TokenInput) domain.ClientID {
	if input.Client != nil {
		return input.Client.ClientID()
	}

	return domain.ClientID{}
}

func resolveUserInfo(user *domain.User) (domain.DisplayName, domain.Email) {
	if user != nil {
		return user.DisplayName(), user.Email()
	}

	return domain.MustDisplayName(defaultDisplayName), domain.MustEmail(defaultEmail)
}

func buildIssuerForInput(input domain.TokenInput) (issuer, version string) {
	tenantID := input.Tenant.TenantID()

	if input.IsV2 {
		return tenantID.AsUrl(input.BaseURL) + "/v2.0", "2.0"
	}

	return tenantID.AsUrl(input.BaseURL), "1.0"
}

func resolveAudience(tenant domain.Tenant, scope string) (domain.IdentifierURI, map[domain.ClientID]bool) {
	targetAudience := domain.MustIdentifierURI("api://default")
	targetAppIDs := make(map[domain.ClientID]bool)

	for scopePart := range strings.SplitSeq(scope, scopeSeparator) {
		if isOIDCScope(scopePart) {
			continue
		}

		matchAudienceForScope(tenant, scopePart, targetAppIDs, &targetAudience)
	}

	return targetAudience, targetAppIDs
}

func matchAudienceForScope(
	tenant domain.Tenant,
	scopePart string,
	targetAppIDs map[domain.ClientID]bool,
	targetAudience *domain.IdentifierURI,
) {
	for _, registration := range tenant.AppRegistrations() {
		if registration.IsAudienceForScope(scopePart) {
			targetAppIDs[registration.ClientID()] = true
			*targetAudience = registration.IdentifierURI()
		}

		matchRoleScopesForScope(registration, scopePart, targetAppIDs, targetAudience)
	}
}

func matchRoleScopesForScope(
	registration domain.AppRegistration,
	scopePart string,
	targetAppIDs map[domain.ClientID]bool,
	targetAudience *domain.IdentifierURI,
) {
	for _, role := range registration.AppRoles() {
		if role.MatchesScope(scopePart) {
			targetAppIDs[registration.ClientID()] = true
			*targetAudience = registration.IdentifierURI()
		}
	}
}

func resolveRoles(
	tenant domain.Tenant,
	client domain.Client,
	user *domain.User,
	targetAppIDs map[domain.ClientID]bool,
	requestedScopes []string,
) []domain.RoleValue {
	resolved := make(map[domain.RoleValue]bool)

	resolveClientAssignmentRoles(client, user, targetAppIDs, resolved)
	resolveScopeMatchedRoles(tenant, targetAppIDs, requestedScopes, resolved)

	roles := make([]domain.RoleValue, emptySliceLen, len(resolved))
	for role := range resolved {
		roles = append(roles, role)
	}

	return roles
}

func resolveClientAssignmentRoles(
	client domain.Client,
	user *domain.User,
	targetAppIDs map[domain.ClientID]bool,
	resolved map[domain.RoleValue]bool,
) {
	if client == nil {
		return
	}

	userGroups := buildUserGroupSet(user)

	for _, assignment := range client.GroupRoleAssignments() {
		addAssignmentRoles(assignment, user, userGroups, targetAppIDs, resolved)
	}
}

func addAssignmentRoles(
	assignment domain.GroupRoleAssignment,
	user *domain.User,
	userGroups map[domain.GroupName]bool,
	targetAppIDs map[domain.ClientID]bool,
	resolved map[domain.RoleValue]bool,
) {
	if user != nil && !userGroups[assignment.GroupName()] {
		return
	}

	appID := assignment.ApplicationID()
	if appID.UUID() != uuid.Nil && !targetAppIDs[appID] {
		return
	}

	for _, roleValue := range assignment.Roles() {
		resolved[roleValue] = true
	}
}

func buildUserGroupSet(user *domain.User) map[domain.GroupName]bool {
	groups := make(map[domain.GroupName]bool)

	if user == nil {
		return groups
	}

	for _, groupName := range user.Groups() {
		groups[groupName] = true
	}

	return groups
}

func resolveScopeMatchedRoles(
	tenant domain.Tenant,
	targetAppIDs map[domain.ClientID]bool,
	requestedScopes []string,
	resolved map[domain.RoleValue]bool,
) {
	for _, registration := range tenant.AppRegistrations() {
		if !targetAppIDs[registration.ClientID()] {
			continue
		}

		matchRolesToScopes(registration, requestedScopes, resolved)
	}
}

func matchRolesToScopes(
	registration domain.AppRegistration,
	requestedScopes []string,
	resolved map[domain.RoleValue]bool,
) {
	for _, role := range registration.AppRoles() {
		if roleMatchesAnyScope(role, requestedScopes) {
			resolved[role.Value()] = true
		}
	}
}

func roleMatchesAnyScope(role domain.Role, requestedScopes []string) bool {
	for _, requestedScope := range requestedScopes {
		if !isOIDCScope(requestedScope) && roleScopesMatchScope(role.Scopes(), requestedScope) {
			return true
		}
	}

	return false
}

func roleScopesMatchScope(scopes []domain.Scope, requestedScope string) bool {
	for _, roleScope := range scopes {
		if roleScope.Value().Matches(requestedScope) {
			return true
		}
	}

	return false
}

func isOIDCScope(scope string) bool {
	switch scope {
	case "openid", "profile", "offline_access", "email":
		return true
	default:
		return false
	}
}

// ResolveAudienceForTest exposes resolveAudience for testing.
func ResolveAudienceForTest(tenant domain.Tenant, scope string) (domain.IdentifierURI, map[domain.ClientID]bool) {
	return resolveAudience(tenant, scope)
}

// ResolveRolesForTest exposes resolveRoles for testing.
func ResolveRolesForTest(
	tenant domain.Tenant,
	client domain.Client,
	user *domain.User,
	targetAppIDs map[domain.ClientID]bool,
	requestedScopes []string,
) []domain.RoleValue {
	return resolveRoles(tenant, client, user, targetAppIDs, requestedScopes)
}
