package app_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"identity/app"
	"identity/domain"
)

func TestIssueToken_Coverage(t *testing.T) {
	t.Parallel()

	key, clientID, userID := setupTokenCoverageBase(t)
	user := domain.NewUser(
		userID, domain.MustUsername("user"), domain.MustPassword("pass"),
		domain.MustDisplayName("User"), domain.MustEmail("user@example.com"), nil,
	)
	appReg := domain.NewAppRegistration(
		domain.MustAppName("App"), clientID,
		domain.MustIdentifierURI("api://app"), nil, nil, nil,
	)
	tenant, _ := domain.NewTenant(
		domain.MustTenantID("11111111-1111-1111-1111-111111111111"),
		domain.MustTenantName("T"), []domain.AppRegistration{appReg},
		nil, []domain.User{user}, nil,
	)

	input := domain.TokenInput{
		Grant:         domain.GrantPassword,
		Tenant:        &tenant,
		User:          &user,
		Client:        nil,
		Scope:         "openid profile offline_access api://app/access",
		IsV2:          true,
		BaseURL:       "https://entra.test",
		Nonce:         "nonce",
		CorrelationID: "corr",
	}

	resp := app.IssueToken(key, input)
	assertTokenResponse(t, resp)

	// V1 Issuance
	input.IsV2 = false

	respV1 := app.IssueToken(key, input)
	if respV1.AccessToken.AsByteArray() == nil {
		t.Fatal("expected V1 access token")
	}
}

func setupTokenCoverageBase(t *testing.T) (*rsa.PrivateKey, domain.ClientID, domain.UserID) {
	t.Helper()

	const rsaKeySize = 2048

	key, _ := rsa.GenerateKey(rand.Reader, rsaKeySize)
	clientID := domain.MustClientID("22222222-2222-2222-2222-222222222222")
	userID := domain.MustUserID("33333333-3333-3333-3333-333333333333")

	return key, clientID, userID
}

func assertTokenResponse(t *testing.T, resp domain.TokenResponse) {
	t.Helper()

	if resp.AccessToken.AsByteArray() == nil {
		t.Fatal("expected access token")
	}

	if resp.IDToken == nil {
		t.Fatal("expected id token")
	}

	if resp.RefreshToken == nil {
		t.Fatal("expected refresh token")
	}
}

func TestIssueAuthCode_Coverage(t *testing.T) {
	t.Parallel()

	const rsaKeySize = 2048

	key, _ := rsa.GenerateKey(rand.Reader, rsaKeySize)
	user := domain.NewUser(
		domain.MustUserID("33333333-3333-3333-3333-333333333333"),
		domain.MustUsername("user"),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		nil,
	)

	code := app.IssueAuthCode(
		key,
		user,
		domain.MustClientID("22222222-2222-2222-2222-222222222222"),
		domain.MustRedirectURL("https://app.test/callback"),
		"openid",
		domain.MustTenantID("11111111-1111-1111-1111-111111111111"),
		"nonce",
	)

	if code == (domain.AuthCode{}) {
		t.Fatal("expected auth code")
	}
}

func TestToken_InternalHelpers_Coverage(t *testing.T) {
	t.Parallel()

	// isOIDCScope
	if !app.IsOIDCScopeForTest("openid") {
		t.Error("expected true for openid")
	}

	if app.IsOIDCScopeForTest("other") {
		t.Error("expected false for other")
	}
}
