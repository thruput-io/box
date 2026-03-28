package http

import (
	"errors"
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

func testTokenHandler(request *http.Request, server *Server) HTTPResponse {
	parts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if len(parts) < 3 {
		return badRequest(errInvalidTestTokenPath)
	}

	tenantIDStr := parts[1]
	appIDStr := parts[2]

	if _, err := app.FindTenant(server.Config, tenantIDStr); err != nil {
		return notFound("tenant not found")
	}

	appUUID, err := uuid.Parse(appIDStr)
	if err != nil {
		return badRequest(errInvalidAppID)
	}

	appID := domain.ClientIDFromUUID(appUUID)
	scope := request.URL.Query().Get("scope")

	if scope == "" {
		scope = "openid"
	}

	raw := app.RawTokenRequest{
		GrantType: "test",
		ClientID:  appID.String(),
		TenantID:  tenantIDStr,
		Scope:     scope,
		IsV2:      true,
		BaseURL:   extractBaseURL(request),
	}

	if username := request.URL.Query().Get("username"); username != "" {
		raw.Username = username
	}

	response, err := app.IssueToken(server.Config, server.Key, raw)
	if err != nil {
		return tokenError(err)
	}

	return okText([]byte(response.AccessToken + "\n"))
}
