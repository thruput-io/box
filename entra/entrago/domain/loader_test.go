package domain

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}

	return path
}

func TestValidateYAML_InvalidForSchema(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaPath := writeTempFile(t, dir, "schema.json", `{"$schema":"http://json-schema.org/draft-07/schema#","type":"object"}`)

	err := validateYAML([]byte("foo\n"), schemaPath)
	if err == nil {
		t.Fatalf("expected validation error")
	} else if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestLoadConfig_Success(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaPath := writeTempFile(t, dir, "schema.json", `{"$schema":"http://json-schema.org/draft-07/schema#","type":"object"}`)

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

	config, err := LoadConfig(yamlPath, schemaPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	tenants := config.Tenants()
	if len(tenants) != 1 {
		t.Fatalf("expected 1 tenant, got %d", len(tenants))
	}

	if tenants[0].Name().String() != "Tenant" {
		t.Fatalf("expected tenant name %q, got %q", "Tenant", tenants[0].Name().String())
	}
}

func TestLoadConfig_ReadError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaPath := writeTempFile(t, dir, "schema.json", `{"$schema":"http://json-schema.org/draft-07/schema#","type":"object"}`)

	_, err := LoadConfig(filepath.Join(dir, "does-not-exist.yaml"), schemaPath)
	if err == nil {
		t.Fatalf("expected read error")
	}
}

func TestBuildRedirectURLs_InvalidURL(t *testing.T) {
	t.Parallel()

	_, err := buildRedirectURLs([]string{""})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestBuildUserGroups_RejectsEmptyGroupName(t *testing.T) {
	t.Parallel()

	_, err := buildUserGroups("user", []string{""})
	if err == nil {
		t.Fatalf("expected error")
	}
}
