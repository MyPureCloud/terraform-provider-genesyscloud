# Automated Dependency Resolution Tests
# Tests for include_filter_resources_by_id with enable_dependency_resolution

param(
    [string]$FlowId = "b84cbae3-7c54-45dc-ade0-7a30fbccf996",
    [string]$FlowName = "Email Decryption Flow"
)

$ErrorActionPreference = "Continue"
$TestResults = @()

function Write-TestHeader {
    param([string]$Title)
    Write-Host "`n================================================================" -ForegroundColor Cyan
    Write-Host "  $Title" -ForegroundColor White
    Write-Host "================================================================" -ForegroundColor Cyan
}

function Write-TestResult {
    param([string]$Name, [bool]$Passed, [string]$Details)
    
    $status = if ($Passed) { "[PASS]" } else { "[FAIL]" }
    $color = if ($Passed) { "Green" } else { "Red" }
    
    Write-Host "`n$status $Name" -ForegroundColor $color
    if ($Details) { Write-Host "  $Details" -ForegroundColor Gray }
    
    $script:TestResults += [PSCustomObject]@{
        Test = $Name
        Passed = $Passed
        Details = $Details
    }
}

function New-TestConfig {
    param([string]$Dir, [string]$FilterType, [string]$FilterValue, [bool]$EnableDeps, [string]$Format = "json")
    
    if (Test-Path $Dir) { Remove-Item $Dir -Recurse -Force }
    New-Item -ItemType Directory -Path $Dir -Force | Out-Null
    
    $filterLine = if ($FilterType -eq "regex") {
        "  include_filter_resources     = [`"$FilterValue`"]"
    } else {
        "  include_filter_resources_by_id = [`"$FilterValue`"]"
    }
    
    $config = @"
terraform {
  required_providers {
    genesyscloud = {
      source = "genesys.com/mypurecloud/genesyscloud"
    }
  }
}

provider "genesyscloud" {}

resource "genesyscloud_tf_export" "test" {
  directory                    = "./export"
  include_state_file           = true
$filterLine
  export_format                = "$Format"
  split_files_by_resource      = true
  enable_dependency_resolution = $($EnableDeps.ToString().ToLower())
}
"@
    
    $config | Out-File -FilePath "$Dir\main.tf" -Encoding utf8
}

function Get-ExportStats {
    param([string]$Dir)
    
    $exportDir = "$Dir\export"
    if (-not (Test-Path $exportDir)) { return $null }
    
    $stateFile = "$exportDir\terraform.tfstate"
    if (-not (Test-Path $stateFile)) {
        return @{ FileCount = 0; ResourceCount = 0 }
    }
    
    try {
        $state = Get-Content $stateFile | ConvertFrom-Json
        return @{
            FileCount = (Get-ChildItem $exportDir -File).Count
            ResourceCount = $state.resources.Count
            ResourceTypes = $state.resources | Select-Object -ExpandProperty type | Sort-Object -Unique
        }
    } catch {
        return @{ FileCount = 0; ResourceCount = 0; Error = $_.Exception.Message }
    }
}

# Check prerequisites
Write-TestHeader "Prerequisites Check"

$prereqOk = $true
if (-not (Get-Command terraform -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: Terraform not found" -ForegroundColor Red
    $prereqOk = $false
}

foreach ($var in @("GENESYSCLOUD_OAUTHCLIENT_ID", "GENESYSCLOUD_OAUTHCLIENT_SECRET", "GENESYSCLOUD_REGION")) {
    if (-not (Test-Path "Env:$var")) {
        Write-Host "ERROR: Missing $var" -ForegroundColor Red
        $prereqOk = $false
    }
}

if (-not $prereqOk) { exit 1 }
Write-Host "All prerequisites met" -ForegroundColor Green

# TEST 1: Basic dependency resolution with ID filter
Write-TestHeader "Test 1: Dependency Resolution Enabled (By ID)"

New-TestConfig -Dir "test1" -FilterType "id" -FilterValue "genesyscloud_flow::$FlowId" -EnableDeps $true

Push-Location test1
$output1 = terraform apply -auto-approve 2>&1
$success1 = $LASTEXITCODE -eq 0
Pop-Location

$stats1 = Get-ExportStats -Dir "test1"
$passed1 = $success1 -and $stats1 -and $stats1.ResourceCount -gt 1

Write-TestResult -Name "Test 1" -Passed $passed1 -Details "$($stats1.ResourceCount) resources, $($stats1.FileCount) files"

# TEST 2: Dependency resolution disabled
Write-TestHeader "Test 2: Dependency Resolution Disabled (By ID)"

New-TestConfig -Dir "test2" -FilterType "id" -FilterValue "genesyscloud_flow::$FlowId" -EnableDeps $false

Push-Location test2
$output2 = terraform apply -auto-approve 2>&1
$success2 = $LASTEXITCODE -eq 0
Pop-Location

$stats2 = Get-ExportStats -Dir "test2"
$passed2 = $success2 -and $stats2 -and $stats2.ResourceCount -eq 1

Write-TestResult -Name "Test 2" -Passed $passed2 -Details "$($stats2.ResourceCount) resources (expected 1)"

# TEST 3: Comparison - Regex vs By-ID
Write-TestHeader "Test 3: Regex vs By-ID Comparison"

New-TestConfig -Dir "test3-regex" -FilterType "regex" -FilterValue "genesyscloud_flow::$FlowName" -EnableDeps $true
New-TestConfig -Dir "test3-byid" -FilterType "id" -FilterValue "genesyscloud_flow::$FlowId" -EnableDeps $true

Push-Location test3-regex
$outputRegex = terraform apply -auto-approve 2>&1
$successRegex = $LASTEXITCODE -eq 0
Pop-Location

Push-Location test3-byid
$outputById = terraform apply -auto-approve 2>&1
$successById = $LASTEXITCODE -eq 0
Pop-Location

$statsRegex = Get-ExportStats -Dir "test3-regex"
$statsById = Get-ExportStats -Dir "test3-byid"

$passed3 = $successRegex -and $successById -and 
           ($statsRegex.ResourceCount -eq $statsById.ResourceCount) -and
           ($statsRegex.FileCount -eq $statsById.FileCount)

$details3 = "Regex: $($statsRegex.ResourceCount) resources | By-ID: $($statsById.ResourceCount) resources"
Write-TestResult -Name "Test 3" -Passed $passed3 -Details $details3

# TEST 4: HCL format
Write-TestHeader "Test 4: HCL Export Format"

New-TestConfig -Dir "test4" -FilterType "id" -FilterValue "genesyscloud_flow::$FlowId" -EnableDeps $true -Format "hcl"

Push-Location test4
$output4 = terraform apply -auto-approve 2>&1
$success4 = $LASTEXITCODE -eq 0
Pop-Location

$hclFiles = Get-ChildItem "test4\export" -Filter "*.tf" -ErrorAction SilentlyContinue
$stats4 = Get-ExportStats -Dir "test4"
$passed4 = $success4 -and $hclFiles.Count -gt 0 -and $stats4.ResourceCount -gt 1

Write-TestResult -Name "Test 4" -Passed $passed4 -Details "$($hclFiles.Count) .tf files created"

# Summary
Write-TestHeader "Test Summary"

$passed = ($TestResults | Where-Object { $_.Passed }).Count
$total = $TestResults.Count
$failed = $total - $passed

Write-Host "`nTotal Tests: $total" -ForegroundColor White
Write-Host "Passed: $passed" -ForegroundColor Green
Write-Host "Failed: $failed" -ForegroundColor $(if ($failed -eq 0) { "Green" } else { "Red" })

Write-Host "`nResults:" -ForegroundColor Cyan
$TestResults | Format-Table -AutoSize

$reportFile = "test-report-$(Get-Date -Format 'yyyyMMdd-HHmmss').json"
$TestResults | ConvertTo-Json | Out-File $reportFile
Write-Host "`nReport saved: $reportFile" -ForegroundColor Gray

if ($failed -gt 0) {
    Write-Host "`nSome tests failed!" -ForegroundColor Yellow
    exit 1
} else {
    Write-Host "`nAll tests passed!" -ForegroundColor Green
    exit 0
}
