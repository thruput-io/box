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
		Config:        newTestConfig(),
		Key:           key,
		LoginTemplate: loginTmpl,
		IndexTemplate: indexTmpl,
	}}
}

func newTestConfig() *domain.Config {
	tenantID := domain.MustTenantID("b5a920d6-7d3c-44fe-baad-4ffed6b8774d")
	clientID := domain.MustClientID("e697b97c-9b4b-487f-9f7a-248386f78864")
	appID := domain.MustClientID("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	userID := domain.MustUserID("6320573e-360a-426c-829d-649a5b3260c8")

	redirectURL, _ := domain.NewRedirectURL(testRedirectURL)
	identifierURI, _ := domain.NewIdentifierURI("api://testapp")

	user := domain.NewUser(
		userID,
		domain.MustUsername(testUsername),
		domain.MustPassword(testPassword),
		domain.MustDisplayName("Test User"),
		domain.MustEmail("test@example.com"),
		nil,
	)

	client := domain.NewClientWithoutSecret(
		domain.MustAppName("TestClient"),
		clientID,
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	registration := domain.NewAppRegistration(
		domain.MustAppName("TestApp"),
		appID,
		identifierURI,
		nil, nil, nil,
	)

	tenant, err := domain.NewTenant(
		tenantID,
		domain.MustTenantName("Default Tenant"),
		[]domain.AppRegistration{registration},
		nil,
		[]domain.User{user},
		[]domain.Client{client},
	)
	if err != nil {
		panic(err)
	}

	config, err := domain.NewConfig([]domain.Tenant{tenant})
	if err != nil {
		panic(err)
	}

	return &config
}

func serverWithClient(t *testing.T, base *identityhttp.Server, client domain.Client) *identityhttp.Server {
	t.Helper()

	tenant := base.App.Config.Tenants()[firstIndex]
	clients := append(tenant.Clients(), client)

	newTenant, err := domain.NewTenant(
		tenant.TenantID(), tenant.Name(),
		tenant.AppRegistrations(), tenant.Groups(),
		tenant.Users(), clients,
	)
	if err != nil {
		t.Fatalf("build tenant: %v", err)
	}

	config, err := domain.NewConfig([]domain.Tenant{newTenant})
	if err != nil {
		t.Fatalf("build config: %v", err)
	}

	return &identityhttp.Server{App: &app.App{
		Config:        &config,
		Key:           base.App.Key,
		LoginTemplate: base.App.LoginTemplate,
		IndexTemplate: base.App.IndexTemplate,
	}}
}
