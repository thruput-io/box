package handlers_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"identity/app"
	"identity/domain"
	"identity/http/handlers"
)

func TestTokenHandler_Coverage(t *testing.T) {
	t.Parallel()

	application, appID := setupTestApplication(t)

	t.Run("InvalidGrantType", func(t *testing.T) {
		t.Parallel()

		form := url.Values{"grant_type": {"invalid"}}
		body := strings.NewReader(form.Encode())
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/token", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp := handlers.ExportTokenHandler(req, application)
		if resp.Status != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.Status)
		}
	})

	t.Run("ClientCredentials_MissingSecret", func(t *testing.T) {
		t.Parallel()

		form := url.Values{
			"grant_type": {"client_credentials"},
			"client_id":  {appID.UUID().String()},
		}
		body := strings.NewReader(form.Encode())
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/token", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp := handlers.ExportTokenHandler(req, application)
		if resp.Status != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.Status)
		}
	})
}

func setupTestApplication(t *testing.T) (*app.App, domain.ClientID) {
	t.Helper()

	const rsaKeySize = 2048

	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		t.Fatal(err)
	}

	tenantID := domain.NewTenantID("11111111-1111-1111-1111-111111111111").MustRight()
	appID := domain.NewClientID("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa").MustRight()
	uID := domain.NewUserID("33333333-3333-3333-3333-333333333333").MustRight()

	user := domain.NewUser(
		uID,
		domain.NewUsername("u").MustRight(),
		domain.NewPassword("p").MustRight(),
		domain.NewDisplayName("D").MustRight(),
		domain.NewEmail("e").MustRight(),
		nil,
	)
	appReg := domain.NewAppRegistration(
		domain.NewAppName("App").MustRight(),
		appID,
		domain.NewIdentifierURI("api://app").MustRight(),
		nil,
		nil,
		nil,
	)
	tenant := domain.NewTenant(
		tenantID,
		domain.NewTenantName("T").MustRight(),
		domain.NewNonEmptyArray(appReg).MustRight(),
		nil,
		domain.NewNonEmptyArray(user).MustRight(),
		nil,
	).MustRight()
	config := domain.NewConfig(domain.NewNonEmptyArray(tenant).MustRight()).MustRight()

	return &app.App{
		Config:        &config,
		Key:           key,
		LoginTemplate: nil,
		IndexTemplate: nil,
	}, appID
}
