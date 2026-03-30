package handlers

import (
	"bytes"
	"net/http"
	"strings"

	"identity/app"
)

func indexHandler(request *http.Request, application *app.App) Response {
	if strings.HasPrefix(request.URL.Path, pathMockUtilsSign) {
		return signTokenHandler(request, application)
	}

	if strings.HasPrefix(request.URL.Path, pathMockUtils) {
		return testTokenHandler(request, application)
	}

	if strings.HasPrefix(request.URL.Path, "/config/") {
		return configHandler(request, application)
	}

	if request.URL.Path != "/" {
		return notFound("not found")
	}

	var buf bytes.Buffer

	err := application.IndexTemplate.Execute(&buf, application.Config)
	if err != nil {
		return internalError("failed to render index")
	}

	return okHTML(buf.Bytes())
}

func healthHandler(_ *http.Request, _ *app.App) Response {
	return Response{Status: http.StatusOK, Body: nil, ContentType: "", Headers: nil}
}
