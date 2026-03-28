package handlers

import (
	"net/http"

	"identity/app"
	"identity/domain"
)

// Exported constants for testing only.
const (
	ContentTypePlain = contentTypePlain
	PathApp          = pathApp
	PathCsharp       = pathCsharp
)

// Exported wrapper functions for testing only.

func OkText(body []byte) Response {
	return okText(body)
}

func InternalError(msg string) Response {
	return internalError(msg)
}

func FromDomainError(domErr *domain.Error) Response {
	return fromDomainError(domErr)
}

func CollectAssignmentRoles(user domain.User, client domain.Client) map[string][]string {
	return collectAssignmentRoles(user, client)
}

func ResolveAppName(tenant domain.Tenant, appIDStr string) string {
	return resolveAppName(tenant, appIDStr)
}

func ResolveDisplayRoles(user domain.User, client *domain.Client, tenant domain.Tenant) []string {
	return resolveDisplayRoles(user, client, tenant)
}

func ResolveTestUser(tenant domain.Tenant, username string) *domain.User {
	return resolveTestUser(tenant, username)
}

func ConfigHandler(request *http.Request, application *app.App) Response {
	return configHandler(request, application)
}

func InvokeTestTokenHandler(request *http.Request, application *app.App) Response {
	return testTokenHandler(request, application)
}

func ParseTenantAndAppID(path, midSegment, suffix string) (domain.TenantID, domain.ClientID, error) {
	return parseTenantAndAppID(path, midSegment, suffix)
}
