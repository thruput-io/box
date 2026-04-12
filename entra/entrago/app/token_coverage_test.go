package app_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/samber/mo"

	"identity/app"
	"identity/domain"
)

func TestIssueToken_Coverage(t *testing.T) {
	t.Parallel()

	key, clientID, userID := setupTokenCoverageBase(t)
	user := domain.NewUser(
		userID,
		mustUsername(t, "user"),
		mustPassword(t, "pass"),
		mustDisplayName(t, "User"),
		mustEmail(t, "user@example.com"),
		nil,
	)
	appReg := domain.NewAppRegistration(
		mustAppName(t, "App"),
		clientID,
		mustIdentifierURI(t, "api://app"),
		nil, nil, nil,
	)
	tenant := domain.NewTenant(
		mustTenantID(t, "11111111-1111-1111-1111-111111111111"),
		mustTenantName(t, "T"),
		domain.NewNonEmptyArray(appReg).MustRight(),
		nil,
		domain.NewNonEmptyArray(user).MustRight(),
		nil,
	).MustRight()

	input := domain.TokenInput{
		Grant:  domain.GrantPassword,
		Tenant: &tenant,
		User:   &user,
		Client: nil,
		Scope: []domain.ScopeValue{
			mustScopeValue(t, "openid"),
			mustScopeValue(t, "profile"),
			mustScopeValue(t, "offline_access"),
			mustScopeValue(t, "api://app/access"),
		},
		IsV2:          true,
		BaseURL:       domain.MustBaseURL("https://entra.test"),
		Nonce:         mo.Some(mustNonce(t, "nonce")),
		CorrelationID: mo.Some(mustCorrelationID(t, "corr")),
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

	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	clientID := mustClientID(t, "22222222-2222-2222-2222-222222222222")
	userID := mustUserID(t, "33333333-3333-3333-3333-333333333333")

	return key, clientID, userID
}

func assertTokenResponse(t *testing.T, resp domain.TokenResponse) {
	t.Helper()

	if resp.AccessToken.AsByteArray() == nil {
		t.Fatal("expected access token")
	}

	if resp.IDToken.IsAbsent() {
		t.Fatal("expected id token")
	}

	if resp.RefreshToken.IsAbsent() {
		t.Fatal("expected refresh token")
	}
}

func TestIssueAuthCode_Coverage(t *testing.T) {
	t.Parallel()

	const rsaKeySize = 2048

	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	user := domain.NewUser(
		mustUserID(t, "33333333-3333-3333-3333-333333333333"),
		mustUsername(t, "user"),
		mustPassword(t, "pass"),
		mustDisplayName(t, "User"),
		mustEmail(t, "user@example.com"),
		nil,
	)

	code := app.IssueAuthCode(
		key,
		user,
		mustClientID(t, "22222222-2222-2222-2222-222222222222"),
		mustRedirectURL(t, "https://app.test/callback"),
		[]domain.ScopeValue{mustScopeValue(t, "openid")},
		mustTenantID(t, "11111111-1111-1111-1111-111111111111"),
		mo.Some(mustNonce(t, "nonce")),
	)

	if code == (domain.AuthCode{}) {
		t.Fatal("expected auth code")
	}
}

func TestToken_InternalHelpers_Coverage(t *testing.T) {
	t.Parallel()

	if !app.IsOIDCScopeForTest(mustScopeValue(t, "openid")) {
		t.Error("expected true for openid")
	}

	if app.IsOIDCScopeForTest(mustScopeValue(t, "other")) {
		t.Error("expected false for other")
	}
}
