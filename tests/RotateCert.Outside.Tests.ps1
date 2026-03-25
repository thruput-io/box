# RotateCert.Outside.Tests.ps1

Describe "rotate-cert target" {
    It "cleans and refreshes localhost certificate trust when localhost setup is present" {
        $makefilePath = Join-Path $PSScriptRoot "../Makefile"
        $makefile = Get-Content -Path $makefilePath -Raw

        $rotateCertBlock = [regex]::Match(
            $makefile,
            '(?ms)^rotate-cert:.*?(?=^[a-zA-Z0-9_.-]+:|\z)'
        ).Value

        $rotateCertBlock | Should -Not -BeNullOrEmpty
        $rotateCertBlock | Should -Match '/etc/resolver/internal'
        $rotateCertBlock | Should -Match 'localhost/clean\.sh'
        $rotateCertBlock | Should -Match 'localhost/setup\.sh'
    }
}