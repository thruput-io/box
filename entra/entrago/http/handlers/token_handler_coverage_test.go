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

	key, _ := rsa.GenerateKey(rand.Reader, rsaKeySize)
	tenantID := domain.MustTenantID("11111111-1111-1111-1111-111111111111")
	appID := domain.MustClientID("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	uID := domain.MustUserID("33333333-3333-3333-3333-333333333333")

	user := domain.NewUser(
		uID, domain.MustUsername("u"), domain.MustPassword("p"),
		domain.MustDisplayName("D"), domain.MustEmail("e"), nil,
	)
	appReg := domain.NewAppRegistration(
		domain.MustAppName("App"), appID,
		domain.MustIdentifierURI("api://app"), nil, nil, nil,
	)
	tenant, _ := domain.NewTenant(
		tenantID, domain.MustTenantName("T"),
		[]domain.AppRegistration{appReg}, nil, []domain.User{user}, nil,
	)
	config, _ := domain.NewConfig([]domain.Tenant{tenant})

	return &app.App{
		Config:        &config,
		Key:           key,
		LoginTemplate: nil,
		IndexTemplate: nil,
	}, appID
}
