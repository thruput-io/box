package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"identity/domain"
	"identity/http/handlers"
)

func TestAdapter_OkTextAndInternalError(t *testing.T) {
	t.Parallel()

	resp := handlers.ExportOkText([]byte("hello"))
	if resp.Status != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Status)
	}

	if resp.ContentType != handlers.ContentTypePlain {
		t.Fatalf("unexpected content type: %q", resp.ContentType)
	}

	resp = handlers.ExportInternalError("boom")
	if resp.Status != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, resp.Status)
	}
}

func TestAdapter_FromDomainError_WritesOAuthError(t *testing.T) {
	t.Parallel()

	resp := handlers.ExportFromDomainError(domain.NewError(domain.ErrCodeInvalidCredentials, "bad"))
	if resp.Status != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, resp.Status)
	}

	if !strings.Contains(string(resp.Body), "\"error\":\"invalid_grant\"") {
		t.Fatalf("unexpected body: %s", string(resp.Body))
	}
}

func TestAdapter_WriteSetsHeadersAndBody(t *testing.T) {
	t.Parallel()

	const (
		testBody    = "x"
		testHeader  = "X-Test"
		testVal     = "1"
		contentType = "text/plain"
	)

	recorder := httptest.NewRecorder()
	handlers.Write(recorder, handlers.Response{
		Status:      http.StatusCreated,
		Body:        []byte(testBody),
		ContentType: contentType,
		Headers:     map[string]string{testHeader: testVal},
	})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, recorder.Code)
	}

	if recorder.Header().Get("Content-Type") != contentType {
		t.Fatalf("unexpected Content-Type: %q", recorder.Header().Get("Content-Type"))
	}

	if recorder.Header().Get(testHeader) != testVal {
		t.Fatalf("unexpected %s: %q", testHeader, recorder.Header().Get(testHeader))
	}

	if recorder.Body.String() != testBody {
		t.Fatalf("unexpected body: %q", recorder.Body.String())
	}
}
