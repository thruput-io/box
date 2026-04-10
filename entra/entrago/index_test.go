package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndex_Root(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %v", recorder.Code)
	}

	body := recorder.Body.String()

	expectedFragments := []string{
		"Entra mock",
		"C# container apps environment variables (msal host configuration)",
		"AzureAd__Instance=https://login.microsoftonline.com/",
		"AzureAd__TenantId=b5a920d6-7d3c-44fe-baad-4ffed6b8774d",
		"AzureAd__Audience=api://testapp",
		"AzureAd__ClientId=aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa",
		"AzureAd__ClientSecret=set-client-secret",
		"data-copy-target=\"csharp-host-config-",
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(body, fragment) {
			t.Errorf("body missing %q", fragment)
		}
	}
}

func TestIndex_NotFound(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/unknown-path", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf(statusFmt+", want 404", recorder.Code)
	}
}

func TestIndex_IndexHTMLPath(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/index.html", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusFmt+", want 200", recorder.Code)
	}

	body := recorder.Body.String()

	for _, fragment := range []string{
		"<title>Entra mock</title>",
		"AzureAd__TenantId=b5a920d6-7d3c-44fe-baad-4ffed6b8774d",
		`data-copy-target="csharp-host-config-`,
	} {
		if !strings.Contains(body, fragment) {
			t.Errorf("body missing %q", fragment)
		}
	}
}

func TestHealth(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/_health", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestCORS_Options(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %v", recorder.Code)
	}

	if recorder.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("missing Access-Control-Allow-Origin header")
	}
}
