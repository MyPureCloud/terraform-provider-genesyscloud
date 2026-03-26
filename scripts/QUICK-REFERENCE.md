# Linux Build - Quick Reference

## 🚀 Quick Fix for Existing Version

```powershell
.\scripts\Fix-LinuxPermissions.ps1 -Version "1.77.2"
```

This script does everything automatically!

---

## 📋 Manual Build Process

### Step 1: Build with WSL

```powershell
.\scripts\Build-LinuxProviderWSL.ps1 -Version "1.77.2"
```

**OR** directly in WSL:

```bash
cd /mnt/c/source/vscodebuild/terraform-provider-genesyscloud-custom/terraform-provider-genesyscloud
./scripts/Build-LinuxProvider.sh 1.77.2
```

### Step 2: Verify

```powershell
.\scripts\Verify-LinuxZip.ps1 -ZipPath "dist/terraform-provider-genesyscloud_1.77.2_linux_amd64.zip"
```

### Step 3: Upload to Azure

```powershell
$v = "1.77.2"  # Your version
$storageAccount = "YOUR_STORAGE_ACCOUNT"
$container = "YOUR_CONTAINER"

# Upload zip
az storage blob upload --account-name $storageAccount --container-name $container `
  --name "registry.terraform.io/mypurecloud/genesyscloud/terraform-provider-genesyscloud_${v}_linux_amd64.zip" `
  --file "terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud/terraform-provider-genesyscloud_${v}_linux_amd64.zip" `
  --overwrite

# Upload manifest
az storage blob upload --account-name $storageAccount --container-name $container `
  --name "registry.terraform.io/mypurecloud/genesyscloud/${v}.json" `
  --file "terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud/${v}.json" `
  --overwrite
```

### Step 4: Test

```bash
rm -rf .terraform/ .terraform.lock.hcl
terraform init
```

---

## ✅ Verification Checklist

| Check | Expected | Bad |
|-------|----------|-----|
| Zip file size | ~66MB ✅ | ~33MB ❌ |
| Verify script | Passes ✅ | Fails ❌ |
| Terraform init | Success ✅ | Checksum error ❌ |
| Provider runs | Works ✅ | "Permission denied" ❌ |

---

## 🔧 Available Scripts

| Script | Purpose | Usage |
|--------|---------|-------|
| `Fix-LinuxPermissions.ps1` | **Quick fix** for existing version | `.\scripts\Fix-LinuxPermissions.ps1 -Version "1.77.2"` |
| `Build-LinuxProviderWSL.ps1` | Build from Windows | `.\scripts\Build-LinuxProviderWSL.ps1 -Version "1.77.2"` |
| `Build-LinuxProvider.sh` | Build in WSL/Linux | `./scripts/Build-LinuxProvider.sh 1.77.2` |
| `Verify-LinuxZip.ps1` | Verify permissions | `.\scripts\Verify-LinuxZip.ps1 -ZipPath "file.zip"` |
| `Build-SinglePlatform.ps1` | ⚠️ For Windows/Mac only | Don't use for Linux! |

---

## 📚 Documentation

- **[LINUX-PERMISSIONS-FIX.md](../LINUX-PERMISSIONS-FIX.md)** - Complete summary
- **[README-LINUX-BUILD.md](./README-LINUX-BUILD.md)** - Detailed guide
- **[README.md](./README.md)** - All build scripts

---

## ❓ Common Issues

### "WSL not found"
```powershell
wsl --install
# Restart computer
```

### Zip is only 33MB
Rebuild with WSL scripts. Windows builds don't preserve permissions.

### "Permission denied" on Linux
Provider doesn't have executable bit. Rebuild with WSL.

### "Checksum mismatch"
Manifest not uploaded or cache issue. Re-upload manifest, clear cache.

---

## 🎯 Remember

- ✅ **DO** use `Build-LinuxProviderWSL.ps1` for Linux
- ❌ **DON'T** use `Build-SinglePlatform.ps1` for Linux
- ✅ Verify file is ~66MB
- ✅ Run verify script before uploading

---

**Need help?** See [LINUX-PERMISSIONS-FIX.md](../LINUX-PERMISSIONS-FIX.md) for troubleshooting.
