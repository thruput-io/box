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

func postLogin(t *testing.T, server interface{ Handler() http.Handler }, form url.Values) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodPost, "/login",
		strings.NewReader(form.Encode()),
	)
	request.Header.Set(headerContentType, headerValueForm)

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	return recorder
}

func TestAuthorize_Success(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.App.Config.Tenants()[firstIndex].Clients()[firstIndex].ClientID().UUID().String()

	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet,
		"/authorize?client_id="+clientID+"&redirect_uri="+testRedirectURL, nil,
	)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestAuthorize_RendersQuickLoginUsers(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.App.Config.Tenants()[firstIndex].Clients()[firstIndex].ClientID().UUID().String()

	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet,
		"/authorize?client_id="+clientID+"&redirect_uri="+testRedirectURL, nil,
	)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(statusFmt, recorder.Code)
	}

	body := recorder.Body.String()

	if !strings.Contains(body, `class="user-item"`) {
		t.Fatalf("expected authorize page to render quick-login user entries")
	}

	if !strings.Contains(body, `name="username"`) {
		t.Fatalf("expected authorize page quick-login controls to submit username")
	}
}

func TestAuthorize_TenantV2AuthorityPath(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	tenantID := server.App.Config.Tenants()[firstIndex].TenantID().UUID().String()
	clientID := server.App.Config.Tenants()[firstIndex].Clients()[firstIndex].ClientID().UUID().String()

	request := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/"+tenantID+"/v2.0/oauth2/v2.0/authorize?client_id="+clientID+"&redirect_uri="+testRedirectURL,
		nil,
	)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for tenant v2 authority path, got %d", recorder.Code)
	}
}

func TestAuthorize_InvalidRedirect(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.App.Config.Tenants()[firstIndex].Clients()[firstIndex].ClientID().UUID().String()

	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet,
		"/authorize?client_id="+clientID+"&redirect_uri=http://evil.example.com", nil,
	)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestAuthorize_ClientNotFound(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet,
		"/authorize?client_id="+uuid.New().String(), nil,
	)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestAuthorize_ParamTooLong(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet,
		"/authorize?p="+strings.Repeat("a", overMaxParamLength), nil,
	)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.App.Config.Tenants()[firstIndex].Clients()[firstIndex].ClientID().UUID().String()

	form := url.Values{}
	form.Set(formUsername, testUsername)
	form.Set(formPassword, testPassword)
	form.Set(formClientID, clientID)
	form.Set(formRedirectURI, testRedirectURL)

	recorder := postLogin(t, server, form)

	if recorder.Code != http.StatusFound {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestLogin_FragmentMode(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.App.Config.Tenants()[firstIndex].Clients()[firstIndex].ClientID().UUID().String()

	form := url.Values{}
	form.Set(formUsername, testUsername)
	form.Set(formPassword, testPassword)
	form.Set(formClientID, clientID)
	form.Set(formRedirectURI, testRedirectURL)
	form.Set(formResponseMode, "fragment")

	recorder := postLogin(t, server, form)

	if recorder.Code != http.StatusFound {
		t.Errorf(statusFmt, recorder.Code)
	}

	if !strings.Contains(recorder.Header().Get("Location"), "#") {
		t.Error("Location should contain # for fragment mode")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.App.Config.Tenants()[firstIndex].Clients()[firstIndex].ClientID().UUID().String()

	form := url.Values{}
	form.Set(formUsername, testUsername)
	form.Set(formPassword, "wrong")
	form.Set(formClientID, clientID)
	form.Set(formRedirectURI, testRedirectURL)

	recorder := postLogin(t, server, form)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestLogin_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/login", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf(statusFmt, recorder.Code)
	}
}
