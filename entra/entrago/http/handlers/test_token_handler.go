package handlers

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

func testTokenHandler(request *http.Request, application *app.App) Response {
	parts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if len(parts) < minTestTokenParts {
		return fromDomainError(domain.NewError(domain.ErrCodeInvalidRequest, errInvalidTestTokenPath.Error()))
	}

	tenantIDStr := parts[testTokenTenantPart]
	appIDStr := parts[testTokenAppPart]

	tenant, err := app.FindTenant(application.Config, tenantIDStr)
	if err != nil {
		return fromDomainError(domain.NewError(domain.ErrCodeTenantNotFound, "tenant not found"))
	}

	appUUID, parseErr := uuid.Parse(appIDStr)
	if parseErr != nil {
		return fromDomainError(domain.NewError(domain.ErrCodeInvalidRequest, errInvalidAppID.Error()))
	}

	scope := request.URL.Query().Get("scope")
	if scope == "" {
		scope = "openid"
	}

	input := domain.TokenInput{
		Grant:         domain.GrantTest,
		Tenant:        tenant,
		Client:        resolveClientFromID(tenant, domain.ClientIDFromUUID(appUUID).String()),
		Scope:         scope,
		IsV2:          true,
		BaseURL:       extractBaseURL(request),
		User:          resolveTestUser(tenant, request.URL.Query().Get("username")),
		Nonce:         emptyValue,
		CorrelationID: emptyValue,
	}

	response := app.IssueToken(application.Key, input)

	return okText([]byte(response.AccessToken + "\n"))
}

func resolveTestUser(tenant domain.Tenant, username string) *domain.User {
	if username != "" {
		if user, found := app.FindUserByID(tenant, username); found {
			return &user
		}
	}

	if len(tenant.Users()) > minValidIndex {
		first := tenant.Users()[minValidIndex]

		return &first
	}

	return nil
}
