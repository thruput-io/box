package handlers_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"identity/app"
	"identity/domain"
	"identity/http/handlers"
)

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	const keySize = 1024

	key, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	return key
}

type tokenHandlerFixture struct {
	application *app.App
	tenantID    domain.TenantID
	appID       domain.ClientID
	userID      domain.UserID
}

func mustAppForTestTokenHandler(t *testing.T) tokenHandlerFixture {
	t.Helper()

	const (
		tenantUUID = "11111111-1111-4111-8111-111111111111"
		appUUID    = "22222222-2222-4222-8222-222222222222"
		userUUID   = "33333333-3333-4333-8333-333333333333"
	)

	tenantID := domain.MustTenantID(tenantUUID)
	appID := domain.MustClientID(appUUID)
	userID := domain.MustUserID(userUUID)

	user, reg, client := setupEntitiesForTokenHandler(t, appID, userID)

	tenant, err := domain.NewTenant(
		tenantID,
		domain.MustTenantName("Tenant"),
		[]domain.AppRegistration{reg},
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

	return tokenHandlerFixture{
		application: &app.App{
			Config:        config,
			Key:           mustRSAKey(t),
			LoginTemplate: nil,
			IndexTemplate: nil,
		},
		tenantID: tenantID,
		appID:    appID,
		userID:   userID,
	}
}

func setupEntitiesForTokenHandler(t *testing.T, appID domain.ClientID, userID domain.UserID) (domain.User, domain.AppRegistration, domain.Client) {
	t.Helper()

	redirectURL, err := domain.NewRedirectURL(testCallbackURI)
	if err != nil {
		t.Fatalf("NewRedirectURL: %v", err)
	}

	appReg := domain.NewAppRegistration(
		domain.MustAppName("App"),
		appID,
		mustIdentifierURI(t, testAppURI),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)
	client := domain.NewClientWithSecret(
		domain.MustAppName("Client"),
		appID,
		domain.MustClientSecret("test-secret"),
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

	return user, appReg, client
}

func TestTestTokenHandler_InvalidPath(t *testing.T) {
	t.Parallel()

	fixture := mustAppForTestTokenHandler(t)
	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://entra.test/test-tokens", nil)

	resp := handlers.ExportInvokeTestTokenHandler(request, fixture.application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtStatus, http.StatusOK, resp.Status)
	}
}

func TestTestTokenHandler_TenantNotFound(t *testing.T) {
	t.Parallel()

	fixture := mustAppForTestTokenHandler(t)
	url := "http://entra.test/test-tokens/nope/" + fixture.appID.RawString()
	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	resp := handlers.ExportInvokeTestTokenHandler(request, fixture.application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtStatus, http.StatusOK, resp.Status)
	}
}

func TestTestTokenHandler_InvalidAppID(t *testing.T) {
	t.Parallel()

	fixture := mustAppForTestTokenHandler(t)
	url := "http://entra.test/test-tokens/" + fixture.tenantID.RawString() + "/nope"
	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	resp := handlers.ExportInvokeTestTokenHandler(request, fixture.application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtStatus, http.StatusOK, resp.Status)
	}
}

func TestTestTokenHandler_Success_DefaultScopeAndUser(t *testing.T) {
	t.Parallel()

	fixture := mustAppForTestTokenHandler(t)
	baseURL := "http://entra.test/test-tokens/"
	url := baseURL + fixture.tenantID.RawString() + "/" + fixture.appID.RawString() + "?username=" + fixture.userID.RawString()
	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	resp := handlers.ExportInvokeTestTokenHandler(request, fixture.application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtStatus, http.StatusOK, resp.Status)
	}

	if resp.ContentType != handlers.ContentTypePlain {
		t.Fatalf("unexpected content type: %q", resp.ContentType)
	}

	body := string(resp.Body)
	if !strings.HasSuffix(body, "\n") {
		t.Fatal("expected trailing newline")
	}

	trimmed := strings.TrimSpace(body)

	const minDots = 2

	if strings.Count(trimmed, ".") < minDots {
		t.Fatalf("expected jwt-like token, got %q", trimmed)
	}
}

func TestResolveTestUser(t *testing.T) {
	t.Parallel()

	fixture := mustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(fixture.application.Config, fixture.tenantID)
	if err != nil {
		t.Fatalf("FindTenantByID: %v", err)
	}

	user := handlers.ExportResolveTestUser(tenant, fixture.userID.RawString())
	if user == nil {
		t.Fatal("expected user")
	}

	if user.ID().RawString() != fixture.userID.RawString() {
		t.Fatal("unexpected user")
	}
}

func TestResolveTestUser_FallbackToFirstWhenRequestedUserMissing(t *testing.T) {
	t.Parallel()

	fixture := mustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(fixture.application.Config, fixture.tenantID)
	if err != nil {
		t.Fatalf("FindTenantByID: %v", err)
	}

	user := handlers.ExportResolveTestUser(tenant, "11111111-1111-4111-8111-111111111111")
	if user == nil {
		t.Fatal("expected fallback user")
	}

	if user.ID().RawString() != fixture.userID.RawString() {
		t.Fatalf("expected fallback to first user %s, got %s", fixture.userID.RawString(), user.ID().RawString())
	}
}
