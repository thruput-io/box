package http

import (
	"crypto/rsa"
	"html/template"
	"net/http"

	"identity/domain"
)

// Server holds the immutable dependencies wired at startup.
type Server struct {
	Config        domain.Config
	Key           *rsa.PrivateKey
	LoginTemplate *template.Template
	IndexTemplate *template.Template
}

// HTTPResponse is the result of a handler — status + body bytes + content type.
type HTTPResponse struct {
	Status      int
	Body        []byte
	ContentType string
	Headers     map[string]string
}

func ok(body []byte, contentType string) HTTPResponse {
	return HTTPResponse{Status: http.StatusOK, Body: body, ContentType: contentType}
}

func okJSON(body []byte) HTTPResponse {
	return ok(body, "application/json")
}

func okText(body []byte) HTTPResponse {
	return ok(body, "text/plain; charset=utf-8")
}

func okHTML(body []byte) HTTPResponse {
	return ok(body, "text/html; charset=utf-8")
}

func badRequest(err error) HTTPResponse {
	return oauthError("invalid_request", err.Error(), http.StatusBadRequest)
}

func unauthorized(err error) HTTPResponse {
	return oauthError("invalid_grant", err.Error(), http.StatusUnauthorized)
}

func invalidClient(err error) HTTPResponse {
	return oauthError("invalid_client", err.Error(), http.StatusUnauthorized)
}

func notFound(msg string) HTTPResponse {
	return HTTPResponse{Status: http.StatusNotFound, Body: []byte(msg), ContentType: "text/plain; charset=utf-8"}
}

func methodNotAllowed() HTTPResponse {
	return HTTPResponse{Status: http.StatusMethodNotAllowed, Body: []byte("method not allowed"), ContentType: "text/plain; charset=utf-8"}
}

func internalError(msg string) HTTPResponse {
	return HTTPResponse{Status: http.StatusInternalServerError, Body: []byte(msg), ContentType: "text/plain; charset=utf-8"}
}

func oauthError(code, description string, status int) HTTPResponse {
	body := []byte(`{"error":"` + code + `","error_description":"` + description + `"}`)

	return HTTPResponse{Status: status, Body: body, ContentType: "application/json"}
}

func write(writer http.ResponseWriter, response HTTPResponse) {
	if response.ContentType != "" {
		writer.Header().Set("Content-Type", response.ContentType)
	}

	for key, value := range response.Headers {
		writer.Header().Set(key, value)
	}

	writer.WriteHeader(response.Status)
	_, _ = writer.Write(response.Body)
}

// Handler returns the root HTTP handler with all routes registered.
func (server *Server) Handler() http.Handler {
	return buildRoutes(server)
}
