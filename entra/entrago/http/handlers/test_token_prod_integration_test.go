package handlers_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"identity/app"
	"identity/config"
	"identity/domain"
	"identity/http/transport"
)

const (
	testAppID         = "aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"
	testTenantID      = "10000000-0000-4000-a000-000000000000"
	pathSeparator     = "/"
	fmtUnexpected     = "unexpected status: %d, body: %s"
	fmtFailedRequest  = "failed to make request: %v"
	fmtFailedReadBody = "failed to read body: %v"
	fmtFailedCreate   = "failed to create request: %v"
	permOwnerRW       = 0o600
)

func TestTestTokenHandler_ProdIntegration_1a(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)
	defer srv.Close()

	tenantID := testTenantID
	appID := testAppID
	clientID := testAppID

	testURL := srv.URL + pathMockUtils + tenantID + pathSeparator + appID + pathSeparator + clientID

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

	tokenStr := strings.TrimSpace(string(body))
	saveActualTokenPayload(t, tokenStr, "testdata/actual-token-1a.json")
	compareTokenWithExpected(t, tokenStr, "testdata/expected-token-1a.json")
}

func TestTestTokenHandler_ProdIntegration_1b(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)
	defer srv.Close()

	tokenContent := loadProdToken(t, "testdata/test-token-all.json")

	var inputClaims map[string]any

	err := json.Unmarshal(tokenContent, &inputClaims)
	if err != nil {
		t.Fatalf("failed to unmarshal input token: %v", err)
	}

	queryParams := url.Values{}

	for k, v := range inputClaims {
		addClaimAsQueryParam(queryParams, k, v)
	}

	baseURL := srv.URL + pathMockUtils + testTenantID + pathSeparator + testAppID + pathSeparator + testAppID
	fullURL := baseURL + "?" + queryParams.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fullURL, nil)
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

	tokenStr := strings.TrimSpace(string(body))
	saveActualTokenPayload(t, tokenStr, "testdata/actual-token-1b.json")
	compareTokenWithExpected(t, tokenStr, "testdata/expected-token-1b.json")
}

func TestTestTokenHandler_ProdIntegration_1c(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)
	defer srv.Close()

	tokenContent := loadProdToken(t, "testdata/test-token-all.json")

	tenantID := testTenantID
	appID := testAppID
	clientID := testAppID

	testURL := srv.URL + pathMockUtils + tenantID + pathSeparator + appID + pathSeparator + clientID

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, testURL, bytes.NewReader(tokenContent),
	)
	if err != nil {
		t.Fatalf(fmtFailedCreate, err)
	}

	req.Header.Set("Content-Type", "application/json")

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

	// 1c should yield same results as 1b if input is same (but without mock-added claims)
	tokenStr := strings.TrimSpace(string(body))
	saveActualTokenPayload(t, tokenStr, "testdata/actual-token-1c.json")
	compareTokenWithExpected(t, tokenStr, "testdata/expected-token-1c.json")
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

	cfg, err := config.LoadConfig(configPath, "")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	return cfg
}

func loadProdToken(t *testing.T, tokenPath string) []byte {
	t.Helper()

	tokenContent, err := os.ReadFile(tokenPath)
	if err != nil {
		panic(err)
	}

	return tokenContent
}

func saveActualTokenPayload(t *testing.T, tokenStr, outputPath string) {
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

	var payload map[string]any

	err = json.Unmarshal(payloadJSON, &payload)
	if err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	indented, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal payload with indent: %v", err)
	}

	// Add newline to match project formatting
	err = os.WriteFile(outputPath, append(indented, '\n'), permOwnerRW)
	if err != nil {
		t.Fatalf("failed to write payload to %s: %v", outputPath, err)
	}
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	cfg := loadConfigForTest(t)
	application := &app.App{
		Config:        cfg,
		Key:           mustRSAKey(t),
		LoginTemplate: nil,
		IndexTemplate: nil,
	}

	server := &transport.Server{App: application}

	return httptest.NewServer(server.Handler())
}
