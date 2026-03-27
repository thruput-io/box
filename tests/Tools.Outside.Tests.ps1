# Tools.Outside.Tests.ps1

Describe "External Infrastructure Validation" {
    BeforeAll {
        $toolsSseUrl = "https://tools.web.internal/sse"

        $clientCertPath = Join-Path $PSScriptRoot "../certs/tools-client.crt"
        $clientKeyPath = Join-Path $PSScriptRoot "../certs/tools-client.key"

        function Invoke-ToolsRequestWithRetry([string]$Url, [int]$MaxAttempts = 2, [int]$DelaySeconds = 0) {
            for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
                try {
                    if (-not (Test-Path -Path $clientCertPath) -or -not (Test-Path -Path $clientKeyPath)) {
                        throw "Missing tools client cert/key. Expected '$clientCertPath' and '$clientKeyPath'. Generate them with ./localhost/generate-tools-client-cert.sh"
                    }

                    # MCP over SSE is a streaming response. We intentionally cap the request time.
                    # curl exits with code 28 on timeout, which is expected as long as we received data.
                    $raw = & curl --silent --show-error --http1.1 --max-time 5 --include --cert $clientCertPath --key $clientKeyPath $Url

                    if (($LASTEXITCODE -ne 0) -and ($LASTEXITCODE -ne 28)) {
                        throw "curl failed with exit code $LASTEXITCODE"
                    }

                    $rawText = ($raw -join "`n")
                    if ($rawText -notmatch "(?m)^HTTP/\d(?:\.\d)?\s+(\d{3})\b") {
                        throw "Unable to parse HTTP status code from curl output."
                    }

                    $statusCode = [int]$Matches[1]
                    # Strip headers (first header block) and keep body.
                    $content = $rawText -replace "(?s)^.*?\r?\n\r?\n", ""
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

    Context "Tools (MCP over SSE)" {
        It "Tools SSE endpoint is reachable from outside the infra network" {
            $response = Invoke-ToolsRequestWithRetry -Url $toolsSseUrl
            $response.StatusCode | Should -Be 200

            # Validate we got an SSE-style payload
            $response.Content | Should -Match "(?m)^event:\s*endpoint\s*$"
            $response.Content | Should -Match "(?m)^data:\s*/messages\?sessionId="
        }
    }
}
