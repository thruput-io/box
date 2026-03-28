package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"identity/app"
	"identity/domain"
)

func mustAppForConfigTests(t *testing.T) (*app.App, domain.TenantID, domain.ClientID, domain.ClientID, domain.ClientID) {
	t.Helper()

	tenantID := domain.MustTenantID("11111111-1111-4111-8111-111111111111")
	tenantName := domain.MustTenantName("Tenant")

	appRegID := domain.MustClientID("22222222-2222-4222-8222-222222222222")
	secretClientID := domain.MustClientID("33333333-3333-4333-8333-333333333333")
	publicClientID := domain.MustClientID("44444444-4444-4444-8444-444444444444")

	redirectURL, err := domain.NewRedirectURL("https://example.com/callback")
	if err != nil {
		t.Fatalf("NewRedirectURL: %v", err)
	}

	registration := domain.NewAppRegistration(
		domain.MustAppName("App"),
		appRegID,
		mustIdentifierURI(t, "api://app"),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)

	user := domain.NewUser(
		domain.MustUserID("55555555-5555-4555-8555-555555555555"),
		domain.MustUsername("user"),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		nil,
	)

	secretClient := domain.NewClient(
		domain.MustAppName("SecretClient"),
		secretClientID,
		domain.NewClientSecret("secret"),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	publicClient := domain.NewClient(
		domain.MustAppName("PublicClient"),
		publicClientID,
		domain.NewClientSecret(""),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	tenant, err := domain.NewTenant(
		tenantID,
		tenantName,
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

	application := &app.App{Config: config}

	return application, tenantID, appRegID, secretClientID, publicClientID
}

func mustIdentifierURI(t *testing.T, raw string) domain.IdentifierURI {
	t.Helper()

	v, err := domain.NewIdentifierURI(raw)
	if err != nil {
		t.Fatalf("NewIdentifierURI(%q): %v", raw, err)
	}

	return v
}

func TestConfigHandler_RawRedirect(t *testing.T) {
	t.Parallel()

	application, _, _, _, _ := mustAppForConfigTests(t)
	request := httptest.NewRequest(http.MethodGet, "http://example.test/config/raw", nil)

	resp := configHandler(request, application)
	if resp.Status != http.StatusMovedPermanently {
		t.Fatalf("expected %d, got %d", http.StatusMovedPermanently, resp.Status)
	}

	if resp.Headers["Location"] != "/DefaultConfig.yaml" {
		t.Fatalf("expected Location /DefaultConfig.yaml, got %q", resp.Headers["Location"])
	}
}

func TestConfigHandler_CsharpAppSuccess(t *testing.T) {
	t.Parallel()

	application, tenantID, appRegID, _, _ := mustAppForConfigTests(t)
	path := "/config/" + tenantID.String() + "/app/" + appRegID.String() + "/csharp"
	request := httptest.NewRequest(http.MethodGet, "http://entra.test"+path, nil)
	request.Host = "entra.test"

	resp := configHandler(request, application)
	if resp.Status != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Status)
	}

	body := string(resp.Body)
	if !strings.Contains(body, "AzureAd__Instance=https://entra.test/") {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestConfigHandler_JsAppSuccess(t *testing.T) {
	t.Parallel()

	application, tenantID, appRegID, _, _ := mustAppForConfigTests(t)
	path := "/config/" + tenantID.String() + "/app/" + appRegID.String() + "/js"
	request := httptest.NewRequest(http.MethodGet, "http://entra.test"+path, nil)
	request.Host = "entra.test"

	resp := configHandler(request, application)
	if resp.Status != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Status)
	}

	body := string(resp.Body)
	if !strings.Contains(body, "msalConfig") {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestConfigHandler_CsharpClientIncludesSecretWhenPresent(t *testing.T) {
	t.Parallel()

	application, tenantID, _, secretClientID, publicClientID := mustAppForConfigTests(t)

	secretPath := "/config/" + tenantID.String() + "/client/" + secretClientID.String() + "/csharp"
	secretReq := httptest.NewRequest(http.MethodGet, "http://entra.test"+secretPath, nil)
	secretReq.Host = "entra.test"

	secretResp := configHandler(secretReq, application)
	if secretResp.Status != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, secretResp.Status)
	}

	if !strings.Contains(string(secretResp.Body), "AzureAd__ClientSecret=secret") {
		t.Fatalf("expected client secret in body")
	}

	publicPath := "/config/" + tenantID.String() + "/client/" + publicClientID.String() + "/csharp"
	publicReq := httptest.NewRequest(http.MethodGet, "http://entra.test"+publicPath, nil)
	publicReq.Host = "entra.test"

	publicResp := configHandler(publicReq, application)
	if publicResp.Status != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, publicResp.Status)
	}

	if strings.Contains(string(publicResp.Body), "AzureAd__ClientSecret=") {
		t.Fatalf("did not expect client secret in body")
	}
}

func TestParseTenantAndAppID_Errors(t *testing.T) {
	t.Parallel()

	_, _, err := parseTenantAndAppID("/config/nope/app/x/csharp", pathApp, pathCsharp)
	if !errors.Is(err, domain.ErrTenantNotFound) {
		t.Fatalf("expected ErrTenantNotFound, got %v", err)
	}

	_, _, err = parseTenantAndAppID(
		"/config/11111111-1111-4111-8111-111111111111/app/nope/csharp",
		pathApp,
		pathCsharp,
	)
	if !errors.Is(err, domain.ErrAppNotFound) {
		t.Fatalf("expected ErrAppNotFound, got %v", err)
	}
}
