package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestAuthorize_Success(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.Config.Tenants()[0].Clients()[0].ClientID().String()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/authorize?client_id="+clientID+"&redirect_uri=http://localhost/callback", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %v", recorder.Code)
	}
}

func TestAuthorize_InvalidRedirect(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.Config.Tenants()[0].Clients()[0].ClientID().String()

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/authorize?client_id="+clientID+"&redirect_uri=http://evil.example.com", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %v", recorder.Code)
	}
}

func TestAuthorize_ClientNotFound(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/authorize?client_id="+uuid.New().String(), nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %v", recorder.Code)
	}
}

func TestAuthorize_ParamTooLong(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/authorize?p="+strings.Repeat("a", 2049), nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %v", recorder.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.Config.Tenants()[0].Clients()[0].ClientID().String()

	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("password", "password")
	form.Set("client_id", clientID)
	form.Set("redirect_uri", "http://localhost/callback")

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusFound {
		t.Errorf("status = %v", recorder.Code)
	}
}

func TestLogin_FragmentMode(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.Config.Tenants()[0].Clients()[0].ClientID().String()

	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("password", "password")
	form.Set("client_id", clientID)
	form.Set("redirect_uri", "http://localhost/callback")
	form.Set("response_mode", "fragment")

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusFound {
		t.Errorf("status = %v", recorder.Code)
	}

	if !strings.Contains(recorder.Header().Get("Location"), "#") {
		t.Error("Location should contain # for fragment mode")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.Config.Tenants()[0].Clients()[0].ClientID().String()

	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("password", "wrong")
	form.Set("client_id", clientID)
	form.Set("redirect_uri", "http://localhost/callback")

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %v", recorder.Code)
	}
}

func TestLogin_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/login", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %v", recorder.Code)
	}
}
