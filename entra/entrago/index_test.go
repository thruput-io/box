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

	if !strings.Contains(recorder.Body.String(), "Entra mock") {
		t.Error("body missing 'Entra mock'")
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
