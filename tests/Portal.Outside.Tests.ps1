# Portal.Outside.Tests.ps1

Describe "External Infrastructure Validation" {
    BeforeAll {
        $portalUrl = "https://portal.web.internal"

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

    Context "Developer Portal" {
        It "Developer portal is reachable from outside the infra network" {
            $response = Invoke-PortalRequestWithRetry -Url $portalUrl
            $response.StatusCode | Should -Be 200
            $response.Content | Should -Match "Developer Portal"
        }
    }
}