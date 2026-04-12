package app

import (
	"crypto/rsa"
	"github.com/google/uuid"
	"github.com/samber/mo"
	"time"

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
func BuildAccessTokenClaims(input domain.TokenInput) domain.Claims {
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
	claims := domain.EmptyClaims()
	claims = claims.Append(user.ID().AsSubject().AsClaim())
	claims = claims.Append(clientID.AsClaim())
	claims = claims.Append(clientID.AsAudClaim())
	claims = claims.Append(tenantID.AsClaim())
	claims = claims.Append(domain.NewJwtNumericDateSeconds(time.Now().Add(authCodeDuration).Unix()).AsExpClaim())
	if c, ok := domain.ScpClaimFromScopeSlice(scope).Get(); ok {
		claims = claims.Append(c)
	}

	if n, ok := nonce.Get(); ok {
		claims = claims.Append(n.AsClaim())
	}

	return domain.MustAuthCode(SignClaims(key, claims))
}

func buildAccessClaims(
	base tokenClaimsBase,
	input domain.TokenInput,
	roles []domain.RoleValue,
) domain.Claims {
	claims := domain.EmptyClaims()

	claims = claims.Append(base.issuer.AsClaim())
	claims = claims.Append(base.audience.AsClaim())
	claims = claims.Append(base.tenantID.AsClaim())
	claims = claims.Append(base.version.AsClaim())
	claims = claims.Append(domain.NewJwtNumericDateSeconds(base.now.Add(accessTokenDuration).Unix()).AsExpClaim())
	claims = claims.Append(domain.NewJwtNumericDateSeconds(base.now.Unix()).AsIatClaim())
	claims = claims.Append(domain.NewJwtNumericDateSeconds(base.now.Unix()).AsNbfClaim())
	claims = claims.Append(domain.TokenUniqueIDFromUUID(uuid.New()).AsUtiClaim())
	if c, ok := domain.RolesClaimFromSlice(roles).Get(); ok {
		claims = claims.Append(c)
	} else {
		claims = claims.Append(domain.EmptyRolesClaim())
	}

	if sub, ok := base.subject.Get(); ok {
		claims = claims.Append(sub.AsClaim())
		claims = claims.Append(sub.AsOidClaim())
	}

	if input.Client != nil {
		claims = claims.Append(input.Client.ClientID().AsClaim())
		claims = claims.Append(domain.Azpacr0Claim())
		claims = claims.Append(input.Client.ClientID().AsAppidClaim())
	}

	if input.User != nil {
		claims = claims.Append(input.User.DisplayName().AsClaim())
		claims = claims.Append(input.User.Email().AsPreferredUsernameClaim())
		claims = claims.Append(input.User.Email().AsClaim())
		claims = claims.Append(input.User.Email().AsUniqueNameClaim())

		if base.audience.Matches(graphClientID) || base.audience.Matches(graphAudience) {
			if c, ok := domain.ScpClaimFromRoleSlice(roles).Get(); ok {
				claims = claims.Append(c)
			}
		} else {
			if c, ok := domain.ScpClaimFromScopeSlice(input.Scope).Get(); ok {
				claims = claims.Append(c)
			}
		}
	}

	if n, ok := input.Nonce.Get(); ok {
		claims = claims.Append(n.AsClaim())
	}

	return claims
}

func buildIDClaims(
	base tokenClaimsBase,
	clientID domain.ClientID,
	nonce mo.Option[domain.Nonce],
	displayName domain.DisplayName,
	email domain.Email,
) domain.Claims {
	claims := domain.EmptyClaims()

	claims = claims.Append(base.issuer.AsClaim())
	claims = claims.Append(clientID.AsAudClaim())
	claims = claims.Append(base.tenantID.AsClaim())
	claims = claims.Append(base.version.AsClaim())
	claims = claims.Append(domain.NewJwtNumericDateSeconds(base.now.Add(accessTokenDuration).Unix()).AsExpClaim())
	claims = claims.Append(domain.NewJwtNumericDateSeconds(base.now.Unix()).AsIatClaim())
	claims = claims.Append(domain.NewJwtNumericDateSeconds(base.now.Unix()).AsNbfClaim())
	claims = claims.Append(displayName.AsClaim())
	claims = claims.Append(email.AsPreferredUsernameClaim())
	claims = claims.Append(email.AsClaim())
	claims = claims.Append(email.AsUniqueNameClaim())

	if sub, ok := base.subject.Get(); ok {
		claims = claims.Append(sub.AsClaim())
		claims = claims.Append(sub.AsOidClaim())
	}

	if n, ok := nonce.Get(); ok {
		claims = claims.Append(n.AsClaim())
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
) domain.Claims {
	claims := domain.EmptyClaims()
	claims = claims.Append(issuer.AsClaim())
	claims = claims.Append(issuer.AsAudClaim())
	claims = claims.Append(domain.NewJwtNumericDateSeconds(now.Add(refreshTokenDuration).Unix()).AsExpClaim())
	claims = claims.Append(domain.NewJwtNumericDateSeconds(now.Unix()).AsIatClaim())
	claims = claims.Append(clientID.AsClaim())
	claims = claims.Append(tenantID.AsClaim())
	if c, ok := domain.ScpClaimFromScopeSlice(scope).Get(); ok {
		claims = claims.Append(c)
	}
	claims = claims.Append(domain.RefreshTokenTypClaim())

	if sub, ok := subject.Get(); ok {
		claims = claims.Append(sub.AsClaim())
	}

	return claims
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

// ResolveRolesForTest exposes resolveRoles for testing.
