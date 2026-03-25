# SetupLocalhost.Outside.Tests.ps1

$composeRoot = Join-Path $PSScriptRoot ".."
$certPath = Join-Path $composeRoot "certs/dev-root-ca.crt"
$hasLocalCert = Test-Path -Path $certPath

Describe "Localhost Setup Validation" {
    BeforeAll {
        $portalUrl = "https://portal.web.internal"
        $portalHost = "portal.web.internal"

        function Invoke-PortalRequestWithRetry([string]$Url, [int]$MaxAttempts = 2, [int]$DelaySeconds = 0) {
            for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
                try {
                    $raw = & curl --silent --show-error --fail-with-body --http1.1 --max-time 5 --write-out "`nSTATUS:%{http_code}" $Url

                    if ($LASTEXITCODE -ne 0) {
                        throw "curl failed with exit code $LASTEXITCODE"
                    }

                    $rawText = ($raw -join "`n")
                    if ($rawText -notmatch "STATUS:(\d{3})\s*$") {
                        throw "Unable to parse status code from curl output."
                    }

                    $statusCode = [int]$Matches[1]
                    $content = $rawText -replace "(\r?\n)?STATUS:\d{3}\s*$", ""
                    return [PSCustomObject]@{
                        StatusCode = $statusCode
                        Content    = $content
                        Scheme     = "https"
                    }
                }
                catch {
                    if ($attempt -eq $MaxAttempts) {
                        throw
                    }

                    Start-Sleep -Seconds $DelaySeconds
                }
            }
        }
    }

    Context "setup-localhost" {
        It "Portal is reachable by URL over HTTPS and returns expected page body" -Skip:(-not $hasLocalCert) {
            try {
                [System.Net.Dns]::GetHostEntry($portalHost) | Out-Null
            }
            catch {
                Set-ItResult -Skipped -Because "Host DNS for *.internal is not configured. Run setup-localhost first."
                return
            }

            $response = Invoke-PortalRequestWithRetry -Url $portalUrl

            $response.StatusCode | Should -Be 200
            $response.Scheme | Should -Be "https"
            $response.Content | Should -Match "Developer Portal"
        }
    }
}
