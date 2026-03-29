package handlers

import (
	"errors"
	"math/rand"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"identity/app"
	"identity/domain"
)

var (
	errInvalidTestTokenPath = errors.New("invalid test-token path: expected /test-tokens/{tenant}/{appId}")
	errInvalidAppID         = errors.New("invalid app ID: must be a UUID")
)

func testTokenHandler(request *http.Request, application *app.App) Response {
	tenants := application.Config.Tenants()
	if len(tenants) == 0 {
		return internalError("no tenants configured")
	}

	parts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")

	tenant := tenants[rand.Intn(len(tenants))]

	if len(parts) > testTokenTenantPart && parts[testTokenTenantPart] != "" {
		if t, err := app.FindTenant(application.Config, parts[testTokenTenantPart]); err == nil {
			tenant = t
		}
	}

	var clientID domain.ClientID

	if len(parts) > testTokenAppPart && parts[testTokenAppPart] != "" {
		if appUUID, err := uuid.Parse(parts[testTokenAppPart]); err == nil {
			clientID = domain.ClientIDFromUUID(appUUID)
		}
	}

	if clientID.UUID() == uuid.Nil {
		if clients := tenant.Clients(); len(clients) > 0 {
			client := clients[rand.Intn(len(clients))]
			clientID = client.ClientID()
		}
	}

	scope := request.URL.Query().Get("scope")
	if scope == "" {
		scope = "openid"
	}

	input := domain.TokenInput{
		Grant:         domain.GrantTest,
		Tenant:        tenant,
		Client:        resolveClientFromID(tenant, clientID),
		Scope:         scope,
		IsV2:          true,
		BaseURL:       extractBaseURL(request),
		User:          resolveTestUser(tenant, request.URL.Query().Get("username")),
		Nonce:         emptyValue,
		CorrelationID: emptyValue,
	}

	response := app.IssueToken(application.Key, input)

	return okText([]byte(response.AccessToken.RawString() + "\n"))
}

func resolveTestUser(tenant domain.Tenant, username string) *domain.User {
	if username != "" {
		if user, found := app.FindUserByID(tenant, username); found {
			return &user
		}
	}

	if len(tenant.Users()) > minValidIndex {
		users := tenant.Users()
		first := users[rand.Intn(len(users))]

		return &first
	}

	return nil
}
