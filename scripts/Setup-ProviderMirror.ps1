#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Sets up a Terraform provider network mirror directory structure
.DESCRIPTION
    Creates the directory structure and metadata files needed to serve 
    a custom Terraform provider via HTTP for automatic installation via terraform init
.PARAMETER OutputDir
    Base directory for the mirror (default: ./provider-mirror)
.PARAMETER ProviderVersion
    Version of the provider (e.g., 1.0.0-custom)
.PARAMETER Namespace
    Provider namespace (default: genesys.com/mypurecloud)
.PARAMETER ProviderName
    Provider name (default: genesyscloud)
.PARAMETER BinaryPath
    Path to the provider binary (default: ./dist/terraform-provider-genesyscloud)
.PARAMETER Platforms
    Platforms to include (default: linux_amd64,darwin_amd64,windows_amd64)
#>

param(
    [string]$OutputDir = "./provider-mirror",
    [string]$ProviderVersion = "1.0.0-custom",
    [string]$Namespace = "genesys.com/mypurecloud",
    [string]$ProviderName = "genesyscloud",
    [string]$BinaryPath = "./dist",
    [string]$Platforms = "linux_amd64,darwin_amd64,windows_amd64"
)

$ErrorActionPreference = "Stop"

Write-Host "Creating Terraform Provider Mirror Structure..." -ForegroundColor Cyan

# Parse namespace
$namespaceParts = $Namespace -split '/'
$hostname = $namespaceParts[0]
$orgname = $namespaceParts[1]

# Create base directory structure
$basePath = Join-Path $OutputDir $hostname $orgname $ProviderName
New-Item -ItemType Directory -Force -Path $basePath | Out-Null

Write-Host "Base path: $basePath" -ForegroundColor Gray

# Process each platform
$platformList = $Platforms -split ','
$versionsData = @()

foreach ($platform in $platformList) {
    Write-Host "`nProcessing platform: $platform" -ForegroundColor Yellow
    
    # Parse platform
    $os, $arch = $platform -split '_'
    
    # Determine binary extension
    $extension = if ($os -eq "windows") { ".exe" } else { "" }
    
    # Find binary file
    $binaryFile = Get-ChildItem -Path $BinaryPath -Filter "terraform-provider-${ProviderName}*${extension}" | 
        Where-Object { 
            $_.Name -match $platform -or 
            ($platformList.Count -eq 1 -and $_.Extension -eq $extension)
        } | 
        Select-Object -First 1
    
    if (-not $binaryFile) {
        Write-Warning "Binary not found for $platform, skipping..."
        continue
    }
    
    Write-Host "  Found binary: $($binaryFile.Name)" -ForegroundColor Gray
    
    # Create platform directory
    $platformPath = Join-Path $basePath $ProviderVersion $platform
    New-Item -ItemType Directory -Force -Path $platformPath | Out-Null
    
    # Copy binary
    $destBinary = Join-Path $platformPath "terraform-provider-${ProviderName}_v${ProviderVersion}${extension}"
    Copy-Item $binaryFile.FullName $destBinary -Force
    Write-Host "  Copied to: $destBinary" -ForegroundColor Gray
    
    # Calculate SHA256
    $hash = (Get-FileHash -Path $destBinary -Algorithm SHA256).Hash.ToLower()
    Write-Host "  SHA256: $hash" -ForegroundColor Gray
    
    # Create SHA256SUM file
    $shasumFile = Join-Path $platformPath "terraform-provider-${ProviderName}_${ProviderVersion}_SHA256SUMS"
    "$hash  terraform-provider-${ProviderName}_v${ProviderVersion}${extension}" | Out-File -FilePath $shasumFile -Encoding ASCII -NoNewline
    
    # Add to versions data
    $versionsData += @{
        os = $os
        arch = $arch
        filename = "terraform-provider-${ProviderName}_v${ProviderVersion}${extension}"
        download_url = "https://yourdomain.com/$hostname/$orgname/$ProviderName/$ProviderVersion/${platform}/terraform-provider-${ProviderName}_v${ProviderVersion}${extension}"
        shasums_url = "https://yourdomain.com/$hostname/$orgname/$ProviderName/$ProviderVersion/${platform}/terraform-provider-${ProviderName}_${ProviderVersion}_SHA256SUMS"
        shasum = $hash
    }
}

# Create versions.json
$versionsJson = @{
    versions = @{
        $ProviderVersion = @{
            protocols = @("5.0")
            platforms = $versionsData | ForEach-Object {
                @{
                    os = $_.os
                    arch = $_.arch
                }
            }
        }
    }
} | ConvertTo-Json -Depth 10

$versionsFile = Join-Path $basePath "versions.json"
$versionsJson | Out-File -FilePath $versionsFile -Encoding UTF8
Write-Host "`nCreated versions.json" -ForegroundColor Green

# Create download metadata for each version
foreach ($versionData in $versionsData) {
    $downloadJson = @{
        protocols = @("5.0")
        os = $versionData.os
        arch = $versionData.arch
        filename = $versionData.filename
        download_url = $versionData.download_url
        shasums_url = $versionData.shasums_url
        shasum = $versionData.shasum
    } | ConvertTo-Json -Depth 10
    
    $platform = "$($versionData.os)_$($versionData.arch)"
    $downloadFile = Join-Path $basePath $ProviderVersion "${platform}.json"
    $downloadJson | Out-File -FilePath $downloadFile -Encoding UTF8
}

Write-Host "`nProvider mirror created successfully!" -ForegroundColor Green
Write-Host "`nNext steps:" -ForegroundColor Cyan
Write-Host "1. Upload the contents of '$OutputDir' to a web server or Azure Blob Storage"
Write-Host "2. Update the download URLs in the JSON files to match your hosting location"
Write-Host "3. Configure Terraform CLI to use the mirror:"
Write-Host ""
Write-Host "   # In ~/.terraformrc or %APPDATA%\terraform.rc:" -ForegroundColor Yellow
Write-Host "   provider_installation {" -ForegroundColor Gray
Write-Host "     network_mirror {" -ForegroundColor Gray
Write-Host "       url = `"https://yourdomain.com/terraform-providers/`"" -ForegroundColor Gray
Write-Host "     }" -ForegroundColor Gray
Write-Host "   }" -ForegroundColor Gray
Write-Host ""
Write-Host "4. Then just run: terraform init" -ForegroundColor Yellow
