package main

import (
	"log"
	"net/http"
)

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
