package app_test

import (
	"errors"
	"fmt"
	"testing"

	"identity/app"
	"identity/domain"
)

func mustConfigWithTenant(t *testing.T, tenant domain.Tenant) *domain.Config {
	t.Helper()

	cfg := domain.NewConfig(domain.NewNonEmptyArray(tenant).MustRight()).MustRight()

	return &cfg
}

type tenantFixture struct {
	tenant domain.Tenant
	client domain.Client
	appReg domain.AppRegistration
	user   domain.User
}

func mustTenantFixture(t *testing.T) tenantFixture {
	t.Helper()

	tenantID := mustTenantID(t, testTenantID)
	tenantName := mustTenantName(t, "Tenant")

	clientID := mustClientID(t, testClientID)
	redirectURL := mustRedirectURL(t, testCallback)
	client := domain.NewClientWithSecret(
		mustAppName(t, "Client"),
		clientID,
		mustClientSecret(t, testSecret),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	appReg := domain.NewAppRegistration(
		mustAppName(t, "App"),
		clientID,
		mustIdentifierURI(t, testApp),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)

	user := domain.NewUser(
		mustUserID(t, testUserID),
		mustUsername(t, testUser),
		mustPassword(t, testPass),
		mustDisplayName(t, "User"),
		mustEmail(t, "user@example.com"),
		[]domain.GroupName{mustGroupName(t, "GroupA")},
	)

	tenant := domain.NewTenant(
		tenantID,
		tenantName,
		domain.NewNonEmptyArray(appReg).MustRight(),
		nil,
		domain.NewNonEmptyArray(user).MustRight(),
		[]domain.Client{client},
	).MustRight()

	return tenantFixture{
		tenant: tenant,
		client: client,
		appReg: appReg,
		user:   user,
	}
}

func TestLookup_TenantSelection(t *testing.T) {
	t.Parallel()

	fixture := mustTenantFixture(t)
	cfg := mustConfigWithTenant(t, fixture.tenant)

	runSelectionTests(t, cfg, fixture.tenant)
}

func runSelectionTests(t *testing.T, cfg *domain.Config, tenant domain.Tenant) {
	t.Helper()

	got, err := app.ExportFindTenant(cfg, "")
	if err != nil {
		t.Fatal(err)
	}

	if got.TenantID() != tenant.TenantID() {
		t.Fatalf("expected first tenant, got %s", got.TenantID().UUID())
	}

	got, err = app.ExportFindTenant(cfg, "common")
	if err != nil {
		t.Fatal(err)
	}

	if got.TenantID() != tenant.TenantID() {
		t.Fatalf("expected first tenant for common, got %s", got.TenantID().UUID())
	}

	got, err = app.ExportFindTenant(cfg, tenant.TenantID().UUID().String())
	if err != nil {
		t.Fatal(err)
	}

	if got.TenantID() != tenant.TenantID() {
		t.Fatalf("expected tenant %s, got %s", tenant.TenantID().UUID(), got.TenantID().UUID())
	}

	const unknownID = "99999999-9999-4999-8999-999999999999"

	_, err = app.ExportFindTenant(cfg, unknownID)
	if !errors.Is(err, domain.ErrTenantNotFound) {
		t.Fatalf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestLookup_TenantByID(t *testing.T) {
	t.Parallel()

	fixture := mustTenantFixture(t)
	cfg := mustConfigWithTenant(t, fixture.tenant)

	got, err := app.ExportFindTenantByID(cfg, fixture.tenant.TenantID())
	if err != nil {
		t.Fatal(err)
	}

	if got.TenantID() != fixture.tenant.TenantID() {
		t.Fatalf("expected tenant %s, got %s", fixture.tenant.TenantID().UUID(), got.TenantID().UUID())
	}
}

func TestLookup_ClientRegistrationAndRedirects(t *testing.T) {
	t.Parallel()

	fixture := mustTenantFixture(t)

	err := validateClientAndApp(t, &fixture.tenant, &fixture.client, &fixture.appReg)
	if err != nil {
		t.Fatal(err)
	}

	allowed, err := app.ExportFindRedirectURLs(&fixture.tenant, fixture.client.ClientID())
	if err != nil {
		t.Fatal(err)
	}

	const expectedAllowedLen = 1

	if len(allowed) != expectedAllowedLen {
		t.Fatalf("expected %d allowed redirect url, got %d", expectedAllowedLen, len(allowed))
	}

	redirectURIStr := allowed[0].Value()

	err = app.ExportValidateRedirectURI(redirectURIStr, allowed)
	if err != nil {
		t.Fatal(err)
	}

	const wrongURI = "https://example.com/nope"

	err = app.ExportValidateRedirectURI(wrongURI, allowed)
	if !errors.Is(err, domain.ErrInvalidRedirectURI) {
		t.Fatalf("expected ErrInvalidRedirectURI, got %v", err)
	}
}

func validateClientAndApp(
	t *testing.T,
	tenant *domain.Tenant,
	client *domain.Client,
	appReg *domain.AppRegistration,
) error {
	t.Helper()

	gotClient, err := app.ExportFindClient(tenant, client.ClientID())
	if err != nil {
		return fmt.Errorf("ExportFindClient: %w", err)
	}

	if gotClient.ClientID() != client.ClientID() {
		t.Fatalf("expected client %s, got %s", client.ClientID().UUID(), gotClient.ClientID().UUID())
	}

	gotReg, err := app.ExportFindAppRegistration(tenant, client.ClientID())
	if err != nil {
		return fmt.Errorf("ExportFindAppRegistration: %w", err)
	}

	if gotReg.IdentifierURI() != appReg.IdentifierURI() {
		t.Fatal("expected identifier URI mismatch")
	}

	return nil
}

func TestLookup_AuthenticateAndSecrets(t *testing.T) {
	t.Parallel()

	fixture := mustTenantFixture(t)

	err := app.ExportValidateClientSecret(&fixture.client, testSecret)
	if err != nil {
		t.Fatal(err)
	}

	err = app.ExportValidateClientSecret(&fixture.client, testWrong)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong secret, got %v", err)
	}

	runAuthTests(t, &fixture.tenant)
}

func runAuthTests(t *testing.T, tenant *domain.Tenant) {
	t.Helper()

	const msgErrCreds = "expected ErrInvalidCredentials for wrong password, got %v"

	err := checkValidAuth(tenant)
	if err != nil {
		t.Fatal(err)
	}

	_, err = app.ExportAuthenticateUser(tenant, testUser, testWrong)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf(msgErrCreds, err)
	}

	_, err = app.ExportAuthenticateUser(tenant, testWrong, testPass)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong username, got %v", err)
	}
}

func checkValidAuth(tenant *domain.Tenant) error {
	_, err := app.ExportAuthenticateUser(tenant, testUser, testPass)
	if err != nil {
		return fmt.Errorf("checkValidAuth: %w", err)
	}

	return nil
}

func TestLookup_UserByID(t *testing.T) {
	t.Parallel()

	fixture := mustTenantFixture(t)

	_, notFound := app.ExportFindUserByID(&fixture.tenant, domain.UserID{})
	if notFound {
		t.Fatal("expected not found")
	}

	got, found := app.ExportFindUserByID(&fixture.tenant, fixture.user.ID())
	if !found {
		t.Fatal("expected found")
	}

	if got.ID() != fixture.user.ID() {
		t.Fatalf("expected user %v, got %v", fixture.user.ID(), got.ID())
	}
}
