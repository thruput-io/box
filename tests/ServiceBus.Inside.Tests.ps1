# ServiceBus.Inside.Tests.ps1

Describe "Service Bus Validation" {
    BeforeAll {
        $serviceBusDns = "my-namespace.servicebus.windows.net"
    }

    It "Can connect to Service Bus AMQP port (socket 5672)" {
        $tcpClient = New-Object System.Net.Sockets.TcpClient
        try {
            $connection = $tcpClient.BeginConnect($serviceBusDns, 5672, $null, $null)
            if (-not $connection.AsyncWaitHandle.WaitOne(2000)) {
                throw "Connection to $serviceBusDns:5672 timed out after 2 second"
            }
            $tcpClient.EndConnect($connection)
            $tcpClient.Connected | Should -Be $true
        } finally {
            $tcpClient.Close()
        }
    }

    It "Sends and receives a message using .NET tool" {
        $connectionString = "Endpoint=sb://$serviceBusDns;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=SAS_KEY_VALUE;UseDevelopmentEmulator=true;"
        $topicName = "test-topic"
        $subscriptionName = "test-subscription"

        Write-Host "DEBUG_LOG: Invoking dot-net-client sb test..."
        $result = dotnet /usr/local/bin/dot-net-client/dot-net-client.dll sb $connectionString $topicName $subscriptionName 2>&1
        Write-Host "DEBUG_LOG: dot-net-client finished. Exit code: $LASTEXITCODE"

        $outputString = $result | Out-String
        Write-Host "DEBUG_LOG: Output from tool:`n$outputString"

        $LASTEXITCODE | Should -Be 0
        $outputString | Should -Match "Sent message successfully"
        $outputString | Should -Match "Received message: Hello Service Bus!"
    }
}
