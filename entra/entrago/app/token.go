package app

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/samber/mo"

	"identity/domain"
)

const (
	accessTokenDuration  = 24 * time.Hour
	refreshTokenDuration = 90 * 24 * time.Hour
	authCodeDuration     = 5 * time.Minute
	accessTokenExpiry    = 3600
	graphClientID        = "00000003-0000-0000-c000-000000000000"
	graphAudience        = "https://graph.microsoft.com"
	emptySliceLen        = 0
	defaultDisplayName   = "Mock User"
	defaultEmail         = "user@example.com"
)

// tokenClaimsBase groups the common fields used to build JWT claim sets.
type tokenClaimsBase struct {
	issuer   domain.Issuer
	subject  mo.Option[domain.Subject]
	audience domain.IdentifierURI
	tenantID domain.TenantID
	version  domain.TokenVersion
	now      time.Time
}

// BuildAccessTokenClaims builds the claims for an access token based on domain inputs.
func BuildAccessTokenClaims(input domain.TokenInput) jwt.MapClaims {
	issuer, version := buildIssuerForInput(input)
	tenantID := input.Tenant.TenantID()
	subject := resolveSubject(input)
	targetAudience, targetAppIDs := resolveAudience(input.Tenant, input.Scope)

	roles := resolveRoles(
		input.Tenant, input.Client, input.User,
		targetAppIDs, input.Scope,
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

	return buildAccessClaims(base, input, roles)
}

// IssueToken issues a signed access token from pre-validated domain inputs.
// This function is pure: inputs are fully validated before calling it, so it cannot fail.
func IssueToken(key *rsa.PrivateKey, input domain.TokenInput) domain.TokenResponse {
	claims := BuildAccessTokenClaims(input)
	accessToken := domain.MustAccessToken(SignClaims(key, claims))

	now := time.Now()

	issuer, version := buildIssuerForInput(input)
	tenantID := input.Tenant.TenantID()
	subject := resolveSubject(input)
	displayName, email := resolveUserInfo(input.User)

	base := tokenClaimsBase{
		issuer:   issuer,
		subject:  subject,
		audience: domain.IdentifierURI{},
		tenantID: tenantID,
		version:  version,
		now:      now,
	}

	response := domain.TokenResponse{
		AccessToken:   accessToken,
		TokenType:     domain.MustTokenType("Bearer"),
		ExpiresIn:     accessTokenExpiry,
		Scope:         input.Scope,
		IDToken:       mo.None[domain.IDToken](),
		RefreshToken:  mo.None[domain.RefreshToken](),
		ClientInfo:    mo.None[domain.ClientInfo](),
		CorrelationID: input.CorrelationID,
	}

	if containsScope(input.Scope, "openid") && input.User != nil {
		idClaims := buildIDClaims(base, clientID(input), input.Nonce, displayName, email)
		idToken := domain.MustIDToken(SignClaims(key, idClaims))
		response.IDToken = mo.Some(idToken)
	}

	if containsScope(input.Scope, "offline_access") || input.Grant == domain.GrantRefreshToken {
		refreshToken := domain.MustRefreshToken(SignClaims(
			key,
			buildRefreshClaims(issuer, subject, clientID(input), tenantID, input.Scope, now),
		))
		response.RefreshToken = mo.Some(refreshToken)
	}

	if input.User != nil {
		clientInfo := BuildClientInfo(input.User.ID(), input.Tenant.TenantID())
		response.ClientInfo = mo.Some(clientInfo)
	}

	return response
}

// IssueAuthCode issues a short-lived signed JWT used as an authorization code.
func IssueAuthCode(
	key *rsa.PrivateKey,
	user domain.User,
	clientID domain.ClientID,
	redirectURI domain.RedirectURL,
	scope []domain.ScopeValue,
	tenantID domain.TenantID,
	nonce mo.Option[domain.Nonce],
) domain.AuthCode {
	claims := jwt.MapClaims{
		claimSub:         user.ID().Value(),
		claimClientID:    clientID.Value(),
		claimRedirectURI: redirectURI.Value(),
		claimScope:       domain.JoinScopeValues(scope),
		claimTenant:      tenantID.Value(),
		claimExp:         time.Now().Add(authCodeDuration).Unix(),
	}

	if n, ok := nonce.Get(); ok {
		claims[claimNonce] = n.Value()
	}

	return domain.MustAuthCode(SignClaims(key, claims))
}

func buildAccessClaims(
	base tokenClaimsBase,
	input domain.TokenInput,
	roles []domain.RoleValue,
) jwt.MapClaims {
	subjectVal := subjectString(base.subject)

	claims := jwt.MapClaims{
		claimIss:   base.issuer.Value(),
		claimSub:   subjectVal,
		claimAud:   base.audience.Value(),
		claimExp:   base.now.Add(accessTokenDuration).Unix(),
		claimIat:   base.now.Unix(),
		claimNbf:   base.now.Unix(),
		claimJti:   uuid.New().String(),
		claimTid:   base.tenantID.Value(),
		claimVer:   base.version.Value(),
		claimOid:   subjectVal,
		claimRoles: roleValuesToStrings(roles),
	}

	if input.Client != nil {
		claims[claimAzp] = input.Client.ClientID().Value()
		claims[claimAzpacls] = claimAzpacls0
		claims[claimAppid] = input.Client.ClientID().Value()
	}

	if input.User != nil {
		input.User.MapClaims(claims)

		if base.audience.Matches(graphClientID) || base.audience.Matches(graphAudience) {
			claims[claimScp] = domain.JoinRoleValues(roles, " ")
		} else {
			claims[claimScp] = domain.JoinScopeValues(input.Scope)
		}
	}

	if n, ok := input.Nonce.Get(); ok {
		claims[claimNonce] = n.Value()
	}

	return claims
}

func buildIDClaims(
	base tokenClaimsBase,
	clientID domain.ClientID,
	nonce mo.Option[domain.Nonce],
	displayName domain.DisplayName,
	email domain.Email,
) jwt.MapClaims {
	subjectVal := subjectString(base.subject)

	claims := jwt.MapClaims{
		claimIss:               base.issuer.Value(),
		claimSub:               subjectVal,
		claimAud:               clientID.Value(),
		claimExp:               base.now.Add(accessTokenDuration).Unix(),
		claimIat:               base.now.Unix(),
		claimNbf:               base.now.Unix(),
		claimTid:               base.tenantID.Value(),
		claimVer:               base.version.Value(),
		claimOid:               subjectVal,
		claimName:              displayName.Value(),
		claimPreferredUsername: email.Value(),
		claimEmail:             email.Value(),
	}

	if n, ok := nonce.Get(); ok {
		claims[claimNonce] = n.Value()
	}

	return claims
}

func buildRefreshClaims(
	issuer domain.Issuer,
	subject mo.Option[domain.Subject],
	clientID domain.ClientID,
	tenantID domain.TenantID,
	scope []domain.ScopeValue,
	now time.Time,
) jwt.MapClaims {
	return jwt.MapClaims{
		claimIss:      issuer.Value(),
		claimSub:      subjectString(subject),
		claimAud:      issuer.Value(),
		claimExp:      now.Add(refreshTokenDuration).Unix(),
		claimIat:      now.Unix(),
		claimClientID: clientID.Value(),
		claimScope:    domain.JoinScopeValues(scope),
		claimTid:      tenantID.Value(),
		claimTyp:      claimRefreshTyp,
	}
}

func resolveSubject(input domain.TokenInput) mo.Option[domain.Subject] {
	if input.User != nil {
		return mo.Some(input.User.ID().AsSubject())
	}

	if input.Client != nil {
		return mo.Some(input.Client.ClientID().AsSubject())
	}

	return mo.None[domain.Subject]()
}

func subjectString(subject mo.Option[domain.Subject]) string {
	if s, ok := subject.Get(); ok {
		return s.Value()
	}

	return ""
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

func buildIssuerForInput(input domain.TokenInput) (domain.Issuer, domain.TokenVersion) {
	tenantID := input.Tenant.TenantID()
	base := tenantID.AsURL(input.BaseURL)

	if input.IsV2 {
		return domain.MustIssuer(base.Value() + "/v2.0"), domain.MustTokenVersion("2.0")
	}

	return base, domain.MustTokenVersion("1.0")
}

func resolveAudience(
	tenant *domain.Tenant,
	scope []domain.ScopeValue,
) (domain.IdentifierURI, map[domain.ClientID]bool) {
	targetAudience := domain.MustIdentifierURI("api://default")
	targetAppIDs := make(map[domain.ClientID]bool)

	for _, scopePart := range scope {
		if isOIDCScope(scopePart) {
			continue
		}

		matchAudienceForScope(tenant, scopePart, targetAppIDs, &targetAudience)
	}

	return targetAudience, targetAppIDs
}

func matchAudienceForScope(
	tenant *domain.Tenant,
	scopePart domain.ScopeValue,
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
	scopePart domain.ScopeValue,
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
	tenant *domain.Tenant,
	client *domain.Client,
	user *domain.User,
	targetAppIDs map[domain.ClientID]bool,
	requestedScopes []domain.ScopeValue,
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
	client *domain.Client,
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
	tenant *domain.Tenant,
	targetAppIDs map[domain.ClientID]bool,
	requestedScopes []domain.ScopeValue,
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
	requestedScopes []domain.ScopeValue,
	resolved map[domain.RoleValue]bool,
) {
	for _, role := range registration.AppRoles() {
		if roleMatchesAnyScope(role, requestedScopes) {
			resolved[role.Value()] = true
		}
	}
}

func roleMatchesAnyScope(role domain.Role, requestedScopes []domain.ScopeValue) bool {
	for _, requestedScope := range requestedScopes {
		if !isOIDCScope(requestedScope) && roleScopesMatchScope(role.Scopes(), requestedScope) {
			return true
		}
	}

	return false
}

func roleScopesMatchScope(scopes []domain.Scope, requestedScope domain.ScopeValue) bool {
	for _, roleScope := range scopes {
		if roleScope.Value().Matches(requestedScope.Value()) {
			return true
		}
	}

	return false
}

func isOIDCScope(scope domain.ScopeValue) bool {
	switch scope.Value() {
	case "openid", "profile", "offline_access", "email":
		return true
	default:
		return false
	}
}

func containsScope(scopes []domain.ScopeValue, name string) bool {
	for _, s := range scopes {
		if s.Value() == name {
			return true
		}
	}

	return false
}

// ResolveAudienceForTest exposes resolveAudience for testing.
func ResolveAudienceForTest(
	tenant *domain.Tenant,
	scope []domain.ScopeValue,
) (domain.IdentifierURI, map[domain.ClientID]bool) {
	return resolveAudience(tenant, scope)
}

// ResolveRolesForTest exposes resolveRoles for testing.
func ResolveRolesForTest(
	tenant *domain.Tenant,
	client *domain.Client,
	user *domain.User,
	targetAppIDs map[domain.ClientID]bool,
	requestedScopes []domain.ScopeValue,
) []domain.RoleValue {
	return resolveRoles(tenant, client, user, targetAppIDs, requestedScopes)
}
