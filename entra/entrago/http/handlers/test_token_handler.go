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

func testTokenHandler(request *http.Request, application *app.App) error {
	tenants := application.Config.Tenants()

	parts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")

	tenant := tenants[0]

	if len(parts) > testTokenTenantPart && parts[testTokenTenantPart] != "" {
		id, idErr := domain.NewTenantID(parts[testTokenTenantPart])
		if idErr != nil {
			return idErr
		}
		ten, err := app.FindTenant(application.Config, id)
		if err != nil {
			return err
		}

		tenant = ten
	}

	var clientIDResult *domain.ClientID = nil

	if len(parts) > testTokenClientPart && parts[testTokenClientPart] != "" {
		clientID, err := domain.NewClientID(parts[testTokenClientPart])
		if err != nil {
			return err
		}
		clientIDResult = &clientID
	}

	var clientResult *domain.Client = nil

	if clientIDResult != nil {
		client, err := app.FindClient(tenant, *clientIDResult)
		if err != nil {
			return err
		}
		clientResult = &client
	} else {
		if len(tenant.Clients()) > 0 {
			clientResult = &tenant.Clients()[0]
		}

		clientResult = &(tenant.AsClient())
		e
	}

	scope := request.URL.Query().Get("scope")
	if scope == "" {
		scope = "openid"
	}

	requestedUsername := request.URL.Query().Get("username")
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

	return okText(response.AccessToken.AsByteArray())
}
