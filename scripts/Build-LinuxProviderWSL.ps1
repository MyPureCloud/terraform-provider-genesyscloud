#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Builds Linux provider using WSL with proper executable permissions
.DESCRIPTION
    Wrapper script that invokes the Build-LinuxProvider.sh script in WSL
    to ensure Linux binaries have proper executable permissions.
.PARAMETER Version
    Version string for the provider (e.g., "1.77.2")
.PARAMETER OutputDir
    Output directory for build artifacts (default: ./dist)
.EXAMPLE
    .\Build-LinuxProviderWSL.ps1 -Version "1.77.2"
.NOTES
    Requires WSL to be installed on Windows
#>

param(
    [Parameter(Mandatory=$false)]
    [string]$Version = "1.0.0-custom",
    
    [Parameter(Mandatory=$false)]
    [string]$OutputDir = "./dist"
)

$ErrorActionPreference = "Stop"

Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "  Linux Provider Build (with Executable Permissions)" -ForegroundColor White
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host ""

# Check if WSL is installed
Write-Host "Checking WSL installation..." -ForegroundColor Yellow
try {
    $wslCheck = wsl --status 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "WSL check failed"
    }
    Write-Host "✅ WSL is installed" -ForegroundColor Green
} catch {
    Write-Host "❌ WSL is not installed or not working properly!" -ForegroundColor Red
    Write-Host ""
    Write-Host "To install WSL:" -ForegroundColor Yellow
    Write-Host "  wsl --install" -ForegroundColor Gray
    Write-Host ""
    Write-Host "After installation, restart your computer and run this script again." -ForegroundColor Yellow
    exit 1
}

# Get the current directory in WSL format
$currentDir = Get-Location
$wslPath = $currentDir.Path -replace '\\', '/' -replace '^([A-Z]):', { '/mnt/' + $_.Groups[1].Value.ToLower() }

Write-Host ""
Write-Host "Build Configuration:" -ForegroundColor Yellow
Write-Host "  Version: $Version" -ForegroundColor Gray
Write-Host "  Output Dir: $OutputDir" -ForegroundColor Gray
Write-Host "  WSL Path: $wslPath" -ForegroundColor Gray
Write-Host ""

# Check if build script exists
$buildScriptPath = Join-Path "scripts" "Build-LinuxProvider.sh"
if (-not (Test-Path $buildScriptPath)) {
    Write-Host "❌ Build script not found: $buildScriptPath" -ForegroundColor Red
    exit 1
}

# Make script executable in WSL
Write-Host "Setting script permissions..." -ForegroundColor Yellow
wsl bash -c "cd '$wslPath' && chmod +x scripts/Build-LinuxProvider.sh"

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to set script permissions" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Script permissions set" -ForegroundColor Green
Write-Host ""
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "  Building in WSL..." -ForegroundColor White
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host ""

# Run the build script in WSL
wsl bash -c "cd '$wslPath' && ./scripts/Build-LinuxProvider.sh '$Version' '$OutputDir'"

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "  Build Completed Successfully!" -ForegroundColor Green
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "The Linux provider has been built with proper executable permissions." -ForegroundColor White
Write-Host ""
Write-Host "Files created:" -ForegroundColor Yellow
Write-Host "  📦 $OutputDir/terraform-provider-genesyscloud_${Version}_linux_amd64.zip" -ForegroundColor Gray
Write-Host "  📄 terraform-local-testing/terraform-provider-mirror/.../${Version}.json" -ForegroundColor Gray
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "  1. Verify the zip file (see scripts/Verify-LinuxZip.ps1)" -ForegroundColor Gray
Write-Host "  2. Upload to your hosting location (Azure Blob, etc.)" -ForegroundColor Gray
Write-Host "  3. Run: terraform init" -ForegroundColor Gray
Write-Host ""
