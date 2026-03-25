# FeDevRouting.Inside.Tests.ps1

Describe "FeDev Routing Validation" {
    BeforeAll {

        function Invoke-RequestWithRetry([string]$Url, [int]$MaxAttempts = 15, [int]$DelaySeconds = 1) {
            for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
                try {
                    return Invoke-WebRequest -Method Get -Uri $Url -TimeoutSec 3
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

    It "fe-dev routing uses different hosts and returns payloads with backend port 5173" {
        $response1 = Invoke-RequestWithRetry -Url "https://test-fe-dev-1.fe-dev.internal"
        $response1.StatusCode | Should -Be 200
        $response1.Content | Should -Be "hello from test-fe-dev-1.fe-dev.internal on port 5173"

        $response2 = Invoke-RequestWithRetry -Url "https://test-fe-dev-2.fe-dev.internal"
        $response2.StatusCode | Should -Be 200
        $response2.Content | Should -Be "hello from test-fe-dev-2.fe-dev.internal on port 5173"
    }
}