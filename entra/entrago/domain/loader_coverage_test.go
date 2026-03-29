package domain_test

import (
	"os"
	"path/filepath"
	"testing"

	"identity/domain"
)

const (
	filePerm = 0o600
)

func TestLoadConfig_ValidationFailed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "invalid.yaml")
	schemaPath := filepath.Join(dir, "schema.json")

	_ = os.WriteFile(configPath, []byte("invalid: yaml: :"), filePerm)
	_ = os.WriteFile(schemaPath, []byte(`{"type": "object"}`), filePerm)

	_, err := domain.LoadConfig(configPath, schemaPath)
	if err == nil {
		t.Fatal("expected error for invalid yaml")
	}
}

func TestNewTenant_Errors(t *testing.T) {
	t.Parallel()

	tenantID := domain.MustTenantID("11111111-1111-1111-1111-111111111111")
	tenantName := domain.MustTenantName("T")

	// No users
	appReg := domain.NewAppRegistration(
		domain.MustAppName("A"),
		domain.MustClientID("22222222-2222-2222-2222-222222222222"),
		domain.MustIdentifierURI("api://a"),
		nil, nil, nil,
	)

	_, err := domain.NewTenant(tenantID, tenantName, []domain.AppRegistration{appReg}, nil, nil, nil)
	if err == nil {
		t.Error("expected error for no users")
	}

	// No app registrations
	uID := domain.MustUserID("33333333-3333-3333-3333-333333333333")
	uName := domain.MustUsername("u")
	uPass := domain.MustPassword("p")
	uDisp := domain.MustDisplayName("D")
	uEmail := domain.MustEmail("e")
	user := domain.NewUser(uID, uName, uPass, uDisp, uEmail, nil)

	_, err = domain.NewTenant(tenantID, tenantName, nil, nil, []domain.User{user}, nil)
	if err == nil {
		t.Error("expected error for no app registrations")
	}
}
