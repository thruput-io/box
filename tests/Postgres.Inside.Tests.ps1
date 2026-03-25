# Postgres.Inside.Tests.ps1

Describe "Postgres Validation" {
    BeforeAll {
        $wildcardDns = "my-app.postgres.database.azure.com"
    }

    It "Postgres: Connectivity via Azure DNS with SSL" {
        $env:PGPASSWORD = "postgres"
        $env:PGCONNECT_TIMEOUT = 1
        $result = psql -h $wildcardDns -U postgres -d postgres -t -c "SELECT 1;" 2>&1 | Out-String
        $result.Trim() | Should -Match "1"
    }
}
