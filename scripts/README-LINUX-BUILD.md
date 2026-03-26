# Building Linux Provider with Executable Permissions

## The Problem

When building the Terraform provider for Linux on Windows and creating a zip file, the Unix executable permissions are **not preserved**. This causes "Permission denied" errors when Terraform tries to run the provider on Linux systems.

### Technical Details

- **Windows zip utilities**: Don't understand or preserve Unix file permissions
- **Result**: Zip file is ~33MB without executable bit
- **Correct result**: Zip file is ~66MB with executable bit preserved
- **Impact**: Provider fails to execute on Linux

## The Solution

Build and package the Linux binary using **WSL (Windows Subsystem for Linux)** which properly handles Unix file permissions.

## Quick Start

### Prerequisites

1. **WSL** installed on Windows
   ```powershell
   wsl --install
   ```

2. **Go** installed in WSL
   ```bash
   # In WSL
   wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **zip** utility in WSL
   ```bash
   sudo apt update && sudo apt install zip -y
   ```

### Build Steps

#### Option 1: Using the Build Script (Recommended)

```bash
# Navigate to repo in WSL
cd /mnt/c/source/vscodebuild/terraform-provider-genesyscloud-custom/terraform-provider-genesyscloud

# Make script executable (first time only)
chmod +x scripts/Build-LinuxProvider.sh

# Build with version
./scripts/Build-LinuxProvider.sh 1.77.2

# Or use default version  
./scripts/Build-LinuxProvider.sh
```

The script will:
- ✅ Build the Linux binary
- ✅ Set `chmod +x` on the binary
- ✅ Create zip with proper permissions
- ✅ Calculate SHA256 and h1 hashes
- ✅ Update manifest files automatically

#### Option 2: Manual Build

```bash
# Navigate to repo in WSL
cd /mnt/c/source/vscodebuild/terraform-provider-genesyscloud-custom/terraform-provider-genesyscloud

# Build for Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o terraform-provider-genesyscloud -ldflags "-s -w" .

# Set executable permission (CRITICAL!)
chmod +x terraform-provider-genesyscloud

# Create zip (preserves permissions)
zip terraform-provider-genesyscloud_1.77.2_linux_amd64.zip terraform-provider-genesyscloud

# Calculate checksums
sha256sum terraform-provider-genesyscloud_1.77.2_linux_amd64.zip
```

### Verify Permissions

```bash
# Check the binary has executable bit
ls -la terraform-provider-genesyscloud

# Should show: -rwxr-xr-x (the 'x' indicates executable)
# ❌ BAD: -rw-r--r-- (no executable permission)
# ✅ GOOD: -rwxr-xr-x (executable permission set)

# Check zip size (should be ~66MB, not ~33MB)
ls -lh terraform-provider-genesyscloud_1.77.2_linux_amd64.zip
```

## Updating Manifest Files

After building, update the manifest with the new checksum:

### 1. Get the h1 hash

The build script outputs this automatically, or calculate manually:

```bash
# Get SHA256
SHA256=$(sha256sum terraform-provider-genesyscloud_1.77.2_linux_amd64.zip | awk '{print $1}')

# Convert to h1 hash (Terraform format)
echo -n "$SHA256" | xxd -r -p | base64 | awk '{print "h1:"$1}'
```

### 2. Update manifest file

Edit: `terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud/1.77.2.json`

```json
{
  "archives": {
    "linux_amd64": {
      "hashes": [
        "h1:NEW_HASH_HERE"
      ],
      "url": "terraform-provider-genesyscloud_1.77.2_linux_amd64.zip"
    }
  }
}
```

### 3. Copy to mirror directory

```bash
cp terraform-provider-genesyscloud_1.77.2_linux_amd64.zip \
   terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud/
```

## Complete Workflow

Here's the full process from build to deployment:

```bash
# 1. Build in WSL with proper permissions
cd /mnt/c/source/vscodebuild/terraform-provider-genesyscloud-custom/terraform-provider-genesyscloud
./scripts/Build-LinuxProvider.sh 1.77.2

# The script automatically:
#   - Builds the binary
#   - Sets executable permissions
#   - Creates zip file
#   - Updates manifest with new checksums
#   - Copies to mirror directory

# 2. Upload to Azure (if using Azure Blob Storage)
# From PowerShell/Windows:
$version = "1.77.2"
$file = "terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud/terraform-provider-genesyscloud_${version}_linux_amd64.zip"

az storage blob upload `
  --account-name <your-storage-account> `
  --container-name <your-container> `
  --name "registry.terraform.io/mypurecloud/genesyscloud/terraform-provider-genesyscloud_${version}_linux_amd64.zip" `
  --file $file `
  --overwrite

# 3. Upload manifest
$manifestFile = "terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud/${version}.json"

az storage blob upload `
  --account-name <your-storage-account> `
  --container-name <your-container> `
  --name "registry.terraform.io/mypurecloud/genesyscloud/${version}.json" `
  --file $manifestFile `
  --overwrite

# 4. Test
terraform init
```

## Troubleshooting

### Permission Denied Error

```
Error: Failed to install provider
Provider "genesys.com/mypurecloud/genesyscloud" v1.77.2 is not available
Failed to execute: fork/exec: permission denied
```

**Solution**: Rebuild using WSL with `chmod +x`

### Wrong Zip Size

- ❌ **~33MB**: No executable permissions (built on Windows)
- ✅ **~66MB**: Executable permissions preserved (built in WSL)

### Checksum Mismatch

```
Error: Failed to install provider
Provider "genesys.com/mypurecloud/genesyscloud" v1.77.2 has invalid checksum
```

**Solution**: Update manifest file with correct h1 hash from new build

## Why This Happens

### Windows File System (NTFS)

- Doesn't have Unix-style executable permissions
- Windows zip tools don't understand or preserve Unix permissions
- The executable bit (`chmod +x`) is not stored in the zip

### Linux File System (ext4, etc.)

- Uses Unix permissions: read (r), write (w), execute (x)
- Files must have `+x` permission to be executed
- Terraform requires the provider binary to be executable

### The Fix

Using WSL provides a Linux environment on Windows that:
- ✅ Understands Unix file permissions
- ✅ Properly sets `chmod +x` on files
- ✅ Preserves permissions when creating zip files
- ✅ Produces zips that work correctly on Linux

## References

- [Terraform Provider Installation](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-installation)
- [Unix File Permissions](https://en.wikipedia.org/wiki/File-system_permissions#Traditional_Unix_permissions)
- [WSL Documentation](https://learn.microsoft.com/en-us/windows/wsl/)
