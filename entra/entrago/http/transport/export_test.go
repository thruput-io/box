package transport

import (
	"net/http"

	"identity/app"
)

// ExportBuildRoutes is for testing buildRoutes from transport_test.
func ExportBuildRoutes(application *app.App) http.Handler {
	return buildRoutes(application)
}

// ExportStatusRecorder is for testing statusRecorder from transport_test.
type ExportStatusRecorder struct {
	http.ResponseWriter

	Status int
}

// WriteHeader implements http.ResponseWriter.
func (r *ExportStatusRecorder) WriteHeader(code int) {
	r.Status = code
	r.ResponseWriter.WriteHeader(code)
}

// ExportCorsMiddleware is for testing corsMiddleware from transport_test.
func ExportCorsMiddleware(next http.Handler) http.Handler {
	return corsMiddleware(next)
}

// ExportLogMiddleware is for testing logMiddleware from transport_test.
func ExportLogMiddleware(next http.Handler) http.Handler {
	return logMiddleware(next)
}
