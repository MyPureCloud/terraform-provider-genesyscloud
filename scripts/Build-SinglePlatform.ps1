#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Builds the Terraform provider for a single platform
.DESCRIPTION
    Simplified build for container/specific platform deployment
.EXAMPLE
    .\Build-SinglePlatform.ps1 -Platform "linux_amd64" -Version "1.0.0-custom"
#>

param(
    [string]$Platform = "linux_amd64",
    [string]$Version = "1.0.0-custom",
    [string]$OutputDir = ".\dist"
)

$ErrorActionPreference = "Stop"

Write-Host "Building Terraform Provider for $Platform" -ForegroundColor Cyan
Write-Host "Version: $Version`n" -ForegroundColor Gray

# Parse platform
$parts = $Platform -split "_"
$os = $parts[0]
$arch = $parts[1]

Write-Host "OS: $os" -ForegroundColor Yellow
Write-Host "Architecture: $arch`n" -ForegroundColor Yellow

# Clean and create output directory
if (Test-Path $OutputDir) {
    Remove-Item "$OutputDir\*" -Recurse -Force
}
New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

# Set environment variables for cross-compilation
$env:GOOS = $os
$env:GOARCH = $arch
$env:CGO_ENABLED = "0"

# Determine file extension
$ext = if ($os -eq "windows") { ".exe" } else { "" }
$binaryName = "terraform-provider-genesyscloud${ext}"
$outputPath = Join-Path $OutputDir $binaryName

Write-Host "Building..." -ForegroundColor Yellow

# Build
go build -o $outputPath -ldflags "-s -w" .

if ($LASTEXITCODE -eq 0) {
    $size = (Get-Item $outputPath).Length / 1MB
    Write-Host "`n✓ Build successful!" -ForegroundColor Green
    Write-Host "  Binary: $outputPath" -ForegroundColor Cyan
    Write-Host "  Size: $([math]::Round($size, 2)) MB`n" -ForegroundColor Gray
    
    # Generate checksum
    $hash = (Get-FileHash $outputPath -Algorithm SHA256).Hash.ToLower()
    "$hash  $binaryName" | Out-File -FilePath "$OutputDir\SHA256SUM" -Encoding ASCII
    Write-Host "✓ Checksum generated: $OutputDir\SHA256SUM" -ForegroundColor Green
    
    Write-Host "`n" ("=" * 70) -ForegroundColor Cyan
    Write-Host "  NEXT STEPS FOR CONTAINER DEPLOYMENT" -ForegroundColor White
    Write-Host ("=" * 70) -ForegroundColor Cyan
    
    Write-Host "`n1. Copy to your container:`n" -ForegroundColor Yellow
    
    $dockerExample = @"
# In your Dockerfile:
COPY dist/terraform-provider-genesyscloud /root/.terraform.d/plugins/genesys.com/mypurecloud/genesyscloud/$Version/linux_amd64/terraform-provider-genesyscloud
RUN chmod +x /root/.terraform.d/plugins/genesys.com/mypurecloud/genesyscloud/$Version/linux_amd64/terraform-provider-genesyscloud
"@
    
    Write-Host $dockerExample -ForegroundColor Gray
    
    Write-Host "`n2. Or use in your terraform config:`n" -ForegroundColor Yellow
    
    $tfConfig = @"
terraform {
  required_providers {
    genesyscloud = {
      source  = "genesys.com/mypurecloud/genesyscloud"
      version = "$Version"
    }
  }
}
"@
    
    Write-Host $tfConfig -ForegroundColor Gray
    Write-Host ""
    
} else {
    Write-Host "`n✗ Build failed!" -ForegroundColor Red
    exit 1
}
