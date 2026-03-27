package main

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

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

	// Ensure request bodies are capped even when this handler is called directly.
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

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
