package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func token(w http.ResponseWriter, r *http.Request) {
	// Ensure we always cap request bodies even when handlers are called directly (e.g., unit tests).
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
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

	switch grantType {
	default:
		sendOAuthError(w, r, "invalid_request", "unsupported grant_type", http.StatusBadRequest)
		return
	case "password":
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
	case "client_credentials":
		if activeClient == nil {
			sendOAuthError(w, r, "invalid_client", "Client not found", http.StatusUnauthorized)
			return
		}
		// Verify secret. Public clients (no secret) cannot use client_credentials
		if activeClient.ClientSecret == "" || subtle.ConstantTimeCompare([]byte(clientSecret), []byte(activeClient.ClientSecret)) != 1 {
			sendOAuthError(w, r, "invalid_client", "Invalid client secret", http.StatusUnauthorized)
			return
		}
	case "authorization_code":
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
	case "refresh_token":
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
		clientInfoJSON, err := json.Marshal(clientInfo)
		if err != nil {
			log.Printf("Error marshaling client_info: %v", err)
		} else {
			response["client_info"] = base64UrlEncode(clientInfoJSON)
		}
	}
	jsonResp, err := json.Marshal(response)
	if err != nil {
		sendOAuthError(w, r, "server_error", "failed to encode token response", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Error writing token response: %v", err)
	}
}
