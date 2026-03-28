package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
)

func discovery(w http.ResponseWriter, r *http.Request) {
	tenant := r.PathValue("tenant")
	if tenant == "" {
		tenant = "common"
	}

	baseURL := getBaseURL(r)
	isV2 := strings.Contains(r.URL.Path, "/v2.0")

	tenantURL := fmt.Sprintf("%s/%s", baseURL, tenant)
	if isV2 {
		tenantURL = fmt.Sprintf("%s/%s/v2.0", baseURL, tenant)
	}

	activeTenant := findTenant(tenant, &configData)
	activeTenantID := activeTenant.TenantID.String()

	issuer := tenantURL
	if !isV2 {
		// v1.0 issuer is typically sts.windows.net, but for mocks we can stay consistent with baseURL
		issuer = fmt.Sprintf("%s/%s", baseURL, activeTenantID)
	}

	resp := DiscoveryResponse{
		TokenEndpoint:                     tenantURL + "/oauth2/token",
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post", "private_key_jwt", "client_secret_basic", "self_signed_tls_client_auth"},
		JwksURI:                           tenantURL + "/discovery/keys",
		ResponseModesSupported:            []string{"query", "fragment", "form_post"},
		SubjectTypesSupported:             []string{"pairwise"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		ResponseTypesSupported:            []string{"code", "id_token", "code id_token", "id_token token"},
		ScopesSupported:                   []string{"openid", "profile", "email", "offline_access"},
		Issuer:                            issuer,
		RequestURIParameterSupported:      false,
		UserinfoEndpoint:                  "https://graph.microsoft.com/oidc/userinfo",
		AuthorizationEndpoint:             tenantURL + "/oauth2/authorize",
		DeviceAuthorizationEndpoint:       tenantURL + "/oauth2/devicecode",
		HTTPLogoutSupported:               true,
		FrontchannelLogoutSupported:       true,
		EndSessionEndpoint:                tenantURL + "/oauth2/logout",
		ClaimsSupported: []string{
			"sub", "iss", "cloud_instance_name", "cloud_instance_host_name",
			"cloud_graph_host_name", "msgraph_host", "aud", "exp", "iat",
			"auth_time", "acr", "nonce", "preferred_username", "name", "tid",
			"ver", "at_hash", "c_hash", "email",
		},
		KerberosEndpoint:                      tenantURL + "/kerberos",
		TLSClientCertificateBoundAccessTokens: true,
		TenantRegionScope:                     "NA",
		CloudInstanceName:                     "microsoftonline.com",
		CloudGraphHostName:                    "graph.windows.net",
		MsgraphHost:                           "graph.microsoft.com",
		RbacURL:                               "https://pas.windows.net",
	}
	resp.MtlsEndpointAliases.TokenEndpoint = fmt.Sprintf("https://mtlsauth.microsoft.com/%s/oauth2/token", tenant)

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding discovery response: %v", err)
	}
}

func callHome(w http.ResponseWriter, r *http.Request) {
	baseURL := getBaseURL(r)
	resp := CallHomeResponse{
		TenantDiscoveryEndpoint: baseURL,
		APIVersion:              "1.1",
	}
	resp.Metadata = append(resp.Metadata, struct {
		PreferredNetwork string   `json:"preferred_network"`
		PreferredCache   string   `json:"preferred_cache"`
		Aliases          []string `json:"aliases"`
	}{
		PreferredNetwork: r.Host,
		PreferredCache:   r.Host,
		Aliases:          []string{r.Host},
	})

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding callHome response: %v", err)
	}
}

func jwks(w http.ResponseWriter, _ *http.Request) {
	n := privateKey.N.Bytes()
	e := big.NewInt(int64(privateKey.E)).Bytes()

	keys := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"use": "sig",
				"kid": "1",
				"alg": "RS256",
				"n":   base64UrlEncode(n),
				"e":   base64UrlEncode(e),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(keys)
	if err != nil {
		log.Printf("Error encoding JWKS response: %v", err)
	}
}
