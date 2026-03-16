# Terraform Provider Network Mirror Setup

Make your custom provider installable via `terraform init` just like the official provider!

## How It Works

Instead of manually copying binaries, you host them on a web server. Terraform downloads them automatically during `terraform init` based on your `.terraformrc` configuration.

## Quick Setup

### Step 1: Build Your Provider

```powershell
# Build for your target platform(s)
.\scripts\Build-SinglePlatform.ps1 -Platform "linux_amd64" -Version "1.0.0-custom"
```

### Step 2: Create Mirror Structure

```powershell
# Create the mirror directory structure
.\scripts\Setup-ProviderMirror.ps1 `
    -OutputDir "./provider-mirror" `
    -ProviderVersion "1.0.0-custom" `
    -BinaryPath "./dist" `
    -Platforms "linux_amd64"
```

This creates:
```
provider-mirror/
  genesys.com/
    mypurecloud/
      genesyscloud/
        versions.json
        1.0.0-custom/
          linux_amd64/
            terraform-provider-genesyscloud_v1.0.0-custom
            terraform-provider-genesyscloud_1.0.0-custom_SHA256SUMS
          linux_amd64.json
```

### Step 3: Deploy to Azure Blob Storage

```powershell
# Deploy to Azure
.\scripts\Deploy-ToAzureBlob.ps1 `
    -MirrorPath "./provider-mirror" `
    -StorageAccount "myterraformstorage" `
    -Container "terraform-providers" `
    -ResourceGroup "my-resource-group"
```

Or manually upload the `provider-mirror` directory contents to any web server.

### Step 4: Update URLs in JSON Files

Edit the JSON files to point to your actual hosting location:

**provider-mirror/genesys.com/mypurecloud/genesyscloud/1.0.0-custom/linux_amd64.json:**
```json
{
  "protocols": ["5.0"],
  "os": "linux",
  "arch": "amd64",
  "filename": "terraform-provider-genesyscloud_v1.0.0-custom",
  "download_url": "https://myterraformstorage.blob.core.windows.net/terraform-providers/genesys.com/mypurecloud/genesyscloud/1.0.0-custom/linux_amd64/terraform-provider-genesyscloud_v1.0.0-custom",
  "shasums_url": "https://myterraformstorage.blob.core.windows.net/terraform-providers/genesys.com/mypurecloud/genesyscloud/1.0.0-custom/linux_amd64/terraform-provider-genesyscloud_1.0.0-custom_SHA256SUMS",
  "shasum": "abc123..."
}
```

Re-upload after editing URLs.

### Step 5: Configure Terraform

Create or edit `~/.terraformrc` (Linux/Mac) or `%APPDATA%\terraform.rc` (Windows):

```hcl
provider_installation {
  network_mirror {
    url = "https://myterraformstorage.blob.core.windows.net/terraform-providers/"
  }
}
```

### Step 6: Use Your Provider

In your Terraform configuration:

```hcl
terraform {
  required_providers {
    genesyscloud = {
      source  = "genesys.com/mypurecloud/genesyscloud"
      version = "1.0.0-custom"
    }
  }
}

provider "genesyscloud" {
  oauthclient_id     = var.oauth_id
  oauthclient_secret = var.oauth_secret
  aws_region         = "us-west-2"
}
```

Then just run:
```bash
terraform init
```

Terraform will automatically download your custom provider from Azure!

## Testing Locally

Before deploying to Azure, test locally:

```powershell
# Start local HTTP server
.\scripts\Serve-ProviderMirror.ps1 -MirrorPath "./provider-mirror" -Port 8080
```

Update your `.terraformrc`:
```hcl
provider_installation {
  network_mirror {
    url = "http://localhost:8080/"
  }
}
```

Run `terraform init` in your test directory.

## Azure Deployment Options

### Option 1: Azure Blob Storage (Simple)

Best for small teams. Files served directly from blob storage.

```powershell
.\scripts\Deploy-ToAzureBlob.ps1 `
    -MirrorPath "./provider-mirror" `
    -StorageAccount "myaccount" `
    -Container "terraform-providers"
```

URL: `https://myaccount.blob.core.windows.net/terraform-providers/`

### Option 2: Azure Static Website (Better)

Enables custom domains and better performance.

```powershell
# Enable static website on storage account
az storage blob service-properties update `
    --account-name myaccount `
    --static-website `
    --index-document index.html

# Upload files
.\scripts\Deploy-ToAzureBlob.ps1 `
    -MirrorPath "./provider-mirror" `
    -StorageAccount "myaccount" `
    -ResourceGroup "my-rg"
```

URL: `https://myaccount.z13.web.core.windows.net/`

### Option 3: Azure CDN (Production)

Adds CDN for global distribution and caching.

```bash
# Create CDN profile
az cdn profile create \
    --name terraform-cdn \
    --resource-group my-rg \
    --sku Standard_Microsoft

# Create CDN endpoint
az cdn endpoint create \
    --name terraform-providers \
    --profile-name terraform-cdn \
    --resource-group my-rg \
    --origin myaccount.blob.core.windows.net \
    --origin-host-header myaccount.blob.core.windows.net
```

URL: `https://terraform-providers.azureedge.net/`

## Container Usage

In your Dockerfile, configure `.terraformrc` before running Terraform:

```dockerfile
FROM hashicorp/terraform:latest

# Configure provider mirror
RUN mkdir -p /root/.terraform.d && \
    echo 'provider_installation {' > /root/.terraformrc && \
    echo '  network_mirror {' >> /root/.terraformrc && \
    echo '    url = "https://myaccount.blob.core.windows.net/terraform-providers/"' >> /root/.terraformrc && \
    echo '  }' >> /root/.terraformrc && \
    echo '}' >> /root/.terraformrc

WORKDIR /workspace
ENTRYPOINT ["terraform"]
```

Now `terraform init` automatically downloads from your mirror!

## CI/CD Pipeline

### Azure DevOps

```yaml
trigger:
  - main

pool:
  vmImage: 'ubuntu-latest'

variables:
  providerVersion: '1.0.0-custom'
  storageAccount: 'myterraformstorage'

stages:
- stage: BuildAndDeploy
  jobs:
  - job: BuildProvider
    steps:
    - task: Go@0
      displayName: 'Build Provider'
      inputs:
        command: 'build'
        arguments: '-o $(Build.ArtifactStagingDirectory)/terraform-provider-genesyscloud'
      env:
        GOOS: linux
        GOARCH: amd64
        CGO_ENABLED: 0
    
    - task: PowerShell@2
      displayName: 'Create Mirror Structure'
      inputs:
        filePath: 'scripts/Setup-ProviderMirror.ps1'
        arguments: >
          -OutputDir "$(Build.ArtifactStagingDirectory)/mirror"
          -ProviderVersion "$(providerVersion)"
          -BinaryPath "$(Build.ArtifactStagingDirectory)"
          -Platforms "linux_amd64"
    
    - task: AzureCLI@2
      displayName: 'Deploy to Azure Blob'
      inputs:
        azureSubscription: 'MyServiceConnection'
        scriptType: 'pscore'
        scriptLocation: 'scriptPath'
        scriptPath: 'scripts/Deploy-ToAzureBlob.ps1'
        arguments: >
          -MirrorPath "$(Build.ArtifactStagingDirectory)/mirror"
          -StorageAccount "$(storageAccount)"
          -Container "terraform-providers"
```

### GitHub Actions

```yaml
name: Deploy Provider Mirror

on:
  push:
    tags:
      - 'v*'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Build Provider
      run: |
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
        go build -o dist/terraform-provider-genesyscloud
    
    - name: Setup PowerShell
      uses: actions/setup-pwsh@v1
    
    - name: Create Mirror
      run: |
        ./scripts/Setup-ProviderMirror.ps1 `
          -OutputDir "./mirror" `
          -ProviderVersion "${GITHUB_REF#refs/tags/v}" `
          -BinaryPath "./dist" `
          -Platforms "linux_amd64"
      shell: pwsh
    
    - uses: azure/login@v1
      with:
        creds: ${{ secrets.AZURE_CREDENTIALS }}
    
    - name: Deploy to Azure
      run: |
        ./scripts/Deploy-ToAzureBlob.ps1 `
          -MirrorPath "./mirror" `
          -StorageAccount "${{ secrets.STORAGE_ACCOUNT }}" `
          -Container "terraform-providers"
      shell: pwsh
```

## Troubleshooting

### Provider Not Found

```
Error: Failed to query available provider packages
```

**Solutions:**
1. Check `.terraformrc` URL is correct and accessible
2. Verify `versions.json` exists at the correct path
3. Test URL in browser: `https://your-url/genesys.com/mypurecloud/genesyscloud/versions.json`

### Checksum Mismatch

```
Error: Failed to verify provider package checksums
```

**Solution:** Re-run `Setup-ProviderMirror.ps1` to regenerate correct checksums.

### CORS Errors (Browser)

If using static website hosting:

```bash
az storage cors add \
    --services b \
    --methods GET HEAD \
    --origins '*' \
    --allowed-headers '*' \
    --account-name myaccount
```

### Wrong Platform Downloaded

Terraform auto-detects your OS/arch. To override:

```bash
# Force specific platform
export TF_PROVIDER_ARCH=amd64
export TF_PROVIDER_OS=linux
terraform init
```

## Directory Structure Reference

```
provider-mirror/
  genesys.com/                    # Hostname from namespace
    mypurecloud/                  # Org name from namespace
      genesyscloud/               # Provider name
        versions.json             # Lists all available versions
        1.0.0-custom/             # Version directory
          linux_amd64/            # Platform directory
            terraform-provider-genesyscloud_v1.0.0-custom
            terraform-provider-genesyscloud_1.0.0-custom_SHA256SUMS
          linux_amd64.json        # Download metadata
          windows_amd64/
            terraform-provider-genesyscloud_v1.0.0-custom.exe
            terraform-provider-genesyscloud_1.0.0-custom_SHA256SUMS
          windows_amd64.json
```

## Network Mirror vs. Filesystem Mirror

**Network Mirror** (Recommended):
- ✅ Works in containers
- ✅ Team-wide access
- ✅ Version control
- ❌ Requires hosting

**Filesystem Mirror** (Local only):
```hcl
provider_installation {
  filesystem_mirror {
    path = "/usr/share/terraform/providers"
  }
}
```
- ✅ No hosting needed
- ❌ Local machine only
- ❌ Manual distribution

## Comparison with Official Registry

| Feature | Official Registry | Network Mirror |
|---------|------------------|----------------|
| Discovery | ✅ terraform.io | ❌ Manual config |
| Version resolution | ✅ Automatic | ✅ Automatic |
| Team distribution | ✅ Automatic | ✅ Via URL |
| Private providers | ❌ Paid | ✅ Free |
| Custom modifications | ❌ No | ✅ Yes |

## Cost Estimates (Azure)

For a small team (~10 developers, 100 init/day):

- **Storage:** ~$0.05/month (1 GB)
- **Bandwidth:** ~$1/month (10 GB)
- **Total:** ~$1-2/month

Much cheaper than Terraform Cloud private registry ($20/month minimum).

---

That's it! Your custom provider is now installable via `terraform init` just like the official one! 🚀
