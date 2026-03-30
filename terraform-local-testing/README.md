# Manual Test for Dependency Resolution Comparison

This directory contains test configurations to manually verify that dependency resolution produces the same results for both filter types.

## Setup

1. Ensure you have the `.terraformrc` file configured (see main README)
2. Make sure environment variables are set:
```powershell
$env:GENESYSCLOUD_OAUTHCLIENT_ID = "your-client-id"
$env:GENESYSCLOUD_OAUTHCLIENT_SECRET = "your-client-secret"
$env:GENESYSCLOUD_REGION = "us-west-2"
```

## Test 1: Export by Regex

```powershell
cd test-regex
terraform init
terraform apply -auto-approve
```

This will export to `./export-regex/`

## Test 2: Export by ID

```powershell
cd ..\test-by-id
terraform init
terraform apply -auto-approve
```

This will export to `./export-by-id/`

## Compare Results

```powershell
# Compare file counts
(Get-ChildItem ./test-regex/export-regex -File).Count
(Get-ChildItem ./test-by-id/export-by-id -File).Count

# Compare resource counts in state files
$regexState = Get-Content ./test-regex/export-regex/terraform.tfstate | ConvertFrom-Json
$byIdState = Get-Content ./test-by-id/export-by-id/terraform.tfstate | ConvertFrom-Json

Write-Host "Regex export resources: $($regexState.resources.Count)"
Write-Host "By-ID export resources: $($byIdState.resources.Count)"

# List resource types
$regexState.resources | Select-Object -ExpandProperty type | Sort-Object -Unique
$byIdState.resources | Select-Object -ExpandProperty type | Sort-Object -Unique
```

## Expected Result

Both exports should contain:
- The same number of files
- The same resource types
- The same resource IDs (dependencies)
- The main flow: `b84cbae3-7c54-45dc-ade0-7a30fbccf996`

## Cleanup

```powershell
terraform destroy -auto-approve
cd ..
Remove-Item -Recurse -Force ./test-regex, ./test-by-id
```
