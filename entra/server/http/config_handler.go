package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"identity/app"
	"identity/domain"
)

func configHandler(request *http.Request, server *Server) HTTPResponse {
	path := request.URL.Path

	switch {
	case path == "/config/raw":
		return configRawHandler(request, server)
	case strings.Contains(path, "/app/") && strings.HasSuffix(path, "/csharp"):
		return configCsharpAppHandler(request, server)
	case strings.Contains(path, "/app/") && strings.HasSuffix(path, "/js"):
		return configJsAppHandler(request, server)
	case strings.Contains(path, "/client/") && strings.HasSuffix(path, "/csharp"):
		return configCsharpClientHandler(request, server)
	default:
		return notFound("config endpoint not found")
	}
}

func configRawHandler(_ *http.Request, _ *Server) HTTPResponse {
	return HTTPResponse{
		Status:      http.StatusMovedPermanently,
		ContentType: "",
		Body:        nil,
		Headers:     map[string]string{"Location": "/DefaultConfig.yaml"},
	}
}

func configCsharpAppHandler(request *http.Request, server *Server) HTTPResponse {
	tenantID, appID, err := parseTenantAndAppID(request.URL.Path, "/app/", "/csharp")
	if err != nil {
		return badRequest(err)
	}

	tenant, err := app.FindTenantByID(server.Config, tenantID)
	if err != nil {
		return notFound("tenant not found")
	}

	registration, err := app.FindAppRegistration(tenant, appID)
	if err != nil {
		return notFound("app registration not found")
	}

	content := fmt.Sprintf(
		"AzureAd__Instance=https://%s/\nAzureAd__TenantId=%s\nAzureAd__ClientId=%s\n",
		request.Host, tenant.TenantID(), registration.ClientID(),
	)

	return HTTPResponse{
		Status:      http.StatusOK,
		Body:        []byte(content),
		ContentType: "text/plain; charset=utf-8",
		Headers:     map[string]string{"Content-Disposition": fmt.Sprintf("inline; filename=%q", registration.Name().String()+"-appsettings.env")},
	}
}

func configJsAppHandler(request *http.Request, server *Server) HTTPResponse {
	tenantID, appID, err := parseTenantAndAppID(request.URL.Path, "/app/", "/js")
	if err != nil {
		return badRequest(err)
	}

	tenant, err := app.FindTenantByID(server.Config, tenantID)
	if err != nil {
		return notFound("tenant not found")
	}

	registration, err := app.FindAppRegistration(tenant, appID)
	if err != nil {
		return notFound("app registration not found")
	}

	content := fmt.Sprintf(
		"const msalConfig = {\n  auth: {\n    clientId: %q,\n    authority: \"https://%s/%s\",\n    knownAuthorities: [%q],\n  },\n};\n",
		registration.ClientID(), request.Host, tenant.TenantID(), request.Host,
	)

	return HTTPResponse{
		Status:      http.StatusOK,
		Body:        []byte(content),
		ContentType: "text/plain; charset=utf-8",
		Headers:     map[string]string{"Content-Disposition": fmt.Sprintf("inline; filename=%q", registration.Name().String()+"-msal-config.js")},
	}
}

func configCsharpClientHandler(request *http.Request, server *Server) HTTPResponse {
	tenantID, clientID, err := parseTenantAndAppID(request.URL.Path, "/client/", "/csharp")
	if err != nil {
		return badRequest(err)
	}

	tenant, err := app.FindTenantByID(server.Config, tenantID)
	if err != nil {
		return notFound("tenant not found")
	}

	client, err := app.FindClient(tenant, clientID)
	if err != nil {
		return notFound("client not found")
	}

	content := fmt.Sprintf(
		"AzureAd__Instance=https://%s/\nAzureAd__TenantId=%s\nAzureAd__ClientId=%s\n",
		request.Host, tenant.TenantID(), client.ClientID(),
	)

	if !client.ClientSecret().IsEmpty() {
		content += fmt.Sprintf("AzureAd__ClientSecret=%s\n", client.ClientSecret())
	}

	return HTTPResponse{
		Status:      http.StatusOK,
		Body:        []byte(content),
		ContentType: "text/plain; charset=utf-8",
		Headers:     map[string]string{"Content-Disposition": fmt.Sprintf("inline; filename=%q", client.Name().String()+"-client.env")},
	}
}

func parseTenantAndAppID(path, mid, suffix string) (domain.TenantID, domain.ClientID, error) {
	path = strings.TrimPrefix(path, "/config/")

	midIndex := strings.Index(path, mid)
	if midIndex < 0 {
		return domain.TenantID{}, domain.ClientID{}, domain.ErrTenantNotFound
	}

	tenantStr := path[:midIndex]
	rest := strings.TrimSuffix(strings.TrimPrefix(path[midIndex:], mid), suffix)

	tenantUUID, err := uuid.Parse(tenantStr)
	if err != nil {
		return domain.TenantID{}, domain.ClientID{}, domain.ErrTenantNotFound
	}

	appUUID, err := uuid.Parse(rest)
	if err != nil {
		return domain.TenantID{}, domain.ClientID{}, domain.ErrAppNotFound
	}

	return domain.TenantIDFromUUID(tenantUUID), domain.ClientIDFromUUID(appUUID), nil
}
