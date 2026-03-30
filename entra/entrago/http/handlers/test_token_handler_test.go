package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"identity/app"
	"identity/domain"
	"identity/http/handlers"
)

const (
	findTenantByIDFmt = "FindTenantByID: %v"
	testUserName      = "user"
)

type tokenHandlerFixture struct {
	application *app.App
	tenantID    domain.TenantID
	appID       domain.ClientID
	userID      domain.UserID
}

func MustAppForTestTokenHandler(t *testing.T) tokenHandlerFixture {
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
			Config:        &config,
			Key:           mustRSAKey(t),
			LoginTemplate: nil,
			IndexTemplate: nil,
		},
		tenantID: tenantID,
		appID:    appID,
		userID:   userID,
	}
}

func setupEntitiesForTokenHandler(
	t *testing.T, appID domain.ClientID, userID domain.UserID,
) (domain.User, domain.AppRegistration, domain.Client) {
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
		domain.MustUsername(testUserName),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		nil,
	)

	return user, appReg, client
}

func TestTestTokenHandler_InvalidPath(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)
	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://entra.test/test-tokens", nil)

	resp := handlers.ExportInvokeTestTokenHandler(request, fixture.application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtStatus, http.StatusOK, resp.Status)
	}
}

func TestTestTokenHandler_TenantNotFound(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)
	url := "http://entra.test/test-tokens/nope/" + parseString(fixture.appID)
	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	resp := handlers.ExportInvokeTestTokenHandler(request, fixture.application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtStatus, http.StatusOK, resp.Status)
	}
}

func TestTestTokenHandler_InvalidAppID(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)
	url := "http://entra.test/test-tokens/" + parseString(fixture.tenantID) + "/nope"
	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	resp := handlers.ExportInvokeTestTokenHandler(request, fixture.application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtStatus, http.StatusOK, resp.Status)
	}
}

func TestTestTokenHandler_Success_DefaultScopeAndUser(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)
	baseURL := "http://entra.test/test-tokens/"
	url := baseURL + parseString(fixture.tenantID) + "/" +
		parseString(fixture.appID) + "?username=" + parseString(fixture.userID)
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

	fixture := MustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(fixture.application.Config, fixture.tenantID)
	if err != nil {
		t.Fatalf(findTenantByIDFmt, err)
	}

	user := handlers.ExportResolveTestUser(tenant, testUserName)
	if user == nil {
		t.Fatal("expected user")
	}

	if user.ID() != fixture.userID {
		t.Fatal("unexpected user")
	}
}

func TestResolveClientFromPart_AppRegistration(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(fixture.application.Config, fixture.tenantID)
	if err != nil {
		t.Fatalf(findTenantByIDFmt, err)
	}

	// appUUID is an AppRegistration in the mock config
	clientID := parseString(fixture.appID)
	client := handlers.ExportResolveClientFromPart(tenant, clientID)

	if client.ClientID().UUID().String() != clientID {
		t.Errorf("expected %s, got %s", clientID, client.ClientID().UUID().String())
	}
}

func TestResolveClientFromPart_NotFound(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(fixture.application.Config, fixture.tenantID)
	if err != nil {
		t.Fatalf(findTenantByIDFmt, err)
	}

	clientID := "00000000-0000-0000-0000-000000000000"
	client := handlers.ExportResolveClientFromPart(tenant, clientID)

	// Fallback to tenant as client
	expected := tenant.TenantID().UUID().String()
	if client.ClientID().UUID().String() != expected {
		t.Errorf("expected %s, got %s", expected, client.ClientID().UUID().String())
	}
}

func TestResolveTestUser_Fallback(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)

	tenant, err := app.FindTenantByID(fixture.application.Config, fixture.tenantID)
	if err != nil {
		t.Fatalf(findTenantByIDFmt, err)
	}

	// 1. Unknown UUID
	user1 := handlers.ExportResolveTestUser(tenant, "00000000-0000-4000-a000-000000000000")
	if user1 == nil || user1.ID() != fixture.userID {
		t.Error("expected fallback for unknown UUID")
	}

	// 2. Unknown name
	user2 := handlers.ExportResolveTestUser(tenant, "unknown-user")
	if user2 == nil || user2.ID() != fixture.userID {
		t.Error("expected fallback for unknown name")
	}

	// 3. Known name
	user3 := handlers.ExportResolveTestUser(tenant, testUserName)
	if user3 == nil || user3.ID() != fixture.userID {
		t.Error("expected match by name")
	}
}

func TestSignTokenHandler_InvalidJSON(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)
	ctx := context.Background()

	// Invalid JSON body
	reqURL := "http://entra.test/test-tokens/sign"
	request := httptest.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader("not-json"))

	resp := handlers.ExportInvokeSignTokenHandler(request, fixture.application)
	if resp.Status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.Status)
	}
}
