package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"identity/domain"
)

func TestAdapter_OkTextAndInternalError(t *testing.T) {
	t.Parallel()

	resp := okText([]byte("hello"))
	if resp.Status != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.Status)
	}

	if resp.ContentType != contentTypePlain {
		t.Fatalf("unexpected content type: %q", resp.ContentType)
	}

	resp = internalError("boom")
	if resp.Status != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, resp.Status)
	}
}

func TestAdapter_FromDomainError_WritesOAuthError(t *testing.T) {
	t.Parallel()

	resp := fromDomainError(domain.NewError(domain.ErrCodeInvalidCredentials, "bad"))
	if resp.Status != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, resp.Status)
	}

	if !strings.Contains(string(resp.Body), "\"error\":\"invalid_grant\"") {
		t.Fatalf("unexpected body: %s", string(resp.Body))
	}
}

func TestAdapter_WriteSetsHeadersAndBody(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	Write(recorder, Response{
		Status:      http.StatusCreated,
		Body:        []byte("x"),
		ContentType: "text/plain",
		Headers:     map[string]string{"X-Test": "1"},
	})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, recorder.Code)
	}

	if recorder.Header().Get("Content-Type") != "text/plain" {
		t.Fatalf("unexpected Content-Type: %q", recorder.Header().Get("Content-Type"))
	}

	if recorder.Header().Get("X-Test") != "1" {
		t.Fatalf("unexpected X-Test: %q", recorder.Header().Get("X-Test"))
	}

	if recorder.Body.String() != "x" {
		t.Fatalf("unexpected body: %q", recorder.Body.String())
	}
}
