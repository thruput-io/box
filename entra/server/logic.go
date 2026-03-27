package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

func loadConfig() Config {
	data, err := os.ReadFile("Config.yaml")
	if err != nil {
		log.Printf("Warning: failed to read Config.yaml, using defaults: %v", err)
		return Config{}
	}

	// Validate against JSON schema to catch configuration errors early
	if err := validateConfig(data); err != nil {
		log.Printf("CRITICAL: Config.yaml validation failed:\n%v", err)
		return Config{}
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Printf("CRITICAL: failed to parse Config.yaml: %v", err)
		return Config{}
	}
	return config
}

func validateConfig(yamlData []byte) error {
	// 1. Unmarshal YAML into a generic structure
	var raw interface{}
	if err := yaml.Unmarshal(yamlData, &raw); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. Load the schema file
	schemaData, err := os.ReadFile("Config.schema.json")
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaData)
	documentLoader := gojsonschema.NewGoLoader(raw)

	// 3. Validate
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		var errMsgs []string
		for _, desc := range result.Errors() {
			errMsgs = append(errMsgs, desc.String())
		}
		return fmt.Errorf("%s", strings.Join(errMsgs, "\n"))
	}

	return nil
}

func findTenant(tenantID string, config *Config) *Tenant {
	if len(config.Tenants) == 0 {
		return &Tenant{TenantID: uuid.MustParse("b5a920d6-7d3c-44fe-baad-4ffed6b8774d")}
	}
	if tenantID == "" || tenantID == "common" {
		return &config.Tenants[0]
	}
	for i := range config.Tenants {
		if config.Tenants[i].TenantID.String() == tenantID {
			return &config.Tenants[i]
		}
	}
	return &config.Tenants[0]
}

func resolveAudience(tenant *Tenant, scope string) (string, map[string]bool) {
	targetAudience := "api://default"
	targetAppIDs := make(map[string]bool)

	if scope == "" {
		return targetAudience, targetAppIDs
	}

	requestedScopes := strings.Split(scope, " ")
	for _, s := range requestedScopes {
		if s == "openid" || s == "profile" || s == "offline_access" || s == "email" {
			continue
		}
		for _, reg := range tenant.AppRegistrations {
			// Exact match or .default suffix
			if reg.ClientID.String() == s || reg.IdentifierURI == s || strings.HasPrefix(s, reg.IdentifierURI+"/") || s == reg.IdentifierURI+"/.default" {
				targetAppIDs[reg.ClientID.String()] = true
				targetAudience = reg.IdentifierURI
			}
			// Check roles and scopes
			for _, role := range reg.AppRoles {
				for _, rs := range role.Scopes {
					if rs.Value == s || strings.HasSuffix(s, "/"+rs.Value) {
						targetAppIDs[reg.ClientID.String()] = true
						targetAudience = reg.IdentifierURI
					}
				}
			}
		}
	}
	return targetAudience, targetAppIDs
}

func resolveRoles(tenant *Tenant, client *Client, user *User, targetAppIDs map[string]bool, requestedScopes []string) []string {
	resolvedRoles := make(map[string]bool)
	userGroups := make(map[string]bool)
	if user != nil {
		for _, g := range user.Groups {
			userGroups[g] = true
		}
	}

	// 1. Resolve from Client's GroupRoleAssignments
	if client != nil {
		for _, gra := range client.GroupRoleAssignments {
			// For user-based flow, check group membership
			if user != nil && !userGroups[gra.GroupName] {
				continue
			}
			// If group name matches or it's a client credentials flow (user is nil)
			if user == nil || userGroups[gra.GroupName] {
				// Filter by target application
				if gra.ApplicationID != uuid.Nil && !targetAppIDs[gra.ApplicationID.String()] {
					continue
				}
				for _, roleVal := range gra.Roles {
					resolvedRoles[roleVal] = true
				}
			}
		}
	}

	// 2. Map requested Scopes to Roles (if application is authorized)
	for _, reg := range tenant.AppRegistrations {
		if !targetAppIDs[reg.ClientID.String()] {
			continue
		}
		for _, role := range reg.AppRoles {
			for _, s := range requestedScopes {
				if s == "openid" || s == "profile" || s == "offline_access" || s == "email" {
					continue
				}
				for _, rs := range role.Scopes {
					if rs.Value == s || strings.HasSuffix(s, "/"+rs.Value) {
						resolvedRoles[role.Value] = true
					}
				}
			}
		}
	}

	roles := []string{}
	for r := range resolvedRoles {
		roles = append(roles, r)
	}
	return roles
}

func base64UrlEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}
