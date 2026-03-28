package transport

import (
	"crypto/rand"
	"crypto/rsa"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"identity/app"
	"identity/domain"
)

const (
	testRedirectURI = "https://localhost/callback"
	testUsername    = "testuser"
	testPassword    = "testpass"
)

func mustTransportTestApp(t *testing.T) *app.App {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve current test file path")
	}

	baseDir := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../.."))
	loginTemplatePath := filepath.Join(baseDir, "templates", "login.html")
	indexTemplatePath := filepath.Join(baseDir, "templates", "index.html")

	loginTemplate, err := template.ParseFiles(loginTemplatePath)
	if err != nil {
		t.Fatalf("parse login template: %v", err)
	}

	indexTemplate, err := template.ParseFiles(indexTemplatePath)
	if err != nil {
		t.Fatalf("parse index template: %v", err)
	}

	redirectURL, err := domain.NewRedirectURL(testRedirectURI)
	if err != nil {
		t.Fatalf("new redirect URL: %v", err)
	}

	identifierURI, err := domain.NewIdentifierURI("api://transport-test")
	if err != nil {
		t.Fatalf("new identifier URI: %v", err)
	}

	tenantID := domain.MustTenantID("b5a920d6-7d3c-44fe-baad-4ffed6b8774d")
	clientID := domain.MustClientID("e697b97c-9b4b-487f-9f7a-248386f78864")
	appID := domain.MustClientID("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	userID := domain.MustUserID("6320573e-360a-426c-829d-649a5b3260c8")

	user := domain.NewUser(
		userID,
		domain.MustUsername(testUsername),
		domain.MustPassword(testPassword),
		domain.MustDisplayName("Test User"),
		domain.MustEmail("test@example.com"),
		nil,
	)

	client := domain.NewClient(
		domain.MustAppName("TestClient"),
		clientID,
		domain.NewClientSecret(""),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	registration := domain.NewAppRegistration(
		domain.MustAppName("TestApp"),
		appID,
		identifierURI,
		nil,
		nil,
		nil,
	)

	tenant, err := domain.NewTenant(
		tenantID,
		domain.MustTenantName("Default Tenant"),
		[]domain.AppRegistration{registration},
		nil,
		[]domain.User{user},
		[]domain.Client{client},
	)
	if err != nil {
		t.Fatalf("new tenant: %v", err)
	}

	config, err := domain.NewConfig([]domain.Tenant{tenant})
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	return &app.App{
		Config:        config,
		Key:           key,
		LoginTemplate: loginTemplate,
		IndexTemplate: indexTemplate,
	}
}

func TestBuildRoutes_HealthEndpointAndCORSHeaders(t *testing.T) {
	t.Parallel()

	handler := buildRoutes(nil)
	request := httptest.NewRequest(http.MethodGet, "/_health", nil)
	request.Header.Set("Origin", "https://example.com")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin to match request origin, got %q", got)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
		t.Fatalf("unexpected Access-Control-Allow-Methods %q", got)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("unexpected Access-Control-Allow-Credentials %q", got)
	}
}

func TestBuildRoutes_CORSPreflightForRegisteredPaths(t *testing.T) {
	t.Parallel()

	handler := buildRoutes(nil)
	paths := []string{
		"/authorize",
		"/common/oauth2/v2.0/authorize",
		"/tenant/oauth2/v2.0/authorize",
		"/tenant/v2.0/oauth2/v2.0/authorize",
		"/tenant/oauth2/v2.0/token",
		"/tenant/v2.0/oauth2/v2.0/token",
		"/tenant/login",
		"/.well-known/openid-configuration",
		"/tenant/v2.0/.well-known/openid-configuration",
		"/discovery/keys",
		"/tenant/discovery/keys",
		"/common/discovery/instance",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodOptions, path, nil)
			request.Header.Set("Origin", "https://example.com")

			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusNoContent {
				t.Fatalf("expected status %d for %s, got %d", http.StatusNoContent, path, recorder.Code)
			}

			if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
				t.Fatalf("expected Access-Control-Allow-Origin for %s, got %q", path, got)
			}
		})
	}
}

func TestBuildRoutes_UnknownPathStillAppliesCORS(t *testing.T) {
	t.Parallel()

	handler := buildRoutes(nil)
	request := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected wildcard CORS origin, got %q", got)
	}
}

func TestBuildRoutes_RouteAliasesDispatchWithoutServerError(t *testing.T) {
	t.Parallel()

	application := mustTransportTestApp(t)
	handler := buildRoutes(application)
	tenant := application.Config.Tenants()[0]
	tenantID := tenant.TenantID().String()
	clientID := tenant.Clients()[0].ClientID().String()
	redirectURI := tenant.Clients()[0].RedirectURLs()[0].String()

	authorizeQuery := "?client_id=" + url.QueryEscape(clientID) + "&redirect_uri=" + url.QueryEscape(redirectURI)

	type routeRequest struct {
		name   string
		method string
		path   string
		body   string
	}

	routes := []routeRequest{
		{name: "authorize bare", method: http.MethodGet, path: "/authorize" + authorizeQuery},
		{name: "authorize oauth2", method: http.MethodGet, path: "/oauth2/authorize" + authorizeQuery},
		{name: "authorize oauth2 v2", method: http.MethodGet, path: "/oauth2/v2.0/authorize" + authorizeQuery},
		{name: "authorize common", method: http.MethodGet, path: "/common/oauth2/authorize" + authorizeQuery},
		{name: "authorize common v2", method: http.MethodGet, path: "/common/oauth2/v2.0/authorize" + authorizeQuery},
		{name: "authorize tenant", method: http.MethodGet, path: "/" + tenantID + "/oauth2/authorize" + authorizeQuery},
		{name: "authorize tenant v2", method: http.MethodGet, path: "/" + tenantID + "/oauth2/v2.0/authorize" + authorizeQuery},
		{name: "authorize tenant prefixed", method: http.MethodGet, path: "/" + tenantID + "/v2.0/oauth2/authorize" + authorizeQuery},
		{name: "authorize tenant prefixed v2", method: http.MethodGet, path: "/" + tenantID + "/v2.0/oauth2/v2.0/authorize" + authorizeQuery},
		{name: "token bare", method: http.MethodPost, path: "/token", body: "grant_type=client_credentials&client_id=" + clientID + "&scope=.default"},
		{name: "token oauth2", method: http.MethodPost, path: "/oauth2/token", body: "grant_type=client_credentials&client_id=" + clientID + "&scope=.default"},
		{name: "token oauth2 v2", method: http.MethodPost, path: "/oauth2/v2.0/token", body: "grant_type=client_credentials&client_id=" + clientID + "&scope=.default"},
		{name: "token common", method: http.MethodPost, path: "/common/oauth2/token", body: "grant_type=client_credentials&client_id=" + clientID + "&scope=.default"},
		{name: "token common v2", method: http.MethodPost, path: "/common/oauth2/v2.0/token", body: "grant_type=client_credentials&client_id=" + clientID + "&scope=.default"},
		{name: "token tenant", method: http.MethodPost, path: "/" + tenantID + "/oauth2/token", body: "grant_type=client_credentials&client_id=" + clientID + "&scope=.default"},
		{name: "token tenant v2", method: http.MethodPost, path: "/" + tenantID + "/oauth2/v2.0/token", body: "grant_type=client_credentials&client_id=" + clientID + "&scope=.default"},
		{name: "token tenant prefixed", method: http.MethodPost, path: "/" + tenantID + "/v2.0/oauth2/token", body: "grant_type=client_credentials&client_id=" + clientID + "&scope=.default"},
		{name: "token tenant prefixed v2", method: http.MethodPost, path: "/" + tenantID + "/v2.0/oauth2/v2.0/token", body: "grant_type=client_credentials&client_id=" + clientID + "&scope=.default"},
		{name: "login bare", method: http.MethodPost, path: "/login", body: "client_id=" + clientID + "&redirect_uri=" + url.QueryEscape(redirectURI) + "&username=" + testUsername + "&password=" + testPassword},
		{name: "login tenant", method: http.MethodPost, path: "/" + tenantID + "/login", body: "client_id=" + clientID + "&redirect_uri=" + url.QueryEscape(redirectURI) + "&username=" + testUsername + "&password=" + testPassword},
		{name: "openid discovery", method: http.MethodGet, path: "/.well-known/openid-configuration"},
		{name: "openid discovery tenant", method: http.MethodGet, path: "/" + tenantID + "/.well-known/openid-configuration"},
		{name: "openid discovery tenant v2", method: http.MethodGet, path: "/" + tenantID + "/v2.0/.well-known/openid-configuration"},
		{name: "jwks", method: http.MethodGet, path: "/discovery/keys"},
		{name: "jwks tenant", method: http.MethodGet, path: "/" + tenantID + "/discovery/keys"},
		{name: "instance", method: http.MethodGet, path: "/common/discovery/instance"},
	}

	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
			if route.method == http.MethodPost {
				request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code == http.StatusNotFound {
				t.Fatalf("expected route %s (%s %s) to be registered", route.name, route.method, route.path)
			}

			if recorder.Code >= http.StatusInternalServerError {
				t.Fatalf("expected non-5xx for %s (%s %s), got %d", route.name, route.method, route.path, recorder.Code)
			}
		})
	}
}

func TestServer_Handler(t *testing.T) {
	t.Parallel()

	server := &Server{}

	handler := server.Handler()
	if handler == nil {
		t.Fatalf("expected non-nil handler")
	}

	request := httptest.NewRequest(http.MethodOptions, "/token", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestStatusRecorder_WriteHeader(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	status := &statusRecorder{ResponseWriter: recorder, status: http.StatusOK}

	status.WriteHeader(http.StatusTeapot)

	if status.status != http.StatusTeapot {
		t.Fatalf("expected stored status %d, got %d", http.StatusTeapot, status.status)
	}

	if recorder.Code != http.StatusTeapot {
		t.Fatalf("expected recorder status %d, got %d", http.StatusTeapot, recorder.Code)
	}
}

func TestCorsMiddleware_OptionsShortCircuitAndDefaultOrigin(t *testing.T) {
	t.Parallel()

	nextCalled := false
	handler := corsMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	}))

	request := httptest.NewRequest(http.MethodOptions, "/anything", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	if nextCalled {
		t.Fatalf("expected next handler not to be called for OPTIONS")
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected wildcard CORS origin, got %q", got)
	}
}

func TestCorsMiddleware_DelegatesToNextForNonOptions(t *testing.T) {
	t.Parallel()

	handler := corsMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusAccepted)
	}))

	request := httptest.NewRequest(http.MethodGet, "/anything", nil)
	request.Header.Set("Origin", "https://example.com")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, recorder.Code)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin to match request origin, got %q", got)
	}
}

func TestLogMiddleware_UsesRecordedStatus(t *testing.T) {
	t.Parallel()

	handler := logMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusCreated)
	}))

	request := httptest.NewRequest(http.MethodGet, "/anything", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, recorder.Code)
	}
}

func TestWithMiddleware_ComposesCORSAndLogging(t *testing.T) {
	t.Parallel()

	handler := withMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusAccepted)
	}))

	request := httptest.NewRequest(http.MethodGet, "/anything", nil)
	request.Header.Set("Origin", "https://example.com")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, recorder.Code)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin to be set by middleware, got %q", got)
	}
}
