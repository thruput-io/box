package main

import (
	"github.com/google/uuid"
)

type OAuthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	TraceID          string `json:"trace_id,omitempty"`
	CorrelationID    string `json:"correlation_id,omitempty"`
	Timestamp        string `json:"timestamp,omitempty"`
}

type Config struct {
	Tenants []Tenant `yaml:"tenants"`
}

type Tenant struct {
	TenantID         uuid.UUID         `yaml:"tenant_id"`
	Name             string            `yaml:"name"`
	AppRegistrations []AppRegistration `yaml:"app_registrations"`
	Groups           []Group           `yaml:"groups"`
	Users            []User            `yaml:"users"`
	Clients          []Client          `yaml:"clients"`
}

type AppRegistration struct {
	Name          string            `yaml:"name"`
	ClientID      uuid.UUID         `yaml:"client_id"`
	IdentifierUri string            `yaml:"identifier_uri"`
	RedirectUrls  []string          `yaml:"redirect_urls"`
	Scopes        []Scope           `yaml:"scopes"`
	AppRoles      []Role            `yaml:"app_roles"`
}

type Scope struct {
	ID          uuid.UUID   `yaml:"id"`
	Value       string `yaml:"value"`
	Description string `yaml:"description"`
}

type Role struct {
	ID          uuid.UUID    `yaml:"id"`
	Value       string  `yaml:"value"`
	Description string  `yaml:"description"`
	Scopes      []Scope `yaml:"scopes"`
}

type Group struct {
	ID   uuid.UUID   `yaml:"id"`
	Name string `yaml:"name"`
}

type User struct {
	ID          uuid.UUID     `yaml:"id"`
	Username    string   `yaml:"username"`
	Password    string   `yaml:"password"`
	DisplayName string   `yaml:"display_name"`
	Email       string   `yaml:"email"`
	Groups      []string `yaml:"groups"`
}

type Client struct {
	Name                 string                `yaml:"name"`
	ClientID             uuid.UUID             `yaml:"client_id"`
	ClientSecret         string                `yaml:"client_secret"`
	RedirectUrls         []string              `yaml:"redirect_urls"`
	GroupRoleAssignments []GroupRoleAssignment `yaml:"group_role_assignments"`
}

type GroupRoleAssignment struct {
	GroupName     string   `yaml:"group_name"`
	Roles         []string `yaml:"roles"`
	ApplicationID uuid.UUID     `yaml:"application_id"`
}

type UserDisplayInfo struct {
	Username    string
	Password    string
	DisplayName string
	Roles       []string
}

type AuthRequest struct {
	ClientID     uuid.UUID
	RedirectURI  string
	State        string
	Scope        string
	ResponseType string
	Tenant       string
	Nonce        string
	ResponseMode string
	Users        []UserDisplayInfo
}

type DiscoveryResponse struct {
	TokenEndpoint                          string   `json:"token_endpoint"`
	TokenEndpointAuthMethodsSupported      []string `json:"token_endpoint_auth_methods_supported"`
	JwksUri                                string   `json:"jwks_uri"`
	ResponseModesSupported                 []string `json:"response_modes_supported"`
	SubjectTypesSupported                  []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported       []string `json:"id_token_signing_alg_values_supported"`
	ResponseTypesSupported                 []string `json:"response_types_supported"`
	ScopesSupported                        []string `json:"scopes_supported"`
	Issuer                                 string   `json:"issuer"`
	RequestUriParameterSupported           bool     `json:"request_uri_parameter_supported"`
	UserinfoEndpoint                       string   `json:"userinfo_endpoint"`
	AuthorizationEndpoint                  string   `json:"authorization_endpoint"`
	DeviceAuthorizationEndpoint            string   `json:"device_authorization_endpoint"`
	HttpLogoutSupported                    bool     `json:"http_logout_supported"`
	FrontchannelLogoutSupported            bool     `json:"frontchannel_logout_supported"`
	EndSessionEndpoint                     string   `json:"end_session_endpoint"`
	ClaimsSupported                        []string `json:"claims_supported"`
	KerberosEndpoint                       string   `json:"kerberos_endpoint"`
	MtlsEndpointAliases                   struct {
		TokenEndpoint string `json:"token_endpoint"`
	} `json:"mtls_endpoint_aliases"`
	TlsClientCertificateBoundAccessTokens bool   `json:"tls_client_certificate_bound_access_tokens"`
	TenantRegionScope                     string `json:"tenant_region_scope"`
	CloudInstanceName                     string `json:"cloud_instance_name"`
	CloudGraphHostName                    string `json:"cloud_graph_host_name"`
	MsgraphHost                           string `json:"msgraph_host"`
	RbacUrl                               string `json:"rbac_url"`
}

type CallHomeResponse struct {
	TenantDiscoveryEndpoint string `json:"tenant_discovery_endpoint"`
	ApiVersion             string `json:"api-version"`
	Metadata               []struct {
		PreferredNetwork string   `json:"preferred_network"`
		PreferredCache   string   `json:"preferred_cache"`
		Aliases          []string `json:"aliases"`
	} `json:"metadata"`
}
