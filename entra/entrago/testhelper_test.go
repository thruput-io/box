package main

import (
	"crypto/rand"
	"crypto/rsa"
	"embed"
	"html/template"
	"testing"

	"identity/app"
	"identity/domain"
	identityhttp "identity/http/transport"
)

//go:embed templates/login.html
var testLoginHTML embed.FS

//go:embed templates/index.html
var testIndexHTML embed.FS

func newTestServer(t *testing.T) *identityhttp.Server {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, rsaKeyBitsTest)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	loginTmpl, err := template.ParseFS(testLoginHTML, "templates/login.html")
	if err != nil {
		t.Fatalf("parse login template: %v", err)
	}

	indexTmpl, err := template.ParseFS(testIndexHTML, "templates/index.html")
	if err != nil {
		t.Fatalf("parse index template: %v", err)
	}

	return &identityhttp.Server{App: &app.App{
		Config:        newTestConfig(t),
		Key:           key,
		LoginTemplate: loginTmpl,
		IndexTemplate: indexTmpl,
	}}
}

func newTestConfig(t *testing.T) *domain.Config {
	t.Helper()

	tenantID := mustTenantID(t, "b5a920d6-7d3c-44fe-baad-4ffed6b8774d")
	clientID := mustClientID(t, "e697b97c-9b4b-487f-9f7a-248386f78864")
	appID := mustClientID(t, "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	userID := mustUserID(t, "6320573e-360a-426c-829d-649a5b3260c8")

	redirectURL := mustRedirectURL(t, testRedirectURL)
	identifierURI := mustIdentifierURI(t, "api://testapp")

	user := domain.NewUser(
		userID,
		mustUsername(t, testUsername),
		mustPassword(t, testPassword),
		mustDisplayName(t, "Test User"),
		mustEmail(t, "test@example.com"),
		nil,
	)

	client := domain.NewClientWithoutSecret(
		mustAppName(t, "TestClient"),
		clientID,
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	registration := domain.NewAppRegistration(
		mustAppName(t, "TestApp"),
		appID,
		identifierURI,
		nil, nil, nil,
	)
	appRegistrations := domain.NewNonEmptyArray(registration).MustRight()
	users := domain.NewNonEmptyArray(user).MustRight()

	tenant := domain.NewTenant(
		tenantID,
		mustTenantName(t, "Default Tenant"),
		appRegistrations,
		nil,
		users,
		[]domain.Client{client},
	).MustRight()

	tenants := domain.NewNonEmptyArray(tenant).MustRight()
	config := domain.NewConfig(tenants).MustRight()

	return &config
}

func serverWithClient(t *testing.T, base *identityhttp.Server, client domain.Client) *identityhttp.Server {
	t.Helper()

	tenant := base.App.Config.Tenants()[firstIndex]
	clients := append(tenant.Clients(), client)

	newTenant := domain.NewTenant(
		tenant.TenantID(), tenant.Name(),
		domain.NewNonEmptyArray(tenant.AppRegistrations()...).MustRight(), tenant.Groups(),
		domain.NewNonEmptyArray(tenant.Users()...).MustRight(), clients,
	).MustRight()

	config := domain.NewConfig(domain.NewNonEmptyArray(newTenant).MustRight()).MustRight()

	return &identityhttp.Server{App: &app.App{
		Config:        &config,
		Key:           base.App.Key,
		LoginTemplate: base.App.LoginTemplate,
		IndexTemplate: base.App.IndexTemplate,
	}}
}
