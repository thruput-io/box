# FeDevRouting.Outside.Tests.ps1

$internalResolverPath = "/etc/resolver/internal"

Describe "fe-dev.internal routing configuration" {
    BeforeAll {
        $script:feDevHosts = @(
            "test-fe-dev-1.fe-dev.internal",
            "test-fe-dev-2.fe-dev.internal"
        )
        $script:internalResolverPath = $internalResolverPath
        $script:wiremockScriptPath = Join-Path $PSScriptRoot "wiremock/wiremock.ps1"
        $script:wiremockScriptPath | Should -Not -BeNullOrEmpty

        & $script:wiremockScriptPath restart -Port 5173

        function Convert-CurlResultToResponse([string]$RawResponse) {
            $parts = "$RawResponse" -split "__STATUS__:", 2
            return @{
                StatusCode = [int]$parts[1].Trim()
                Content = $parts[0].TrimEnd()
            }
        }

        function Invoke-FeDevRequestWithInternalResolver([string]$HostName, [int]$MaxAttempts = 2, [int]$DelaySeconds = 0) {
            for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
                $raw = & curl --silent --show-error --insecure --fail-with-body --max-time 5 -w "`n__STATUS__:%{http_code}" "https://$HostName/"
                if ($LASTEXITCODE -eq 0) {
                    return Convert-CurlResultToResponse -RawResponse $raw
                }

                if ($attempt -eq $MaxAttempts) {
                    throw "Request failed for https://$HostName after $MaxAttempts attempts (exit code $LASTEXITCODE)."
                }

                Start-Sleep -Seconds $DelaySeconds
            }
        }

        function Invoke-FeDevRequestWithoutLocalhostOverrides([string]$HostName, [int]$MaxAttempts = 2, [int]$DelaySeconds = 0) {
            $resolveArg = "$HostName`:443:127.0.0.1"
            for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
                $raw = & curl --silent --show-error --insecure --fail-with-body --max-time 5 --resolve $resolveArg -w "`n__STATUS__:%{http_code}" "https://$HostName/"
                if ($LASTEXITCODE -eq 0) {
                    return Convert-CurlResultToResponse -RawResponse $raw
                }

                if ($attempt -eq $MaxAttempts) {
                    throw "Request failed for https://$HostName after $MaxAttempts attempts (exit code $LASTEXITCODE)."
                }

                Start-Sleep -Seconds $DelaySeconds
            }
        }
    }

    AfterAll {
        & $script:wiremockScriptPath stop
    }

    Context "when /etc/resolver/internal exists" -Skip:(-not (Test-Path $internalResolverPath)) {
        It "invokes https calls on both fe-dev hosts" {
            $response1 = Invoke-FeDevRequestWithInternalResolver -HostName $script:feDevHosts[0]
            $response2 = Invoke-FeDevRequestWithInternalResolver -HostName $script:feDevHosts[1]

            $response1.StatusCode | Should -Be 200
            $response1.Content | Should -Be "hello from $($script:feDevHosts[0]) on port 5173"

            $response2.StatusCode | Should -Be 200
            $response2.Content | Should -Be "hello from $($script:feDevHosts[1]) on port 5173"
        }
    }

    Context "when using localhost with hostname and port" {
        It "returns expected responses for both fe-dev hosts" {
            $response1 = Invoke-FeDevRequestWithoutLocalhostOverrides -HostName $script:feDevHosts[0]
            $response2 = Invoke-FeDevRequestWithoutLocalhostOverrides -HostName $script:feDevHosts[1]

            $response1.StatusCode | Should -Be 200
            $response1.Content | Should -Be "hello from $($script:feDevHosts[0]) on port 5173"

            $response2.StatusCode | Should -Be 200
            $response2.Content | Should -Be "hello from $($script:feDevHosts[1]) on port 5173"
        }
    }

    Context "after WireMock is shut down when using localhost with hostname and port" {
        BeforeAll {
            & $script:wiremockScriptPath stop
        }

        It "throws for both fe-dev hosts" {
            { Invoke-FeDevRequestWithoutLocalhostOverrides -HostName $script:feDevHosts[0] -MaxAttempts 2 -DelaySeconds 0 } | Should -Throw
            { Invoke-FeDevRequestWithoutLocalhostOverrides -HostName $script:feDevHosts[1] -MaxAttempts 2 -DelaySeconds 0 } | Should -Throw
        }
    }
}