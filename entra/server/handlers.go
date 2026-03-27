package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.URL.Path == "/_health" {
			if strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug") {
				log.Printf("[DEBUG] %s %s %s", r.Method, r.URL.Path, time.Since(start))
			}
			return
		}
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := indexTemplate.Execute(w, configData); err != nil {
		log.Printf("failed to render index template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func maxBytesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Limit request body to 1MB to prevent DoS
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-client-ver, x-client-sku, x-client-os, x-client-cpu, x-client-os-ver, x-ms-client-request-id, client-request-id")
		w.Header().Set("Access-Control-Expose-Headers", "x-ms-request-id, client-request-id")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validateRedirectURI(redirectURI string, allowedURIs []string) bool {
	if redirectURI == "" {
		log.Printf("Redirect URI is empty")
		return false
	}
	if len(allowedURIs) == 0 {
		return true
	}
	for _, allowed := range allowedURIs {
		if redirectURI == allowed {
			return true
		}
	}
	log.Printf("Redirect URI %s not in allowed list: %v", redirectURI, allowedURIs)
	return false
}

func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

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
		TokenEndpoint:                     fmt.Sprintf("%s/oauth2/token", tenantURL),
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post", "private_key_jwt", "client_secret_basic", "self_signed_tls_client_auth"},
		JwksURI:                           fmt.Sprintf("%s/discovery/keys", tenantURL),
		ResponseModesSupported:            []string{"query", "fragment", "form_post"},
		SubjectTypesSupported:             []string{"pairwise"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		ResponseTypesSupported:            []string{"code", "id_token", "code id_token", "id_token token"},
		ScopesSupported:                   []string{"openid", "profile", "email", "offline_access"},
		Issuer:                            issuer,
		RequestURIParameterSupported:      false,
		UserinfoEndpoint:                  "https://graph.microsoft.com/oidc/userinfo",
		AuthorizationEndpoint:             fmt.Sprintf("%s/oauth2/authorize", tenantURL),
		DeviceAuthorizationEndpoint:       fmt.Sprintf("%s/oauth2/devicecode", tenantURL),
		HTTPLogoutSupported:               true,
		FrontchannelLogoutSupported:       true,
		EndSessionEndpoint:                fmt.Sprintf("%s/oauth2/logout", tenantURL),
		ClaimsSupported: []string{
			"sub", "iss", "cloud_instance_name", "cloud_instance_host_name",
			"cloud_graph_host_name", "msgraph_host", "aud", "exp", "iat",
			"auth_time", "acr", "nonce", "preferred_username", "name", "tid",
			"ver", "at_hash", "c_hash", "email",
		},
		KerberosEndpoint:                      fmt.Sprintf("%s/kerberos", tenantURL),
		TLSClientCertificateBoundAccessTokens: true,
		TenantRegionScope:                     "NA",
		CloudInstanceName:                     "microsoftonline.com",
		CloudGraphHostName:                    "graph.windows.net",
		MsgraphHost:                           "graph.microsoft.com",
		RbacURL:                               "https://pas.windows.net",
	}
	resp.MtlsEndpointAliases.TokenEndpoint = fmt.Sprintf("https://mtlsauth.microsoft.com/%s/oauth2/token", tenant)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding callHome response: %v", err)
	}
}

func sendOAuthError(w http.ResponseWriter, r *http.Request, errCode string, desc string, status int) {
	correlationID := r.Header.Get("client-request-id")
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("client-request-id", correlationID)
	w.Header().Set("x-ms-request-id", correlationID)
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(OAuthError{
		Error:            errCode,
		ErrorDescription: desc,
		TraceID:          uuid.New().String(),
		CorrelationID:    correlationID,
		Timestamp:        time.Now().Format("2006-01-02 15:04:05Z"),
	}); err != nil {
		log.Printf("Error encoding OAuth error: %v", err)
	}
}

func token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendOAuthError(w, r, "invalid_request", err.Error(), http.StatusBadRequest)
		return
	}

	// Basic length validation to prevent extremely long strings from being processed
	for key, values := range r.Form {
		for _, v := range values {
			if len(v) > 2048 {
				sendOAuthError(w, r, "invalid_request", fmt.Sprintf("field %s is too long", key), http.StatusBadRequest)
				return
			}
		}
	}

	grantType := r.Form.Get("grant_type")
	clientID, _ := uuid.Parse(r.Form.Get("client_id"))
	username := r.Form.Get("username")
	scope := r.Form.Get("scope")
	clientSecret := r.Form.Get("client_secret")

	// Determine tenant and version
	isV2 := strings.Contains(r.URL.Path, "/v2.0")
	tenantID := r.PathValue("tenant")
	if tenantID == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 && parts[0] != "token" && parts[0] != "common" && parts[0] != "oauth2" {
			tenantID = parts[0]
		}
	}
	activeTenant := findTenant(tenantID, &configData)
	activeTenantID := activeTenant.TenantID.String()

	baseURL := getBaseURL(r)
	issuer := fmt.Sprintf("%s/%s/v2.0", baseURL, activeTenantID)
	version := "2.0"
	if !isV2 {
		issuer = fmt.Sprintf("%s/%s", baseURL, activeTenantID)
		version = "1.0"
	}

	// Default values
	sub := clientID.String()
	var roles []string
	name := "Mock User"
	email := "user@example.com"
	nonce := ""

	var activeUser *User
	var activeClient *Client

	// Find Client
	if clientID != uuid.Nil {
		for i := range activeTenant.Clients {
			if activeTenant.Clients[i].ClientID == clientID {
				activeClient = &activeTenant.Clients[i]
				break
			}
		}
	}

	if grantType == "password" {
		if username == "" {
			sendOAuthError(w, r, "invalid_request", "missing username", http.StatusBadRequest)
			return
		}
		for _, u := range activeTenant.Users {
			if u.Username == username && subtle.ConstantTimeCompare([]byte(u.Password), []byte(r.Form.Get("password"))) == 1 {
				activeUser = &u
				break
			}
		}
		if activeUser == nil {
			sendOAuthError(w, r, "invalid_grant", "Invalid username or password", http.StatusUnauthorized)
			return
		}
		// Verify secret if provided (for confidential clients)
		if activeClient != nil && activeClient.ClientSecret != "" && clientSecret != "" && subtle.ConstantTimeCompare([]byte(activeClient.ClientSecret), []byte(clientSecret)) != 1 {
			sendOAuthError(w, r, "invalid_client", "Invalid client secret", http.StatusUnauthorized)
			return
		}
	} else if grantType == "client_credentials" {
		if activeClient == nil {
			sendOAuthError(w, r, "invalid_client", "Client not found", http.StatusUnauthorized)
			return
		}
		// Verify secret. Public clients (no secret) cannot use client_credentials
		if activeClient.ClientSecret == "" || subtle.ConstantTimeCompare([]byte(clientSecret), []byte(activeClient.ClientSecret)) != 1 {
			sendOAuthError(w, r, "invalid_client", "Invalid client secret", http.StatusUnauthorized)
			return
		}
	} else if grantType == "authorization_code" {
		code := r.Form.Get("code")
		t, err := jwt.Parse(code, func(_ *jwt.Token) (interface{}, error) {
			return privateKey.Public(), nil
		})
		if err != nil || !t.Valid {
			sendOAuthError(w, r, "invalid_grant", "Invalid or expired authorization code", http.StatusUnauthorized)
			return
		}
		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			sendOAuthError(w, r, "invalid_grant", "Invalid claims in authorization code", http.StatusUnauthorized)
			return
		}
		if r.Form.Get("redirect_uri") != "" && r.Form.Get("redirect_uri") != claims["redirect_uri"].(string) {
			sendOAuthError(w, r, "invalid_grant", "redirect_uri mismatch", http.StatusUnauthorized)
			return
		}
		sub = claims["sub"].(string)
		if clientID == uuid.Nil {
			clientID, _ = uuid.Parse(claims["client_id"].(string))
		}
		if scope == "" {
			if s, ok := claims["scope"].(string); ok {
				scope = s
			}
		}
		if n, ok := claims["nonce"].(string); ok {
			nonce = n
		}
		// Find user
		for i := range activeTenant.Users {
			if activeTenant.Users[i].ID.String() == sub {
				activeUser = &activeTenant.Users[i]
				break
			}
		}
	} else if grantType == "refresh_token" {
		refreshToken := r.Form.Get("refresh_token")
		t, err := jwt.Parse(refreshToken, func(_ *jwt.Token) (interface{}, error) {
			return privateKey.Public(), nil
		})
		if err != nil || !t.Valid {
			sendOAuthError(w, r, "invalid_grant", "Invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok || claims["typ"] != "Refresh" {
			sendOAuthError(w, r, "invalid_grant", "Invalid refresh token", http.StatusUnauthorized)
			return
		}
		sub = claims["sub"].(string)
		if clientID == uuid.Nil {
			clientID, _ = uuid.Parse(claims["client_id"].(string))
		}
		if scope == "" {
			if s, ok := claims["scope"].(string); ok {
				scope = s
			}
		}
		// Find user
		for i := range activeTenant.Users {
			if activeTenant.Users[i].ID.String() == sub {
				activeUser = &activeTenant.Users[i]
				break
			}
		}
		// If it was a client-only refresh token, sub will be client ID
		if activeUser == nil {
			for i := range activeTenant.Clients {
				if activeTenant.Clients[i].ClientID.String() == sub {
					activeClient = &activeTenant.Clients[i]
					break
				}
			}
		}
	} else {
		sendOAuthError(w, r, "unsupported_grant_type", fmt.Sprintf("grant_type %s is not supported", grantType), http.StatusBadRequest)
		return
	}

	if activeUser != nil {
		sub = activeUser.ID.String()
		name = activeUser.DisplayName
		email = activeUser.Email
	}

	targetAudience, targetAppIDs := resolveAudience(activeTenant, scope)
	roles = resolveRoles(activeTenant, activeClient, activeUser, targetAppIDs, strings.Split(scope, " "))

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": issuer,
		"sub": sub,
		"aud": targetAudience,
		"exp": now.Add(24 * time.Hour).Unix(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"jti": uuid.New().String(),
		"tid": activeTenantID,
		"ver": version,
		"oid": sub,
	}

	if activeClient != nil {
		claims["azp"] = activeClient.ClientID.String()
		claims["azpacls"] = "0"
		// appid is often present in v2.0 tokens for backward compatibility
		claims["appid"] = activeClient.ClientID.String()
	}

	if activeUser != nil {
		claims["name"] = name
		claims["preferred_username"] = email
		claims["email"] = email
		claims["unique_name"] = email
		// For user flows, scp contains the scopes granted
		if targetAudience == "00000003-0000-0000-c000-000000000000" || targetAudience == "https://graph.microsoft.com" {
			claims["scp"] = strings.Join(roles, " ")
		} else {
			// Include original scope if it was a user flow
			claims["scp"] = scope
		}
	}

	claims["roles"] = roles

	if nonce != "" {
		claims["nonce"] = nonce
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "1"
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		sendOAuthError(w, r, "server_error", "failed to sign token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"access_token": tokenString,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        scope,
	}

	// Generate ID Token if openid scope is present
	requestedScopesList := strings.Split(scope, " ")
	hasOpenID := false
	hasOfflineAccess := false
	for _, s := range requestedScopesList {
		if s == "openid" {
			hasOpenID = true
		}
		if s == "offline_access" {
			hasOfflineAccess = true
		}
	}

	if hasOpenID && activeUser != nil {
		idClaims := jwt.MapClaims{
			"iss":                issuer,
			"sub":                sub,
			"aud":                clientID.String(),
			"exp":                now.Add(24 * time.Hour).Unix(),
			"iat":                now.Unix(),
			"nbf":                now.Unix(),
			"tid":                activeTenantID,
			"ver":                version,
			"oid":                sub,
			"name":               name,
			"preferred_username": email,
			"email":              email,
		}
		if nonce != "" {
			idClaims["nonce"] = nonce
		}
		idToken := jwt.NewWithClaims(jwt.SigningMethodRS256, idClaims)
		idToken.Header["kid"] = "1"
		idTokenString, _ := idToken.SignedString(privateKey)
		response["id_token"] = idTokenString
	}

	// Generate Refresh Token if offline_access is present or it's a refresh_token grant
	if hasOfflineAccess || grantType == "refresh_token" {
		refreshClaims := jwt.MapClaims{
			"iss":       issuer,
			"sub":       sub,
			"aud":       issuer,
			"exp":       now.Add(90 * 24 * time.Hour).Unix(),
			"iat":       now.Unix(),
			"client_id": clientID.String(),
			"scope":     scope,
			"tid":       activeTenantID,
			"typ":       "Refresh",
		}
		refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
		refreshToken.Header["kid"] = "1"
		refreshTokenString, _ := refreshToken.SignedString(privateKey)
		response["refresh_token"] = refreshTokenString
	}

	correlationID := r.Header.Get("client-request-id")
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("client-request-id", correlationID)
	w.Header().Set("x-ms-request-id", correlationID)
	if activeUser != nil {
		clientInfo := map[string]string{
			"uid":  activeUser.ID.String(),
			"utid": activeTenantID,
		}
		clientInfoJSON, _ := json.Marshal(clientInfo)
		response["client_info"] = base64UrlEncode(clientInfoJSON)
	}
	jsonResp, _ := json.Marshal(response)
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Error writing token response: %v", err)
	}
}

func jwks(w http.ResponseWriter, _ *http.Request) {
	n := privateKey.N.Bytes()
	e := big.NewInt(int64(privateKey.E)).Bytes()

	keys := map[string]interface{}{
		"keys": []map[string]interface{}{
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
	if err := json.NewEncoder(w).Encode(keys); err != nil {
		log.Printf("Error encoding JWKS response: %v", err)
	}
}

func authorize(w http.ResponseWriter, r *http.Request) {
	// Basic query parameter length validation
	for key, values := range r.URL.Query() {
		for _, v := range values {
			if len(v) > 2048 {
				sendOAuthError(w, r, "invalid_request", fmt.Sprintf("parameter %s is too long", key), http.StatusBadRequest)
				return
			}
		}
	}

	clientID, _ := uuid.Parse(r.URL.Query().Get("client_id"))
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	scope := r.URL.Query().Get("scope")
	responseType := r.URL.Query().Get("response_type")
	nonce := r.URL.Query().Get("nonce")
	responseMode := r.URL.Query().Get("response_mode")

	tenantID := ""
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) > 0 && parts[0] != "authorize" && parts[0] != "common" && parts[0] != "oauth2" && parts[0] != "v2.0" {
		tenantID = parts[0]
	}

	activeTenant := findTenant(tenantID, &configData)
	var allowedURIs []string
	found := false
	for _, c := range activeTenant.Clients {
		if c.ClientID == clientID {
			allowedURIs = c.RedirectUrls
			found = true
			break
		}
	}
	if !found {
		for _, reg := range activeTenant.AppRegistrations {
			if reg.ClientID == clientID {
				allowedURIs = reg.RedirectUrls
				found = true
				break
			}
		}
	}
	if !found {
		sendOAuthError(w, r, "invalid_request", "Client not found", http.StatusBadRequest)
		return
	}

	if !validateRedirectURI(redirectURI, allowedURIs) {
		sendOAuthError(w, r, "invalid_request", "invalid_redirect_uri", http.StatusBadRequest)
		return
	}

	// Build user display list for the login page
	var usersDisplay []UserDisplayInfo
	if activeTenant != nil {
		// Find the client we are logging into
		var currentClient *Client
		for _, c := range activeTenant.Clients {
			if c.ClientID == clientID {
				currentClient = &c
				break
			}
		}

		for _, u := range activeTenant.Users {
			var roleDescriptions []string
			if currentClient != nil {
				// Map to collect roles per application
				appRoles := make(map[uuid.UUID][]string)
				for _, gName := range u.Groups {
					for _, gra := range currentClient.GroupRoleAssignments {
						if gra.GroupName == gName {
							appRoles[gra.ApplicationID] = append(appRoles[gra.ApplicationID], gra.Roles...)
						}
					}
				}

				for appID, roles := range appRoles {
					// Find application name
					appName := appID.String()
					for _, reg := range activeTenant.AppRegistrations {
						if reg.ClientID == appID {
							appName = reg.Name
							break
						}
					}
					roleDescriptions = append(roleDescriptions, fmt.Sprintf("%s: %s", appName, strings.Join(roles, ", ")))
				}
			}

			usersDisplay = append(usersDisplay, UserDisplayInfo{
				Username:    u.Username,
				Password:    u.Password,
				DisplayName: u.DisplayName,
				Roles:       roleDescriptions,
			})
		}
	}

	data := AuthRequest{
		ClientID:     clientID,
		RedirectURI:  redirectURI,
		State:        state,
		Scope:        scope,
		ResponseType: responseType,
		Tenant:       tenantID,
		Nonce:        nonce,
		ResponseMode: responseMode,
		Users:        usersDisplay,
	}

	if err := loginTemplate.Execute(w, data); err != nil {
		sendOAuthError(w, r, "server_error", "failed to render login page", http.StatusInternalServerError)
		return
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendOAuthError(w, r, "invalid_request", err.Error(), http.StatusBadRequest)
		return
	}

	// Basic length validation
	for key, values := range r.Form {
		for _, v := range values {
			if len(v) > 2048 {
				sendOAuthError(w, r, "invalid_request", fmt.Sprintf("field %s is too long", key), http.StatusBadRequest)
				return
			}
		}
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")
	clientID := r.Form.Get("client_id")
	redirectURI := r.Form.Get("redirect_uri")
	state := r.Form.Get("state")
	scope := r.Form.Get("scope")
	tenantID := r.Form.Get("tenant")
	nonce := r.Form.Get("nonce")
	responseMode := r.Form.Get("response_mode")

	activeTenant := findTenant(tenantID, &configData)

	var authenticatedUser *User
	var allowedURIs []string
	clientFound := false

	if activeTenant != nil {
		for _, u := range activeTenant.Users {
			if u.Username == username && subtle.ConstantTimeCompare([]byte(u.Password), []byte(password)) == 1 {
				authenticatedUser = &u
				break
			}
		}
		cID, _ := uuid.Parse(clientID)
		for _, c := range activeTenant.Clients {
			if c.ClientID == cID {
				allowedURIs = c.RedirectUrls
				clientFound = true
				break
			}
		}
		if !clientFound {
			for _, reg := range activeTenant.AppRegistrations {
				if reg.ClientID == cID {
					allowedURIs = reg.RedirectUrls
					clientFound = true
					break
				}
			}
		}
	}

	if !clientFound {
		sendOAuthError(w, r, "invalid_request", "Client not found", http.StatusBadRequest)
		return
	}

	if !validateRedirectURI(redirectURI, allowedURIs) {
		sendOAuthError(w, r, "invalid_request", "invalid_redirect_uri", http.StatusBadRequest)
		return
	}

	if authenticatedUser == nil {
		sendOAuthError(w, r, "invalid_grant", "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate auth code (as a JWT signed by our private key)
	claims := jwt.MapClaims{
		"sub":          authenticatedUser.ID.String(),
		"client_id":    clientID,
		"redirect_uri": redirectURI,
		"scope":        scope,
		"tenant":       tenantID,
		"nonce":        nonce,
		"exp":          time.Now().Add(5 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	authCode, err := token.SignedString(privateKey)
	if err != nil {
		sendOAuthError(w, r, "server_error", "failed to generate auth code", http.StatusInternalServerError)
		return
	}

	// Redirect back
	target, err := url.Parse(redirectURI)
	if err != nil {
		sendOAuthError(w, r, "invalid_request", "invalid redirect_uri", http.StatusBadRequest)
		return
	}

	values := url.Values{}
	values.Set("code", authCode)
	if state != "" {
		values.Set("state", state)
	}

	var finalRedirectURL string
	if responseMode == "fragment" {
		target.RawQuery = ""
		finalRedirectURL = target.String() + "#" + values.Encode()
	} else {
		q := target.Query()
		for k, v := range values {
			q.Set(k, v[0])
		}
		target.RawQuery = q.Encode()
		finalRedirectURL = target.String()
	}

	log.Printf("Redirecting to: %s", finalRedirectURL)
	http.Redirect(w, r, finalRedirectURL, http.StatusFound)
}
