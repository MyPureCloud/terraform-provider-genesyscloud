# Build Scripts

This directory contains scripts for building the Terraform provider binary.

## Quick Reference

### Building for Linux (Recommended for Production)

**⚠️ Important: Linux binaries MUST be built using WSL to have proper executable permissions!**

```powershell
# Windows - Use WSL wrapper
.\scripts\Build-LinuxProviderWSL.ps1 -Version "1.77.2"
```

```bash
# Or directly in WSL
./scripts/Build-LinuxProvider.sh 1.77.2
```

### Building for Other Platforms

```powershell
# Windows - Local development only
.\scripts\Build-SinglePlatform.ps1 -Platform "windows_amd64" -Version "1.0.0-custom"

# macOS/Darwin
.\scripts\Build-SinglePlatform.ps1 -Platform "darwin_amd64" -Version "1.0.0-custom"
```

### Verifying Linux Builds

```powershell
# Verify a zip file has proper permissions
.\scripts\Verify-LinuxZip.ps1 -ZipPath "dist/terraform-provider-genesyscloud_1.77.2_linux_amd64.zip"
```

## Scripts Overview

### Build-LinuxProviderWSL.ps1

**Purpose**: Build Linux provider with proper executable permissions (Windows host)

**Usage**:
```powershell
.\scripts\Build-LinuxProviderWSL.ps1 -Version "1.77.2" [-OutputDir "./dist"]
```

**Requirements**:
- Windows with WSL installed
- Go installed in WSL
- zip utility in WSL

**What it does**:
1. Checks WSL is available
2. Invokes Build-LinuxProvider.sh in WSL
3. Ensures executable permissions are set
4. Creates zip with preserved permissions
5. Updates manifest files automatically

### Build-LinuxProvider.sh

**Purpose**: Build script that runs in Linux/WSL

**Usage**:
```bash
./scripts/Build-LinuxProvider.sh [VERSION] [OUTPUT_DIR]

# Examples
./scripts/Build-LinuxProvider.sh
./scripts/Build-LinuxProvider.sh 1.77.2
./scripts/Build-LinuxProvider.sh 1.77.2 ./custom-output
```

**What it does**:
1. Builds Linux AMD64 binary
2. Sets `chmod +x` on binary
3. Creates zip while preserving permissions
4. Calculates SHA256 and h1 hashes
5. Updates mirror directory and manifest files
6. Outputs all paths and checksums

### Build-SinglePlatform.ps1

**Purpose**: Build for any platform (cross-compilation)

**⚠️ Warning**: Do NOT use for production Linux builds! Use Build-LinuxProviderWSL.ps1 instead.

**Usage**:
```powershell
.\scripts\Build-SinglePlatform.ps1 -Platform "PLATFORM" -Version "VERSION" [-OutputDir "./dist"]

# Examples
.\scripts\Build-SinglePlatform.ps1 -Platform "windows_amd64" -Version "1.77.2"
.\scripts\Build-SinglePlatform.ps1 -Platform "darwin_amd64" -Version "1.77.2"
.\scripts\Build-SinglePlatform.ps1 -Platform "linux_arm64" -Version "1.77.2"
```

**Supported Platforms**:
- `windows_amd64`
- `darwin_amd64` (macOS Intel)
- `darwin_arm64` (macOS Apple Silicon)
- `linux_amd64`
- `linux_arm64`

**Limitations**:
- Linux builds from Windows lack executable permissions
- Will display warnings for Linux builds
- Better for Windows/macOS builds

### Verify-LinuxZip.ps1

**Purpose**: Verify that a zip file has proper Unix executable permissions

**Usage**:
```powershell
.\scripts\Verify-LinuxZip.ps1 -ZipPath "path/to/file.zip"

# Example
.\scripts\Verify-LinuxZip.ps1 -ZipPath "dist/terraform-provider-genesyscloud_1.77.2_linux_amd64.zip"
```

**What it checks**:
- File size (should be ~66MB with permissions, ~33MB without)
- Extracts zip in WSL
- Checks Unix file permissions
- Verifies executable bit is set

**Exit codes**:
- `0`: Verification passed (has executable permissions)
- `1`: Verification failed (missing executable permissions)

## The Linux Permissions Issue

### The Problem

When you build a Linux binary on Windows and create a zip file using Windows tools:
- ❌ **The Unix executable permission bit is NOT preserved**
- ❌ File size is ~33MB (compressed without permission metadata)
- ❌ Linux systems cannot execute the binary → "Permission denied"

When you build in WSL/Linux with proper `chmod +x`:
- ✅ **The executable permission IS preserved in the zip**
- ✅ File size is ~66MB (includes permission metadata)
- ✅ Linux systems can execute the binary

### Why This Happens

1. **Windows file system (NTFS)** doesn't have Unix-style permissions
2. **Windows zip tools** don't understand or store Unix permission bits
3. **Linux requires** the executable bit (`+x`) to run files
4. **Terraform** requires the provider binary to be executable

### The Solution

**Use WSL** - Windows Subsystem for Linux provides:
- Real Linux file system with Unix permissions
- `chmod` command that actually sets the executable bit
- Linux `zip` utility that preserves permissions in the archive

## Complete Workflow

### For Version 1.77.2

```powershell
# 1. Build in WSL with proper permissions
.\scripts\Build-LinuxProviderWSL.ps1 -Version "1.77.2"

# 2. Verify the build
.\scripts\Verify-LinuxZip.ps1 -ZipPath "dist/terraform-provider-genesyscloud_1.77.2_linux_amd64.zip"

# 3. Upload to Azure (if using Azure Blob Storage)
$version = "1.77.2"
az storage blob upload `
  --account-name <storage-account> `
  --container-name <container> `
  --name "registry.terraform.io/mypurecloud/genesyscloud/terraform-provider-genesyscloud_${version}_linux_amd64.zip" `
  --file "terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud/terraform-provider-genesyscloud_${version}_linux_amd64.zip" `
  --overwrite

# 4. Upload manifest
az storage blob upload `
  --account-name <storage-account> `
  --container-name <container> `
  --name "registry.terraform.io/mypurecloud/genesyscloud/${version}.json" `
  --file "terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud/${version}.json" `
  --overwrite

# 5. Test
terraform init
```

## Troubleshooting

### "WSL not found" Error

**Solution**: Install WSL
```powershell
wsl --install
# Restart computer after installation
```

### "Permission denied" when running provider on Linux

**Cause**: Binary was built without executable permissions

**Solution**: Rebuild using Build-LinuxProviderWSL.ps1

### Zip file is only ~33MB

**Cause**: Built on Windows without WSL

**Solution**: Rebuild using Build-LinuxProviderWSL.ps1

**Expected**: ~66MB (includes permission metadata)

### "Checksum mismatch" in Terraform

**Cause**: Manifest file has old checksum

**Solution**: The build scripts automatically update the manifest. If you see this:
1. Check the manifest file was uploaded
2. Verify the h1 hash matches the build output
3. Clear Terraform cache: `rm -rf .terraform/`

## Additional Documentation

- [README-LINUX-BUILD.md](./README-LINUX-BUILD.md) - Detailed Linux build documentation
- [NETWORK-MIRROR-SETUP.md](./NETWORK-MIRROR-SETUP.md) - Provider mirror setup guide

## Provider Mirror Structure

After building, the provider mirror structure looks like:

```
terraform-local-testing/
  terraform-provider-mirror/
    registry.terraform.io/
      mypurecloud/
        genesyscloud/
          index.json                                              # Version index
          1.77.2.json                                             # Manifest with checksums
          terraform-provider-genesyscloud_1.77.2_linux_amd64.zip  # Provider binary
```

**Key Files**:

- `index.json`: Lists available versions
- `{version}.json`: Contains checksums (h1 hashes) for validation
- `*.zip`: The actual provider binary (must have executable permissions)

## Scripts for Advanced Usage

### Setup-ProviderMirror.ps1

Creates network mirror structure for multiple platforms

### Serve-ProviderMirror.ps1

Local HTTP server for testing provider installation

### generate_env.py

Generates environment configuration files

## References

- [Terraform Provider Installation](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-installation)
- [Network Mirror Protocol](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol)
- [WSL Documentation](https://learn.microsoft.com/en-us/windows/wsl/)
