# Tools.Outside.Tests.ps1

Describe "External Infrastructure Validation" {
    BeforeAll {
        $toolsSseUrl = "https://tools.web.internal/sse"
        $toolsMessagesBaseUrl = "https://tools.web.internal/messages"

        function Invoke-ToolsRequestWithRetry([string]$Url, [int]$MaxAttempts = 2, [int]$DelaySeconds = 0) {
            for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
                try {
                    # MCP over SSE is a streaming response. We intentionally cap the request time.
                    # curl exits with code 28 on timeout, which is expected as long as we received data.
                    $raw = & curl --silent --show-error --http1.1 --max-time 5 --include $Url

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

        function New-ToolsMcpSessionId([int]$MaxAttempts = 2) {
            # IMPORTANT: The tools server keeps sessions only while the SSE stream is open.
            # This function starts a long-lived curl process, reads the SSE output until it
            # finds the sessionId, and returns both the sessionId and the process so the caller
            # can keep it open while posting messages.
            for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
                $proc = $null
                try {
                    $args = @(
                        "--silent",
                        "--show-error",
                        "--http1.1",
                        "--no-buffer",
                        "--max-time", "15",
                        $toolsSseUrl
                    )

                    $psi = [System.Diagnostics.ProcessStartInfo]::new()
                    $psi.FileName = "curl"
                    foreach ($a in $args) { [void]$psi.ArgumentList.Add($a) }
                    $psi.RedirectStandardOutput = $true
                    $psi.RedirectStandardError = $true
                    $psi.UseShellExecute = $false

                    $proc = [System.Diagnostics.Process]::new()
                    $proc.StartInfo = $psi
                    [void]$proc.Start()

                    $deadline = [DateTime]::UtcNow.AddSeconds(6)
                    while (-not $proc.HasExited -and [DateTime]::UtcNow -lt $deadline) {
                        $line = $proc.StandardOutput.ReadLine()
                        if ($null -eq $line) {
                            Start-Sleep -Milliseconds 50
                            continue
                        }

                        if ($line -match "^data:\s*/messages\?sessionId=([^\s]+)\s*$") {
                            return [PSCustomObject]@{
                                SessionId = $Matches[1]
                                Process   = $proc
                            }
                        }
                    }

                    $stderr = $proc.StandardError.ReadToEnd()
                    throw "Timed out waiting for sessionId from tools SSE stream. stderr: ${stderr}"
                }
                catch {
                    if ($proc -and -not $proc.HasExited) {
                        try { $proc.Kill() } catch { }
                    }

                    if ($attempt -eq $MaxAttempts) {
                        throw
                    }
                }
            }
        }

        function Invoke-ToolsMcpMessage([string]$SessionId, [string]$JsonRpcBody) {
            if (-not $SessionId) {
                throw "SessionId must be provided"
            }

            $url = "${toolsMessagesBaseUrl}?sessionId=${SessionId}"

            $curlArgs = @(
                "--silent",
                "--show-error",
                "--http1.1",
                "--max-time", "5",
                "--header", "Content-Type: application/json",
                "--request", "POST",
                "--data", $JsonRpcBody,
                "--write-out", "`nHTTPSTATUS:%{http_code}`n",
                $url
            )

            $raw = & curl @curlArgs

            if ($LASTEXITCODE -ne 0) {
                throw "curl failed posting MCP message with exit code $LASTEXITCODE"
            }

            $text = ($raw -join "`n")
            if ($text -notmatch "(?m)^HTTPSTATUS:(\d{3})\s*$") {
                throw "Unable to parse HTTP status code from curl output."
            }

            return [PSCustomObject]@{
                StatusCode = [int]$Matches[1]
                Body       = ($text -replace "(?ms)\nHTTPSTATUS:\d{3}\s*$", "")
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

        It "Tools MCP message endpoint accepts JSON-RPC after establishing a session" {
            $session = New-ToolsMcpSessionId

            # We intentionally do not assert on the JSON-RPC *response* content here, because
            # in the legacy MCP HTTP+SSE transport the response is delivered on the SSE stream.
            # This is a connectivity test that verifies:
            # - session establishment works
            # - /messages is reachable and recognizes the session
            # - requests are accepted (no 404/500)
            try {
                $json = '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
                $post = Invoke-ToolsMcpMessage -SessionId $session.SessionId -JsonRpcBody $json

                # Depending on server/SDK version, this can be 200, 202, or 204.
                $post.StatusCode | Should -BeIn @(200, 202, 204)
            }
            finally {
                if ($session.Process -and -not $session.Process.HasExited) {
                    try { $session.Process.Kill() } catch { }
                }
            }
        }
    }
}
