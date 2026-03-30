# Automated Dependency Resolution Tests

This directory contains automated tests for validating that `enable_dependency_resolution` works correctly with `include_filter_resources_by_id`.

## Test Suite

The test suite validates the following scenarios:

### ✓ Test 1: Basic Dependency Resolution (By ID)
Verifies that using `include_filter_resources_by_id` with `enable_dependency_resolution = true` successfully exports a flow and all its dependencies.

**Expected:** Multiple resources exported (flow + dependencies)

### ✓ Test 2: Dependency Resolution Disabled
Verifies that when `enable_dependency_resolution = false`, only the specified flow is exported without dependencies.

**Expected:** Only 1 resource (the flow itself)

### ✓ Test 3: Empty Filter List
Validates handling of an empty `include_filter_resources_by_id` array.

**Expected:** No resources exported or appropriate error

### ✓ Test 4: HCL Export Format
Tests that dependency resolution works with HCL export format (not just JSON).

**Expected:** .tf files created with multiple resources

### ✓ Test 5: Multiple Resources
Validates that multiple resource IDs can be specified and their dependencies are resolved.

**Expected:** All specified resources and their dependencies exported

### ✓ Test 6: All Features Combined
Tests that dependency resolution works correctly when combined with other export features:
- `export_computed`
- `log_permission_errors`
- `ignore_cyclic_deps`
- `split_files_by_resource`

**Expected:** Successful export with all features enabled

### ✓ Test 7: Regex vs By-ID Comparison
Compares exports using `include_filter_resources` (regex) vs `include_filter_resources_by_id` to ensure they produce identical results.

**Expected:** Same number of resources, files, and resource types

## Running the Tests

### Prerequisites

Ensure environment variables are set:

```powershell
$env:GENESYSCLOUD_OAUTHCLIENT_ID = "your-client-id"
$env:GENESYSCLOUD_OAUTHCLIENT_SECRET = "your-client-secret"
$env:GENESYSCLOUD_REGION = "us-west-2"
```

### Run All Tests

```powershell
cd automated-tests
.\Run-DependencyResolutionTests.ps1
```

### Run with Custom Flow

```powershell
.\Run-DependencyResolutionTests.ps1 -FlowId "your-flow-id" -FlowName "Your Flow Name"
```

## Test Output

The script will:
1. Run each test in sequence
2. Display real-time progress
3. Show pass/fail status for each test
4. Generate a summary report
5. Export detailed JSON report: `test-report-YYYYMMDD-HHMMSS.json`

### Sample Output

```
==================================================================
  Test 1: Basic Dependency Resolution (By ID)
==================================================================

[✓ PASS] Test 1
  Successfully exported flow with dependencies
  Details: 14 resources, 12 files

==================================================================
  Test 2: Dependency Resolution Disabled (By ID)
==================================================================

[✓ PASS] Test 2
  Dependency resolution correctly disabled
  Details: 1 resources (expected 1)

...

==================================================================
  Test Summary
==================================================================

Total Tests: 7
Passed: 7
Failed: 0
Success Rate: 100%
```

## Test Directories

Each test creates its own directory:
- `test1-basic-deps-by-id/`
- `test2-no-deps-by-id/`
- `test3-empty-filter/`
- `test4-hcl-format/`
- `test5-multiple-resources/`
- `test6-all-features/`
- `test7-comparison-regex/`
- `test7-comparison-byid/`

## Cleanup

After reviewing results, clean up test directories:

```powershell
Remove-Item test* -Recurse -Force
```

Or keep them for manual inspection of exported resources.

## Troubleshooting

### Test Failures

If tests fail, check:
1. Environment variables are set correctly
2. OAuth client has required permissions
3. Flow ID exists in your org
4. Provider binary is up to date (run `go build` in provider root)

### Permission Errors

If you see permission errors, ensure your OAuth client has:
- Flow read permissions
- Permissions to read all dependency resource types (queues, integrations, users, etc.)

### Timeout Issues

For complex flows with many dependencies, tests may take 30-60 seconds. This is normal.

## Adding New Tests

To add a new test case:

1. Create a new test section in the script
2. Use `New-TestDirectory` to create test folder
3. Generate Terraform config
4. Use `Invoke-TerraformExport` to run test
5. Use `Get-ExportStats` to analyze results
6. Use `Write-TestResult` to record outcome

Example:

```powershell
Write-TestHeader "Test X: Your Test Name"

$testXDir = New-TestDirectory "testX-description"
@"
terraform {
  required_providers {
    genesyscloud = {
      source = "genesys.com/mypurecloud/genesyscloud"
    }
  }
}

provider "genesyscloud" {}

resource "genesyscloud_tf_export" "test" {
  # Your test configuration
}
"@ | Out-File -FilePath (Join-Path $testXDir "main.tf") -Encoding utf8

$resultX = Invoke-TerraformExport -ConfigPath $testXDir -TestName "TestX"
$statsX = Get-ExportStats -ExportDir (Join-Path $testXDir "export")

# Validate and record result
Write-TestResult -TestName "Test X" -Passed $yourCondition -Message "Your message"
```

## CI/CD Integration

To use in CI/CD pipelines:

```yaml
- name: Run Dependency Resolution Tests
  run: |
    pwsh automated-tests/Run-DependencyResolutionTests.ps1
  env:
    GENESYSCLOUD_OAUTHCLIENT_ID: ${{ secrets.OAUTH_ID }}
    GENESYSCLOUD_OAUTHCLIENT_SECRET: ${{ secrets.OAUTH_SECRET }}
    GENESYSCLOUD_REGION: "us-west-2"
```

The script exits with code 0 on success, 1 on failure.
