package main

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"identity/domain"
)

func mustIdentifierURI(t *testing.T, raw string) domain.IdentifierURI {
	t.Helper()

	v, err := domain.NewIdentifierURI(raw)
	if err != nil {
		t.Fatalf("NewIdentifierURI(%q): %v", raw, err)
	}

	return v
}

func minimalConfig(t *testing.T) domain.Config {
	t.Helper()

	redirectURL, err := domain.NewRedirectURL("https://example.com/callback")
	if err != nil {
		t.Fatalf("NewRedirectURL: %v", err)
	}

	registration := domain.NewAppRegistration(
		domain.MustAppName("App"),
		domain.MustClientID("22222222-2222-4222-8222-222222222222"),
		mustIdentifierURI(t, "api://app"),
		[]domain.RedirectURL{redirectURL},
		nil,
		nil,
	)

	user := domain.NewUser(
		domain.MustUserID("33333333-3333-4333-8333-333333333333"),
		domain.MustUsername("user"),
		domain.MustPassword("pass"),
		domain.MustDisplayName("User"),
		domain.MustEmail("user@example.com"),
		nil,
	)

	tenant, err := domain.NewTenant(
		domain.MustTenantID("11111111-1111-4111-8111-111111111111"),
		domain.MustTenantName("Tenant"),
		[]domain.AppRegistration{registration},
		nil,
		[]domain.User{user},
		nil,
	)
	if err != nil {
		t.Fatalf("NewTenant: %v", err)
	}

	config, err := domain.NewConfig([]domain.Tenant{tenant})
	if err != nil {
		t.Fatalf("NewConfig: %v", err)
	}

	return config
}

func TestResolveConfigPath(t *testing.T) {
	t.Parallel()

	path := resolveConfigPath(func(string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	if path != defaultConfigPath {
		t.Fatalf("expected %q, got %q", defaultConfigPath, path)
	}

	path = resolveConfigPath(func(string) (os.FileInfo, error) { return nil, nil })
	if path != configPath {
		t.Fatalf("expected %q, got %q", configPath, path)
	}

	path = resolveConfigPath(func(string) (os.FileInfo, error) { return nil, errors.New("boom") })
	if path != configPath {
		t.Fatalf("expected %q, got %q", configPath, path)
	}
}

func TestResolveAddr(t *testing.T) {
	t.Parallel()

	if got := resolveAddr(func(string) string { return "" }); got != defaultAddr {
		t.Fatalf("expected %q, got %q", defaultAddr, got)
	}

	if got := resolveAddr(func(string) string { return "1234" }); got != ":1234" {
		t.Fatalf("expected %q, got %q", ":1234", got)
	}
}

func TestLoadTemplates(t *testing.T) {
	t.Parallel()

	login, index, err := loadTemplates()
	if err != nil {
		t.Fatalf("loadTemplates: %v", err)
	}

	if login == nil || index == nil {
		t.Fatalf("expected templates")
	}
}

func TestRun_Success(t *testing.T) {
	t.Parallel()

	err := run(runDeps{
		keyBits: 1024,
		stat:    func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		getenv:  func(string) string { return "" },
		loadConfig: func(path string) (domain.Config, error) {
			if path != defaultConfigPath {
				t.Fatalf("expected config path %q, got %q", defaultConfigPath, path)
			}

			return minimalConfig(t), nil
		},
		listen: func(server *http.Server) error {
			if server.Addr != defaultAddr {
				t.Fatalf("expected addr %q, got %q", defaultAddr, server.Addr)
			}

			return nil
		},
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestRun_LoadConfigError(t *testing.T) {
	t.Parallel()

	expected := errors.New("boom")

	err := run(runDeps{
		keyBits: 1024,
		stat:    func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		getenv:  func(string) string { return "" },
		loadConfig: func(string) (domain.Config, error) {
			return domain.Config{}, expected
		},
		listen: func(*http.Server) error { return nil },
	})

	if !errors.Is(err, expected) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestRun_ListenError(t *testing.T) {
	t.Parallel()

	expected := errors.New("listen boom")

	err := run(runDeps{
		keyBits: 1024,
		stat:    func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		getenv:  func(string) string { return "" },
		loadConfig: func(string) (domain.Config, error) {
			return minimalConfig(t), nil
		},
		listen: func(*http.Server) error { return expected },
	})

	if !errors.Is(err, expected) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestRun_KeyGenerationError(t *testing.T) {
	t.Parallel()

	err := run(runDeps{
		keyBits: 0,
		stat:    func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		getenv:  func(string) string { return "" },
		loadConfig: func(string) (domain.Config, error) {
			return minimalConfig(t), nil
		},
		listen: func(*http.Server) error { return nil },
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestDefaultLoadConfig(t *testing.T) {
	t.Parallel()

	config, err := defaultLoadConfig(defaultConfigPath)
	if err != nil {
		t.Fatalf("defaultLoadConfig: %v", err)
	}

	if len(config.Tenants()) == 0 {
		t.Fatalf("expected tenants")
	}
}
