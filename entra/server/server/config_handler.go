package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"identity/app"
	"identity/domain"
)

const (
	pathApp                  = "/app/"
	pathClient               = "/client/"
	msgTenantNotFound        = "tenant not found"
	headerContentDisposition = "Content-Disposition"
)

func configHandler(request *http.Request, server *Server) Response {
	path := request.URL.Path

	switch {
	case path == "/config/raw":
		return configRawHandler(request, server)
	case strings.Contains(path, pathApp) && strings.HasSuffix(path, pathCsharp):
		return configCsharpAppHandler(request, server)
	case strings.Contains(path, pathApp) && strings.HasSuffix(path, pathJS):
		return configJsAppHandler(request, server)
	case strings.Contains(path, pathClient) && strings.HasSuffix(path, pathCsharp):
		return configCsharpClientHandler(request, server)
	default:
		return notFound("config endpoint not found")
	}
}

func configRawHandler(_ *http.Request, _ *Server) Response {
	return Response{
		Status:      http.StatusMovedPermanently,
		ContentType: "",
		Body:        nil,
		Headers:     map[string]string{"Location": "/DefaultConfig.yaml"},
	}
}

func configCsharpAppHandler(request *http.Request, server *Server) Response {
	tenantID, appID, err := parseTenantAndAppID(request.URL.Path, pathApp, pathCsharp)
	if err != nil {
		return badRequest(err)
	}

	tenant, err := app.FindTenantByID(server.Config, tenantID)
	if err != nil {
		return notFound(msgTenantNotFound)
	}

	registration, err := app.FindAppRegistration(tenant, appID)
	if err != nil {
		return notFound("app registration not found")
	}

	content := fmt.Sprintf(
		"AzureAd__Instance=https://%s/\nAzureAd__TenantId=%s\nAzureAd__ClientId=%s\n",
		request.Host, tenant.TenantID(), registration.ClientID(),
	)
	disposition := fmt.Sprintf(fmtDisposition, registration.Name().String()+"-appsettings.env")

	return Response{
		Status:      http.StatusOK,
		Body:        []byte(content),
		ContentType: contentTypePlain,
		Headers:     map[string]string{headerContentDisposition: disposition},
	}
}

func configJsAppHandler(request *http.Request, server *Server) Response {
	tenantID, appID, err := parseTenantAndAppID(request.URL.Path, pathApp, pathJS)
	if err != nil {
		return badRequest(err)
	}

	tenant, err := app.FindTenantByID(server.Config, tenantID)
	if err != nil {
		return notFound(msgTenantNotFound)
	}

	registration, err := app.FindAppRegistration(tenant, appID)
	if err != nil {
		return notFound("app registration not found")
	}

	msalFmt := "const msalConfig = {\n  auth: {\n    clientId: %q," +
		"\n    authority: \"https://%s/%s\",\n    knownAuthorities: [%q],\n  },\n};\n"
	content := fmt.Sprintf(msalFmt, registration.ClientID(), request.Host, tenant.TenantID(), request.Host)
	disposition := fmt.Sprintf(fmtDisposition, registration.Name().String()+"-msal-config.js")

	return Response{
		Status:      http.StatusOK,
		Body:        []byte(content),
		ContentType: contentTypePlain,
		Headers:     map[string]string{headerContentDisposition: disposition},
	}
}

func configCsharpClientHandler(request *http.Request, server *Server) Response {
	tenantID, clientID, err := parseTenantAndAppID(request.URL.Path, pathClient, pathCsharp)
	if err != nil {
		return badRequest(err)
	}

	tenant, err := app.FindTenantByID(server.Config, tenantID)
	if err != nil {
		return notFound(msgTenantNotFound)
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

	disposition := fmt.Sprintf(fmtDisposition, client.Name().String()+"-client.env")

	return Response{
		Status:      http.StatusOK,
		Body:        []byte(content),
		ContentType: contentTypePlain,
		Headers:     map[string]string{headerContentDisposition: disposition},
	}
}

func parseTenantAndAppID(path, midSegment, suffix string) (domain.TenantID, domain.ClientID, error) {
	path = strings.TrimPrefix(path, "/config/")

	midIndex := strings.Index(path, midSegment)
	if midIndex < minValidIndex {
		return domain.TenantID{}, domain.ClientID{}, domain.ErrTenantNotFound
	}

	tenantStr := path[:midIndex]
	rest := strings.TrimSuffix(strings.TrimPrefix(path[midIndex:], midSegment), suffix)

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
