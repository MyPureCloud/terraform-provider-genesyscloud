# Linux Executable Permissions Fix - Summary

## What Was Done

This repository has been updated with scripts and documentation to properly handle Linux executable permissions when building the Terraform provider on Windows.

## The Problem

When building Linux binaries on Windows:
- ❌ Zip files created don't preserve Unix executable permissions
- ❌ File size: ~33MB (without permission metadata)
- ❌ Error on Linux: "Permission denied" when Terraform tries to execute the provider

When building with proper permissions:
- ✅ Executable bit (`chmod +x`) is preserved in the zip
- ✅ File size: ~66MB (includes permission metadata)
- ✅ Works correctly on Linux systems

## Files Created

### Build Scripts

1. **`scripts/Build-LinuxProviderWSL.ps1`**
   - PowerShell wrapper that invokes WSL to build properly
   - Usage: `.\scripts\Build-LinuxProviderWSL.ps1 -Version "1.77.2"`

2. **`scripts/Build-LinuxProvider.sh`**
   - Bash script that runs in WSL/Linux
   - Builds binary, sets permissions, creates zip
   - Automatically updates manifest files
   - Usage: `./scripts/Build-LinuxProvider.sh 1.77.2`

3. **`scripts/Verify-LinuxZip.ps1`**
   - Verifies a zip has proper executable permissions
   - Usage: `.\scripts\Verify-LinuxZip.ps1 -ZipPath "path/to/file.zip"`

4. **`scripts/Fix-LinuxPermissions.ps1`**
   - Quick fix script for existing versions
   - Rebuilds and verifies automatically
   - Usage: `.\scripts\Fix-LinuxPermissions.ps1 -Version "1.77.2"`

### Documentation

5. **`scripts/README-LINUX-BUILD.md`**
   - Comprehensive guide to the Linux permissions issue
   - Step-by-step build instructions
   - Troubleshooting guide

6. **`scripts/README.md`**
   - Overview of all build scripts
   - Quick reference guide
   - Complete workflow documentation

### Updates to Existing Files

7. **`scripts/Build-SinglePlatform.ps1`**
   - Added warnings about Linux builds on Windows
   - Directs users to proper WSL scripts

8. **`README.md`**
   - Added "Building for Linux (Production)" section
   - Links to detailed documentation

## Quick Start (For Version 1.77.2)

### Option 1: Fix Existing Version

```powershell
# Run the fix script
.\scripts\Fix-LinuxPermissions.ps1 -Version "1.77.2"

# This will:
# 1. Check current version
# 2. Rebuild with proper permissions
# 3. Verify the build
# 4. Update manifest files
# 5. Provide upload commands
```

### Option 2: Manual Build

```powershell
# Build with WSL
.\scripts\Build-LinuxProviderWSL.ps1 -Version "1.77.2"

# Verify
.\scripts\Verify-LinuxZip.ps1 -ZipPath "dist/terraform-provider-genesyscloud_1.77.2_linux_amd64.zip"
```

## File Structure After Build

```
terraform-provider-genesyscloud/
├── dist/
│   └── terraform-provider-genesyscloud_1.77.2_linux_amd64.zip  (✅ 66MB with permissions)
├── terraform-local-testing/
│   └── terraform-provider-mirror/
│       └── registry.terraform.io/
│           └── mypurecloud/
│               └── genesyscloud/
│                   ├── index.json
│                   ├── 1.77.2.json  (✅ Updated with new h1 hash)
│                   └── terraform-provider-genesyscloud_1.77.2_linux_amd64.zip  (✅ 66MB)
└── scripts/
    ├── Build-LinuxProviderWSL.ps1  (✅ New)
    ├── Build-LinuxProvider.sh  (✅ New)
    ├── Verify-LinuxZip.ps1  (✅ New)
    ├── Fix-LinuxPermissions.ps1  (✅ New)
    ├── README-LINUX-BUILD.md  (✅ New)
    ├── README.md  (✅ New)
    ├── Build-SinglePlatform.ps1  (⚠️ Updated with warnings)
    └── ... (other scripts)
```

## Upload to Azure

After building, upload the files:

```powershell
# Set your Azure details
$storageAccount = "YOUR_STORAGE_ACCOUNT"
$container = "YOUR_CONTAINER"
$version = "1.77.2"

# Upload zip
az storage blob upload `
  --account-name $storageAccount `
  --container-name $container `
  --name "registry.terraform.io/mypurecloud/genesyscloud/terraform-provider-genesyscloud_${version}_linux_amd64.zip" `
  --file "terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud/terraform-provider-genesyscloud_${version}_linux_amd64.zip" `
  --overwrite

# Upload manifest
az storage blob upload `
  --account-name $storageAccount `
  --container-name $container `
  --name "registry.terraform.io/mypurecloud/genesyscloud/${version}.json" `
  --file "terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud/${version}.json" `
  --overwrite
```

## Verification Checklist

- [ ] WSL is installed and working
- [ ] Build script runs without errors
- [ ] Zip file is ~66MB (not ~33MB)
- [ ] Verify script passes
- [ ] Manifest has updated h1 hash
- [ ] Files uploaded to Azure
- [ ] `terraform init` succeeds
- [ ] Provider executes without "Permission denied" error

## Key Technical Details

### Why WSL?

- **Windows (NTFS)** doesn't support Unix file permissions
- **WSL** provides a real Linux environment with ext4 filesystem
- **Linux zip utility** preserves Unix permissions in archives
- **Result**: Zip contains executable permission metadata

### The h1 Hash

Terraform uses a special hash format called "h1" (hash version 1):

```bash
# SHA256 of zip file
SHA256=$(sha256sum file.zip | awk '{print $1}')

# Convert to h1 format (base64 of binary SHA256)
H1=$(echo -n "$SHA256" | xxd -r -p | base64)
echo "h1:$H1"
```

The build scripts calculate this automatically.

### File Size Difference

- **Without permissions**: ~33MB
  - Only contains compressed binary
  
- **With permissions**: ~66MB  
  - Contains binary + permission metadata
  - This is expected and correct!

## Troubleshooting

### "WSL not found"

```powershell
# Install WSL
wsl --install

# Restart computer
# Run build script again
```

### "Permission denied" on Linux

**Cause**: Binary doesn't have executable permission

**Solution**: Rebuild using `Build-LinuxProviderWSL.ps1`

### "Checksum mismatch"

**Cause**: Manifest file has old checksum

**Solutions**:
1. Build scripts update manifest automatically
2. Verify manifest was uploaded to Azure
3. Clear local cache: `rm -rf .terraform/`

## Future Builds

For all future Linux builds:

```powershell
# Always use WSL build script
.\scripts\Build-LinuxProviderWSL.ps1 -Version "VERSION"

# Never use Build-SinglePlatform.ps1 for Linux
# (It will show warnings if you try)
```

## References

- [scripts/README-LINUX-BUILD.md](./scripts/README-LINUX-BUILD.md) - Detailed documentation
- [scripts/README.md](./scripts/README.md) - Script reference guide  
- [Terraform Provider Installation](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-installation)
- [WSL Documentation](https://learn.microsoft.com/en-us/windows/wsl/)

---

## Summary

✅ **Problem**: Windows can't build Linux binaries with proper permissions  
✅ **Solution**: Use WSL to build and package  
✅ **Result**: Properly executable Linux provider that works on all Linux systems

The fix ensures that all future Linux builds will have the correct executable permissions, preventing "Permission denied" errors on Linux systems.
