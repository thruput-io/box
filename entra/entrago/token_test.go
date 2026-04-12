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

func postToken(t *testing.T, server interface{ Handler() http.Handler }, form url.Values) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(
		context.Background(), http.MethodPost, tokenEndpoint,
		strings.NewReader(form.Encode()),
	)
	request.Header.Set(headerContentType, headerValueForm)

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	return recorder
}

func TestToken_PasswordGrant(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.App.Config.Tenants()[firstIndex].Clients()[firstIndex].ClientID().UUID().String()

	form := url.Values{}
	form.Set(formGrantType, grantTypePassword)
	form.Set(formUsername, testUsername)
	form.Set(formPassword, testPassword)
	form.Set(formClientID, clientID)

	recorder := postToken(t, server, form)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusBodyFmt, recorder.Code, recorder.Body.String())
	}
}

func TestToken_PasswordGrant_WrongPassword(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	form := url.Values{}
	form.Set(formGrantType, grantTypePassword)
	form.Set(formUsername, testUsername)
	form.Set(formPassword, "wrong")

	recorder := postToken(t, server, form)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestToken_PasswordGrant_MissingUsername(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	form := url.Values{}
	form.Set(formGrantType, grantTypePassword)

	recorder := postToken(t, server, form)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestToken_ClientCredentials(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	secretClientID := mustClientID(t, "cccccccc-cccc-4ccc-accc-cccccccccccc")
	redirectURL := mustRedirectURL(t, testRedirectURL)
	secretClient := domain.NewClientWithSecret(
		mustAppName(t, "SecretClient"),
		secretClientID,
		mustClientSecret(t, testClientSecret),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	srv := serverWithClient(t, server, secretClient)

	form := url.Values{}
	form.Set(formGrantType, "client_credentials")
	form.Set(formClientID, secretClientID.UUID().String())
	form.Set(formClientSecret, testClientSecret)

	recorder := postToken(t, srv, form)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusBodyFmt, recorder.Code, recorder.Body.String())
	}
}

func TestToken_ClientCredentials_WrongSecret(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	secretClientID := mustClientID(t, "cccccccc-cccc-4ccc-accc-cccccccccccc")
	redirectURL := mustRedirectURL(t, testRedirectURL)
	secretClient := domain.NewClientWithSecret(
		mustAppName(t, "SecretClient"),
		secretClientID,
		mustClientSecret(t, testClientSecret),
		[]domain.RedirectURL{redirectURL},
		nil,
	)

	srv := serverWithClient(t, server, secretClient)

	form := url.Values{}
	form.Set(formGrantType, "client_credentials")
	form.Set(formClientID, secretClientID.UUID().String())
	form.Set(formClientSecret, "wrong")

	recorder := postToken(t, srv, form)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestToken_AuthorizationCode(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.App.Config.Tenants()[firstIndex].Clients()[firstIndex].ClientID().UUID().String()

	claims := jwt.MapClaims{
		"sub": "user1",
		"azp": clientID,
		"scp": "openid",
		"exp": time.Now().Add(5 * time.Minute).Unix(),
	}

	authCode, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(server.App.Key)
	if err != nil {
		t.Fatalf("sign auth code: %v", err)
	}

	form := url.Values{}
	form.Set(formGrantType, "authorization_code")
	form.Set(formCode, authCode)
	form.Set(formRedirectURI, testRedirectURL)

	recorder := postToken(t, server, form)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusBodyFmt, recorder.Code, recorder.Body.String())
	}
}

func TestToken_RefreshToken(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	clientID := server.App.Config.Tenants()[firstIndex].Clients()[firstIndex].ClientID().UUID().String()

	claims := jwt.MapClaims{
		"sub": "user1",
		"azp": clientID,
		"scp": "openid",
		"exp": time.Now().Add(5 * time.Minute).Unix(),
		"typ": "Refresh",
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(server.App.Key)
	if err != nil {
		t.Fatalf("sign refresh token: %v", err)
	}

	form := url.Values{}
	form.Set(formGrantType, "refresh_token")
	form.Set(formRefreshToken, refreshToken)

	recorder := postToken(t, server, form)

	if recorder.Code != http.StatusOK {
		t.Errorf(statusBodyFmt, recorder.Code, recorder.Body.String())
	}
}

func TestToken_UnsupportedGrant(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	form := url.Values{}
	form.Set(formGrantType, "implicit")

	recorder := postToken(t, server, form)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(statusFmt, recorder.Code)
	}
}

func TestToken_ParamTooLong(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)

	form := url.Values{}
	form.Set(formGrantType, strings.Repeat("a", overMaxParamLength))

	recorder := postToken(t, server, form)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf(statusFmt, recorder.Code)
	}
}
