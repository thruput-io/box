package transport_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"identity/app"
	"identity/domain"
	"identity/http/transport"
)

func mustAppForRoutes(t *testing.T) *app.App {
	t.Helper()

	tenantID := mustTenantID(t, "11111111-1111-4111-8111-111111111111")
	appID := mustClientID(t, "22222222-2222-4222-8222-222222222222")
	redirectURL := mustRedirectURL(t, testCallbackURI)

	registration := domain.NewAppRegistration(
		mustAppName(t, "App"),
		appID,
		mustIdentifierURI(t, testAppURI),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)

	user := domain.NewUser(
		mustUserID(t, "33333333-3333-4333-8333-333333333333"),
		mustUsername(t, testUsername),
		mustPassword(t, testPassword),
		mustDisplayName(t, "User"),
		mustEmail(t, "user@example.com"),
		nil,
	)

	tenant := domain.NewTenant(
		tenantID,
		mustTenantName(t, "Tenant"),
		domain.NewNonEmptyArray(registration).MustRight(),
		nil,
		domain.NewNonEmptyArray(user).MustRight(),
		nil,
	).MustRight()

	config := domain.NewConfig(domain.NewNonEmptyArray(tenant).MustRight()).MustRight()

	return &app.App{
		Config:        &config,
		Key:           mustRSAKey(t),
		LoginTemplate: nil,
		IndexTemplate: nil,
	}
}

func TestBuildRoutes_ReturnsHandler(t *testing.T) {
	t.Parallel()

	application := mustAppForRoutes(t)
	handler := transport.ExportBuildRoutes(application)

	if handler == nil {
		t.Fatal("expected handler")
	}
}

func TestCorsMiddleware_AddsHeaders(t *testing.T) {
	t.Parallel()

	inner := &testHandler{called: false}
	middleware := transport.ExportCorsMiddleware(inner)

	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodOptions, exampleCom, nil)
	recorder := httptest.NewRecorder()

	middleware.ServeHTTP(recorder, request)

	if inner.called {
		t.Fatal("inner handler should not be called for OPTIONS request")
	}

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected origin *, got %q", got)
	}
}

func TestLogMiddleware_RecordsStatus(t *testing.T) {
	t.Parallel()

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	logMiddleware := transport.ExportLogMiddleware(inner)

	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, exampleCom, nil)
	recorder := httptest.NewRecorder()

	logMiddleware.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusTeapot {
		t.Fatalf("expected %d, got %d", http.StatusTeapot, recorder.Code)
	}
}

func TestStatusRecorder_WriteHeader(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()

	const initialStatus = 0

	sr := &transport.ExportStatusRecorder{ResponseWriter: recorder, Status: initialStatus}

	sr.WriteHeader(http.StatusAccepted)

	if sr.Status != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, sr.Status)
	}

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected recorder code %d, got %d", http.StatusAccepted, recorder.Code)
	}
}

func TestServer_Handler(t *testing.T) {
	t.Parallel()

	application := mustAppForRoutes(t)
	server := transport.Server{App: application}

	if server.Handler() == nil {
		t.Fatal("expected handler")
	}
}
