#!/usr/bin/env pwsh
param(
    [Parameter(Position=0)]
    [ValidateSet('install', 'start', 'stop', 'restart', 'status')]
    [string]$Command,

    [ValidateRange(1, 65535)]
    [int]$Port = 8085
)

$ErrorActionPreference = "Stop"

$WIREMOCK_PORTS = @(8085, 5173)
$STATUS_PORTS = if ($PSBoundParameters.ContainsKey("Port")) { @($Port) } else { $WIREMOCK_PORTS }
$SCRIPT_DIR = Split-Path -Parent -Resolve $MyInvocation.MyCommand.Path
$WIREMOCK_DIR = $SCRIPT_DIR
$LOG_DIR = Join-Path $WIREMOCK_DIR "log"

# Ensure ~/.dotnet/tools is in PATH for Start-Process to find dotnet-wiremock
$dotnetToolsPath = Join-Path $Home ".dotnet/tools"
if (Test-Path $dotnetToolsPath) {
    if ($env:PATH -notlike "*$dotnetToolsPath*") {
        $env:PATH = "$dotnetToolsPath$([IO.Path]::PathSeparator)$env:PATH"
    }
}

# Create directory structure
New-Item -ItemType Directory -Force -Path (Join-Path $WIREMOCK_DIR "__admin/mappings") | Out-Null
New-Item -ItemType Directory -Force -Path $LOG_DIR | Out-Null

function Test-WireMockPort([int]$TargetPort) {
    try {
        $adminUrl = "http://localhost:${TargetPort}/__admin/mappings"
        $response = Invoke-WebRequest -Method Get -Uri $adminUrl -TimeoutSec 2
        if ($response.StatusCode -ne 200) {
            return $false
        }

        $contentType = "$($response.Headers["Content-Type"])"
        if ($contentType -notlike "application/json*") {
            return $false
        }

        $null = $response.Content | ConvertFrom-Json -ErrorAction Stop
        return $true
    }
    catch {
        return $false
    }
}

function Install-WireMock {
    # Check if already installed
    $installed = dotnet tool list --global | Select-String "dotnet-wiremock"
    if ($installed) {
        Write-Host "WireMock.Net is already installed as global tool"
        return
    }

    Write-Host "Installing WireMock.Net as global tool..."

    try {
        dotnet tool install --global WireMock.Net.StandAlone
        Write-Host "✓ WireMock.Net installed as global tool"
    } catch {
        Write-Error "Failed to install WireMock.Net: $_"
    }
}

function Start-WireMock {
    # Check if WireMock.Net is installed
    $installed = dotnet tool list --global | Select-String "wiremock.net.standalone"
    if (-not $installed) {
        Install-WireMock
    }

    $targetPorts = $WIREMOCK_PORTS | Sort-Object -Unique
    $healthyPorts = @()
    foreach ($targetPort in $targetPorts) {
        if (Test-WireMockPort -TargetPort $targetPort) {
            $healthyPorts += $targetPort
        }
    }

    if ($healthyPorts.Count -eq $targetPorts.Count) {
        Write-Host "WireMock is already serving ports: $($targetPorts -join ', ')"
        return
    }

    if (Get-Process -Name "dotnet-wiremock" -ErrorAction SilentlyContinue) {
        Stop-WireMock
        Start-Sleep -Milliseconds 500
    }

    foreach ($targetPort in $targetPorts) {
        $stdoutFile = Join-Path $LOG_DIR "wiremock.${targetPort}.stdout"
        $stderrFile = Join-Path $LOG_DIR "wiremock.${targetPort}.stderr"

        Write-Host "Starting WireMock on port ${targetPort}..."

        Start-Process -FilePath "dotnet-wiremock" -ArgumentList "--Port $targetPort --WatchStaticMappings true --ReadStaticMappings true --WireMockLogger WireMockConsoleLogger" -WorkingDirectory $WIREMOCK_DIR -NoNewWindow -RedirectStandardOutput $stdoutFile -RedirectStandardError $stderrFile
    }

    $pendingPorts = @($targetPorts)
    for ($attempt = 0; $attempt -lt 20 -and $pendingPorts.Count -gt 0; $attempt++) {
        Start-Sleep -Milliseconds 250
        $pendingPorts = @($pendingPorts | Where-Object { -not (Test-WireMockPort -TargetPort $_) })
    }

    if ($pendingPorts.Count -gt 0) {
        throw "WireMock failed to start on port(s): $($pendingPorts -join ', '). Check if ports are already in use."
    }
}

function Stop-WireMock {
    Write-Host "Stopping WireMock..."
    Get-Process -Name "dotnet-wiremock" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

    # Wait and verify all processes are killed
    for ($i = 0; $i -lt 10; $i++) {
        $remainingProcesses = Get-Process -Name "dotnet-wiremock" -ErrorAction SilentlyContinue
        $respondingPorts = @()
        foreach ($targetPort in $WIREMOCK_PORTS) {
            if (Test-WireMockPort -TargetPort $targetPort) {
                $respondingPorts += $targetPort
            }
        }

        if ((-not $remainingProcesses) -and $respondingPorts.Count -eq 0) {
            return
        }
        Start-Sleep -Milliseconds 500
        # Force kill any remaining processes
        $remainingProcesses | Stop-Process -Force -ErrorAction SilentlyContinue
    }
}

function Get-WireMockStatus {
    $processes = Get-Process -Name "dotnet-wiremock" -ErrorAction SilentlyContinue

    if (-not $processes) {
        Write-Host "WireMock is not running"
        return $false
    }

    $portsToCheck = $STATUS_PORTS | Sort-Object -Unique
    $unhealthyPorts = @()

    foreach ($targetPort in $portsToCheck) {
        if (-not (Test-WireMockPort -TargetPort $targetPort)) {
            $unhealthyPorts += $targetPort
        }
    }

    if ($unhealthyPorts.Count -eq 0) {
        Write-Host "WireMock is running (PID: $($processes.Id -join ', '))"
        foreach ($targetPort in $portsToCheck) {
            Write-Host "  Admin: http://localhost:${targetPort}/__admin/mappings"
        }
        return $true
    }

    Write-Host "WireMock process exists but is not responding on port(s): $($unhealthyPorts -join ', ')"
    return $false
}

# Main command handling
switch ($Command) {
    'install' { Install-WireMock }
    'start' { Start-WireMock }
    'stop' { Stop-WireMock }
    'restart' {
        Stop-WireMock
        Start-Sleep -Seconds 1
        Start-WireMock
    }
    'status' { Get-WireMockStatus }
    default {
        Write-Host "Usage: $($MyInvocation.MyCommand.Name) {install|start|stop|restart|status} [-Port <1-65535>]"
        exit 1
    }
}
