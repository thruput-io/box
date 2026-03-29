package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

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
		Scope:         request.Form.Get("scope"),
		Nonce:         request.Form.Get("nonce"),
		IsV2:          isV2,
		BaseURL:       extractBaseURL(request),
		CorrelationID: correlationID(request),
	}

	switch grant {
	case domain.GrantPassword, domain.GrantTest:
		return buildPasswordInput(base, request, tenant)
	case domain.GrantClientCredentials:
		return buildClientCredentialsInput(base, request, tenant)
	case domain.GrantAuthorizationCode:
		return buildAuthCodeInput(base, request, tenant, application)
	case domain.GrantRefreshToken:
		return buildRefreshTokenInput(base, request, tenant, application)
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
	user, err := app.AuthenticateUser(tenant, request.Form.Get("username"), request.Form.Get("password"))
	if err != nil {
		return domain.TokenInput{}, domain.NewError(domain.ErrCodeInvalidCredentials, "invalid username or password")
	}

	base.User = &user
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
	if secretRaw != "" {
		s, err := domain.NewClientSecret(secretRaw)
		if err != nil {
			return domain.TokenInput{}, domain.NewError(domain.ErrCodeInvalidRequest, "invalid client secret format")
		}
		secret = &s
	}

	if err := app.ValidateClientSecret(client, secret); err != nil {
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

	if base.Scope == emptyValue {
		base.Scope = parsed.scope
	}

	base.Nonce = parsed.nonce
	base.Client = resolveClientFromID(tenant, firstOf(request.Form.Get(formKeyClientID), parsed.clientID))

	if user, found := app.FindUserByID(tenant, parsed.subject); found {
		base.User = &user
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

	if base.Scope == emptyValue {
		base.Scope = parsed.scope
	}

	base.Client = resolveClientFromID(tenant, firstOf(request.Form.Get(formKeyClientID), parsed.clientID))

	if user, found := app.FindUserByID(tenant, parsed.subject); found {
		base.User = &user
	}

	return base, nil
}

func resolveClientFromForm(tenant domain.Tenant, clientIDStr, clientSecret string) domain.Client {
	client := resolveClientFromID(tenant, clientIDStr)
	if client == nil {
		return nil
	}

	var secret *domain.ClientSecret
	if clientSecret != "" {
		s, err := domain.NewClientSecret(clientSecret)
		if err != nil {
			return nil
		}
		secret = &s
	}

	if err := client.Validate(secret); err != nil {
		return nil
	}

	return client
}

func resolveClientFromID(tenant domain.Tenant, clientIDStr string) domain.Client {
	if clientIDStr == emptyValue {
		return nil
	}

	clientID, err := domain.NewClientID(clientIDStr)
	if err != nil {
		return nil
	}

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
		"scope":        response.Scope,
	}

	if response.IDToken != nil {
		body["id_token"] = *response.IDToken
	}

	if response.RefreshToken != nil {
		body[jsonKeyRefreshToken] = *response.RefreshToken
	}

	if response.ClientInfo != nil {
		body["client_info"] = *response.ClientInfo
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return internalError("failed to encode token response")
	}

	return Response{
		Status:      http.StatusOK,
		Body:        encoded,
		ContentType: "application/json",
		Headers: map[string]string{
			"Client-Request-Id": response.CorrelationID,
			"X-Ms-Request-Id":   response.CorrelationID,
		},
	}
}

func correlationID(request *http.Request) string {
	headerID := request.Header.Get("Client-Request-Id")
	if headerID != emptyValue {
		return headerID
	}

	return uuid.New().String()
}
