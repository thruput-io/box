# Identity.Inside.Tests.ps1

Describe "Identity Verification" {
    BeforeAll {
        $baseUrl = "https://login.microsoftonline.com"
        $tenantId = "10000000-0000-4000-a000-000000000000"

        $backendClient = "dddddddd-dddd-4ddd-addd-dddddddddddd"
        $backendSecret = "mm-backend-secret"
        $frontendClient = "e697b97c-9b4b-487f-9f7a-248386f78864"

        $mmbApi = "api://aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"
        $customerApi = "api://bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb"
        $healerApi = "api://cccccccc-cccc-4ccc-accc-cccccccccccc"
        $leadApi = "api://eeeeeeee-eeee-4eee-aeee-eeeeeeeeeeee"
        $fuzzApi = "api://ffffffff-ffff-4fff-afff-ffffffffffff"
        $graphApi = "https://graph.microsoft.com"
        $redirectUri = "https://msal-client.web.internal/"

        # Helper to decode JWT payload
        function Get-JwtPayload($token) {
            $tokenParts = $token.Split('.')
            $base64Payload = $tokenParts[1].Replace('-', '+').Replace('_', '/')
            $base64Payload = $base64Payload.PadRight($base64Payload.Length + (4 - $base64Payload.Length % 4) % 4, '=')
            $payloadJson = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($base64Payload))
            return $payloadJson | ConvertFrom-Json
        }
    }

    Context "Identity: Microsoft Entra Mock (Verifications)" {
        It "Backend Client (Client Credentials) - CustomerService .default -> read role" {
            $body = @{
                grant_type = "client_credentials"
                client_id = $backendClient
                client_secret = $backendSecret
                scope = "$customerApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $response.access_token
            $payload.roles | Should -HaveCount 1
            $payload.roles | Should -Contain "read"
            $payload.aud | Should -Be $customerApi
        }

        It "Backend Client (Client Credentials) - HealerService .default -> read role" {
            $body = @{
                grant_type = "client_credentials"
                client_id = $backendClient
                client_secret = $backendSecret
                scope = "$healerApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $response.access_token
            $payload.roles | Should -HaveCount 1
            $payload.roles | Should -Contain "read"
            $payload.aud | Should -Be $healerApi
        }

        It "Backend Client (Client Credentials) - LadderService .default -> read and write roles" {
            $body = @{
                grant_type = "client_credentials"
                client_id = $backendClient
                client_secret = $backendSecret
                scope = "$leadApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $response.access_token
            $payload.roles | Should -HaveCount 2
            $payload.roles | Should -Contain "read"
            $payload.roles | Should -Contain "write"
            $payload.aud | Should -Be $leadApi
        }

        It "Backend Client (Client Credentials) - FuzzService .default -> no roles" {
            $body = @{
                grant_type = "client_credentials"
                client_id = $backendClient
                client_secret = $backendSecret
                scope = "$fuzzApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $response.access_token
            $payload.roles | Should -BeNullOrEmpty
            $payload.aud | Should -Be $fuzzApi
        }

        It "Backend Client (Client Credentials) - Auth Code flow should not be permitted" {
            $body = @{
                grant_type = "authorization_code"
                client_id = $backendClient
                code = "any-code"
                redirect_uri = "https://not-registered"
            }
            try {
                Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -ErrorAction Stop -TimeoutSec 1
                fail "Should have failed as auth code flow is not permitted for this client"
            } catch {
                $_.Exception.Response.StatusCode | Should -Be 401
            }
        }

        It "Frontend Client (Password Flow) - healer user + MMB .default -> read and write role" {
            $body = @{
                grant_type = "password"
                username = "healer"
                password = "healerPassword_123"
                client_id = $frontendClient
                scope = "$mmbApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $response.access_token
            $payload.roles | Should -HaveCount 2
            $payload.roles | Should -Contain "read"
            $payload.roles | Should -Contain "write"
            $payload.aud | Should -Be $mmbApi
        }

        It "Frontend Client (Password Flow) - diego user + MMB .default -> admin role" {
            $body = @{
                grant_type = "password"
                username = "diego"
                password = "diegoPassword_123"
                client_id = $frontendClient
                scope = "$mmbApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $response.access_token
            $payload.roles | Should -HaveCount 1
            $payload.roles | Should -Contain "admin"
            $payload.aud | Should -Be $mmbApi
        }

        It "Frontend Client (Password Flow) - anyuser user + MMB .default -> no role" {
            $body = @{
                grant_type = "password"
                username = "anyuser"
                password = "anyPassword123"
                client_id = $frontendClient
                scope = "$mmbApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $response.access_token
            $payload.roles | Should -BeNullOrEmpty
            $payload.aud | Should -Be $mmbApi
        }

        It "Frontend Client (Password Flow) - callcenter user + MMB .default -> call-center role" {
            $body = @{
                grant_type = "password"
                username = "callcenter"
                password = "password123"
                client_id = $frontendClient
                scope = "$mmbApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $response.access_token
            $payload.roles | Should -HaveCount 1
            $payload.roles | Should -Contain "call-center"
            $payload.aud | Should -Be $mmbApi
        }

        It "Frontend Client (Password Flow) - HealerService .default should not be permitted" {
            $body = @{
                grant_type = "password"
                username = "healer"
                password = "healerPassword_123"
                client_id = $frontendClient
                scope = "$healerApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $response.access_token
            $payload.roles | Should -BeNullOrEmpty
            $payload.aud | Should -Be $healerApi
        }

        It "Frontend Client (Credentials Flow) - should be permitted giving the response entra would give" {
            $body = @{
                grant_type = "client_credentials"
                client_id = $frontendClient
                client_secret = "arbitrary-secret"
                scope = "$mmbApi/.default"
            }
            try {
                Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -ErrorAction Stop -TimeoutSec 1
                fail "Should have failed or not returned a token"
            } catch {
                $_.Exception.Response.StatusCode | Should -Be 401
            }
        }
    }

    Context "Identity: Interactive Flow (Authorization Code)" {
        It "Should show the login page" {
            $authorizeUrl = "$baseUrl/$tenantId/oauth2/v2.0/authorize?client_id=$frontendClient&redirect_uri=$redirectUri&response_type=code&scope=$mmbApi/.default&state=1234&nonce=abcd"
            $response = Invoke-WebRequest -Uri $authorizeUrl -MaximumRedirection 0 -TimeoutSec 1
            $response.StatusCode | Should -Be 200
            $response.Content | Should -Match "Sign in"
            $response.Content | Should -Match "username"
            $response.Content | Should -Match "password"
        }

        It "Should authenticate and redirect with code" {
            $loginUrl = "$baseUrl/login"
            $body = @{
                username = "diego"
                password = "diegoPassword_123"
                client_id = $frontendClient
                redirect_uri = $redirectUri
                state = "1234"
                scope = "$mmbApi/.default"
                tenant = $tenantId
                nonce = "abcd"
            }
            try {
                $response = Invoke-WebRequest -Uri $loginUrl -Method Post -Body $body -MaximumRedirection 0 -TimeoutSec 1
                fail "Should have redirected (302)"
            } catch {
                $response = $_.Exception.Response
                $response.StatusCode | Should -Be 302
            }
            $response.Headers.Location | Should -Match "code="
            $response.Headers.Location | Should -Match "state=1234"
            
            # Extract code
            $location = $response.Headers.Location
            $code = ($location -split 'code=')[1] -split '&' | Select-Object -First 1
            $code | Should -Not -BeNullOrEmpty

            # Exchange code for token
            $tokenBody = @{
                grant_type = "authorization_code"
                client_id = $frontendClient
                code = $code
                redirect_uri = $redirectUri
            }
            $tokenResponse = Invoke-RestMethod -Method Post -Uri "$baseUrl/$tenantId/oauth2/v2.0/token" -Body $tokenBody -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $tokenResponse.access_token | Should -Not -BeNullOrEmpty
            
            $payload = Get-JwtPayload $tokenResponse.access_token
            $payload.roles | Should -Contain "admin"
            $payload.aud | Should -Be $mmbApi
            $payload.nonce | Should -Be "abcd"
        }

        It "Should fail with invalid redirect_uri" {
            $authorizeUrl = "$baseUrl/$tenantId/oauth2/v2.0/authorize?client_id=$frontendClient&redirect_uri=https://evil.com&response_type=code&scope=$mmbApi/.default"
            try {
                Invoke-WebRequest -Uri $authorizeUrl -MaximumRedirection 0 -ErrorAction Stop -TimeoutSec 1
                throw "Should have failed with 400 Bad Request"
            } catch {
                $_.Exception.Response.StatusCode | Should -Be 400
                $_.Exception.Message | Should -Match "400"
            }
        }

        It "Should authenticate and redirect with code for Graph User.Read" {
            $loginUrl = "$baseUrl/login"
            $body = @{
                username = "healer"
                password = "healerPassword_123"
                client_id = $frontendClient
                redirect_uri = $redirectUri
                state = "5678"
                scope = "$graphApi/User.Read"
                tenant = $tenantId
            }
            try {
                $response = Invoke-WebRequest -Uri $loginUrl -Method Post -Body $body -MaximumRedirection 0 -TimeoutSec 1
                fail "Should have redirected (302)"
            } catch {
                $response = $_.Exception.Response
                $response.StatusCode | Should -Be 302
            }
            $location = $response.Headers.Location
            $code = ($location -split 'code=')[1] -split '&' | Select-Object -First 1
            
            $tokenBody = @{
                grant_type = "authorization_code"
                client_id = $frontendClient
                code = $code
                redirect_uri = $redirectUri
            }
            $tokenResponse = Invoke-RestMethod -Method Post -Uri "$baseUrl/$tenantId/oauth2/v2.0/token" -Body $tokenBody -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $tokenResponse.access_token
            $payload.roles | Should -HaveCount 1
            $payload.roles | Should -Contain "User.Read"
            $payload.aud | Should -Be $graphApi
            $payload.scp | Should -Match "User.Read"
        }
    }

    Context "Identity: Entra ID Protocol Nuances" {
        It "Should return JSON error response for invalid credentials" {
            $body = @{
                grant_type = "password"
                username = "diego"
                password = "WRONG_PASSWORD"
                client_id = $frontendClient
            }
            try {
                Invoke-RestMethod -Uri "$baseUrl/token" -Method Post -Body $body -ContentType "application/x-www-form-urlencoded" -ErrorAction Stop -TimeoutSec 1
                throw "Should have failed with 401"
            } catch {
                $_.Exception.Response.StatusCode | Should -Be 401
                # In PowerShell 7, Invoke-RestMethod puts the JSON error body in ErrorDetails.Message
                $errorObj = $_.ErrorDetails.Message | ConvertFrom-Json
                $errorObj.error | Should -Be "invalid_grant"
                $errorObj.correlation_id | Should -Not -BeNullOrEmpty
                $errorObj.trace_id | Should -Not -BeNullOrEmpty
            }
        }

        It "Token should contain oid, ver, tid, and azp claims" {
            $body = @{
                grant_type = "client_credentials"
                client_id = $backendClient
                client_secret = $backendSecret
                scope = "$customerApi/.default"
            }
            # Use v2.0 endpoint for v2.0 token
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/$tenantId/oauth2/v2.0/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $response.access_token
            
            $payload.oid | Should -Be $backendClient
            $payload.tid | Should -Be $tenantId
            $payload.ver | Should -Be "2.0"
            $payload.azp | Should -Be $backendClient
            $payload.appid | Should -Be $backendClient
        }

        It "Token for user should contain preferred_username and name" {
             $body = @{
                grant_type = "password"
                username = "diego"
                password = "diegoPassword_123"
                client_id = $frontendClient
                scope = "$mmbApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $response.access_token
            
            $payload.name | Should -Be "Diego Admin"
            $payload.preferred_username | Should -Be "user@abroad.com"
            $payload.email | Should -Be "user@abroad.com"
            $payload.oid | Should -Not -BeNullOrEmpty
        }

        It "Should support v1.0 discovery and token endpoints" {
            $v1DiscoveryUrl = "$baseUrl/$tenantId/.well-known/openid-configuration"
            $response = Invoke-RestMethod -Method Get -Uri $v1DiscoveryUrl -TimeoutSec 1
            $response.issuer | Should -Not -Match "/v2.0"
            
            $body = @{
                grant_type = "client_credentials"
                client_id = $backendClient
                client_secret = $backendSecret
                scope = "$customerApi/.default"
            }
            # Use v1.0 token endpoint
            $tokenResponse = Invoke-RestMethod -Method Post -Uri "$baseUrl/$tenantId/oauth2/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $payload = Get-JwtPayload $tokenResponse.access_token
            $payload.ver | Should -Be "1.0"
            $payload.iss | Should -Not -Match "/v2.0"
        }
    }

    Context "Identity: Refresh Tokens" {
        It "Should return a refresh token when offline_access is requested (Password Flow)" {
            $body = @{
                grant_type = "password"
                username = "diego"
                password = "diegoPassword_123"
                client_id = $frontendClient
                scope = "openid offline_access $mmbApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $response.refresh_token | Should -Not -BeNullOrEmpty
            $response.access_token | Should -Not -BeNullOrEmpty
            $response.id_token | Should -Not -BeNullOrEmpty
        }

        It "Should use refresh token to get a new access token" {
            # 1. Get initial tokens
            $body = @{
                grant_type = "password"
                username = "healer"
                password = "healerPassword_123"
                client_id = $frontendClient
                scope = "openid offline_access $mmbApi/.default"
            }
            $response = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $refreshToken = $response.refresh_token
            $oldAccessToken = $response.access_token

            # 2. Refresh
            $refreshBody = @{
                grant_type = "refresh_token"
                client_id = $frontendClient
                refresh_token = $refreshToken
            }
            $refreshResponse = Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $refreshBody -ContentType "application/x-www-form-urlencoded" -TimeoutSec 1
            $refreshResponse.access_token | Should -Not -BeNullOrEmpty
            $refreshResponse.refresh_token | Should -Not -BeNullOrEmpty
            $refreshResponse.access_token | Should -Not -Be $oldAccessToken

            $payload = Get-JwtPayload $refreshResponse.access_token
            $payload.roles | Should -Contain "read"
            $payload.aud | Should -Be $mmbApi
        }

        It "Should fail with invalid refresh token" {
            $body = @{
                grant_type = "refresh_token"
                client_id = $frontendClient
                refresh_token = "invalid-token"
            }
            try {
                Invoke-RestMethod -Method Post -Uri "$baseUrl/token" -Body $body -ContentType "application/x-www-form-urlencoded" -ErrorAction Stop -TimeoutSec 1
                throw "Should have failed"
            } catch {
                $_.Exception.Response.StatusCode | Should -Be 401
                $errorObj = $_.ErrorDetails.Message | ConvertFrom-Json
                $errorObj.error | Should -Be "invalid_grant"
            }
        }
    }
}
