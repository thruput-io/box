package app_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"testing"

	"identity/app"
	"identity/domain"
)

func TestIssueAuthCode_Claims(t *testing.T) {
	t.Parallel()

	fixture := mustTenantFixture(t)
	key := mustRSAKey(t)

	const nonceStr = "nonce-123"

	code := app.ExportIssueAuthCode(
		key,
		fixture.user,
		fixture.client.ClientID(),
		testCallback,
		"openid profile",
		fixture.tenant.TenantID().String(),
		nonceStr,
	)

	claims, err := app.ExportParseSignedToken(key, code)
	if err != nil {
		t.Fatal(err)
	}

	verifyAuthCodeClaims(
		t,
		claims,
		fixture.user.ID().String(),
		fixture.client.ClientID().String(),
		nonceStr,
		fixture.tenant.TenantID().String(),
	)
}

func verifyAuthCodeClaims(t *testing.T, claims map[string]any, sub, clientID, nonce, tenant string) {
	t.Helper()

	verifySubAndClient(t, claims, sub, clientID)
	verifyOtherClaims(t, claims, nonce, tenant)

	if _, hasExp := claims[app.ClaimExp].(float64); !hasExp {
		t.Fatal("expected exp claim")
	}
}

func verifySubAndClient(t *testing.T, claims map[string]any, sub, clientID string) {
	t.Helper()

	gotSub, subFound := claims[app.ClaimSub].(string)
	if !subFound || gotSub != sub {
		t.Fatalf("expected sub %q, got %q", sub, gotSub)
	}

	gotCID, cidFound := claims[app.ClaimClientID].(string)
	if !cidFound || gotCID != clientID {
		t.Fatalf("expected client_id %q, got %q", clientID, gotCID)
	}
}

func verifyOtherClaims(t *testing.T, claims map[string]any, nonce, tenant string) {
	t.Helper()

	verifyAuthURIAndScope(t, claims)

	gotTID, tidFound := claims[app.ClaimTenant].(string)
	if !tidFound || gotTID != tenant {
		t.Fatalf("expected tenant %q, got %q", tenant, gotTID)
	}

	gotNonce, nonceFound := claims[app.ClaimNonce].(string)
	if !nonceFound || gotNonce != nonce {
		t.Fatalf("expected nonce %q, got %q", nonce, gotNonce)
	}
}

func verifyAuthURIAndScope(t *testing.T, claims map[string]any) {
	t.Helper()

	gotURI, uriFound := claims[app.ClaimRedirectURI].(string)
	if !uriFound || gotURI != testCallback {
		t.Fatalf("expected redirect_uri %q, got %q", testCallback, gotURI)
	}

	const oidcScope = "openid profile"

	gotScp, scpFound := claims[app.ClaimScope].(string)
	if !scpFound || gotScp != oidcScope {
		t.Fatalf("expected scope %q, got %q", oidcScope, gotScp)
	}
}

func TestParseSignedToken_InvalidToken(t *testing.T) {
	t.Parallel()

	key := mustRSAKey(t)

	_, err := app.ExportParseSignedToken(key, "not-a-token")
	if err == nil {
		t.Fatal("expected invalid token error")
	}

	if !strings.Contains(err.Error(), "invalid token") {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func TestParseSignedToken_InvalidSignature(t *testing.T) {
	t.Parallel()

	keyA := mustRSAKey(t)
	keyB := mustRSAKey(t)

	const otherSub = "subject"

	token := app.ExportSignClaims(keyA, map[string]any{app.ClaimSub: otherSub})

	_, err := app.ExportParseSignedToken(keyB, token)
	if err == nil {
		t.Fatal("expected invalid token error")
	}
}

func TestBuildClientInfo(t *testing.T) {
	t.Parallel()

	encoded := app.BuildClientInfo("user-id", "tenant-id")

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}

	var payload map[string]string

	err = json.Unmarshal(decoded, &payload)
	if err != nil {
		t.Fatal(err)
	}

	if payload["uid"] != "user-id" {
		t.Fatalf("expected uid user-id, got %q", payload["uid"])
	}

	if payload["utid"] != "tenant-id" {
		t.Fatalf("expected utid tenant-id, got %q", payload["utid"])
	}
}

func TestPublicKey(t *testing.T) {
	t.Parallel()

	key := mustRSAKey(t)

	jwks := app.PublicKey(key)

	expectedN := base64.RawURLEncoding.EncodeToString(key.N.Bytes())
	expectedE := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())

	if jwks.N != expectedN {
		t.Fatalf("expected N %q, got %q", expectedN, jwks.N)
	}

	if jwks.E != expectedE {
		t.Fatalf("expected E %q, got %q", expectedE, jwks.E)
	}
}

func TestLookup_NotFoundErrors(t *testing.T) {
	t.Parallel()

	fixture := mustTenantFixture(t)
	cfg := mustConfigWithTenant(t, fixture.tenant)

	const missingID = "99999999-9999-4999-8999-999999999999"

	_, err := app.ExportFindTenantByID(cfg, domain.MustTenantID(missingID))
	if !errors.Is(err, domain.ErrTenantNotFound) {
		t.Fatalf("expected ErrTenantNotFound, got %v", err)
	}

	missingClientID := domain.MustClientID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")

	_, err = app.ExportFindClient(fixture.tenant, missingClientID)
	if !errors.Is(err, domain.ErrClientNotFound) {
		t.Fatalf("expected ErrClientNotFound, got %v", err)
	}

	_, err = app.ExportFindAppRegistration(fixture.tenant, missingClientID)
	if !errors.Is(err, domain.ErrAppNotFound) {
		t.Fatalf("expected ErrAppNotFound, got %v", err)
	}
}
