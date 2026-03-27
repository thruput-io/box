package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.URL.Path == "/_health" {
			if strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug") {
				log.Printf("[DEBUG] %s %q %s", r.Method, r.URL.Path, time.Since(start))
			}
			return
		}
		log.Printf("%s %q %s", r.Method, r.URL.Path, time.Since(start))
	})
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
