# Wiremock.Inside.Tests.ps1

Describe "Wiremock Validation" {
    It "Wiremock on localhost is routed correctly" {
        $url1 = "https://test-mock-1.mock.internal/"
        $response1 = Invoke-WebRequest -Method Get -Uri $url1 -TimeoutSec 1
        $response1.StatusCode | Should -Be 200
        $response1.Content | Should -Be "hello from test-mock-1.mock.internal on port 8085"

        $url2 = "https://test-mock-2.mock.internal/"
        $response2 = Invoke-WebRequest -Method Get -Uri $url2 -TimeoutSec 1
        $response2.StatusCode | Should -Be 200
        $response2.Content | Should -Be "hello from test-mock-2.mock.internal on port 8085"

        $url404 = "https://not-there.mock.internal/"
        { Invoke-WebRequest -Method Get -Uri $url404 -TimeoutSec 1 } | Should -Throw
    }
}
