package handlers

import (
	"errors"
	"net/http"
	"strings"

	"identity/app"
	"identity/domain"
)

var (
	errInvalidTestTokenPath = errors.New("invalid test-token path: expected /test-tokens/{tenant}/{appId}")
	errInvalidAppID         = errors.New("invalid app ID: must be a UUID")
)

func testTokenHandler(request *http.Request, application *app.App) Response {
	tenants := application.Config.Tenants()

	parts := strings.Split(strings.Trim(request.URL.Path, pathSeparator), pathSeparator)

	tenant := tenants[0]

	if len(parts) > testTokenTenantPart && parts[testTokenTenantPart] != emptyValue {
		id, idErr := domain.NewTenantID(parts[testTokenTenantPart])
		if idErr == nil {
			ten, err := app.FindTenantByID(application.Config, id)
			if err == nil {
				tenant = ten
			}
		}
	}

	var clientResult domain.Client = tenant.AsClient()

	if len(parts) > testTokenClientPart && parts[testTokenClientPart] != emptyValue {
		clientID, err := domain.NewClientID(parts[testTokenClientPart])
		if err == nil {
			client, err := app.FindClient(tenant, clientID)
			if err == nil {
				clientResult = client
			}
		}
	} else if len(tenant.Clients()) > 0 {
		clientResult = tenant.Clients()[0]
	}

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

func resolveTestUser(tenant domain.Tenant, username string) *domain.User {
	if username != emptyValue {
		uname, err := domain.NewUsername(username)
		if err == nil {
			for _, user := range tenant.Users() {
				if user.Username() == uname {
					return &user
				}
			}
		}
	}

	if len(tenant.Users()) > 0 {
		return &tenant.Users()[0]
	}

	return nil
}
