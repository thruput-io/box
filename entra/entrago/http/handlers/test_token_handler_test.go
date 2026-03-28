package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"identity/app"
	"identity/domain"
)

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	return key
}

func mustAppForTestTokenHandler(t *testing.T) (*app.App, domain.TenantID, domain.ClientID, domain.UserID) {
	t.Helper()

	tenantID := domain.MustTenantID("11111111-1111-4111-8111-111111111111")
	appID := domain.MustClientID("22222222-2222-4222-8222-222222222222")
	userID := domain.MustUserID("33333333-3333-4333-8333-333333333333")

	redirectURL, err := domain.NewRedirectURL("https://example.com/callback")
	if err != nil {
		t.Fatalf("NewRedirectURL: %v", err)
	}

	registration := domain.NewAppRegistration(
		domain.MustAppName("App"),
		appID,
		mustIdentifierURI(t, "api://app"),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)

	client := domain.NewClient(
		domain.MustAppName("Client"),
		appID,
		domain.NewClientSecret(""),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	user := domain.NewUser(
		userID,
		domain.MustUsername("user"),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		nil,
	)

	tenant, err := domain.NewTenant(
		tenantID,
		domain.MustTenantName("Tenant"),
		[]domain.AppRegistration{registration},
		nil,
		[]domain.User{user},
		[]domain.Client{client},
	)
	if err != nil {
		t.Fatalf("NewTenant: %v", err)
	}

	config, err := domain.NewConfig([]domain.Tenant{tenant})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}

	return &app.App{Config: config, Key: mustRSAKey(t)}, tenantID, appID, userID
}

func TestTestTokenHandler_InvalidPath(t *testing.T) {
	t.Parallel()

	application, _, _, _ := mustAppForTestTokenHandler(t)
	request := httptest.NewRequest(http.MethodGet, "http://entra.test/test-tokens", nil)

	resp := testTokenHandler(request, application)
	if resp.Status != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.Status)
	}
}

func TestTestTokenHandler_TenantNotFound(t *testing.T) {
	t.Parallel()

	application, _, appID, _ := mustAppForTestTokenHandler(t)
	request := httptest.NewRequest(http.MethodGet, "http://entra.test/test-tokens/nope/"+appID.String(), nil)

	resp := testTokenHandler(request, application)
	if resp.Status != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, resp.Status)
	}
}

func TestTestTokenHandler_InvalidAppID(t *testing.T) {
	t.Parallel()

	application, tenantID, _, _ := mustAppForTestTokenHandler(t)
	request := httptest.NewRequest(http.MethodGet, "http://entra.test/test-tokens/"+tenantID.String()+"/nope", nil)

	resp := testTokenHandler(request, application)
	if resp.Status != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.Status)
	}
}

func TestTestTokenHandler_Success_DefaultScopeAndUser(t *testing.T) {
	t.Parallel()

	application, tenantID, appID, userID := mustAppForTestTokenHandler(t)
	request := httptest.NewRequest(
		http.MethodGet,
		"http://entra.test/test-tokens/"+tenantID.String()+"/"+appID.String()+"?username="+userID.String(),
		nil,
	)

	resp := testTokenHandler(request, application)
	if resp.Status != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Status)
	}

	if resp.ContentType != contentTypePlain {
		t.Fatalf("unexpected content type: %q", resp.ContentType)
	}

	body := string(resp.Body)
	if !strings.HasSuffix(body, "\n") {
		t.Fatalf("expected trailing newline")
	}

	trimmed := strings.TrimSpace(body)
	if strings.Count(trimmed, ".") < 2 {
		t.Fatalf("expected jwt-like token, got %q", trimmed)
	}
}

func TestResolveTestUser(t *testing.T) {
	t.Parallel()

	application, tenantID, _, userID := mustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(application.Config, tenantID)
	if err != nil {
		t.Fatalf("FindTenantByID: %v", err)
	}

	user := resolveTestUser(tenant, userID.String())
	if user == nil {
		t.Fatalf("expected user")
	}

	if user.ID().String() != userID.String() {
		t.Fatalf("unexpected user")
	}
}
