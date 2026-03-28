package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// testToken issues a signed JWT with default claims merged with any query parameters.
// Dispatched from the index handler to avoid wildcard conflicts in the mux.
// Path: /test-tokens/{tenantId}/{appId}/{clientId}.
func testToken(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/test-tokens/")

	parts := strings.SplitN(trimmed, "/", 3)
	if len(parts) != 3 {
		http.Error(w, "usage: /test-tokens/{tenantId}/{appId}/{clientId}", http.StatusBadRequest)

		return
	}

	tenantID, appID, clientID := parts[0], parts[1], parts[2]

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "invalid tenantId", http.StatusBadRequest)

		return
	}

	aid, err := uuid.Parse(appID)
	if err != nil {
		http.Error(w, "invalid appId", http.StatusBadRequest)

		return
	}

	cid, err := uuid.Parse(clientID)
	if err != nil {
		http.Error(w, "invalid clientId", http.StatusBadRequest)

		return
	}

	tenant := findTenantStrict(tid, &configData)
	if tenant == nil {
		http.Error(w, "tenant not found", http.StatusNotFound)

		return
	}

	claims := jwt.MapClaims{
		"iss": tid.String(),
		"aud": aid.String(),
		"sub": cid.String(),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
		"tid": tid.String(),
		"oid": cid.String(),
	}

	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			claims[key] = values[0]
		} else {
			claims[key] = values
		}
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := tok.SignedString(privateKey)
	if err != nil {
		http.Error(w, "failed to sign token", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(map[string]string{"access_token": signed}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
