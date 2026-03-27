package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	for _, allowed := range allowedURIs {
		if redirectURI == allowed {
			return true
		}
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
