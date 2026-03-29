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

// ExportOkText is for testing okText from handlers_test.
func ExportOkText(body []byte) Response {
	return okText(body)
}

// ExportInternalError is for testing internalError from handlers_test.
func ExportInternalError(msg string) Response {
	return internalError(msg)
}

// ExportFromDomainError is for testing fromDomainError from handlers_test.
func ExportFromDomainError(domErr *domain.Error) Response {
	return fromDomainError(domErr)
}

// ExportCollectAssignmentRoles is for testing collectAssignmentRoles from handlers_test.
func ExportCollectAssignmentRoles(user domain.User, client domain.Client) map[domain.ClientID][]string {
	return collectAssignmentRoles(user, client)
}

// ExportResolveAppName is for testing resolveAppName from handlers_test.
func ExportResolveAppName(tenant domain.Tenant, appID domain.ClientID) string {
	return resolveAppName(tenant, appID)
}

// ExportResolveDisplayRoles is for testing resolveDisplayRoles from handlers_test.
func ExportResolveDisplayRoles(user domain.User, client domain.Client, tenant domain.Tenant) []string {
	return resolveDisplayRoles(user, client, tenant)
}

// ExportResolveTestUser is for testing resolveTestUser from handlers_test.
func ExportResolveTestUser(tenant domain.Tenant, username string) *domain.User {
	return resolveTestUser(tenant, username)
}

// ExportConfigHandler is for testing configHandler from handlers_test.
func ExportConfigHandler(request *http.Request, application *app.App) Response {
	return configHandler(request, application)
}

// ExportInvokeTestTokenHandler is for testing testTokenHandler from handlers_test.
func ExportInvokeTestTokenHandler(request *http.Request, application *app.App) Response {
	return testTokenHandler(request, application)
}

// ExportParseTenantAndAppID is for testing parseTenantAndAppID from handlers_test.
func ExportParseTenantAndAppID(path, midSegment, suffix string) (domain.TenantID, domain.ClientID, error) {
	return parseTenantAndAppID(path, midSegment, suffix)
}

// ExportResolveClientFromID is for testing resolveClientFromID from handlers_test.
func ExportResolveClientFromID(tenant domain.Tenant, clientID string) domain.Client {
	id, _ := domain.NewClientID(clientID)

	return resolveClientFromID(tenant, id)
}

// ExportResolveClientFromForm is for testing resolveClientFromForm from handlers_test.
func ExportResolveClientFromForm(tenant domain.Tenant, clientID, clientSecret string) domain.Client {
	return resolveClientFromForm(tenant, clientID, clientSecret)
}

// ExportFirstOf is for testing firstOf from handlers_test.
func ExportFirstOf(primary, fallback string) string {
	return firstOf(primary, fallback)
}

// ExportCorrelationID is for testing correlationID from handlers_test.
func ExportCorrelationID(request *http.Request) string {
	return correlationID(request)
}
