package main

import (
	"log"
	"net/http"
	"strings"
)

func health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func index(w http.ResponseWriter, r *http.Request) {
	// Dispatch /test-tokens/ here to avoid wildcard conflicts in the mux.
	if strings.HasPrefix(r.URL.Path, "/test-tokens/") {
		testToken(w, r)

		return
	}

	if r.URL.Path != "/" {
		http.NotFound(w, r)

		return
	}

	err := indexTemplate.Execute(w, configData)

	if err != nil {
		log.Printf("failed to render index template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
