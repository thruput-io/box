package app_test

import (
	"errors"
	"fmt"
	"testing"

	"identity/app"
	"identity/domain"
)

func mustConfigWithTenant(t *testing.T, tenant domain.Tenant) domain.Config {
	t.Helper()

	cfg, err := domain.NewConfig([]domain.Tenant{tenant})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}

	return cfg
}

type tenantFixture struct {
	tenant domain.Tenant
	client domain.Client
	appReg domain.AppRegistration
	user   domain.User
}

func mustTenantFixture(t *testing.T) tenantFixture {
	t.Helper()

	tenantID := domain.MustTenantID(testTenantID)
	tenantName := domain.MustTenantName("Tenant")

	clientID := domain.MustClientID(testClientID)
	redirectURL := mustRedirectURL(t, testCallback)
	client := domain.NewClient(
		domain.MustAppName("Client"),
		clientID,
		domain.NewClientSecret(testSecret),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	appReg := domain.NewAppRegistration(
		domain.MustAppName("App"),
		clientID,
		mustIdentifierURI(t, testApp),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)

	user := domain.NewUser(
		domain.MustUserID(testUserID),
		domain.MustUsername(testUser),
		domain.MustPassword(testPass),
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

func runSelectionTests(t *testing.T, cfg domain.Config, tenant domain.Tenant) {
	t.Helper()

	got, err := app.ExportFindTenant(cfg, "")
	if err != nil {
		t.Fatal(err)
	}

	if got.TenantID().String() != tenant.TenantID().String() {
		t.Fatalf("expected first tenant, got %s", got.TenantID())
	}

	got, err = app.ExportFindTenant(cfg, "common")
	if err != nil {
		t.Fatal(err)
	}

	if got.TenantID().String() != tenant.TenantID().String() {
		t.Fatalf("expected first tenant for common, got %s", got.TenantID())
	}

	got, err = app.ExportFindTenant(cfg, tenant.TenantID().String())
	if err != nil {
		t.Fatal(err)
	}

	if got.TenantID().String() != tenant.TenantID().String() {
		t.Fatalf("expected tenant %s, got %s", tenant.TenantID(), got.TenantID())
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

	if got.TenantID().String() != fixture.tenant.TenantID().String() {
		t.Fatalf("expected tenant %s, got %s", fixture.tenant.TenantID(), got.TenantID())
	}
}

func TestLookup_ClientRegistrationAndRedirects(t *testing.T) {
	t.Parallel()

	fixture := mustTenantFixture(t)

	err := validateClientAndApp(t, fixture.tenant, fixture.client, fixture.appReg)
	if err != nil {
		t.Fatal(err)
	}

	allowed, err := app.ExportFindRedirectURLs(fixture.tenant, fixture.client.ClientID())
	if err != nil {
		t.Fatal(err)
	}

	const expectedAllowedLen = 1

	if len(allowed) != expectedAllowedLen {
		t.Fatalf("expected %d allowed redirect url, got %d", expectedAllowedLen, len(allowed))
	}

	err = app.ExportValidateRedirectURI(allowed[0].String(), allowed)
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
	tenant domain.Tenant,
	client domain.Client,
	appReg domain.AppRegistration,
) error {
	t.Helper()

	gotClient, err := app.ExportFindClient(tenant, client.ClientID())
	if err != nil {
		return fmt.Errorf("ExportFindClient: %w", err)
	}

	if gotClient.ClientID().String() != client.ClientID().String() {
		t.Fatalf("expected client %s, got %s", client.ClientID(), gotClient.ClientID())
	}

	gotReg, err := app.ExportFindAppRegistration(tenant, client.ClientID())
	if err != nil {
		return fmt.Errorf("ExportFindAppRegistration: %w", err)
	}

	if gotReg.IdentifierURI().String() != appReg.IdentifierURI().String() {
		t.Fatalf("expected identifier URI %s, got %s", appReg.IdentifierURI(), gotReg.IdentifierURI())
	}

	return nil
}

func TestLookup_AuthenticateAndSecrets(t *testing.T) {
	t.Parallel()

	fixture := mustTenantFixture(t)

	err := app.ExportValidateClientSecret(fixture.client, testSecret)
	if err != nil {
		t.Fatal(err)
	}

	err = app.ExportValidateClientSecret(fixture.client, testWrong)
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong secret, got %v", err)
	}

	runAuthTests(t, fixture.tenant)
}

func runAuthTests(t *testing.T, tenant domain.Tenant) {
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

func checkValidAuth(tenant domain.Tenant) error {
	_, err := app.ExportAuthenticateUser(tenant, testUser, testPass)

	return err
}

func TestLookup_UserByID(t *testing.T) {
	t.Parallel()

	fixture := mustTenantFixture(t)

	_, notFound := app.ExportFindUserByID(fixture.tenant, "not-a-user")
	if notFound {
		t.Fatal("expected not found")
	}

	got, found := app.ExportFindUserByID(fixture.tenant, fixture.user.ID().String())
	if !found {
		t.Fatal("expected found")
	}

	if got.ID().String() != fixture.user.ID().String() {
		t.Fatalf("expected user %s, got %s", fixture.user.ID(), got.ID())
	}
}
