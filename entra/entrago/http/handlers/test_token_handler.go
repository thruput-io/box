package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"identity/app"
	"identity/domain"
)

const (
	firstIndex = 0
	emptySize  = 0
)

func signTokenHandler(request *http.Request, application *app.App) Response {
	var tokenData []byte

	var err error

	if request.Method == http.MethodPost {
		tokenData, err = io.ReadAll(request.Body)
		if err != nil {
			return internalError("failed to read body")
		}
	}

	if len(tokenData) == emptySize {
		tokenData = []byte(request.URL.Query().Get("token"))
	}

	if len(tokenData) == emptySize {
		return badRequest(domain.NewError(domain.ErrCodeInvalidRequest, "missing token"))
	}

	var claims jwt.MapClaims

	err = json.Unmarshal(tokenData, &claims)
	if err != nil {
		return badRequest(domain.NewError(domain.ErrCodeInvalidRequest, "invalid json claims"))
	}

	signed := app.SignClaims(application.Key, claims)

	return okText([]byte(signed))
}

func testTokenHandler(request *http.Request, application *app.App) Response {
	parts := strings.Split(strings.Trim(request.URL.Path, pathSeparator), pathSeparator)
	tenant := resolveTestTenant(application.Config, parts)
	clientResult := resolveTestClient(tenant, parts)

	scope := request.URL.Query().Get("scope")
	if scope == emptyValue {
		scope = "openid"
	}

	input := domain.TokenInput{
		Grant:         domain.GrantTest,
		Tenant:        tenant,
		Client:        clientResult,
		Scope:         scope,
		IsV2:          true,
		BaseURL:       extractBaseURL(request),
		User:          resolveTestUser(tenant, request.URL.Query().Get("username")),
		Nonce:         emptyValue,
		CorrelationID: emptyValue,
	}

	response := app.IssueToken(application.Key, input)

	return okText(response.AccessToken.AsByteArray())
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
	id, idErr := domain.NewTenantID(part)
	if idErr != nil {
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
	clientID, err := domain.NewClientID(part)
	if err != nil {
		c := tenant.AsClient()

		return &c
	}

	client, err := app.FindClient(*tenant, clientID)
	if err != nil {
		c := tenant.AsClient()

		return &c
	}

	return client
}

func resolveTestUser(tenant *domain.Tenant, username string) *domain.User {
	if username == emptyValue {
		return resolveDefaultUser(tenant)
	}

	uname, err := domain.NewUsername(username)
	if err != nil {
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
