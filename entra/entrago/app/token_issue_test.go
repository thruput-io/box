package app

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"identity/domain"
)

func mustRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	return key
}

func parseJWT(t *testing.T, raw string, key *rsa.PrivateKey) jwt.MapClaims {
	t.Helper()

	parsed, err := jwt.Parse(raw, func(_ *jwt.Token) (any, error) { return &key.PublicKey, nil })
	if err != nil {
		t.Fatalf("jwt.Parse: %v", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("expected MapClaims")
	}

	return claims
}

func TestIssueToken_OpenIDNonceRefreshAndClientInfo(t *testing.T) {
	t.Parallel()

	tenant, client, _, user := mustTenantWithClientAndRegistration(t)
	key := mustRSAKey(t)

	input := domain.TokenInput{
		Grant:         domain.GrantPassword,
		Tenant:        tenant,
		User:          &user,
		Client:        &client,
		Scope:         "openid offline_access api://app/access",
		Nonce:         "nonce",
		IsV2:          true,
		BaseURL:       "https://login.microsoftonline.com",
		CorrelationID: "corr",
	}

	resp := IssueToken(key, input)

	if resp.AccessToken == "" {
		t.Fatalf("expected access token")
	}

	if resp.IDToken == nil {
		t.Fatalf("expected id token")
	}

	if resp.RefreshToken == nil {
		t.Fatalf("expected refresh token")
	}

	if resp.ClientInfo == nil {
		t.Fatalf("expected client info")
	}

	if resp.CorrelationID != "corr" {
		t.Fatalf("expected correlation id corr, got %q", resp.CorrelationID)
	}

	accessClaims := parseJWT(t, resp.AccessToken, key)
	if got, ok := accessClaims["scp"].(string); !ok || got != input.Scope {
		t.Fatalf("expected access scp %q, got %#v", input.Scope, accessClaims["scp"])
	}

	idClaims := parseJWT(t, *resp.IDToken, key)
	if got, ok := idClaims["nonce"].(string); !ok || got != "nonce" {
		t.Fatalf("expected id_token nonce %q, got %#v", "nonce", idClaims["nonce"])
	}

	if gotAud, ok := idClaims["aud"].(string); !ok || gotAud != client.ClientID().String() {
		t.Fatalf("expected id_token aud %q, got %#v", client.ClientID().String(), idClaims["aud"])
	}
}

func TestIssueToken_GraphAudienceUsesRolesForScp(t *testing.T) {
	t.Parallel()

	graphClientID := domain.MustClientID("00000003-0000-0000-c000-000000000000")
	redirectURL := mustRedirectURL(t, "https://example.com/callback")

	scope := domain.NewScope(
		domain.MustScopeID("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"),
		mustScopeValue(t, "User.Read"),
		mustScopeDescription(t, "desc"),
	)
	role := domain.NewRole(
		domain.MustRoleID("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb"),
		mustRoleValue(t, "GraphRole"),
		mustRoleDescription(t, "desc"),
		[]domain.Scope{scope},
	)

	graphRegistration := domain.NewAppRegistration(
		domain.MustAppName("Graph"),
		graphClientID,
		mustIdentifierURI(t, "https://graph.microsoft.com"),
		[]domain.RedirectURL{redirectURL},
		[]domain.Scope{scope},
		[]domain.Role{role},
	)

	client := domain.NewClient(
		domain.MustAppName("Client"),
		graphClientID,
		domain.NewClientSecret("secret"),
		[]domain.RedirectURL{redirectURL},
		nil,
	)
	user := domain.NewUser(
		domain.MustUserID("33333333-3333-4333-8333-333333333333"),
		domain.MustUsername("user"),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		nil,
	)

	tenant, err := domain.NewTenant(
		domain.MustTenantID("11111111-1111-4111-8111-111111111111"),
		domain.MustTenantName("Tenant"),
		[]domain.AppRegistration{graphRegistration},
		nil,
		[]domain.User{user},
		[]domain.Client{client},
	)
	if err != nil {
		t.Fatalf("NewTenant: %v", err)
	}

	key := mustRSAKey(t)
	input := domain.TokenInput{
		Grant:   domain.GrantPassword,
		Tenant:  tenant,
		User:    &user,
		Client:  &client,
		Scope:   "https://graph.microsoft.com/User.Read",
		IsV2:    true,
		BaseURL: "https://login.microsoftonline.com",
	}

	resp := IssueToken(key, input)
	accessClaims := parseJWT(t, resp.AccessToken, key)

	gotScp, ok := accessClaims["scp"].(string)
	if !ok {
		t.Fatalf("expected scp string, got %#v", accessClaims["scp"])
	}

	if gotScp == input.Scope {
		t.Fatalf("expected graph scp to be roles-based, not the requested scope")
	}

	if gotScp != "GraphRole" {
		t.Fatalf("expected scp %q, got %q", "GraphRole", gotScp)
	}

	if gotAud, ok := accessClaims["aud"].(string); !ok || gotAud != "https://graph.microsoft.com" {
		t.Fatalf("expected aud %q, got %#v", "https://graph.microsoft.com", accessClaims["aud"])
	}

	if _, err := jwt.Parse(resp.AccessToken, func(_ *jwt.Token) (any, error) { return &key.PublicKey, nil }); err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			t.Fatalf("signature invalid")
		}
	}
}
