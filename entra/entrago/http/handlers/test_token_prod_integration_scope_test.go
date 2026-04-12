package handlers_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
)

const (
	testFrontendClientID = "e697b97c-9b4b-487f-9f7a-248386f78864"
	emptyLen             = 0
)

// TestTestTokenHandler_ProdIntegration_2a verifies that
// GET /mock-utils/{tenantId}/{appId}/{clientId}?scope=read
// returns a valid JWT with the requested scope and roles resolved by scope matching alone,
// when the clientId resolves to an app registration (not a configured client with assignments).
func TestTestTokenHandler_ProdIntegration_2a(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)
	defer srv.Close()

	testURL := srv.URL + pathMockUtils +
		testTenantID + pathSeparator +
		testAppID + pathSeparator +
		testAppID + "?scope=read"

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testURL, nil)
	if err != nil {
		t.Fatalf(fmtFailedCreate, err)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf(fmtFailedRequest, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf(fmtUnexpected, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(fmtFailedReadBody, err)
	}

	claims := decodeTokenClaims(t, strings.TrimSpace(string(body)))

	assertClaimString(t, claims, "scp", "read")
	assertClaimString(t, claims, "tid", testTenantID)
	assertClaimString(t, claims, "azp", testAppID)
	assertClaimRoles(t, claims, []string{"read"})
}

// TestTestTokenHandler_ProdIntegration_2b verifies that
// GET /mock-utils/{tenantId}/{appId}/{clientId}?scope=read
// returns a valid JWT where roles are expanded by the client's configured groupRoleAssignments
// when the clientId is a configured client in the config file.
// The healer user belongs to the "healers" group which is assigned both read and write
// for the MyAppBackend application — so both roles appear in the token.
func TestTestTokenHandler_ProdIntegration_2b(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)
	defer srv.Close()

	testURL := srv.URL + pathMockUtils +
		testTenantID + pathSeparator +
		testAppID + pathSeparator +
		testFrontendClientID + "?scope=read"

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testURL, nil)
	if err != nil {
		t.Fatalf(fmtFailedCreate, err)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf(fmtFailedRequest, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf(fmtUnexpected, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(fmtFailedReadBody, err)
	}

	claims := decodeTokenClaims(t, strings.TrimSpace(string(body)))

	assertClaimString(t, claims, "scp", "read")
	assertClaimString(t, claims, "tid", testTenantID)
	assertClaimString(t, claims, "azp", testFrontendClientID)
	assertClaimRoles(t, claims, []string{"read", "write"})
}

func decodeTokenClaims(t *testing.T, tokenStr string) map[string]any {
	t.Helper()

	const expectedParts = 3

	parts := strings.Split(tokenStr, ".")
	if len(parts) != expectedParts {
		t.Fatalf("invalid token format: expected 3 parts, got %d", len(parts))
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode token payload: %v", err)
	}

	var claims map[string]any

	err = json.Unmarshal(payloadJSON, &claims)
	if err != nil {
		t.Fatalf("failed to unmarshal token claims: %v", err)
	}

	return claims
}

func assertClaimString(t *testing.T, claims map[string]any, key, expected string) {
	t.Helper()

	raw, exists := claims[key]
	if !exists {
		t.Errorf("claim %q: missing from token", key)

		return
	}

	actual, isString := raw.(string)
	if !isString {
		t.Errorf("claim %q: expected string %q, got %T(%v)", key, expected, raw, raw)

		return
	}

	if actual != expected {
		t.Errorf("claim %q: expected %q, got %q", key, expected, actual)
	}
}

func assertClaimRoles(t *testing.T, claims map[string]any, expectedRoles []string) {
	t.Helper()

	raw, exists := claims["roles"]
	if !exists {
		t.Error("claim \"roles\": missing from token")

		return
	}

	rawSlice, isAnySlice := raw.([]any)
	if !isAnySlice {
		t.Errorf("claim \"roles\": expected array, got %T(%v)", raw, raw)

		return
	}

	actual := make([]string, emptyLen, len(rawSlice))

	for _, v := range rawSlice {
		role, isString := v.(string)
		if !isString {
			t.Errorf("claim \"roles\": non-string element %T(%v)", v, v)

			return
		}

		actual = append(actual, role)
	}

	slices.Sort(actual)

	expected := make([]string, len(expectedRoles))
	copy(expected, expectedRoles)
	slices.Sort(expected)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("claim \"roles\": expected %v, got %v", expected, actual)
	}
}
