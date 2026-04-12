package handlers

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/samber/mo"

	"identity/app"
	"identity/domain"
)

func addAzureVerifiedMockUtilsClaims(claims jwt.MapClaims) {
	claims["azpacls"] = "0"
	claims["jti"] = uuid.NewString()
}

func signTokenHandler(request *http.Request, application *app.App) Response {
	tokenData, err := extractTokenData(request)
	if err != nil {
		return internalError("failed to read body")
	}

	if len(tokenData) == emptySize {
		return badRequest(domain.NewError(domain.ErrCodeInvalidRequest, "missing token"))
	}

	var claims jwt.MapClaims

	err = json.Unmarshal(tokenData, &claims)
	if err != nil {
		return badRequest(domain.NewError(domain.ErrCodeInvalidRequest, "invalid json claims"))
	}

	signed := app.SignMapClaims(application.Key, claims)

	return okText([]byte(signed + "\n"))
}

func extractTokenData(request *http.Request) ([]byte, error) {
	var tokenData []byte

	if request.Method == http.MethodPost {
		data, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}

		tokenData = data
	}

	if len(tokenData) == emptySize {
		tokenData = []byte(request.URL.Query().Get("token"))
	}

	return tokenData, nil
}

func testTokenHandler(request *http.Request, application *app.App) Response {
	parts := strings.Split(strings.Trim(request.URL.Path, pathSeparator), pathSeparator)
	tenant := resolveTestTenant(application.Config, parts)
	clientResult := resolveTestClient(tenant, parts)

	tokenData, err := extractTokenData(request)
	if err == nil && len(tokenData) > emptySize {
		return signWithOverrides(application.Key, tokenData, tenant, clientResult)
	}

	scopeStr := request.URL.Query().Get("scope")
	if scopeStr == emptyValue {
		scopeStr = "openid"
	}

	input := domain.TokenInput{
		Grant:         domain.GrantTest,
		Tenant:        tenant,
		Client:        clientResult,
		Scope:         parseScopeValues(scopeStr),
		IsV2:          true,
		BaseURL:       extractBaseURL(request),
		User:          resolveTestUser(tenant, request.URL.Query().Get("username")),
		Nonce:         mo.None[domain.Nonce](),
		CorrelationID: mo.None[domain.CorrelationID](),
	}

	query := request.URL.Query()
	if len(query) == emptySize {
		mapClaims := app.BuildAccessTokenClaims(input).ToJWT()
		addAzureVerifiedMockUtilsClaims(mapClaims)
		delete(mapClaims, "azpacr")
		delete(mapClaims, "uti")

		signed := app.SignMapClaims(application.Key, mapClaims)

		return okText([]byte(signed + "\n"))
	}

	return issueTokenWithQueryOverrides(application, input, query)
}

func issueTokenWithQueryOverrides(
	application *app.App,
	input domain.TokenInput,
	query url.Values,
) Response {
	mapClaims := app.BuildAccessTokenClaims(input).ToJWT()

	// Add/Override with any additional query parameters as requested in user_interface.md
	for key, values := range query {
		if key == "scope" || key == "username" || key == "token" {
			continue
		}

		if len(values) == singleValueSize {
			mapClaims[key] = values[firstIndex]
		} else {
			mapClaims[key] = values
		}
	}

	addAzureVerifiedMockUtilsClaims(mapClaims)

	signed := app.SignMapClaims(application.Key, mapClaims)

	return okText([]byte(signed + "\n"))
}

func signWithOverrides(
	key *rsa.PrivateKey,
	tokenData []byte,
	tenant *domain.Tenant,
	client *domain.Client,
) Response {
	var claims jwt.MapClaims

	err := json.Unmarshal(tokenData, &claims)
	if err != nil {
		return badRequest(domain.NewError(domain.ErrCodeInvalidRequest, "invalid json claims"))
	}

	// Always override tid, aud, azp from URL segments if they are provided.
	claims["tid"] = tenant.TenantID().UUID().String()
	claims["aud"] = client.ClientID().UUID().String()
	claims["azp"] = client.ClientID().UUID().String()

	signed := app.SignMapClaims(key, claims)

	return okText([]byte(signed + "\n"))
}

func resolveTestTenant(config *domain.Config, parts []string) *domain.Tenant {
	tenants := config.Tenants()
	tenant := &tenants[firstIndex]

	if len(parts) > testTokenTenantPart && parts[testTokenTenantPart] != emptyValue {
		tenant = resolveTenantFromPart(config, parts[testTokenTenantPart], tenant)
	}

	return tenant
}

func resolveTenantFromPart(config *domain.Config, part string, defaultTenant *domain.Tenant) *domain.Tenant {
	id, ok := domain.NewTenantID(part).Right()
	if !ok {
		return defaultTenant
	}

	ten, err := app.FindTenantByID(config, id)
	if err != nil {
		return defaultTenant
	}

	return ten
}

func resolveTestClient(tenant *domain.Tenant, parts []string) *domain.Client {
	if len(parts) > testTokenClientPart && parts[testTokenClientPart] != emptyValue {
		return resolveClientFromPart(tenant, parts[testTokenClientPart])
	}

	clients := tenant.Clients()
	if len(clients) > emptySliceSize {
		return &clients[firstIndex]
	}

	c := tenant.AsClient()

	return &c
}

func resolveClientFromPart(tenant *domain.Tenant, part string) *domain.Client {
	clientID, ok := domain.NewClientID(part).Right()
	if !ok {
		c := tenant.AsClient()

		return &c
	}

	client, err := app.FindClient(*tenant, clientID)
	if err == nil {
		return client
	}

	// Also check app registrations as they can be audiences too.
	reg, regErr := app.FindAppRegistration(*tenant, clientID)
	if regErr == nil {
		client := domain.NewClientWithoutSecret(
			reg.Name(),
			reg.ClientID(),
			reg.RedirectURLs(),
			nil, // No assignments for default registrations
		)

		return &client
	}

	c := tenant.AsClient()

	return &c
}

func resolveTestUser(tenant *domain.Tenant, username string) *domain.User {
	if username == emptyValue {
		return resolveDefaultUser(tenant)
	}

	uname, ok := domain.NewUsername(username).Right()
	if !ok {
		return resolveDefaultUser(tenant)
	}

	for _, user := range tenant.Users() {
		if user.Username() == uname {
			return &user
		}
	}

	return resolveDefaultUser(tenant)
}

func resolveDefaultUser(tenant *domain.Tenant) *domain.User {
	users := tenant.Users()
	if len(users) > emptySliceSize {
		return &users[firstIndex]
	}

	return nil
}
