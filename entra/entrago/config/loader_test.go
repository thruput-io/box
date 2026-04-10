package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"identity/config"
	"identity/domain"
)

const (
	filePermission = 0o600
	schemaFileName = "schema.json"
	schemaJSON     = `{"$schema":"http://json-schema.org/draft-07/schema#","type":"object"}`
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)

	err := os.WriteFile(path, []byte(content), filePermission)
	if err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}

	return path
}

func TestValidateYAML_InvalidForSchema(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaPath := writeTempFile(t, dir, schemaFileName, schemaJSON)

	err := config.ExportValidateYAML([]byte("foo\n"), schemaPath)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !errors.Is(err, domain.ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestLoadConfig_Success(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaPath := writeTempFile(t, dir, schemaFileName, schemaJSON)
	yamlPath := writeTempFile(t, dir, "config.yaml", `tenants:
  - tenantId: "11111111-1111-4111-8111-111111111111"
    name: "Tenant"
    appRegistrations:
      - name: "App"
        clientId: "22222222-2222-4222-8222-222222222222"
        identifierUri: "api://app"
        redirectUrls:
          - "https://example.com/callback"
        scopes: []
        appRoles: []
    groups: []
    users:
      - id: "33333333-3333-4333-8333-333333333333"
        username: "user"
        password: "pass"
        displayName: "User"
        email: "user@example.com"
        groups: []
    clients: []
`)

	cfg, err := config.LoadConfig(yamlPath, schemaPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	tenants := cfg.Tenants()

	if len(tenants) != expectedTenants {
		t.Fatalf("expected 1 tenant, got %d", len(tenants))
	}

	if tenants[0].Name() != domain.MustTenantName("Tenant") {
		t.Fatalf("expected tenant name %v, got %v", domain.MustTenantName("Tenant"), tenants[0].Name())
	}
}

func TestLoadConfig_ReadError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaPath := writeTempFile(t, dir, schemaFileName, schemaJSON)

	_, err := config.LoadConfig(filepath.Join(dir, "does-not-exist.yaml"), schemaPath)
	if err == nil {
		t.Fatal("expected read error")
	}
}

func TestBuildRedirectURLs_InvalidURL(t *testing.T) {
	t.Parallel()

	urls, err := config.ExportBuildRedirectURLs([]string{""})
	if err == nil {
		t.Fatal("expected error")
	}

	if len(urls) != zeroLength {
		t.Fatal("expected no urls on error")
	}
}

func TestBuildUserGroups_RejectsEmptyGroupName(t *testing.T) {
	t.Parallel()

	groups, err := config.ExportBuildUserGroups("user", []string{""})
	if err == nil {
		t.Fatal("expected error")
	}

	if len(groups) != zeroLength {
		t.Fatal("expected no groups on error")
	}
}
