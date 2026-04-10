package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"identity/app"
)

func discoveryHandler(request *http.Request, application *app.App) Response {
	tenantIDStr := request.PathValue("tenant")
	if tenantIDStr == "" {
		tenantIDStr = "common"
	}

	isV2 := strings.Contains(request.URL.Path, "/v2.0")

	tenant, err := app.FindTenant(application.Config, tenantIDStr)
	if err != nil {
		return badRequest(err)
	}

	baseURLStr := extractBaseURL(request).Value()
	tenantIDResolved := tenant.TenantID().Value()

	tenantURL := baseURLStr + pathSeparator + tenantIDStr
	if isV2 {
		tenantURL = baseURLStr + pathSeparator + tenantIDStr + "/v2.0"
	}

	issuer := tenantURL
	if !isV2 {
		issuer = baseURLStr + pathSeparator + tenantIDResolved
	}

	body, err := json.Marshal(buildDiscoveryDocument(tenantURL, issuer, tenantIDStr))
	if err != nil {
		return internalError("failed to encode discovery document")
	}

	return okJSON(body)
}

func jwksHandler(_ *http.Request, application *app.App) Response {
	jwks := app.PublicKey(application.Key)

	body, err := json.Marshal(map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"use": "sig",
				"kid": "1",
				"alg": "RS256",
				"n":   jwks.N,
				"e":   jwks.E,
			},
		},
	})
	if err != nil {
		return internalError("failed to encode JWKS")
	}

	return okJSON(body)
}

func callHomeHandler(request *http.Request, _ *app.App) Response {
	baseURL := extractBaseURL(request).Value()

	body, err := json.Marshal(map[string]any{
		"tenant_discovery_endpoint": baseURL,
		"api-version":               "1.1",
		"metadata": []map[string]any{
			{
				"preferred_network": request.Host,
				"preferred_cache":   request.Host,
				"aliases":           []string{request.Host},
			},
		},
	})
	if err != nil {
		return internalError("failed to encode call-home response")
	}

	return okJSON(body)
}

func buildDiscoveryDocument(tenantURL, issuer, tenant string) map[string]any {
	return map[string]any{
		"token_endpoint": tenantURL + "/oauth2/token",
		"token_endpoint_auth_methods_supported": []string{
			"client_secret_post", "private_key_jwt", "client_secret_basic", "self_signed_tls_client_auth",
		},
		"jwks_uri":                              tenantURL + "/discovery/keys",
		"response_modes_supported":              []string{"query", "fragment", "form_post"},
		"subject_types_supported":               []string{"pairwise"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"response_types_supported":              []string{"code", "id_token", "code id_token", "id_token token"},
		"scopes_supported":                      []string{"openid", "profile", "email", "offline_access"},
		"issuer":                                issuer,
		"request_uri_parameter_supported":       false,
		"userinfo_endpoint":                     "https://graph.microsoft.com/oidc/userinfo",
		"authorization_endpoint":                tenantURL + "/oauth2/authorize",
		"device_authorization_endpoint":         tenantURL + "/oauth2/devicecode",
		"http_logout_supported":                 true,
		"frontchannel_logout_supported":         true,
		"end_session_endpoint":                  tenantURL + "/oauth2/logout",
		"claims_supported": []string{
			"sub", "iss", "cloud_instance_name", "cloud_instance_host_name", "cloud_graph_host_name",
			"msgraph_host", "aud", "exp", "iat", "auth_time", "acr", "nonce", "preferred_username",
			"name", "tid", "ver", "at_hash", "c_hash", "email",
		},
		"kerberos_endpoint":     tenantURL + "/kerberos",
		"tenant_region_scope":   "NA",
		"cloud_instance_name":   "microsoftonline.com",
		"cloud_graph_host_name": "graph.windows.net",
		"msgraph_host":          "graph.microsoft.com",
		"rbac_url":              "https://pas.windows.net",
		"tls_client_certificate_bound_access_tokens": true,
		"mtls_endpoint_aliases": map[string]string{
			"token_endpoint": fmt.Sprintf("https://mtlsauth.microsoft.com/%s/oauth2/token", tenant),
		},
	}
}
