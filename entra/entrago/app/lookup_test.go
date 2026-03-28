package app

import (
	"errors"
	"testing"

	"identity/domain"
)

func mustRedirectURL(t *testing.T, raw string) domain.RedirectURL {
	t.Helper()

	url, err := domain.NewRedirectURL(raw)
	if err != nil {
		t.Fatalf("NewRedirectURL(%q): %v", raw, err)
	}

	return url
}

func mustConfigWithTenant(t *testing.T, tenant domain.Tenant) domain.Config {
	t.Helper()

	cfg, err := domain.NewConfig([]domain.Tenant{tenant})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}

	return cfg
}

func mustTenantWithClientAndRegistration(t *testing.T) (domain.Tenant, domain.Client, domain.AppRegistration, domain.User) {
	t.Helper()

	tenantID := domain.MustTenantID("11111111-1111-4111-8111-111111111111")
	tenantName := domain.MustTenantName("Tenant")

	clientID := domain.MustClientID("22222222-2222-4222-8222-222222222222")
	redirectURL := mustRedirectURL(t, "https://example.com/callback")
	client := domain.NewClient(
		domain.MustAppName("Client"),
		clientID,
		domain.NewClientSecret("secret"),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	appReg := domain.NewAppRegistration(
		domain.MustAppName("App"),
		clientID,
		mustIdentifierURI(t, "api://app"),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)

	user := domain.NewUser(
		domain.MustUserID("33333333-3333-4333-8333-333333333333"),
		domain.MustUsername("user"),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		[]domain.GroupName{mustGroupName(t, "GroupA")},
	)

	tenant, err := domain.NewTenant(
		tenantID,
		tenantName,
		[]domain.AppRegistration{appReg},
		nil,
		[]domain.User{user},
		[]domain.Client{client},
	)
	if err != nil {
		t.Fatalf("NewTenant: %v", err)
	}

	return tenant, client, appReg, user
}

func TestLookup_TenantSelection(t *testing.T) {
	t.Parallel()

	tenant, _, _, _ := mustTenantWithClientAndRegistration(t)
	cfg := mustConfigWithTenant(t, tenant)

	got, err := FindTenant(cfg, "")
	if err != nil {
		t.Fatalf("FindTenant(\"\"): %v", err)
	}

	if got.TenantID().String() != tenant.TenantID().String() {
		t.Fatalf("expected first tenant, got %s", got.TenantID())
	}

	got, err = FindTenant(cfg, "common")
	if err != nil {
		t.Fatalf("FindTenant(common): %v", err)
	}

	if got.TenantID().String() != tenant.TenantID().String() {
		t.Fatalf("expected first tenant for common, got %s", got.TenantID())
	}

	got, err = FindTenant(cfg, tenant.TenantID().String())
	if err != nil {
		t.Fatalf("FindTenant(by id): %v", err)
	}

	if got.TenantID().String() != tenant.TenantID().String() {
		t.Fatalf("expected tenant %s, got %s", tenant.TenantID(), got.TenantID())
	}

	_, err = FindTenant(cfg, "99999999-9999-4999-8999-999999999999")
	if !errors.Is(err, domain.ErrTenantNotFound) {
		t.Fatalf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestLookup_TenantByID(t *testing.T) {
	t.Parallel()

	tenant, _, _, _ := mustTenantWithClientAndRegistration(t)
	cfg := mustConfigWithTenant(t, tenant)

	got, err := FindTenantByID(cfg, tenant.TenantID())
	if err != nil {
		t.Fatalf("FindTenantByID: %v", err)
	}

	if got.TenantID().String() != tenant.TenantID().String() {
		t.Fatalf("expected tenant %s, got %s", tenant.TenantID(), got.TenantID())
	}
}

func TestLookup_ClientRegistrationAndRedirects(t *testing.T) {
	t.Parallel()

	tenant, client, appReg, _ := mustTenantWithClientAndRegistration(t)

	gotClient, err := FindClient(tenant, client.ClientID())
	if err != nil {
		t.Fatalf("FindClient: %v", err)
	}

	if gotClient.ClientID().String() != client.ClientID().String() {
		t.Fatalf("expected client %s, got %s", client.ClientID(), gotClient.ClientID())
	}

	gotReg, err := FindAppRegistration(tenant, client.ClientID())
	if err != nil {
		t.Fatalf("FindAppRegistration: %v", err)
	}

	if gotReg.IdentifierURI().String() != appReg.IdentifierURI().String() {
		t.Fatalf("expected identifier URI %s, got %s", appReg.IdentifierURI(), gotReg.IdentifierURI())
	}

	allowed, err := FindRedirectURLs(tenant, client.ClientID())
	if err != nil {
		t.Fatalf("FindRedirectURLs: %v", err)
	}

	if len(allowed) != 1 {
		t.Fatalf("expected 1 allowed redirect url, got %d", len(allowed))
	}

	if err := ValidateRedirectURI(allowed[0].String(), allowed); err != nil {
		t.Fatalf("ValidateRedirectURI(allowed): %v", err)
	}

	if err := ValidateRedirectURI("https://example.com/nope", allowed); !errors.Is(err, domain.ErrInvalidRedirectURI) {
		t.Fatalf("expected ErrInvalidRedirectURI, got %v", err)
	}
}

func TestLookup_AuthenticateAndSecrets(t *testing.T) {
	t.Parallel()

	tenant, client, _, _ := mustTenantWithClientAndRegistration(t)

	if _, err := AuthenticateUser(tenant, "user", "pass"); err != nil {
		t.Fatalf("AuthenticateUser(valid): %v", err)
	}

	if _, err := AuthenticateUser(tenant, "user", "wrong"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}

	if _, err := AuthenticateUser(tenant, "wrong", "pass"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong username, got %v", err)
	}

	err := ValidateClientSecret(client, "secret")
	if err != nil {
		t.Fatalf("ValidateClientSecret(valid): %v", err)
	}

	err = ValidateClientSecret(client, "wrong")

	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong secret, got %v", err)
	}
}

func TestLookup_UserByID(t *testing.T) {
	t.Parallel()

	tenant, _, _, user := mustTenantWithClientAndRegistration(t)

	_, ok := FindUserByID(tenant, "not-a-user")
	if ok {
		t.Fatalf("expected not found")
	}

	got, ok := FindUserByID(tenant, user.ID().String())
	if !ok {
		t.Fatalf("expected found")
	}

	if got.ID().String() != user.ID().String() {
		t.Fatalf("expected user %s, got %s", user.ID(), got.ID())
	}
}
