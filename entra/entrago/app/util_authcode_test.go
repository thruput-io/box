package app

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"testing"

	"identity/domain"
)

func TestIssueAuthCode_Claims(t *testing.T) {
	t.Parallel()

	tenant, client, _, user := mustTenantWithClientAndRegistration(t)
	key := mustRSAKey(t)

	code := IssueAuthCode(
		key,
		user,
		client.ClientID(),
		"https://example.com/callback",
		"openid profile",
		tenant.TenantID().String(),
		"nonce-123",
	)

	claims, err := ParseSignedToken(key, code)
	if err != nil {
		t.Fatalf("ParseSignedToken: %v", err)
	}

	if got, _ := claims[claimSub].(string); got != user.ID().String() {
		t.Fatalf("expected sub %q, got %q", user.ID().String(), got)
	}

	if got, _ := claims[claimClientID].(string); got != client.ClientID().String() {
		t.Fatalf("expected client_id %q, got %q", client.ClientID().String(), got)
	}

	if got, _ := claims[claimRedirectURI].(string); got != "https://example.com/callback" {
		t.Fatalf("expected redirect_uri %q, got %q", "https://example.com/callback", got)
	}

	if got, _ := claims[claimScope].(string); got != "openid profile" {
		t.Fatalf("expected scope %q, got %q", "openid profile", got)
	}

	if got, _ := claims[claimTenant].(string); got != tenant.TenantID().String() {
		t.Fatalf("expected tenant %q, got %q", tenant.TenantID().String(), got)
	}

	if got, _ := claims[claimNonce].(string); got != "nonce-123" {
		t.Fatalf("expected nonce %q, got %q", "nonce-123", got)
	}

	if _, ok := claims[claimExp].(float64); !ok {
		t.Fatalf("expected exp claim")
	}
}

func TestParseSignedToken_InvalidToken(t *testing.T) {
	t.Parallel()

	key := mustRSAKey(t)

	_, err := ParseSignedToken(key, "not-a-token")
	if err == nil {
		t.Fatalf("expected invalid token error")
	}

	if !strings.Contains(err.Error(), "invalid token") {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func TestParseSignedToken_InvalidSignature(t *testing.T) {
	t.Parallel()

	keyA := mustRSAKey(t)
	keyB := mustRSAKey(t)

	token := SignClaims(keyA, map[string]any{claimSub: "subject"})

	_, err := ParseSignedToken(keyB, token)
	if err == nil {
		t.Fatalf("expected invalid token error")
	}
}

func TestBuildClientInfo(t *testing.T) {
	t.Parallel()

	encoded := BuildClientInfo("user-id", "tenant-id")

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("DecodeString: %v", err)
	}

	var payload map[string]string

	err = json.Unmarshal(decoded, &payload)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
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

	jwks := PublicKey(key)

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

	tenant, _, _, _ := mustTenantWithClientAndRegistration(t)
	cfg := mustConfigWithTenant(t, tenant)

	_, err := FindTenantByID(cfg, domain.MustTenantID("99999999-9999-4999-8999-999999999999"))
	if !errors.Is(err, domain.ErrTenantNotFound) {
		t.Fatalf("expected ErrTenantNotFound, got %v", err)
	}

	missingClientID := domain.MustClientID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")

	_, err = FindClient(tenant, missingClientID)
	if !errors.Is(err, domain.ErrClientNotFound) {
		t.Fatalf("expected ErrClientNotFound, got %v", err)
	}

	_, err = FindAppRegistration(tenant, missingClientID)
	if !errors.Is(err, domain.ErrAppNotFound) {
		t.Fatalf("expected ErrAppNotFound, got %v", err)
	}
}
