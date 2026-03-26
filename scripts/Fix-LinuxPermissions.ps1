#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Quick fix for existing version with permission issues
.DESCRIPTION
    Rebuilds an existing version with proper executable permissions and updates the manifest
.PARAMETER Version
    Version to rebuild (e.g., "1.77.2")
.EXAMPLE
    .\Fix-LinuxPermissions.ps1 -Version "1.77.2"
#>

param(
    [Parameter(Mandatory=$true)]
    [string]$Version
)

$ErrorActionPreference = "Stop"

Clear-Host
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "  Fix Linux Provider Executable Permissions" -ForegroundColor White
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Version: $Version" -ForegroundColor Yellow
Write-Host ""

# Check if version exists
$mirrorDir = "terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud"
$zipFile = Join-Path $mirrorDir "terraform-provider-genesyscloud_${Version}_linux_amd64.zip"

if (Test-Path $zipFile) {
    Write-Host "Found existing zip: $zipFile" -ForegroundColor Gray
    $oldSize = [math]::Round((Get-Item $zipFile).Length / 1MB, 2)
    Write-Host "Current size: $oldSize MB" -ForegroundColor Gray
    
    if ($oldSize -lt 50) {
        Write-Host "❌ Size indicates missing executable permissions!" -ForegroundColor Red
    } else {
        Write-Host "⚠️  Size looks OK, but verifying..." -ForegroundColor Yellow
    }
    Write-Host ""
    
    # Verify first
    Write-Host "Checking current zip..." -ForegroundColor Yellow
    try {
        & "$PSScriptRoot\Verify-LinuxZip.ps1" -ZipPath $zipFile
        if ($LASTEXITCODE -eq 0) {
            Write-Host ""
            Write-Host "✅ Current version already has proper permissions!" -ForegroundColor Green
            Write-Host "No rebuild needed." -ForegroundColor Green
            Write-Host ""
            exit 0
        }
    } catch {
        # Verification failed, continue with rebuild
    }
    Write-Host ""
}

Write-Host "Starting rebuild process..." -ForegroundColor Cyan
Write-Host ""

# Step 1: Build with WSL
Write-Host "Step 1: Building with WSL..." -ForegroundColor Yellow
Write-Host ("=" * 70) -ForegroundColor Gray

& "$PSScriptRoot\Build-LinuxProviderWSL.ps1" -Version $Version

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Step 2: Verifying build..." -ForegroundColor Yellow
Write-Host ("=" * 70) -ForegroundColor Gray

$newZipPath = "dist/terraform-provider-genesyscloud_${Version}_linux_amd64.zip"

& "$PSScriptRoot\Verify-LinuxZip.ps1" -ZipPath $newZipPath

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "❌ Verification failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "  ✅ FIX COMPLETE!" -ForegroundColor Green
Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host ""

$newSize = [math]::Round((Get-Item $newZipPath).Length / 1MB, 2)

Write-Host "Results:" -ForegroundColor Yellow
if (Test-Path $zipFile) {
    Write-Host "  Old size: $oldSize MB (no permissions)" -ForegroundColor Gray
}
Write-Host "  New size: $newSize MB (with permissions)" -ForegroundColor Green
Write-Host ""

Write-Host "Files updated:" -ForegroundColor Yellow
Write-Host "  ✅ $mirrorDir/terraform-provider-genesyscloud_${Version}_linux_amd64.zip" -ForegroundColor Green
Write-Host "  ✅ $mirrorDir/${Version}.json (manifest with new checksum)" -ForegroundColor Green
Write-Host ""

Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. Upload the new zip file to your hosting location:" -ForegroundColor White
Write-Host ""
Write-Host "   az storage blob upload \`" -ForegroundColor Gray
Write-Host "     --account-name YOUR_STORAGE_ACCOUNT \`" -ForegroundColor Gray
Write-Host "     --container-name YOUR_CONTAINER \`" -ForegroundColor Gray
Write-Host "     --name `"registry.terraform.io/mypurecloud/genesyscloud/terraform-provider-genesyscloud_${Version}_linux_amd64.zip`" \`" -ForegroundColor Gray
Write-Host "     --file `"$mirrorDir/terraform-provider-genesyscloud_${Version}_linux_amd64.zip`" \`" -ForegroundColor Gray
Write-Host "     --overwrite" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Upload the updated manifest:" -ForegroundColor White
Write-Host ""
Write-Host "   az storage blob upload \`" -ForegroundColor Gray
Write-Host "     --account-name YOUR_STORAGE_ACCOUNT \`" -ForegroundColor Gray
Write-Host "     --container-name YOUR_CONTAINER \`" -ForegroundColor Gray
Write-Host "     --name `"registry.terraform.io/mypurecloud/genesyscloud/${Version}.json`" \`" -ForegroundColor Gray
Write-Host "     --file `"$mirrorDir/${Version}.json`" \`" -ForegroundColor Gray
Write-Host "     --overwrite" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Clear Terraform cache and test:" -ForegroundColor White
Write-Host ""
Write-Host "   rm -rf .terraform/ .terraform.lock.hcl" -ForegroundColor Gray
Write-Host "   terraform init" -ForegroundColor Gray
Write-Host ""
