package http

import (
	"log"
	"net/http"
	"time"
)

func buildRoutes(server *Server) http.Handler {
	mux := http.NewServeMux()

	adapt := func(handler func(*http.Request, *Server) HTTPResponse) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			write(writer, handler(request, server))
		}
	}

	mux.HandleFunc("/health", adapt(healthHandler))
	mux.HandleFunc("/", adapt(indexHandler))

	// OAuth2 / OIDC — tenant-scoped
	mux.HandleFunc("/{tenant}/oauth2/v2.0/authorize", adapt(authorizeHandler))
	mux.HandleFunc("/{tenant}/oauth2/authorize", adapt(authorizeHandler))
	mux.HandleFunc("/{tenant}/oauth2/v2.0/token", adapt(tokenHandler))
	mux.HandleFunc("/{tenant}/oauth2/token", adapt(tokenHandler))
	mux.HandleFunc("/{tenant}/login", adapt(loginHandler))

	// OAuth2 / OIDC — common and bare paths
	mux.HandleFunc("/common/oauth2/v2.0/authorize", adapt(authorizeHandler))
	mux.HandleFunc("/common/oauth2/authorize", adapt(authorizeHandler))
	mux.HandleFunc("/common/oauth2/v2.0/token", adapt(tokenHandler))
	mux.HandleFunc("/common/oauth2/token", adapt(tokenHandler))
	mux.HandleFunc("/oauth2/v2.0/authorize", adapt(authorizeHandler))
	mux.HandleFunc("/oauth2/authorize", adapt(authorizeHandler))
	mux.HandleFunc("/authorize", adapt(authorizeHandler))
	mux.HandleFunc("/oauth2/v2.0/token", adapt(tokenHandler))
	mux.HandleFunc("/oauth2/token", adapt(tokenHandler))
	mux.HandleFunc("/token", adapt(tokenHandler))
	mux.HandleFunc("/login", adapt(loginHandler))

	// Discovery
	mux.HandleFunc("/{tenant}/v2.0/.well-known/openid-configuration", adapt(discoveryHandler))
	mux.HandleFunc("/{tenant}/.well-known/openid-configuration", adapt(discoveryHandler))
	mux.HandleFunc("/.well-known/openid-configuration", adapt(discoveryHandler))
	mux.HandleFunc("/{tenant}/discovery/keys", adapt(jwksHandler))
	mux.HandleFunc("/discovery/keys", adapt(jwksHandler))
	mux.HandleFunc("/common/discovery/instance", adapt(callHomeHandler))

	return withMiddleware(mux)
}

func withMiddleware(next http.Handler) http.Handler {
	return logMiddleware(corsMiddleware(next))
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}

		next.ServeHTTP(recorder, request)

		log.Printf("%s %s %d %s", request.Method, request.URL.Path, recorder.status, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		writer.Header().Set("Access-Control-Allow-Origin", origin)
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Client-Request-Id")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)

			return
		}

		next.ServeHTTP(writer, request)
	})
}

type statusRecorder struct {
	status int
	http.ResponseWriter
}

func (recorder *statusRecorder) WriteHeader(status int) {
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}
