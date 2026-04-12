package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/samber/mo"

	"identity/app"
	"identity/domain"
)

func tokenHandler(request *http.Request, application *app.App) Response {
	err := request.ParseForm()
	if err != nil {
		return fromDomainError(domain.NewError(domain.ErrCodeInvalidRequest, err.Error()))
	}

	paramErr := validateParamLengths(request.Form)
	if paramErr != nil {
		return fromDomainError(domain.NewError(domain.ErrCodeInvalidRequest, paramErr.Error()))
	}

	inputE := parseTokenRequest(request, application)
	if domErr, ok := inputE.Left(); ok {
		return fromDomainError(domErr)
	}

	return encodeTokenResponse(app.IssueToken(application.Key, inputE.MustRight()))
}

const (
	errMsgInvalidCredentials  = "invalid username or password"
	errMsgRedirectURIMismatch = "redirect_uri mismatch"
)

func parseTokenRequest(request *http.Request, application *app.App) mo.Either[domain.Error, domain.TokenInput] {
	grantTypeStr := request.Form.Get("grant_type")
	tenantIDStr := extractTenantID(request)
	isV2 := strings.Contains(request.URL.Path, "/v2.0")

	tenant, err := app.FindTenant(application.Config, tenantIDStr)
	if err != nil {
		return mo.Left[domain.Error, domain.TokenInput](domain.ErrTenantNotFound)
	}

	grantE := parseGrantType(grantTypeStr)

	grant, ok := grantE.Right()
	if !ok {
		return mo.Left[domain.Error, domain.TokenInput](grantE.MustLeft())
	}

	base := domain.TokenInput{
		Grant:         grant,
		Tenant:        tenant,
		User:          nil,
		Client:        nil,
		Scope:         parseScopeValues(request.Form.Get("scope")),
		Nonce:         parseOptionalNonce(request.Form.Get("nonce")),
		IsV2:          isV2,
		BaseURL:       extractBaseURL(request),
		CorrelationID: parseOptionalCorrelationID(request),
	}

	switch grant {
	case domain.GrantPassword, domain.GrantTest:
		return buildPasswordInput(base, request, *tenant)
	case domain.GrantClientCredentials:
		return buildClientCredentialsInput(base, request, *tenant)
	case domain.GrantAuthorizationCode:
		return buildAuthCodeInput(base, request, *tenant, application)
	case domain.GrantRefreshToken:
		return buildRefreshTokenInput(base, request, *tenant, application)
	default:
		return mo.Left[domain.Error, domain.TokenInput](domain.ErrUnsupportedGrantType)
	}
}

func parseGrantType(raw string) mo.Either[domain.Error, domain.GrantType] {
	switch raw {
	case "password":
		return mo.Right[domain.Error, domain.GrantType](domain.GrantPassword)
	case "test":
		return mo.Right[domain.Error, domain.GrantType](domain.GrantTest)
	case "client_credentials":
		return mo.Right[domain.Error, domain.GrantType](domain.GrantClientCredentials)
	case "authorization_code":
		return mo.Right[domain.Error, domain.GrantType](domain.GrantAuthorizationCode)
	case "refresh_token":
		return mo.Right[domain.Error, domain.GrantType](domain.GrantRefreshToken)
	default:
		return mo.Left[domain.Error, domain.GrantType](
			domain.NewError(domain.ErrCodeUnsupportedGrantType, "unsupported grant_type: "+raw),
		)
	}
}

func buildPasswordInput(
	base domain.TokenInput, request *http.Request, tenant domain.Tenant,
) mo.Either[domain.Error, domain.TokenInput] {
	username, ok := domain.NewUsername(request.Form.Get("username")).Right()
	if !ok {
		return mo.Left[domain.Error, domain.TokenInput](
			domain.NewError(domain.ErrCodeInvalidCredentials, errMsgInvalidCredentials),
		)
	}

	password, ok := domain.NewPassword(request.Form.Get("password")).Right()
	if !ok {
		return mo.Left[domain.Error, domain.TokenInput](
			domain.NewError(domain.ErrCodeInvalidCredentials, errMsgInvalidCredentials),
		)
	}

	user, err := app.AuthenticateUser(tenant, username, password)
	if err != nil {
		return mo.Left[domain.Error, domain.TokenInput](
			domain.NewError(domain.ErrCodeInvalidCredentials, errMsgInvalidCredentials),
		)
	}

	base.User = user
	base.Client = resolveClientFromForm(tenant, request.Form.Get("client_id"), request.Form.Get("client_secret"))

	return mo.Right[domain.Error, domain.TokenInput](base)
}

func buildClientCredentialsInput(
	base domain.TokenInput, request *http.Request, tenant domain.Tenant,
) mo.Either[domain.Error, domain.TokenInput] {
	clientID, ok := domain.NewClientID(request.Form.Get("client_id")).Right()
	if !ok {
		return mo.Left[domain.Error, domain.TokenInput](
			domain.NewError(domain.ErrCodeClientNotFound, "invalid client_id"),
		)
	}

	client, err := app.FindClient(tenant, clientID)
	if err != nil {
		return mo.Left[domain.Error, domain.TokenInput](domain.ErrClientNotFound)
	}

	var secret *domain.ClientSecret

	secretRaw := request.Form.Get("client_secret")
	if secretRaw != emptyValue {
		clientSecret, ok := domain.NewClientSecret(secretRaw).Right()
		if !ok {
			return mo.Left[domain.Error, domain.TokenInput](
				domain.NewError(domain.ErrCodeInvalidRequest, "invalid client secret format"),
			)
		}

		secret = &clientSecret
	}

	err = app.ValidateClientSecret(*client, secret)
	if err != nil {
		return mo.Left[domain.Error, domain.TokenInput](
			domain.NewError(domain.ErrCodeInvalidCredentials, "invalid client secret"),
		)
	}

	base.Client = client

	return mo.Right[domain.Error, domain.TokenInput](base)
}

func buildAuthCodeInput(
	base domain.TokenInput, request *http.Request, tenant domain.Tenant, application *app.App,
) mo.Either[domain.Error, domain.TokenInput] {
	parsedE := parseAuthCode(application.Key, request.Form.Get("code"))
	if parsedE.IsLeft() {
		return mo.Left[domain.Error, domain.TokenInput](parsedE.MustLeft())
	}

	parsed := parsedE.MustRight()

	clientIDFromClaims := claimsClientIDString(parsed)

	if result := checkRedirectURI(
		tenant,
		request.Form.Get("redirect_uri"),
		firstOf(request.Form.Get(formKeyClientID), clientIDFromClaims),
	); result.IsLeft() {
		return mo.Left[domain.Error, domain.TokenInput](result.MustLeft())
	}

	if len(base.Scope) == emptySize {
		base.Scope = parsed.ScopeValues()
	}

	base.Nonce = parsed.Nonce()

	clientID, ok := domain.NewClientID(firstOf(request.Form.Get(formKeyClientID), clientIDFromClaims)).Right()
	if ok {
		base.Client = resolveClientFromID(tenant, clientID)
	}

	base.User = resolveUserFromClaims(tenant, parsed)

	return mo.Right[domain.Error, domain.TokenInput](base)
}

func buildRefreshTokenInput(
	base domain.TokenInput, request *http.Request, tenant domain.Tenant, application *app.App,
) mo.Either[domain.Error, domain.TokenInput] {
	parsed, ok := parseRefreshToken(application.Key, request.Form.Get("refresh_token")).Right()
	if !ok {
		return mo.Left[domain.Error, domain.TokenInput](
			domain.NewError(domain.ErrCodeInvalidGrant, errMsgInvalidRefreshToken),
		)
	}

	if len(base.Scope) == emptySize {
		base.Scope = parsed.ScopeValues()
	}

	clientID, ok := domain.NewClientID(firstOf(request.Form.Get(formKeyClientID), claimsClientIDString(parsed))).Right()
	if ok {
		base.Client = resolveClientFromID(tenant, clientID)
	}

	base.User = resolveUserFromClaims(tenant, parsed)

	return mo.Right[domain.Error, domain.TokenInput](base)
}

func resolveUserFromClaims(tenant domain.Tenant, claims domain.Claims) *domain.User {
	sub, found := claims.Subject().Get()
	if !found {
		return nil
	}

	userID, ok := domain.NewUserID(sub.Value()).Right()
	if !ok {
		return nil
	}

	user, found := app.FindUserByID(tenant, userID)
	if !found {
		return nil
	}

	return user
}

func checkRedirectURI(tenant domain.Tenant, requestedRedirect, clientIDStr string) mo.Either[domain.Error, struct{}] {
	if requestedRedirect == emptyValue {
		return mo.Right[domain.Error, struct{}](struct{}{})
	}

	redirectURI, ok := domain.NewRedirectURL(requestedRedirect).Right()
	if !ok {
		return mo.Left[domain.Error, struct{}](
			domain.NewError(domain.ErrCodeInvalidGrant, errMsgRedirectURIMismatch),
		)
	}

	clientID, ok := domain.NewClientID(clientIDStr).Right()
	if !ok {
		return mo.Left[domain.Error, struct{}](
			domain.NewError(domain.ErrCodeInvalidGrant, errMsgRedirectURIMismatch),
		)
	}

	allowedURLs, err := app.FindRedirectURLs(tenant, clientID)
	if err != nil {
		return mo.Left[domain.Error, struct{}](
			domain.NewError(domain.ErrCodeInvalidGrant, errMsgRedirectURIMismatch),
		)
	}

	err = app.ValidateRedirectURI(redirectURI, allowedURLs)
	if err != nil {
		return mo.Left[domain.Error, struct{}](
			domain.NewError(domain.ErrCodeInvalidGrant, errMsgRedirectURIMismatch),
		)
	}

	return mo.Right[domain.Error, struct{}](struct{}{})
}

func claimsClientIDString(claims domain.Claims) string {
	clientID, found := claims.AuthorizedPartyClientID().Get()
	if !found {
		return emptyValue
	}

	return clientID.Value()
}

func resolveClientFromForm(tenant domain.Tenant, clientIDStr, clientSecret string) *domain.Client {
	clientID, ok := domain.NewClientID(clientIDStr).Right()
	if !ok {
		return nil
	}

	client := resolveClientFromID(tenant, clientID)
	if client == nil {
		return nil
	}

	var secret *domain.ClientSecret

	if clientSecret != emptyValue {
		s, ok := domain.NewClientSecret(clientSecret).Right()
		if !ok {
			return nil
		}

		secret = &s
	}

	if client.Validate(secret).IsLeft() {
		return nil
	}

	return client
}

func resolveClientFromID(tenant domain.Tenant, clientID domain.ClientID) *domain.Client {
	client, err := app.FindClient(tenant, clientID)
	if err != nil {
		return nil
	}

	return client
}

func firstOf(primary, fallback string) string {
	if primary != emptyValue {
		return primary
	}

	return fallback
}

func encodeTokenResponse(response domain.TokenResponse) Response {
	body := map[string]any{
		"access_token": response.AccessToken,
		"token_type":   response.TokenType,
		"expires_in":   response.ExpiresIn,
		"scope":        domain.JoinScopeValues(response.Scope),
	}

	if idToken, ok := response.IDToken.Get(); ok {
		body["id_token"] = idToken
	}

	if refreshToken, ok := response.RefreshToken.Get(); ok {
		body[jsonKeyRefreshToken] = refreshToken
	}

	if clientInfo, ok := response.ClientInfo.Get(); ok {
		body["client_info"] = clientInfo
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return internalError("failed to encode token response")
	}

	headers := map[string]string{}
	if corrID, ok := response.CorrelationID.Get(); ok {
		headers["Client-Request-Id"] = corrID.Value()
		headers["X-Ms-Request-Id"] = corrID.Value()
	}

	return Response{
		Status:      http.StatusOK,
		Body:        encoded,
		ContentType: "application/json",
		Headers:     headers,
	}
}

func parseOptionalCorrelationID(request *http.Request) mo.Option[domain.CorrelationID] {
	headerID := request.Header.Get("Client-Request-Id")
	if headerID != emptyValue {
		c, ok := domain.NewCorrelationID(headerID).Right()
		if ok {
			return mo.Some(c)
		}
	}

	return mo.Some(domain.NewCorrelationID(uuid.New().String()).MustRight())
}
