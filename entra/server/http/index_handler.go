package http

import (
	"bytes"
	"net/http"
	"strings"
)

func indexHandler(request *http.Request, server *Server) HTTPResponse {
	if strings.HasPrefix(request.URL.Path, "/test-tokens/") {
		return testTokenHandler(request, server)
	}

	if strings.HasPrefix(request.URL.Path, "/config/") {
		return configHandler(request, server)
	}

	if request.URL.Path != "/" {
		return notFound("not found")
	}

	var buf bytes.Buffer

	err := server.IndexTemplate.Execute(&buf, server.Config)
	if err != nil {
		return internalError("failed to render index")
	}

	return okHTML(buf.Bytes())
}

func healthHandler(_ *http.Request, _ *Server) HTTPResponse {
	return HTTPResponse{Status: http.StatusOK, Body: nil, ContentType: "", Headers: nil}
}
