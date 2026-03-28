# Entra mock
## Entra endpoints
Describe the endpoints that are available and how they will respond 

## Test endpoints
-   Endpoint to retrieve a test token with defaulted if not configured claims plus any query parameters as claim GET /test-tokens/{tenantId}/{appId}/{clientId}?scope=read,scope=write
-   Endpoint for just signing provided token POST /test-token/sign?token=token

## Current configuration

### App registration as a header, 
#### links to get configuration for it in different formats:
-   C# container apps environment variables (msal host configuration)
-   JavaScript msal configuration

#### Claims and roles for it
-   Then user groups that has roles for it 
-   Users that has role for it with link to start login to any registered redirect url
-  clients that have access to it with a link to retrieve 
-   token for app by default scope and C# environment variables for Client configuration

