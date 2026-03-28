package transport

import (
	"log"
	nethttp "net/http"
	"time"

	"identity/app"
	"identity/http/handlers"
)

// Server is the HTTP host wiring for the identity mock.
// It holds the application dependency-bag and exposes the root handler.
type Server struct {
	App *app.App
}

// Handler returns the root HTTP handler with all routes registered.
func (server *Server) Handler() nethttp.Handler {
	return buildRoutes(server.App)
}

func buildRoutes(application *app.App) nethttp.Handler {
	mux := nethttp.NewServeMux()

	adapt := func(handler func(*nethttp.Request, *app.App) handlers.Response) nethttp.HandlerFunc {
		return func(writer nethttp.ResponseWriter, request *nethttp.Request) {
			handlers.Write(writer, handler(request, application))
		}
	}

	mux.HandleFunc("/_health", adapt(handlers.Health))
	mux.HandleFunc("/", adapt(handlers.Index))

	// OAuth2 / OIDC — tenant-scoped
	mux.HandleFunc("/{tenant}/oauth2/v2.0/authorize", adapt(handlers.Authorize))
	mux.HandleFunc("/{tenant}/oauth2/authorize", adapt(handlers.Authorize))
	mux.HandleFunc("/{tenant}/oauth2/v2.0/token", adapt(handlers.Token))
	mux.HandleFunc("/{tenant}/oauth2/token", adapt(handlers.Token))
	mux.HandleFunc("/{tenant}/login", adapt(handlers.Login))

	// OAuth2 / OIDC — common and bare paths
	mux.HandleFunc("/common/oauth2/v2.0/authorize", adapt(handlers.Authorize))
	mux.HandleFunc("/common/oauth2/authorize", adapt(handlers.Authorize))
	mux.HandleFunc("/common/oauth2/v2.0/token", adapt(handlers.Token))
	mux.HandleFunc("/common/oauth2/token", adapt(handlers.Token))
	mux.HandleFunc("/oauth2/v2.0/authorize", adapt(handlers.Authorize))
	mux.HandleFunc("/oauth2/authorize", adapt(handlers.Authorize))
	mux.HandleFunc("/authorize", adapt(handlers.Authorize))
	mux.HandleFunc("/oauth2/v2.0/token", adapt(handlers.Token))
	mux.HandleFunc("/oauth2/token", adapt(handlers.Token))
	mux.HandleFunc("/token", adapt(handlers.Token))
	mux.HandleFunc("/login", adapt(handlers.Login))

	// Discovery
	mux.HandleFunc("/{tenant}/v2.0/.well-known/openid-configuration", adapt(handlers.Discovery))
	mux.HandleFunc("/{tenant}/.well-known/openid-configuration", adapt(handlers.Discovery))
	mux.HandleFunc("/.well-known/openid-configuration", adapt(handlers.Discovery))
	mux.HandleFunc("/{tenant}/discovery/keys", adapt(handlers.JWKS))
	mux.HandleFunc("/discovery/keys", adapt(handlers.JWKS))
	mux.HandleFunc("/common/discovery/instance", adapt(handlers.CallHome))

	return withMiddleware(mux)
}

func withMiddleware(next nethttp.Handler) nethttp.Handler {
	return logMiddleware(corsMiddleware(next))
}

func logMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(writer nethttp.ResponseWriter, request *nethttp.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: writer, status: nethttp.StatusOK}

		next.ServeHTTP(recorder, request)

		log.Printf("%s %s %d %s", request.Method, request.URL.Path, recorder.status, time.Since(start))
	})
}

func corsMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(writer nethttp.ResponseWriter, request *nethttp.Request) {
		origin := request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		writer.Header().Set("Access-Control-Allow-Origin", origin)
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Client-Request-Id")
		writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if request.Method == nethttp.MethodOptions {
			writer.WriteHeader(nethttp.StatusNoContent)

			return
		}

		next.ServeHTTP(writer, request)
	})
}

type statusRecorder struct {
	nethttp.ResponseWriter

	status int
}

func (recorder *statusRecorder) WriteHeader(status int) {
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}
