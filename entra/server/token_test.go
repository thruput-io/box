package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"identity/domain"
)

func TestToken_PasswordGrant(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.Config.Tenants()[0].Clients()[0].ClientID().String()

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", "testuser")
	form.Set("password", "password")
	form.Set("client_id", clientID)

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %v, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestToken_PasswordGrant_WrongPassword(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", "testuser")
	form.Set("password", "wrong")

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %v", recorder.Code)
	}
}

func TestToken_PasswordGrant_MissingUsername(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	form := url.Values{}
	form.Set("grant_type", "password")

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %v", recorder.Code)
	}
}

func TestToken_ClientCredentials(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	secretClientID := domain.MustClientID("cccccccc-cccc-4ccc-accc-cccccccccccc")
	redirectURL, _ := domain.NewRedirectURL("http://localhost/callback")
	secretClient := domain.NewClient(
		domain.MustAppName("SecretClient"),
		secretClientID,
		domain.NewClientSecret("secret"),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	srv := serverWithClient(t, server, secretClient)

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", secretClientID.String())
	form.Set("client_secret", "secret")

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %v, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestToken_ClientCredentials_WrongSecret(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	secretClientID := domain.MustClientID("cccccccc-cccc-4ccc-accc-cccccccccccc")
	redirectURL, _ := domain.NewRedirectURL("http://localhost/callback")
	secretClient := domain.NewClient(
		domain.MustAppName("SecretClient"),
		secretClientID,
		domain.NewClientSecret("secret"),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	srv := serverWithClient(t, server, secretClient)

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", secretClientID.String())
	form.Set("client_secret", "wrong")

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %v", recorder.Code)
	}
}

func TestToken_AuthorizationCode(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.Config.Tenants()[0].Clients()[0].ClientID().String()

	claims := jwt.MapClaims{
		"sub":          "user1",
		"client_id":    clientID,
		"redirect_uri": "http://localhost/callback",
		"scope":        "openid",
		"exp":          time.Now().Add(5 * time.Minute).Unix(),
	}
	authCode, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(server.Key)

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", authCode)
	form.Set("redirect_uri", "http://localhost/callback")

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %v, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestToken_RefreshToken(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.Config.Tenants()[0].Clients()[0].ClientID().String()

	claims := jwt.MapClaims{
		"sub":       "user1",
		"client_id": clientID,
		"exp":       time.Now().Add(5 * time.Minute).Unix(),
		"typ":       "Refresh",
	}
	refreshToken, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(server.Key)

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %v, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestToken_UnsupportedGrant(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	form := url.Values{}
	form.Set("grant_type", "implicit")

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %v", recorder.Code)
	}
}

func TestToken_ParamTooLong(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	form := url.Values{}
	form.Set("grant_type", strings.Repeat("a", 2049))

	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %v", recorder.Code)
	}
}
