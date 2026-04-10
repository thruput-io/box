# Identity.Index.Inside.Tests.ps1

Describe "Identity Index Page Verification" {
    BeforeAll {
        $baseUrl = "https://login.microsoftonline.com"
    }

    Context "Identity: index page" {
        It "GET / returns index html with documented endpoints" {
            $response = Invoke-WebRequest -Method Get -Uri "$baseUrl/" -TimeoutSec 1

            $response.StatusCode | Should -Be 200
            $response.Headers["Content-Type"] | Should -Match "text/html"
            $response.Content | Should -Match "<title>Entra mock</title>"
            $response.Content | Should -Match "/\.well-known/openid-configuration"
            $response.Content | Should -Match "<code>/token</code>"
            $response.Content | Should -Match "/authorize"
            $response.Content | Should -Match "/mock-utils/\{tenantId\}/\{appId\}/\{clientId\}\?scope=read,scope=write"
            $response.Content | Should -Match "/mock-utils/sign\?token=token"
            $response.Content | Should -Match "C# container apps environment variables \(msal host configuration\)"
            $response.Content | Should -Match "AzureAd__Instance=https://login\.microsoftonline\.com/"
            $response.Content | Should -Match "AzureAd__TenantId=10000000-0000-4000-a000-000000000000"
            $response.Content | Should -Match "AzureAd__Audience=api://aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"
            $response.Content | Should -Match "AzureAd__ClientId=aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"
            $response.Content | Should -Match "AzureAd__ClientSecret=set-client-secret"
            $response.Content | Should -Match 'data-copy-target="csharp-host-config-'
        }

        It "GET /index.html routes to the index page" {
            $response = Invoke-WebRequest -Method Get -Uri "$baseUrl/index.html" -TimeoutSec 1

            $response.StatusCode | Should -Be 200
            $response.Headers["Content-Type"] | Should -Match "text/html"
            $response.Content | Should -Match "<title>Entra mock</title>"
            $response.Content | Should -Match "AzureAd__TenantId=10000000-0000-4000-a000-000000000000"
            $response.Content | Should -Match 'data-copy-target="csharp-host-config-'
        }
    }
}