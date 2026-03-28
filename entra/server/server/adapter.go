package server

import (
	"crypto/rsa"
	"html/template"
	"net/http"

	"identity/domain"
)

const contentTypePlain = "text/plain; charset=utf-8"

// Server holds the immutable dependencies wired at startup.
type Server struct {
	Config        domain.Config
	Key           *rsa.PrivateKey
	LoginTemplate *template.Template
	IndexTemplate *template.Template
}

// Response is the result of a handler — status + body bytes + content type.
type Response struct {
	Status      int
	Body        []byte
	ContentType string
	Headers     map[string]string
}

func ok(body []byte, contentType string) Response {
	return Response{Status: http.StatusOK, Body: body, ContentType: contentType, Headers: nil}
}

func okJSON(body []byte) Response {
	return ok(body, "application/json")
}

func okText(body []byte) Response {
	return ok(body, "text/plain; charset=utf-8")
}

func okHTML(body []byte) Response {
	return ok(body, "text/html; charset=utf-8")
}

func badRequest(err error) Response {
	return oauthError("invalid_request", err.Error(), http.StatusBadRequest)
}

func notFound(msg string) Response {
	return Response{
		Status: http.StatusNotFound, Body: []byte(msg),
		ContentType: contentTypePlain, Headers: nil,
	}
}

func methodNotAllowed() Response {
	return Response{
		Status: http.StatusMethodNotAllowed, Body: []byte("method not allowed"),
		ContentType: contentTypePlain, Headers: nil,
	}
}

func internalError(msg string) Response {
	return Response{
		Status: http.StatusInternalServerError, Body: []byte(msg),
		ContentType: contentTypePlain, Headers: nil,
	}
}

func fromDomainError(domErr *domain.Error) Response {
	httpStatus := domainErrorToHTTPStatus(domErr.Code)

	return oauthError(string(domErr.Code), domErr.Message, httpStatus)
}

func domainErrorToHTTPStatus(code domain.ErrorCode) int {
	switch code {
	case domain.ErrCodeInvalidCredentials,
		domain.ErrCodeClientNotFound,
		domain.ErrCodeInvalidGrant:
		return http.StatusUnauthorized
	case domain.ErrCodeTenantNotFound,
		domain.ErrCodeAppNotFound,
		domain.ErrCodeUserNotFound:
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}

func oauthError(code, description string, status int) Response {
	body := []byte(`{"error":"` + code + `","error_description":"` + description + `"}`)

	return Response{Status: status, Body: body, ContentType: "application/json", Headers: nil}
}

func write(writer http.ResponseWriter, response Response) {
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
