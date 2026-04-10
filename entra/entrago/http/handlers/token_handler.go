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

	input, domErr := parseTokenRequest(request, application)
	if domErr != nil {
		return fromDomainError(domErr)
	}

	return encodeTokenResponse(app.IssueToken(application.Key, input))
}

func parseTokenRequest(request *http.Request, application *app.App) (domain.TokenInput, *domain.Error) {
	grantTypeStr := request.Form.Get("grant_type")
	tenantIDStr := extractTenantID(request)
	isV2 := strings.Contains(request.URL.Path, "/v2.0")

	tenant, err := app.FindTenant(application.Config, tenantIDStr)
	if err != nil {
		return domain.TokenInput{}, domain.NewError(domain.ErrCodeTenantNotFound, "tenant not found")
	}

	grant, domErr := parseGrantType(grantTypeStr)
	if domErr != nil {
		return domain.TokenInput{}, domErr
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
		return domain.TokenInput{}, domain.NewError(domain.ErrCodeUnsupportedGrantType, "unsupported grant_type")
	}
}

func parseGrantType(raw string) (domain.GrantType, *domain.Error) {
	switch raw {
	case "password":
		return domain.GrantPassword, nil
	case "test":
		return domain.GrantTest, nil
	case "client_credentials":
		return domain.GrantClientCredentials, nil
	case "authorization_code":
		return domain.GrantAuthorizationCode, nil
	case "refresh_token":
		return domain.GrantRefreshToken, nil
	default:
		return "", domain.NewError(domain.ErrCodeUnsupportedGrantType, "unsupported grant_type: "+raw)
	}
}

func buildPasswordInput(
	base domain.TokenInput, request *http.Request, tenant domain.Tenant,
) (domain.TokenInput, *domain.Error) {
	username, _ := domain.NewUsername(request.Form.Get("username"))
	password, _ := domain.NewPassword(request.Form.Get("password"))

	user, err := app.AuthenticateUser(tenant, username, password)
	if err != nil {
		return domain.TokenInput{}, domain.NewError(domain.ErrCodeInvalidCredentials, "invalid username or password")
	}

	base.User = user
	base.Client = resolveClientFromForm(tenant, request.Form.Get("client_id"), request.Form.Get("client_secret"))

	return base, nil
}

func buildClientCredentialsInput(
	base domain.TokenInput, request *http.Request, tenant domain.Tenant,
) (domain.TokenInput, *domain.Error) {
	clientIDStr := request.Form.Get("client_id")

	clientID, err := domain.NewClientID(clientIDStr)
	if err != nil {
		return domain.TokenInput{}, domain.NewError(domain.ErrCodeClientNotFound, "invalid client_id")
	}

	client, err := app.FindClient(tenant, clientID)
	if err != nil {
		return domain.TokenInput{}, domain.NewError(domain.ErrCodeClientNotFound, "client not found")
	}

	var secret *domain.ClientSecret

	secretRaw := request.Form.Get("client_secret")
	if secretRaw != emptyValue {
		s, err := domain.NewClientSecret(secretRaw)
		if err != nil {
			return domain.TokenInput{}, domain.NewError(domain.ErrCodeInvalidRequest, "invalid client secret format")
		}

		secret = &s
	}

	err = app.ValidateClientSecret(*client, secret)
	if err != nil {
		return domain.TokenInput{}, domain.NewError(domain.ErrCodeInvalidCredentials, "invalid client secret")
	}

	base.Client = client

	return base, nil
}

func buildAuthCodeInput(
	base domain.TokenInput, request *http.Request, tenant domain.Tenant, application *app.App,
) (domain.TokenInput, *domain.Error) {
	parsed, parseErr := parseAuthCode(application.Key, request.Form.Get("code"))
	if parseErr != nil {
		return domain.TokenInput{}, parseErr
	}

	requestedRedirect := request.Form.Get("redirect_uri")
	if requestedRedirect != emptyValue && requestedRedirect != parsed.redirectURI {
		return domain.TokenInput{}, domain.NewError(domain.ErrCodeInvalidGrant, "redirect_uri mismatch")
	}

	if len(base.Scope) == emptySize {
		base.Scope = parsed.scope
	}

	base.Nonce = parsed.nonce
	clientID, _ := domain.NewClientID(firstOf(request.Form.Get(formKeyClientID), parsed.clientID))
	base.Client = resolveClientFromID(tenant, clientID)

	userID, _ := domain.NewUserID(parsed.subject)
	if user, found := app.FindUserByID(tenant, userID); found {
		base.User = user
	}

	return base, nil
}

func buildRefreshTokenInput(
	base domain.TokenInput, request *http.Request, tenant domain.Tenant, application *app.App,
) (domain.TokenInput, *domain.Error) {
	parsed, parseErr := parseRefreshToken(application.Key, request.Form.Get("refresh_token"))
	if parseErr != nil {
		return domain.TokenInput{}, parseErr
	}

	if len(base.Scope) == emptySize {
		base.Scope = parsed.scope
	}

	clientID, _ := domain.NewClientID(firstOf(request.Form.Get(formKeyClientID), parsed.clientID))
	base.Client = resolveClientFromID(tenant, clientID)

	userID, _ := domain.NewUserID(parsed.subject)
	if user, found := app.FindUserByID(tenant, userID); found {
		base.User = user
	}

	return base, nil
}

func resolveClientFromForm(tenant domain.Tenant, clientIDStr, clientSecret string) *domain.Client {
	clientID, err := domain.NewClientID(clientIDStr)
	if err != nil {
		return nil
	}

	client := resolveClientFromID(tenant, clientID)
	if client == nil {
		return nil
	}

	var secret *domain.ClientSecret

	if clientSecret != emptyValue {
		s, err := domain.NewClientSecret(clientSecret)
		if err != nil {
			return nil
		}

		secret = &s
	}

	err = client.Validate(secret)
	if err != nil {
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
		c, err := domain.NewCorrelationID(headerID)
		if err == nil {
			return mo.Some(c)
		}
	}

	return mo.Some(domain.MustCorrelationID(uuid.New().String()))
}
