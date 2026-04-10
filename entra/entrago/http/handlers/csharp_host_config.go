package handlers

import (
	"fmt"

	"identity/domain"
)

// CsharpHostConfig holds the domain values needed to render the
// C# MSAL host environment variable block.
type CsharpHostConfig struct {
	tenantID      domain.TenantID
	identifierURI domain.IdentifierURI
	clientID      domain.ClientID
}

// NewCsharpHostConfig creates a CsharpHostConfig from the domain values
// required for the C# MSAL host environment variable block.
func NewCsharpHostConfig(
	tenantID domain.TenantID,
	identifierURI domain.IdentifierURI,
	clientID domain.ClientID,
) CsharpHostConfig {
	return CsharpHostConfig{
		tenantID:      tenantID,
		identifierURI: identifierURI,
		clientID:      clientID,
	}
}

// Fragment builds the C# MSAL host environment variable block.
func (c CsharpHostConfig) Fragment() string {
	tenantID := c.tenantID.Value()
	audience := c.identifierURI.Value()
	clientID := c.clientID.Value()

	const tmpl = "AzureAd__Instance=https://login.microsoftonline.com/\n" +
		"AzureAd__TenantId=%s\n" +
		"AzureAd__Audience=%s\n" +
		"AzureAd__ClientId=%s\n" +
		"AzureAd__ClientSecret=set-client-secret"

	return fmt.Sprintf(tmpl, tenantID, audience, clientID)
}
