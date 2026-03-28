package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
)

func validateRedirectURI(redirectURI string, allowedURIs []string) bool {
	if redirectURI == "" {
		log.Printf("Redirect URI is empty")

		return false
	}

	if len(allowedURIs) == 0 {
		return true
	}

	if slices.Contains(allowedURIs, redirectURI) {
		return true
	}

	log.Printf("Redirect URI %q not in allowed list: %v", redirectURI, allowedURIs)

	return false
}

func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func sendOAuthError(w http.ResponseWriter, r *http.Request, errCode string, desc string, status int) {
	correlationID := r.Header.Get("Client-Request-Id")
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Client-Request-Id", correlationID)
	w.Header().Set("X-Ms-Request-Id", correlationID)
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(OAuthError{
		Error:            errCode,
		ErrorDescription: desc,
		TraceID:          uuid.New().String(),
		CorrelationID:    correlationID,
		Timestamp:        time.Now().Format("2006-01-02 15:04:05Z"),
	})
	if err != nil {
		log.Printf("Error encoding OAuth error: %v", err)
	}
}
