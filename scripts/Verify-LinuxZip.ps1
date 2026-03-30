#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Verifies that a Linux provider zip has proper executable permissions
.DESCRIPTION
    Extracts and checks the zip file using WSL to verify Unix permissions are preserved
.PARAMETER ZipPath
    Path to the zip file to verify
.EXAMPLE
    .\Verify-LinuxZip.ps1 -ZipPath "dist/terraform-provider-genesyscloud_1.77.2_linux_amd64.zip"
#>

param(
    [Parameter(Mandatory = $true)]
    [string]$ZipPath
)

$ErrorActionPreference = "Stop"

Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "  Linux Provider Zip Verification" -ForegroundColor White
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host ""

# Check if file exists
if (-not (Test-Path $ZipPath)) {
    Write-Host "❌ File not found: $ZipPath" -ForegroundColor Red
    exit 1
}

$zipFile = Get-Item $ZipPath
$zipSizeMB = [math]::Round($zipFile.Length / 1MB, 2)

Write-Host "File: $($zipFile.Name)" -ForegroundColor Yellow
Write-Host "Size: $zipSizeMB MB" -ForegroundColor Gray
Write-Host "Path: $($zipFile.FullName)" -ForegroundColor Gray
Write-Host ""

# Size check (rough heuristic)
if ($zipSizeMB -lt 50) {
    Write-Host "⚠️  WARNING: Zip file seems small ($zipSizeMB MB)" -ForegroundColor Yellow
    Write-Host "   Expected size: ~66MB (with executable permissions)" -ForegroundColor Gray
    Write-Host "   Smaller size (~33MB) usually means no executable bit" -ForegroundColor Gray
    Write-Host ""
}

# Check if WSL is available
try {
    $null = wsl --version 2>&1
    $hasWSL = $LASTEXITCODE -eq 0
}
catch {
    $hasWSL = $false
}

if (-not $hasWSL) {
    Write-Host "⚠️  WSL not available - cannot verify Unix permissions" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "To install WSL and verify properly:" -ForegroundColor Gray
    Write-Host "  wsl --install" -ForegroundColor Gray
    Write-Host ""
    exit 0
}

# Convert path to WSL format
$fullPath = Resolve-Path $ZipPath
$wslPath = $fullPath.Path -replace '\\', '/' -replace '^([A-Z]):', { '/mnt/' + $_.Groups[1].Value.ToLower() }

Write-Host "Verifying in WSL..." -ForegroundColor Yellow
Write-Host ""

# Create temp directory and extract
$verifyScript = @"
set -e
TEMP_DIR=`$(mktemp -d)
trap 'rm -rf `$TEMP_DIR' EXIT

echo "Extracting zip file..."
unzip -q '$wslPath' -d `$TEMP_DIR

echo ""
echo "Checking permissions..."
cd `$TEMP_DIR

# Find the binary
BINARY=`$(find . -type f -name 'terraform-provider-genesyscloud*' | head -n1)

if [ -z "`$BINARY" ]; then
    echo "❌ No provider binary found in zip!"
    exit 1
fi

echo "Binary: `$BINARY"
echo ""

# Check permissions
PERMS=`$(ls -la `$BINARY)
echo "Permissions: `$PERMS"
echo ""

# Check if executable
if [ -x "`$BINARY" ]; then
    echo "✅ PASS - Binary has executable permission (+x)"
    echo ""
    echo "This zip file is correctly built with Unix permissions."
    echo "It will work on Linux systems."
    exit 0
else
    echo "❌ FAIL - Binary does NOT have executable permission"
    echo ""
    echo "This zip was likely created on Windows without WSL."
    echo "It will fail on Linux with 'Permission denied' error."
    echo ""
    echo "Solution: Rebuild using WSL"
    echo "  ./scripts/Build-LinuxProviderWSL.ps1 -Version <version>"
    exit 1
fi
"@

# Run verification in WSL
$verifyScript | wsl bash

$exitCode = $LASTEXITCODE

Write-Host ""
Write-Host "=====================================================================" -ForegroundColor Cyan

if ($exitCode -eq 0) {
    Write-Host "  ✅ VERIFICATION PASSED" -ForegroundColor Green
    Write-Host "=====================================================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "The zip file has proper Unix executable permissions." -ForegroundColor White
    Write-Host "It will work correctly on Linux systems." -ForegroundColor White
}
else {
    Write-Host "  ❌ VERIFICATION FAILED" -ForegroundColor Red
    Write-Host "=====================================================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "The zip file does NOT have proper executable permissions." -ForegroundColor Red
    Write-Host "Rebuild using: .\scripts\Build-LinuxProviderWSL.ps1" -ForegroundColor Yellow
}

Write-Host ""

exit $exitCode
