package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"identity/app"
	"identity/domain"
	"identity/http/handlers"
)

type configFixture struct {
	application    *app.App
	tenantID       domain.TenantID
	appRegID       domain.ClientID
	secretClientID domain.ClientID
	publicClientID domain.ClientID
}

func mustAppForConfigTests(t *testing.T) configFixture {
	t.Helper()

	const (
		tenantUUID = "11111111-1111-4111-8111-111111111111"
		appRegUUID = "22222222-2222-4222-8222-222222222222"
		secretUUID = "33333333-3333-4333-8333-333333333333"
		publicUUID = "44444444-4444-4444-8444-444444444444"
		userUUID   = "55555555-5555-4555-8555-555555555555"
	)

	tenantID := domain.MustTenantID(tenantUUID)
	appRegID := domain.MustClientID(appRegUUID)
	secretClientID := domain.MustClientID(secretUUID)
	publicClientID := domain.MustClientID(publicUUID)

	user, secretClient, publicClient := setupCoreEntities(t, userUUID, secretClientID, publicClientID)
	registration := setupRegistration(t, appRegID)

	tenant, err := domain.NewTenant(
		tenantID,
		domain.MustTenantName("Tenant"),
		[]domain.AppRegistration{registration},
		nil,
		[]domain.User{user},
		[]domain.Client{secretClient, publicClient},
	)
	if err != nil {
		t.Fatalf("NewTenant: %v", err)
	}

	config, err := domain.NewConfig([]domain.Tenant{tenant})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}

	return configFixture{
		application: &app.App{
			Config:        &config,
			Key:           nil,
			LoginTemplate: nil,
			IndexTemplate: nil,
		},
		tenantID:       tenantID,
		appRegID:       appRegID,
		secretClientID: secretClientID,
		publicClientID: publicClientID,
	}
}

func setupCoreEntities(
	t *testing.T,
	userUUID string,
	secretCID, publicCID domain.ClientID,
) (user domain.User, secretClient domain.Client, publicClient domain.Client) {
	t.Helper()

	redirectURL, err := domain.NewRedirectURL(testCallbackURI)
	if err != nil {
		t.Fatalf("NewRedirectURL: %v", err)
	}

	user = domain.NewUser(
		domain.MustUserID(userUUID),
		domain.MustUsername("user"),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		nil,
	)
	secretClient = domain.NewClientWithSecret(
		domain.MustAppName("SecretClient"),
		secretCID,
		domain.MustClientSecret("secret"),
		[]domain.RedirectURL{redirectURL},
		nil,
	)
	publicClient = domain.NewClientWithoutSecret(
		domain.MustAppName("PublicClient"),
		publicCID,
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	return user, secretClient, publicClient
}

func setupRegistration(t *testing.T, appRegID domain.ClientID) domain.AppRegistration {
	t.Helper()

	redirectURL, err := domain.NewRedirectURL(testCallbackURI)
	if err != nil {
		t.Fatalf("NewRedirectURL: %v", err)
	}

	return domain.NewAppRegistration(
		domain.MustAppName("App"),
		appRegID,
		mustIdentifierURI(t, testAppURI),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)
}

func TestConfigHandler_RawRedirect(t *testing.T) {
	t.Parallel()

	fixture := mustAppForConfigTests(t)
	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://example.test/config/raw", nil)

	resp := handlers.ExportConfigHandler(request, fixture.application)
	if resp.Status != http.StatusMovedPermanently {
		t.Fatalf(fmtStatus, http.StatusMovedPermanently, resp.Status)
	}

	if resp.Headers["Location"] != "/DefaultConfig.yaml" {
		t.Fatalf("expected Location /DefaultConfig.yaml, got %q", resp.Headers["Location"])
	}
}

func TestConfigHandler_CsharpAppSuccess(t *testing.T) {
	t.Parallel()

	fixture := mustAppForConfigTests(t)
	path := pathConfig + fixture.tenantID.UUID().String() + "/app/" + fixture.appRegID.UUID().String() + pathCsharp

	const testHost = "entra.test"

	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://"+testHost+path, nil)
	request.Host = testHost

	resp := handlers.ExportConfigHandler(request, fixture.application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtStatus, http.StatusOK, resp.Status)
	}

	body := string(resp.Body)
	if !strings.Contains(body, "AzureAd__Instance=https://entra.test/") {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestConfigHandler_JsAppSuccess(t *testing.T) {
	t.Parallel()

	fixture := mustAppForConfigTests(t)
	path := pathConfig + fixture.tenantID.UUID().String() + "/app/" + fixture.appRegID.UUID().String() + "/js"

	const testHost = "entra.test"

	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://"+testHost+path, nil)
	request.Host = testHost

	resp := handlers.ExportConfigHandler(request, fixture.application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtStatus, http.StatusOK, resp.Status)
	}

	body := string(resp.Body)
	if !strings.Contains(body, "msalConfig") {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestConfigHandler_CsharpClientIncludesSecretWhenPresent(t *testing.T) {
	t.Parallel()

	fixture := mustAppForConfigTests(t)

	const testHost = "entra.test"

	tenantID := fixture.tenantID.UUID().String()
	secretCID := fixture.secretClientID.UUID().String()
	secretPath := pathConfig + tenantID + "/client/" + secretCID + pathCsharp

	ctx := context.Background()
	secretReq := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://"+testHost+secretPath, nil)
	secretReq.Host = testHost

	secretResp := handlers.ExportConfigHandler(secretReq, fixture.application)
	if secretResp.Status != http.StatusOK {
		t.Fatalf(fmtStatus, http.StatusOK, secretResp.Status)
	}

	if !strings.Contains(string(secretResp.Body), "AzureAd__ClientSecret=secret") {
		t.Fatal("expected client secret in body")
	}

	publicCID := fixture.publicClientID.UUID().String()
	publicPath := pathConfig + tenantID + "/client/" + publicCID + pathCsharp
	publicReq := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://"+testHost+publicPath, nil)
	publicReq.Host = testHost

	publicResp := handlers.ExportConfigHandler(publicReq, fixture.application)
	if publicResp.Status != http.StatusOK {
		t.Fatalf(fmtStatus, http.StatusOK, publicResp.Status)
	}

	if strings.Contains(string(publicResp.Body), "AzureAd__ClientSecret=") {
		t.Fatal("did not expect client secret in body")
	}
}

func TestParseTenantAndAppID_Errors(t *testing.T) {
	t.Parallel()

	_, _, err := handlers.ExportParseTenantAndAppID("/config/nope/app/x/csharp", handlers.PathApp, handlers.PathCsharp)
	if !errors.Is(err, domain.ErrTenantNotFound) {
		t.Fatalf("expected ErrTenantNotFound, got %v", err)
	}

	const validTID = "11111111-1111-4111-8111-111111111111"

	_, _, err = handlers.ExportParseTenantAndAppID(
		"/config/"+validTID+"/app/nope/csharp",
		handlers.PathApp,
		handlers.PathCsharp,
	)
	if !errors.Is(err, domain.ErrAppNotFound) {
		t.Fatalf("expected ErrAppNotFound, got %v", err)
	}
}
