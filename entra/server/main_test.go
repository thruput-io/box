package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func requireNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func cleanupRemove(t *testing.T, path string) {
	t.Helper()
	t.Cleanup(func() {
		err := os.Remove(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			t.Errorf("remove %s: %v", path, err)
		}
	})
}

func init() {
	// Pre-load templates for tests using the embedded FS
	var err error

	loginTemplate, err = template.ParseFS(loginHTML, "templates/login.html")
	if err != nil {
		panic(err)
	}

	indexTemplate, err = template.ParseFS(indexHTML, "templates/index.html")
	if err != nil {
		panic(err)
	}
}

func setupTestConfig() {
	configData = Config{
		Tenants: []Tenant{
			{
				TenantID: uuid.MustParse("b5a920d6-7d3c-44fe-baad-4ffed6b8774d"),
				Name:     "Default Tenant",
				Users: []User{
					{
						ID:       uuid.New(),
						Username: "testuser",
						Password: "password",
					},
				},
				Clients: []Client{
					{
						ClientID:     uuid.New(),
						RedirectUrls: []string{"http://localhost/callback"},
					},
				},
			},
		},
	}
}

func TestHandlers(t *testing.T) {
	// Setup global state for tests
	var err error

	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	setupTestConfig()

	t.Run("Health", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/_health", nil)
		w := httptest.NewRecorder()
		health(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("health() status = %v, want %v", w.Code, http.StatusOK)
		}
	})

	t.Run("Discovery", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/.well-known/openid-configuration", nil)
		w := httptest.NewRecorder()
		discovery(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("discovery() status = %v, want %v", w.Code, http.StatusOK)
		}

		var resp DiscoveryResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		if err != nil {
			t.Errorf("failed to decode discovery response: %v", err)
		}
	})

	t.Run("Index", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
		w := httptest.NewRecorder()
		index(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("index() status = %v, want %v", w.Code, http.StatusOK)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Entra mock") {
			t.Errorf("index() body missing title")
		}

		if !strings.Contains(body, "Test endpoints") {
			t.Errorf("index() body missing Test endpoints section")
		}

		if !strings.Contains(body, "Current configuration") {
			t.Errorf("index() body missing Current configuration section")
		}

		if !strings.Contains(body, "Default Tenant") {
			t.Errorf("index() body missing tenant name")
		}
	})

	t.Run("IndexNotFound", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/not-found", nil)
		w := httptest.NewRecorder()
		index(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("index() status = %v, want %v", w.Code, http.StatusNotFound)
		}
	})

	t.Run("Authorize", func(t *testing.T) {
		u1 := uuid.New()
		configData.Tenants[0].Clients = append(configData.Tenants[0].Clients, Client{
			ClientID:     u1,
			RedirectUrls: []string{"http://localhost/callback"},
		})
		configData.Tenants[0].AppRegistrations = append(configData.Tenants[0].AppRegistrations, AppRegistration{
			Name:         "AppReg",
			ClientID:     uuid.New(),
			RedirectUrls: []string{"http://localhost/callback2"},
		})

		t.Run("SuccessClient", func(t *testing.T) {
			req := httptest.NewRequestWithContext(context.Background(), "GET", "/authorize?client_id="+u1.String()+"&redirect_uri=http://localhost/callback", nil)
			w := httptest.NewRecorder()
			authorize(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("authorize() status = %v, want %v", w.Code, http.StatusOK)
			}
		})

		t.Run("SuccessAppReg", func(t *testing.T) {
			clientID := configData.Tenants[0].AppRegistrations[len(configData.Tenants[0].AppRegistrations)-1].ClientID
			req := httptest.NewRequestWithContext(context.Background(), "GET", "/authorize?client_id="+clientID.String()+"&redirect_uri=http://localhost/callback2", nil)
			w := httptest.NewRecorder()
			authorize(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("authorize() status = %v", w.Code)
			}
		})

		t.Run("InvalidRedirect", func(t *testing.T) {
			req := httptest.NewRequestWithContext(context.Background(), "GET", "/authorize?client_id="+u1.String()+"&redirect_uri=http://wrong", nil)
			w := httptest.NewRecorder()
			authorize(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("authorize() status = %v", w.Code)
			}
		})
	})

	t.Run("Login", func(t *testing.T) {
		u1 := uuid.New()
		user := User{
			ID:       u1,
			Username: "loginuser",
			Password: "password",
		}
		configData.Tenants[0].Users = append(configData.Tenants[0].Users, user)

		t.Run("Success", func(t *testing.T) {
			form := url.Values{}
			form.Add("username", "loginuser")
			form.Add("password", "password")
			form.Add("client_id", configData.Tenants[0].Clients[0].ClientID.String())
			form.Add("redirect_uri", "http://localhost/callback")

			req := httptest.NewRequestWithContext(context.Background(), "POST", "/login", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			login(w, req)

			if w.Code != http.StatusFound {
				t.Errorf("login() status = %v, want %v", w.Code, http.StatusFound)
			}
		})

		t.Run("SuccessFragment", func(t *testing.T) {
			form := url.Values{}
			form.Add("username", "loginuser")
			form.Add("password", "password")
			form.Add("client_id", configData.Tenants[0].Clients[0].ClientID.String())
			form.Add("redirect_uri", "http://localhost/callback")
			form.Add("response_mode", "fragment")

			req := httptest.NewRequestWithContext(context.Background(), "POST", "/login", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			login(w, req)

			if w.Code != http.StatusFound {
				t.Errorf("login() status = %v", w.Code)
			}

			if !strings.Contains(w.Header().Get("Location"), "#") {
				t.Error("Location should contain # for fragment response_mode")
			}
		})

		t.Run("InvalidPassword", func(t *testing.T) {
			form := url.Values{}
			form.Add("username", "loginuser")
			form.Add("password", "wrong")
			form.Add("client_id", configData.Tenants[0].Clients[0].ClientID.String())
			form.Add("redirect_uri", "http://localhost/callback")

			req := httptest.NewRequestWithContext(context.Background(), "POST", "/login", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			login(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("login() status = %v", w.Code)
			}
		})
	})

	t.Run("TokenPassword", func(t *testing.T) {
		form := url.Values{}
		form.Add("grant_type", "password")
		form.Add("username", "testuser")
		form.Add("password", "password")
		form.Add("client_id", configData.Tenants[0].Clients[0].ClientID.String())

		req := httptest.NewRequestWithContext(context.Background(), "POST", "/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		token(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("token() status = %v, want %v, body=%s", w.Code, http.StatusOK, w.Body.String())
		}
	})

	t.Run("TokenClientCredentials", func(t *testing.T) {
		u1 := uuid.New()
		configData.Tenants[0].Clients = append(configData.Tenants[0].Clients, Client{
			ClientID:     u1,
			ClientSecret: "secret",
		})

		form := url.Values{}
		form.Add("grant_type", "client_credentials")
		form.Add("client_id", u1.String())
		form.Add("client_secret", "secret")

		req := httptest.NewRequestWithContext(context.Background(), "POST", "/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		token(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("token() status = %v, want %v, body=%s", w.Code, http.StatusOK, w.Body.String())
		}
	})

	t.Run("CallHome", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/discovery/instance", nil)
		w := httptest.NewRecorder()
		callHome(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("callHome() status = %v, want %v", w.Code, http.StatusOK)
		}
	})

	t.Run("CorsMiddleware", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := corsMiddleware(next)
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("corsMiddleware() status = %v, want %v", w.Code, http.StatusNoContent)
		}

		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("corsMiddleware() Origin = %v, want *", w.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("DiscoveryV2", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/mytenant/v2.0/.well-known/openid-configuration", nil)
		w := httptest.NewRecorder()
		discovery(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("discovery() status = %v", w.Code)
		}
	})

	t.Run("DiscoveryCommon", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/common/.well-known/openid-configuration", nil)
		w := httptest.NewRecorder()
		discovery(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("discovery() status = %v", w.Code)
		}
	})

	t.Run("JWKS", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/discovery/keys", nil)
		w := httptest.NewRecorder()
		jwks(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("jwks() status = %v, want %v", w.Code, http.StatusOK)
		}
	})

	t.Run("GetBaseURL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-Proto", "https")

		base := getBaseURL(req)
		if !strings.HasPrefix(base, "https://") {
			t.Errorf("getBaseURL() should be https, got %s", base)
		}
	})

	t.Run("TokenAuthCode", func(t *testing.T) {
		// Generate an auth code JWT
		claims := jwt.MapClaims{
			"sub":          "user1",
			"client_id":    configData.Tenants[0].Clients[0].ClientID.String(),
			"redirect_uri": "http://localhost/callback",
			"scope":        "openid",
			"exp":          time.Now().Add(5 * time.Minute).Unix(),
		}
		tokenJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		authCode, _ := tokenJWT.SignedString(privateKey)

		form := url.Values{}
		form.Add("grant_type", "authorization_code")
		form.Add("code", authCode)
		form.Add("redirect_uri", "http://localhost/callback")

		req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		token(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("token() status = %v, body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("TokenRefreshToken", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":       "user1",
			"client_id": configData.Tenants[0].Clients[0].ClientID.String(),
			"exp":       time.Now().Add(5 * time.Minute).Unix(),
			"typ":       "Refresh",
		}
		tokenJWT := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		refreshToken, _ := tokenJWT.SignedString(privateKey)

		form := url.Values{}
		form.Add("grant_type", "refresh_token")
		form.Add("refresh_token", refreshToken)

		req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		token(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("token() status = %v, body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("TokenV2", func(t *testing.T) {
		form := url.Values{}
		form.Add("grant_type", "password")
		form.Add("username", "testuser")
		form.Add("password", "password")
		form.Add("client_id", configData.Tenants[0].Clients[0].ClientID.String())
		form.Add("scope", "openid offline_access")

		req := httptest.NewRequest(http.MethodPost, "/mytenant/oauth2/v2.0/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		token(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("token() status = %v", w.Code)
		}
	})

	t.Run("TokenErrors", func(t *testing.T) {
		t.Run("InvalidForm", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader("!!invalid!!"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			token(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %v", w.Code)
			}
		})

		t.Run("FieldTooLong", func(t *testing.T) {
			form := url.Values{}
			form.Add("grant_type", strings.Repeat("a", 2049))
			req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			token(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %v", w.Code)
			}
		})

		t.Run("PasswordMissingUsername", func(t *testing.T) {
			form := url.Values{}
			form.Add("grant_type", "password")
			req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			token(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %v", w.Code)
			}
		})

		t.Run("PasswordInvalidUser", func(t *testing.T) {
			form := url.Values{}
			form.Add("grant_type", "password")
			form.Add("username", "nosuchuser")
			form.Add("password", "p")
			req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			token(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("status = %v", w.Code)
			}
		})

		t.Run("ClientCredentialsMissingClient", func(t *testing.T) {
			form := url.Values{}
			form.Add("grant_type", "client_credentials")
			form.Add("client_id", uuid.New().String())
			req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			token(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("status = %v", w.Code)
			}
		})
	})

	t.Run("AuthorizeErrors", func(t *testing.T) {
		t.Run("ParamTooLong", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/authorize?p="+strings.Repeat("a", 2049), nil)
			w := httptest.NewRecorder()
			authorize(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %v", w.Code)
			}
		})

		t.Run("ClientNotFound", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/authorize?client_id="+uuid.New().String(), nil)
			w := httptest.NewRecorder()
			authorize(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %v", w.Code)
			}
		})
	})

	t.Run("LoginErrors", func(t *testing.T) {
		t.Run("InvalidMethod", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/login", nil)
			w := httptest.NewRecorder()
			login(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("status = %v", w.Code)
			}
		})

		t.Run("InvalidForm", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("!!invalid!!"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			login(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %v", w.Code)
			}
		})
	})

	t.Run("Middleware", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		t.Run("Logger", func(t *testing.T) {
			handler := logger(next)
			req := httptest.NewRequest(http.MethodGet, "/_health", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("logger() status = %v", w.Code)
			}
		})

		t.Run("MaxBytes", func(t *testing.T) {
			handler := maxBytesMiddleware(next)
			body := strings.Repeat("a", 1024*1024+1)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK { // Middleware itself doesn't return error, but limits reader
				t.Errorf("maxBytesMiddleware() status = %v", w.Code)
			}
		})
	})

	t.Run("ValidateRedirectURI", func(t *testing.T) {
		if validateRedirectURI("", []string{"http://test"}) {
			t.Error("Empty redirect URI should be invalid")
		}

		if !validateRedirectURI("http://test", []string{}) {
			t.Error("Empty allowed list should allow all")
		}

		if validateRedirectURI("http://wrong", []string{"http://test"}) {
			t.Error("Wrong redirect URI should be invalid")
		}
	})

	t.Run("CorsMiddlewarePost", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := corsMiddleware(next)
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %v", w.Code)
		}
	})

	t.Run("TokenMoreGrants", func(t *testing.T) {
		t.Run("AuthCodeRedirectMismatch", func(t *testing.T) {
			claims := jwt.MapClaims{
				"sub":          "u1",
				"redirect_uri": "http://ok",
				"exp":          time.Now().Add(5 * time.Minute).Unix(),
			}
			tk, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
			requireNoError(t, err)

			form := url.Values{}
			form.Add("grant_type", "authorization_code")
			form.Add("code", tk)
			form.Add("redirect_uri", "http://wrong")
			req := httptest.NewRequestWithContext(context.Background(), "POST", "/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			token(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("status = %v", w.Code)
			}
		})
	})

	t.Run("LoadResources", func(t *testing.T) {
		// PKCS8 key
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		requireNoError(t, err)
		keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
		requireNoError(t, err)

		pemBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes}
		requireNoError(t, os.WriteFile("test8.key", pem.EncodeToMemory(pemBlock), 0o600))
		cleanupRemove(t, "test8.key")

		// PKCS1 key
		key1Bytes := x509.MarshalPKCS1PrivateKey(key)
		pemBlock1 := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: key1Bytes}
		requireNoError(t, os.WriteFile("test1.key", pem.EncodeToMemory(pemBlock1), 0o600))
		cleanupRemove(t, "test1.key")

		requireNoError(t, os.WriteFile("Config.yaml", []byte("tenants: []"), 0o600))
		cleanupRemove(t, "Config.yaml")
		requireNoError(t, os.WriteFile("login.html", []byte("<html></html>"), 0o600))
		cleanupRemove(t, "login.html")
		requireNoError(t, os.WriteFile("index.html", []byte("<html></html>"), 0o600))
		cleanupRemove(t, "index.html")

		loadResources("test8.key")
		loadResources("test1.key")
		loadResources("non-existent")
		requireNoError(t, os.WriteFile("invalid.key", []byte("invalid"), 0o600))
		cleanupRemove(t, "invalid.key")
		loadResources("invalid.key")
	})

	t.Run("DiscoveryV2Tenant", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/v2.0/.well-known/openid-configuration", nil)
		w := httptest.NewRecorder()
		discovery(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %v", w.Code)
		}
	})

	t.Run("ResolveRolesExtra", func(t *testing.T) {
		tenant := &Tenant{
			AppRegistrations: []AppRegistration{
				{
					ClientID:      uuid.MustParse("33333333-3333-4333-a333-333333333333"),
					IdentifierURI: "api://app1",
					AppRoles: []Role{
						{
							Value:  "Role1",
							Scopes: []Scope{{Value: "s1"}},
						},
					},
				},
			},
		}
		// Test unresolvedRoles map paths
		roles := resolveRoles(tenant, nil, nil, map[string]bool{"33333333-3333-4333-a333-333333333333": true}, []string{"s1"})
		if len(roles) != 1 || roles[0] != "Role1" {
			t.Errorf("got %v", roles)
		}

		// targetAppIDs mismatch
		_ = resolveRoles(tenant, nil, nil, map[string]bool{"other": true}, []string{"s1"})
	})

	t.Run("ValidateConfigMore", func(_ *testing.T) {
		// Schema validation failure
		yamlData := []byte("tenants: [{name: 'T1'}]") // Missing required tenant_id
		_ = validateConfig(yamlData)
	})

	t.Run("TokenFinalErrors", func(t *testing.T) {
		t.Run("PasswordWrongSecret", func(t *testing.T) {
			setupTestConfig()

			u := uuid.New()
			configData.Tenants[0].Clients = append(configData.Tenants[0].Clients, Client{
				ClientID:     u,
				ClientSecret: "secret",
			})
			form := url.Values{}
			form.Add("grant_type", "password")
			form.Add("username", "testuser")
			form.Add("password", "password")
			form.Add("client_id", u.String())
			form.Add("client_secret", "wrong")
			req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			token(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("status = %v", w.Code)
			}
		})

		t.Run("AuthCodeNoRedirectParam", func(t *testing.T) {
			setupTestConfig()

			claims := jwt.MapClaims{
				"sub":          "u1",
				"redirect_uri": "http://ok",
				"client_id":    configData.Tenants[0].Clients[0].ClientID.String(),
				"exp":          time.Now().Add(5 * time.Minute).Unix(),
			}
			tk, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
			requireNoError(t, err)

			form := url.Values{}
			form.Add("grant_type", "authorization_code")
			form.Add("code", tk)
			req := httptest.NewRequestWithContext(context.Background(), "POST", "/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			token(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("status = %v, body = %s", w.Code, w.Body.String())
			}
		})
	})

	t.Run("Router", func(t *testing.T) {
		mux := setupRouter()
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/_health", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Router _health failed: %v", w.Code)
		}
	})

	t.Run("LoadConfig", func(t *testing.T) {
		// Mock files
		requireNoError(t, os.WriteFile("Config.yaml", []byte("tenants: []"), 0o600))
		cleanupRemove(t, "Config.yaml")

		config := loadConfig()
		if len(config.Tenants) != 0 {
			t.Errorf("Expected 0 tenants, got %d", len(config.Tenants))
		}
	})

	t.Run("LoadConfigNotFound", func(_ *testing.T) {
		// Ensure Config.yaml does not exist
		_ = os.Remove("Config.yaml")

		config := loadConfig()
		if len(config.Tenants) != 0 {
			t.Error("Should return empty config when file not found")
		}
	})

	t.Run("LoggerDebug", func(t *testing.T) {
		requireNoError(t, os.Setenv("LOG_LEVEL", "debug"))
		t.Cleanup(func() {
			requireNoError(t, os.Unsetenv("LOG_LEVEL"))
		})

		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := logger(next)
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/_health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	})

	t.Run("IndexTemplateError", func(t *testing.T) {
		// Store current template
		oldTmpl := indexTemplate

		defer func() { indexTemplate = oldTmpl }()

		// Create a failing template
		indexTemplate = template.Must(template.New("fail").Parse("{{.NoSuchField}}"))
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
		w := httptest.NewRecorder()
		index(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("status = %v", w.Code)
		}
	})

	t.Run("TokenNonceState", func(t *testing.T) {
		setupTestConfig()

		form := url.Values{}
		form.Add("grant_type", "password")
		form.Add("username", "testuser")
		form.Add("password", "password")
		form.Add("client_id", configData.Tenants[0].Clients[0].ClientID.String())
		form.Add("nonce", "my-nonce")

		req := httptest.NewRequestWithContext(context.Background(), "POST", "/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		token(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %v", w.Code)
		}
	})

	t.Run("AuthorizeUsersDisplay", func(t *testing.T) {
		setupTestConfig()

		u1 := configData.Tenants[0].Clients[0].ClientID
		// Add groups and roles for display
		configData.Tenants[0].Users[0].Groups = []string{"group1"}
		configData.Tenants[0].Clients[0].GroupRoleAssignments = []GroupRoleAssignment{
			{
				GroupName:     "group1",
				Roles:         []string{"Admin"},
				ApplicationID: u1,
			},
		}

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/authorize?client_id="+u1.String()+"&redirect_uri=http://localhost/callback", nil)
		w := httptest.NewRecorder()
		authorize(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %v", w.Code)
		}
	})

	t.Run("DiscoveryTenantNotFound", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/nosuchtenant/.well-known/openid-configuration", nil)
		w := httptest.NewRecorder()
		discovery(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %v", w.Code)
		}
	})

	t.Run("CallHomeHost", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/discovery/instance", nil)
		req.Host = "custom.host"
		w := httptest.NewRecorder()
		callHome(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %v", w.Code)
		}
	})

	t.Run("TokenFinalScenarios", func(t *testing.T) {
		t.Run("RefreshClient", func(t *testing.T) {
			setupTestConfig()

			u := uuid.New()
			configData.Tenants[0].Clients = append(configData.Tenants[0].Clients, Client{
				ClientID: u,
			})
			claims := jwt.MapClaims{
				"sub":       u.String(),
				"client_id": u.String(),
				"exp":       time.Now().Add(5 * time.Minute).Unix(),
				"typ":       "Refresh",
			}
			tk, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
			requireNoError(t, err)

			form := url.Values{}
			form.Add("grant_type", "refresh_token")
			form.Add("refresh_token", tk)
			req := httptest.NewRequestWithContext(context.Background(), "POST", "/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			token(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("status = %v", w.Code)
			}
		})
	})

	t.Run("LoadConfigFailures", func(t *testing.T) {
		// Invalid schema validation failure path in loadConfig
		requireNoError(t, os.WriteFile("Config.yaml", []byte("tenants: [{name: 'Invalid'}]"), 0o600))
		cleanupRemove(t, "Config.yaml")

		_ = loadConfig()

		// Invalid YAML failure path in loadConfig
		requireNoError(t, os.WriteFile("Config.yaml", []byte("!!invalid"), 0o600))

		_ = loadConfig()
	})

	t.Run("DiscoveryDeepPaths", func(_ *testing.T) {
		paths := []string{
			"/tenant/oauth2/authorize",
			"/common/oauth2/authorize",
			"/v2.0/oauth2/authorize",
			"/tenant/v2.0/oauth2/authorize",
		}
		for _, p := range paths {
			req := httptest.NewRequestWithContext(context.Background(), "GET", p+"?client_id="+uuid.New().String()+"&redirect_uri=http://ok", nil)
			w := httptest.NewRecorder()
			authorize(w, req)
		}
	})

	t.Run("TokenDeepPaths", func(_ *testing.T) {
		paths := []string{
			"/tenant/oauth2/token",
			"/common/oauth2/token",
			"/tenant/oauth2/v2.0/token",
			"/common/oauth2/v2.0/token",
		}
		for _, p := range paths {
			req := httptest.NewRequestWithContext(context.Background(), "POST", p, strings.NewReader("grant_type=password&username=u&password=p&client_id="+uuid.New().String()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			token(w, req)
		}
	})

	t.Run("DiscoveryDeepPaths2", func(_ *testing.T) {
		paths := []string{
			"/.well-known/openid-configuration",
			"/v2.0/.well-known/openid-configuration",
			"/tenant/.well-known/openid-configuration",
			"/tenant/v2.0/.well-known/openid-configuration",
			"/common/.well-known/openid-configuration",
			"/common/v2.0/.well-known/openid-configuration",
		}
		for _, p := range paths {
			req := httptest.NewRequestWithContext(context.Background(), "GET", p, nil)
			w := httptest.NewRecorder()
			discovery(w, req)
		}
	})
}

func TestResolveAudience(t *testing.T) {
	u1 := uuid.MustParse("11111111-1111-4111-a111-111111111111")
	u2 := uuid.MustParse("22222222-2222-4222-a222-222222222222")
	tenant := &Tenant{
		AppRegistrations: []AppRegistration{
			{
				Name:          "App1",
				ClientID:      u1,
				IdentifierURI: "api://app1",
			},
			{
				Name:          "App2",
				ClientID:      u2,
				IdentifierURI: "api://app2",
				AppRoles: []Role{
					{
						Value: "Role1",
						Scopes: []Scope{
							{Value: "scope1"},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name         string
		scope        string
		wantAudience string
		wantAppIDs   map[string]bool
	}{
		{
			name:         "Empty scope",
			scope:        "",
			wantAudience: "api://default",
			wantAppIDs:   map[string]bool{},
		},
		{
			name:         "Exact ClientID match",
			scope:        u1.String(),
			wantAudience: "api://app1",
			wantAppIDs:   map[string]bool{u1.String(): true},
		},
		{
			name:         "Exact IdentifierUri match",
			scope:        "api://app1",
			wantAudience: "api://app1",
			wantAppIDs:   map[string]bool{u1.String(): true},
		},
		{
			name:         ".default suffix",
			scope:        "api://app1/.default",
			wantAudience: "api://app1",
			wantAppIDs:   map[string]bool{u1.String(): true},
		},
		{
			name:         "Scope match",
			scope:        "scope1",
			wantAudience: "api://app2",
			wantAppIDs:   map[string]bool{u2.String(): true},
		},
		{
			name:         "Scope with URI prefix",
			scope:        "api://app2/scope1",
			wantAudience: "api://app2",
			wantAppIDs:   map[string]bool{u2.String(): true},
		},
		{
			name:         "Multiple scopes",
			scope:        "openid profile api://app1/.default",
			wantAudience: "api://app1",
			wantAppIDs:   map[string]bool{u1.String(): true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAud, gotAppIDs := resolveAudience(tenant, tt.scope)
			if gotAud != tt.wantAudience {
				t.Errorf("resolveAudience() gotAud = %v, want %v", gotAud, tt.wantAudience)
			}

			if !reflect.DeepEqual(gotAppIDs, tt.wantAppIDs) {
				t.Errorf("resolveAudience() gotAppIDs = %v, want %v", gotAppIDs, tt.wantAppIDs)
			}
		})
	}
}

func TestResolveRoles(t *testing.T) {
	api1 := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	tenant := &Tenant{
		AppRegistrations: []AppRegistration{
			{
				ClientID:      api1,
				IdentifierURI: "api://api-1",
				AppRoles: []Role{
					{
						Value: "Admin",
						Scopes: []Scope{
							{Value: "admin-scope"},
						},
					},
					{
						Value: "Reader",
						Scopes: []Scope{
							{Value: "read-scope"},
						},
					},
				},
			},
		},
	}

	client := &Client{
		GroupRoleAssignments: []GroupRoleAssignment{
			{
				GroupName:     "group-1",
				Roles:         []string{"Reader"},
				ApplicationID: api1,
			},
			{
				GroupName:     "group-2",
				Roles:         []string{"Admin"},
				ApplicationID: api1,
			},
		},
	}

	user := &User{
		Groups: []string{"group-1"},
	}

	tests := []struct {
		name            string
		client          *Client
		user            *User
		targetAppIDs    map[string]bool
		requestedScopes []string
		wantRoles       []string
	}{
		{
			name:            "User role from group",
			client:          client,
			user:            user,
			targetAppIDs:    map[string]bool{api1.String(): true},
			requestedScopes: []string{"api://api-1/.default"},
			wantRoles:       []string{"Reader"},
		},
		{
			name:            "Client credentials (no user)",
			client:          client,
			user:            nil,
			targetAppIDs:    map[string]bool{api1.String(): true},
			requestedScopes: []string{"api://api-1/.default"},
			wantRoles:       []string{"Admin", "Reader"},
		},
		{
			name:            "Scope to Role mapping",
			client:          nil,
			user:            nil,
			targetAppIDs:    map[string]bool{api1.String(): true},
			requestedScopes: []string{"admin-scope"},
			wantRoles:       []string{"Admin"},
		},
		{
			name:            "Filter by targetAppIDs",
			client:          client,
			user:            nil,
			targetAppIDs:    map[string]bool{"other-api": true},
			requestedScopes: []string{"api://api-1/.default"},
			wantRoles:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveRoles(tenant, tt.client, tt.user, tt.targetAppIDs, tt.requestedScopes)
			sort.Strings(got)
			sort.Strings(tt.wantRoles)

			if !reflect.DeepEqual(got, tt.wantRoles) {
				t.Errorf("resolveRoles() = %v, want %v", got, tt.wantRoles)
			}
		})
	}
}

func TestNewHandlers(t *testing.T) {
	var err error

	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	requireNoError(t, err)
	setupTestConfig()

	tid := configData.Tenants[0].TenantID
	appID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa")
	clientID := configData.Tenants[0].Clients[0].ClientID

	configData.Tenants[0].AppRegistrations = []AppRegistration{
		{
			Name:          "TestApp",
			ClientID:      appID,
			IdentifierURI: "api://testapp",
		},
	}

	t.Run("TestToken_OK", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/test-tokens/"+tid.String()+"/"+appID.String()+"/"+clientID.String(), nil)
		w := httptest.NewRecorder()
		testToken(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("testToken status = %v, body = %s", w.Code, w.Body.String())
		}

		if !strings.Contains(w.Body.String(), "access_token") {
			t.Errorf("testToken body missing access_token: %s", w.Body.String())
		}
	})

	t.Run("TestToken_BadTenant", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/test-tokens/not-a-uuid/"+appID.String()+"/"+clientID.String(), nil)
		w := httptest.NewRecorder()
		testToken(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %v", w.Code)
		}
	})

	t.Run("TestToken_TenantNotFound", func(t *testing.T) {
		missing := uuid.New()
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/test-tokens/"+missing.String()+"/"+appID.String()+"/"+clientID.String(), nil)
		w := httptest.NewRecorder()
		testToken(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %v", w.Code)
		}
	})

	t.Run("TestToken_MissingSegments", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/test-tokens/"+tid.String(), nil)
		w := httptest.NewRecorder()
		testToken(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %v", w.Code)
		}
	})

	t.Run("ConfigRaw", func(t *testing.T) {
		requireNoError(t, os.WriteFile("Config.yaml", []byte("tenants: []"), 0o600))
		cleanupRemove(t, "Config.yaml")

		req := httptest.NewRequestWithContext(context.Background(), "GET", "/config/raw", nil)
		w := httptest.NewRecorder()
		configRaw(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("configRaw status = %v", w.Code)
		}
	})

	t.Run("ConfigCsharpApp_OK", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/config/"+tid.String()+"/app/"+appID.String()+"/csharp", nil)
		w := httptest.NewRecorder()
		configCsharpApp(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("configCsharpApp status = %v, body = %s", w.Code, w.Body.String())
		}

		if !strings.Contains(w.Body.String(), "AzureAd__Instance") {
			t.Errorf("missing AzureAd__Instance in body: %s", w.Body.String())
		}
	})

	t.Run("ConfigCsharpApp_BadTenant", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/config/bad/app/"+appID.String()+"/csharp", nil)
		w := httptest.NewRecorder()
		configCsharpApp(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %v", w.Code)
		}
	})

	t.Run("ConfigCsharpApp_TenantNotFound", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/config/"+uuid.New().String()+"/app/"+appID.String()+"/csharp", nil)
		w := httptest.NewRecorder()
		configCsharpApp(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %v", w.Code)
		}
	})

	t.Run("ConfigCsharpApp_AppNotFound", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/config/"+tid.String()+"/app/"+uuid.New().String()+"/csharp", nil)
		w := httptest.NewRecorder()
		configCsharpApp(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %v", w.Code)
		}
	})

	t.Run("ConfigJsApp_OK", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/config/"+tid.String()+"/app/"+appID.String()+"/js", nil)
		w := httptest.NewRecorder()
		configJsApp(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("configJsApp status = %v, body = %s", w.Code, w.Body.String())
		}

		if !strings.Contains(w.Body.String(), "msalConfig") {
			t.Errorf("missing msalConfig in body: %s", w.Body.String())
		}
	})

	t.Run("ConfigCsharpClient_OK", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/config/"+tid.String()+"/client/"+clientID.String()+"/csharp", nil)
		w := httptest.NewRecorder()
		configCsharpClient(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("configCsharpClient status = %v, body = %s", w.Code, w.Body.String())
		}

		if !strings.Contains(w.Body.String(), "AzureAd__Instance") {
			t.Errorf("missing AzureAd__Instance in body: %s", w.Body.String())
		}
	})

	t.Run("ConfigCsharpClient_NotFound", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/config/"+tid.String()+"/client/"+uuid.New().String()+"/csharp", nil)
		w := httptest.NewRecorder()
		configCsharpClient(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %v", w.Code)
		}
	})

	t.Run("IndexDispatchesTestToken", func(t *testing.T) {
		req := httptest.NewRequestWithContext(context.Background(), "GET",
			"/test-tokens/"+tid.String()+"/"+appID.String()+"/"+clientID.String(), nil)
		w := httptest.NewRecorder()
		index(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("index dispatch to testToken status = %v, body = %s", w.Code, w.Body.String())
		}
	})
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "Valid config",
			yaml: `
tenants:
  - tenant_id: "b5a920d6-7d3c-44fe-baad-4ffed6b8774d"
    name: "Default Tenant"
    app_registrations:
      - name: "App1"
        client_id: "33333333-3333-4333-a333-333333333333"
        identifier_uri: "api://app1"
        scopes: []
        app_roles: []
    groups: []
    users: []
    clients: []
`,
			wantErr: false,
		},
		{
			name: "Invalid UUID",
			yaml: `
tenants:
  - tenant_id: "invalid-uuid"
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
