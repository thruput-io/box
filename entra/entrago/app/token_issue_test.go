package app_test

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"identity/app"
	"identity/domain"
)

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, testKeySize)
	if err != nil {
		t.Fatal(err)
	}

	return key
}

type rawStringer interface {
	RawString() string
}

func parseJWT(t *testing.T, token rawStringer, key *rsa.PrivateKey) jwt.MapClaims {
	t.Helper()

	parsed, err := jwt.Parse(token.RawString(), func(_ *jwt.Token) (any, error) { return &key.PublicKey, nil })
	if err != nil {
		t.Fatalf("jwt.Parse: %v", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}

	return claims
}

func TestIssueToken_OpenIDNonceRefreshAndClientInfo(t *testing.T) {
	t.Parallel()

	fixture := mustTenantFixture(t)
	key := mustRSAKey(t)

	input := domain.TokenInput{
		Grant:         domain.GrantPassword,
		Tenant:        fixture.tenant,
		User:          &fixture.user,
		Client:        fixture.client,
		Scope:         testScope,
		Nonce:         testNonce,
		IsV2:          true,
		BaseURL:       "https://login.microsoftonline.com",
		CorrelationID: testCorr,
	}

	resp := app.ExportIssueToken(key, input)

	verifyResponse(t, resp, testCorr)

	accessClaims := parseJWT(t, resp.AccessToken, key)
	if got, ok := accessClaims["scp"].(string); !ok || got != input.Scope {
		t.Fatalf("expected access scp %q, got %#v", input.Scope, accessClaims["scp"])
	}

	idClaims := parseJWT(t, *resp.IDToken, key)
	if got, ok := idClaims["nonce"].(string); !ok || got != testNonce {
		t.Fatalf("expected id_token nonce %q, got %#v", testNonce, idClaims["nonce"])
	}

	if gotAud, ok := idClaims["aud"].(string); !ok || gotAud != fixture.client.ClientID().UUID().String() {
		t.Fatalf("expected id_token aud %q, got %#v", fixture.client.ClientID().UUID().String(), idClaims["aud"])
	}
}

func verifyResponse(t *testing.T, resp domain.TokenResponse, corr string) {
	t.Helper()

	if resp.IDToken == nil {
		t.Fatal("expected id token")
	}

	if resp.RefreshToken == nil {
		t.Fatal("expected refresh token")
	}

	if resp.ClientInfo == nil {
		t.Fatal("expected client info")
	}

	if resp.CorrelationID != corr {
		t.Fatalf("expected correlation id %s, got %q", corr, resp.CorrelationID)
	}
}

type graphTestFixture struct {
	client domain.Client
	user   domain.User
	tenant domain.Tenant
}

func TestIssueToken_GraphAudienceUsesRolesForScp(t *testing.T) {
	t.Parallel()

	fixture := setupGraphTest(t)
	key := mustRSAKey(t)
	input := domain.TokenInput{
		Grant:         domain.GrantPassword,
		Tenant:        fixture.tenant,
		User:          &fixture.user,
		Client:        fixture.client,
		Scope:         "https://graph.microsoft.com/User.Read",
		Nonce:         testNonce,
		IsV2:          true,
		BaseURL:       "https://login.microsoftonline.com",
		CorrelationID: testCorr,
	}

	resp := app.ExportIssueToken(key, input)
	accessClaims := parseJWT(t, resp.AccessToken, key)

	verifyGraphClaims(t, accessClaims, input.Scope, "https://graph.microsoft.com", "GraphRole")

	const (
		jwtErrFmt = "jwt.Parse: %v"
		sigErr    = "signature invalid"
	)

	_, err := jwt.Parse(resp.AccessToken.RawString(), func(_ *jwt.Token) (any, error) {
		return &key.PublicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			t.Fatal(sigErr)
		}

		t.Fatalf(jwtErrFmt, err)
	}
}

func setupGraphTest(t *testing.T) graphTestFixture {
	t.Helper()

	const (
		graphID   = "00000003-0000-0000-c000-000000000000"
		scopeID   = "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"
		roleID    = "bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb"
		graphURI  = "https://graph.microsoft.com"
		graphRole = "GraphRole"
	)

	graphClientID := domain.MustClientID(graphID)
	redirectURL := mustRedirectURL(t, testCallback)
	scope := domain.NewScope(
		domain.MustScopeID(scopeID),
		mustScopeValue(t, "User.Read"),
		mustScopeDescription(t, "desc"),
	)
	role := domain.NewRole(
		domain.MustRoleID(roleID),
		mustRoleValue(t, graphRole),
		mustRoleDescription(t, "desc"),
		[]domain.Scope{scope},
	)
	reg := domain.NewAppRegistration(
		domain.MustAppName("Graph"),
		graphClientID,
		mustIdentifierURI(t, graphURI),
		[]domain.RedirectURL{redirectURL},
		[]domain.Scope{scope},
		[]domain.Role{role},
	)
	client := domain.NewClientWithSecret(
		domain.MustAppName("Client"),
		graphClientID,
		domain.MustClientSecret(testSecret),
		[]domain.RedirectURL{redirectURL},
		nil,
	)
	user := domain.NewUser(
		domain.MustUserID(testUserID),
		domain.MustUsername(testUser),
		domain.MustPassword(testPass),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		nil,
	)

	tenant, err := domain.NewTenant(
		domain.MustTenantID(testTenantID),
		domain.MustTenantName("Tenant"),
		[]domain.AppRegistration{reg},
		nil,
		[]domain.User{user},
		[]domain.Client{client},
	)
	if err != nil {
		t.Fatalf("NewTenant: %v", err)
	}

	return graphTestFixture{client: client, user: user, tenant: tenant}
}

func verifyGraphClaims(t *testing.T, claims jwt.MapClaims, scope, aud, role string) {
	t.Helper()

	gotScp, ok := claims["scp"].(string)
	if !ok {
		t.Fatalf("expected scp string, got %#v", claims["scp"])
	}

	if gotScp == scope {
		t.Fatal("expected graph scp to be roles-based, not the requested scope")
	}

	if gotScp != role {
		t.Fatalf("expected scp %q, got %q", role, gotScp)
	}

	if gotAud, ok := claims["aud"].(string); !ok || gotAud != aud {
		t.Fatalf("expected aud %q, got %#v", aud, claims["aud"])
	}
}
