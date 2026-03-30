package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"identity/http/handlers"
)

const (
	testSignURL = testServerURL + pathMockUtilsSign
	emptyString = ""
)

func TestSignTokenHandler_WithProdToken(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)

	tokenJSON, err := os.ReadFile("testdata/test-token.json")
	if err != nil {
		t.Fatalf("failed to read testdata/test-token.json: %v", err)
	}

	ctx := context.Background()

	body := strings.NewReader(string(tokenJSON))

	request := httptest.NewRequestWithContext(ctx, http.MethodPost, testSignURL, body)

	resp := handlers.ExportInvokeSignTokenHandler(request, fixture.application)
	if resp.Status != http.StatusOK {
		t.Fatalf("unexpected status: %d, body: %s", resp.Status, string(resp.Body))
	}

	signedToken := string(resp.Body)
	if signedToken == emptyString {
		t.Fatal("expected signed token, got empty response")
	}

	verifySignedToken(t, fixture, tokenJSON, signedToken)
}

func verifySignedToken(t *testing.T, _ tokenHandlerFixture, tokenJSON []byte, signedToken string) {
	t.Helper()

	claims := jwt.MapClaims{}

	_, _, err := jwt.NewParser().ParseUnverified(signedToken, claims)
	if err != nil {
		t.Fatalf("failed to parse signed token: %v", err)
	}

	parsedClaims := claims

	var originalClaims map[string]any

	err = json.Unmarshal(tokenJSON, &originalClaims)
	if err != nil {
		t.Fatalf("failed to unmarshal original claims: %v", err)
	}

	for claimKey, claimValue := range originalClaims {
		actual, ok := parsedClaims[claimKey]
		if !ok {
			t.Errorf("claim %s missing in signed token", claimKey)

			continue
		}

		if fmtValue(actual) != fmtValue(claimValue) {
			t.Errorf("claim %s mismatch: expected %v, got %v", claimKey, claimValue, actual)
		}
	}
}

func TestSignTokenHandler_WithQueryParam(t *testing.T) {
	t.Parallel()

	fixture := MustAppForTestTokenHandler(t)

	const tokenJSON = `{"sub":"user123","scp":"read write"}`

	ctx := context.Background()

	baseURL := testServerURL + pathMockUtilsSign + "?token="

	requestURL := baseURL + url.QueryEscape(tokenJSON)

	request := httptest.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)

	resp := handlers.ExportInvokeSignTokenHandler(request, fixture.application)
	if resp.Status != http.StatusOK {
		t.Fatalf("unexpected status: %d, body: %s", resp.Status, string(resp.Body))
	}

	signedToken := string(resp.Body)
	claims := jwt.MapClaims{}

	_, _, err := jwt.NewParser().ParseUnverified(signedToken, claims)
	if err != nil {
		t.Fatalf("failed to parse signed token: %v", err)
	}

	parsedClaims := claims

	if parsedClaims["sub"] != "user123" || parsedClaims["scp"] != "read write" {
		t.Errorf("unexpected claims: %v", parsedClaims)
	}
}

func fmtValue(v any) string {
	valStr := fmt.Sprintf("%v", v)

	noBracketOpen := strings.ReplaceAll(valStr, "[", emptyString)

	noBracket := strings.ReplaceAll(noBracketOpen, "]", emptyString)

	return strings.TrimSpace(strings.ToLower(noBracket))
}
