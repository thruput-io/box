package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscovery_Root(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet, "/.well-known/openid-configuration", nil,
	)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestDiscovery_TenantScoped(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	tenantID := server.App.Config.Tenants()[firstIndex].TenantID().UUID().String()

	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet,
		"/"+tenantID+"/.well-known/openid-configuration", nil,
	)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestDiscovery_V2(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	tenantID := server.App.Config.Tenants()[firstIndex].TenantID().UUID().String()

	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet,
		"/"+tenantID+"/v2.0/.well-known/openid-configuration", nil,
	)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestJWKS(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet, "/discovery/keys", nil,
	)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestCallHome(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodGet, "/common/discovery/instance", nil,
	)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusFmt, recorder.Code)
	}
}
