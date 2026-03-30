#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Starts a local HTTP server to serve the Terraform provider mirror
.DESCRIPTION
    Simple HTTP server for testing provider installation locally
.PARAMETER MirrorPath
    Path to the provider mirror directory (default: ./provider-mirror)
.PARAMETER Port
    Port to listen on (default: 8080)
#>

param(
    [string]$MirrorPath = "./provider-mirror",
    [int]$Port = 8080
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $MirrorPath)) {
    Write-Error "Mirror path not found: $MirrorPath"
    exit 1
}

Write-Host "Starting provider mirror HTTP server..." -ForegroundColor Cyan
Write-Host "  Path: $MirrorPath" -ForegroundColor Gray
Write-Host "  Port: $Port" -ForegroundColor Gray
Write-Host ""
Write-Host "Configure Terraform with:" -ForegroundColor Yellow
Write-Host "  provider_installation {" -ForegroundColor Gray
Write-Host "    network_mirror {" -ForegroundColor Gray
Write-Host "      url = `"http://localhost:$Port/`"" -ForegroundColor Gray
Write-Host "    }" -ForegroundColor Gray
Write-Host "  }" -ForegroundColor Gray
Write-Host ""
Write-Host "Press Ctrl+C to stop..." -ForegroundColor Yellow
Write-Host ""

# Start Python HTTP server
Push-Location $MirrorPath
try {
    if (Get-Command python -ErrorAction SilentlyContinue) {
        python -m http.server $Port
    } elseif (Get-Command python3 -ErrorAction SilentlyContinue) {
        python3 -m http.server $Port
    } else {
        Write-Error "Python not found. Install Python or use another HTTP server."
    }
} finally {
    Pop-Location
}
