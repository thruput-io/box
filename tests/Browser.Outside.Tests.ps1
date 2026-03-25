# Browser.Outside.Tests.ps1

Describe "External Infrastructure Validation" {
    BeforeAll {
        $browserUrl = "https://browser.web.internal"

        function Invoke-BrowserRequestWithRetry([string]$Url, [int]$MaxAttempts = 2, [int]$DelaySeconds = 0) {
            for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
                $raw = & curl --silent --show-error --fail-with-body --max-time 5 -w "`n__STATUS__:%{http_code}" $Url
                if ($LASTEXITCODE -eq 0) {
                    $parts = "$raw" -split "__STATUS__:", 2
                    return @{
                        StatusCode = [int]$parts[1].Trim()
                        Content = $parts[0].TrimEnd()
                    }
                }

                if ($attempt -eq $MaxAttempts) {
                    throw "Request failed for $Url after $MaxAttempts attempts (exit code $LASTEXITCODE)."
                }

                Start-Sleep -Seconds $DelaySeconds
            }
        }
    }

    Context "Firefox Browser" {
        It "Firefox web UI is reachable from outside the infra network" {
            $response = Invoke-BrowserRequestWithRetry -Url $browserUrl
            $response.StatusCode | Should -Be 200
        }
    }
}
