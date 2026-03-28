package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// configRaw serves the raw Config.yaml file.
func configRaw(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "Config.yaml")
}

// configCsharpApp serves C# container apps environment variables for an app registration (MSAL host config).
// Path: /config/{tenantId}/app/{appId}/csharp.
func configCsharpApp(w http.ResponseWriter, r *http.Request) {
	tenantID, appID, ok := parseTenantApp(w, r, "/config/", "/app/", "/csharp")
	if !ok {
		return
	}

	tenant := findTenantStrict(tenantID, &configData)
	if tenant == nil {
		http.Error(w, "tenant not found", http.StatusNotFound)

		return
	}

	var app *AppRegistration

	for i := range tenant.AppRegistrations {
		if tenant.AppRegistrations[i].ClientID == appID {
			app = &tenant.AppRegistrations[i]

			break
		}
	}

	if app == nil {
		http.Error(w, "app registration not found", http.StatusNotFound)

		return
	}

	host := r.Host

	var sb strings.Builder

	fmt.Fprintf(&sb, "AzureAd__Instance=https://%s/\n", host)
	fmt.Fprintf(&sb, "AzureAd__TenantId=%s\n", tenant.TenantID)
	fmt.Fprintf(&sb, "AzureAd__ClientId=%s\n", app.ClientID)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s-appsettings.env\"", app.Name))
	_, _ = fmt.Fprint(w, sb.String())
}

// configJsApp serves JavaScript MSAL configuration for an app registration.
// Path: /config/{tenantId}/app/{appId}/js.
func configJsApp(w http.ResponseWriter, r *http.Request) {
	tenantID, appID, ok := parseTenantApp(w, r, "/config/", "/app/", "/js")
	if !ok {
		return
	}

	tenant := findTenantStrict(tenantID, &configData)
	if tenant == nil {
		http.Error(w, "tenant not found", http.StatusNotFound)

		return
	}

	var app *AppRegistration

	for i := range tenant.AppRegistrations {
		if tenant.AppRegistrations[i].ClientID == appID {
			app = &tenant.AppRegistrations[i]

			break
		}
	}

	if app == nil {
		http.Error(w, "app registration not found", http.StatusNotFound)

		return
	}

	host := r.Host

	var sb strings.Builder

	sb.WriteString("const msalConfig = {\n")
	sb.WriteString("  auth: {\n")
	fmt.Fprintf(&sb, "    clientId: \"%s\",\n", app.ClientID)
	fmt.Fprintf(&sb, "    authority: \"https://%s/%s\",\n", host, tenant.TenantID)
	fmt.Fprintf(&sb, "    knownAuthorities: [\"%s\"],\n", host)
	sb.WriteString("  },\n")
	sb.WriteString("};\n")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s-msal-config.js\"", app.Name))
	_, _ = fmt.Fprint(w, sb.String())
}

// configCsharpClient serves C# environment variables for a client (client credentials config).
// Path: /config/{tenantId}/client/{clientId}/csharp.
func configCsharpClient(w http.ResponseWriter, r *http.Request) {
	tenantID, clientID, ok := parseTenantApp(w, r, "/config/", "/client/", "/csharp")
	if !ok {
		return
	}

	tenant := findTenantStrict(tenantID, &configData)
	if tenant == nil {
		http.Error(w, "tenant not found", http.StatusNotFound)

		return
	}

	var client *Client

	for i := range tenant.Clients {
		if tenant.Clients[i].ClientID == clientID {
			client = &tenant.Clients[i]

			break
		}
	}

	if client == nil {
		http.Error(w, "client not found", http.StatusNotFound)

		return
	}

	host := r.Host

	var sb strings.Builder

	fmt.Fprintf(&sb, "AzureAd__Instance=https://%s/\n", host)
	fmt.Fprintf(&sb, "AzureAd__TenantId=%s\n", tenant.TenantID)
	fmt.Fprintf(&sb, "AzureAd__ClientId=%s\n", client.ClientID)

	if client.ClientSecret != "" {
		fmt.Fprintf(&sb, "AzureAd__ClientSecret=%s\n", client.ClientSecret)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s-client.env\"", client.Name))
	_, _ = fmt.Fprint(w, sb.String())
}

// parseTenantApp extracts two UUIDs from a path of the form:
// {prefix}{tenantId}{mid}{secondId}{suffix}.
func parseTenantApp(w http.ResponseWriter, r *http.Request, prefix, mid, suffix string) (uuid.UUID, uuid.UUID, bool) {
	path := strings.TrimPrefix(r.URL.Path, prefix)

	midIdx := strings.Index(path, mid)
	if midIdx < 0 {
		http.Error(w, "invalid path", http.StatusBadRequest)

		return uuid.UUID{}, uuid.UUID{}, false
	}

	tenantStr := path[:midIdx]
	rest := strings.TrimSuffix(strings.TrimPrefix(path[midIdx:], mid), suffix)

	tid, err := uuid.Parse(tenantStr)
	if err != nil {
		http.Error(w, "invalid tenantId", http.StatusBadRequest)

		return uuid.UUID{}, uuid.UUID{}, false
	}

	sid, err := uuid.Parse(rest)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)

		return uuid.UUID{}, uuid.UUID{}, false
	}

	return tid, sid, true
}
