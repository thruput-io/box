package handlers_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"identity/app"
	"identity/domain"
	"identity/http/handlers"
)

const (
	testAppID       = "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"
	testTenantID    = "10000000-0000-4000-a000-000000000000"
	pathSeparator   = "/"
	testTokenPrefix = "/test-tokens/"
	fmtUnexpected   = "unexpected status: %d, body: %s"
)

func TestTestTokenHandler_ProdIntegration_1a(t *testing.T) {
	t.Parallel()

	config := loadConfigForTest(t)
	tenantID := testTenantID
	appID := testAppID
	clientID := testAppID

	application := &app.App{
		Config:        config,
		Key:           mustRSAKey(t),
		LoginTemplate: nil,
		IndexTemplate: nil,
	}

	testURL := testTokenPrefix + tenantID + pathSeparator + appID + pathSeparator + clientID
	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)

	resp := handlers.ExportInvokeTestTokenHandler(request, application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtUnexpected, resp.Status, string(resp.Body))
	}

	compareTokenWithExpected(t, string(resp.Body), "testdata/expected-token-1a.json")
}

func TestTestTokenHandler_ProdIntegration_1b(t *testing.T) {
	t.Parallel()

	config := loadConfigForTest(t)
	tokenContent := loadProdToken(t, "testdata/test-token-all.json")

	var inputClaims map[string]any

	err := json.Unmarshal(tokenContent, &inputClaims)
	if err != nil {
		t.Fatalf("failed to unmarshal input token: %v", err)
	}

	tenantID := testTenantID
	appID := testAppID
	clientID := testAppID

	application := &app.App{
		Config:        config,
		Key:           mustRSAKey(t),
		LoginTemplate: nil,
		IndexTemplate: nil,
	}

	baseURL := testTokenPrefix + tenantID + pathSeparator + appID + pathSeparator + clientID
	queryParams := url.Values{}

	for k, v := range inputClaims {
		addClaimAsQueryParam(queryParams, k, v)
	}

	fullURL := baseURL + "?" + queryParams.Encode()
	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)

	resp := handlers.ExportInvokeTestTokenHandler(request, application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtUnexpected, resp.Status, string(resp.Body))
	}

	compareTokenWithExpected(t, string(resp.Body), "testdata/expected-token-1b.json")
}

func TestTestTokenHandler_ProdIntegration_1c(t *testing.T) {
	t.Parallel()

	config := loadConfigForTest(t)
	tokenContent := loadProdToken(t, "testdata/test-token-all.json")

	tenantID := testTenantID
	appID := testAppID
	clientID := testAppID

	application := &app.App{
		Config:        config,
		Key:           mustRSAKey(t),
		LoginTemplate: nil,
		IndexTemplate: nil,
	}

	testURL := testTokenPrefix + tenantID + pathSeparator + appID + pathSeparator + clientID
	ctx := context.Background()
	request := httptest.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(tokenContent))
	request.Header.Set("Content-Type", "application/json")

	resp := handlers.ExportInvokeTestTokenHandler(request, application)
	if resp.Status != http.StatusOK {
		t.Fatalf(fmtUnexpected, resp.Status, string(resp.Body))
	}

	// 1c should yield same results as 1b if input is same (but without mock-added claims)
	compareTokenWithExpected(t, string(resp.Body), "testdata/expected-token-1c.json")
}

func addClaimAsQueryParam(queryParams url.Values, key string, val any) {
	// Skip dynamic claims that we want to be generated or we handle with placeholders
	if key == "iat" || key == "exp" || key == "nbf" {
		return
	}

	// Map 'scp' to 'scope' because testTokenHandler uses 'scope' query param
	queryKey := key
	if key == "scp" {
		queryKey = "scope"
	}

	if arr, ok := val.([]any); ok {
		for _, item := range arr {
			queryParams.Add(queryKey, fmt.Sprintf("%v", item))
		}
	} else {
		queryParams.Set(queryKey, fmt.Sprintf("%v", val))
	}
}

func compareTokenWithExpected(t *testing.T, tokenStr string, expectedPath string) {
	t.Helper()

	const expectedParts = 3

	parts := strings.Split(tokenStr, ".")
	if len(parts) != expectedParts {
		t.Fatalf("invalid token format: expected 3 parts, got %d", len(parts))
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	var actual map[string]any

	err = json.Unmarshal(payloadJSON, &actual)
	if err != nil {
		t.Fatalf("failed to unmarshal actual payload: %v", err)
	}

	expectedData, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read expected file %s: %v", expectedPath, err)
	}

	var expected map[string]any

	err = json.Unmarshal(expectedData, &expected)
	if err != nil {
		t.Fatalf("failed to unmarshal expected payload: %v", err)
	}

	compareClaims(t, actual, expected)
}

func compareClaims(t *testing.T, actual, expected map[string]any) {
	t.Helper()

	checkExpectedClaims(t, actual, expected)
	checkUnexpectedClaims(t, actual, expected)
}

func checkExpectedClaims(t *testing.T, actual, expected map[string]any) {
	t.Helper()

	for key, expectedVal := range expected {
		actualVal, ok := actual[key]
		if !ok {
			t.Errorf("missing claim: %s", key)

			continue
		}

		if expectedVal == "<DYNAMIC>" {
			continue
		}

		compareSingleClaim(t, key, actualVal, expectedVal)
	}
}

func checkUnexpectedClaims(t *testing.T, actual, expected map[string]any) {
	t.Helper()

	for key, val := range actual {
		if _, ok := expected[key]; !ok {
			t.Errorf("unexpected extra claim: %s (value: %v)", key, val)
		}
	}
}

func compareSingleClaim(t *testing.T, key string, actualVal, expectedVal any) {
	t.Helper()

	if reflect.DeepEqual(expectedVal, actualVal) {
		return
	}

	expectedBytes, err := json.Marshal(expectedVal)
	if err != nil {
		t.Errorf("failed to marshal expected for %s: %v", key, err)

		return
	}

	actualBytes, err := json.Marshal(actualVal)
	if err != nil {
		t.Errorf("failed to marshal actual for %s: %v", key, err)

		return
	}

	if string(expectedBytes) != string(actualBytes) {
		t.Errorf("claim mismatch for %s:\nexp: %v\nact: %v",
			key, string(expectedBytes), string(actualBytes))
	}
}

func loadConfigForTest(t *testing.T) *domain.Config {
	t.Helper()

	configPath := "DefaultConfig.yaml"

	_, err := os.Stat(configPath)
	if err != nil {
		configPath = "../../DefaultConfig.yaml"
	}

	config, err := domain.LoadConfig(configPath, "")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	return config
}

func loadProdToken(t *testing.T, tokenPath string) []byte {
	t.Helper()

	tokenContent, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Skip("skipping test: testdata/test-token.json not found")
	}

	return tokenContent
}
