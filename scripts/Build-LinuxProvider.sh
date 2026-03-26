#!/bin/bash
#
# Build-LinuxProvider.sh
# Builds the Terraform provider for Linux with proper executable permissions
# MUST be run in WSL to preserve Unix file permissions in zip
#

set -e

VERSION="${1:-1.0.0-custom}"
OUTPUT_DIR="${2:-./dist}"
MIRROR_DIR="./terraform-local-testing/terraform-provider-mirror/registry.terraform.io/mypurecloud/genesyscloud"

echo "======================================================================"
echo "  Building Linux Provider with Executable Permissions"
echo "======================================================================"
echo ""
echo "Version: $VERSION"
echo "Output Directory: $OUTPUT_DIR"
echo ""

# Clean and create output directory
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Build for Linux AMD64
echo "Building binary..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
    -o "$OUTPUT_DIR/terraform-provider-genesyscloud" \
    -ldflags "-s -w" \
    .

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

# Set executable permission
chmod +x "$OUTPUT_DIR/terraform-provider-genesyscloud"

# Verify executable bit is set
if [ ! -x "$OUTPUT_DIR/terraform-provider-genesyscloud" ]; then
    echo "❌ Failed to set executable permission!"
    exit 1
fi

BINARY_SIZE=$(stat -f%z "$OUTPUT_DIR/terraform-provider-genesyscloud" 2>/dev/null || stat -c%s "$OUTPUT_DIR/terraform-provider-genesyscloud")
echo "✅ Build successful! ($(numfmt --to=iec-i --suffix=B $BINARY_SIZE || echo "$BINARY_SIZE bytes"))"

# Create zip file (preserves permissions)
ZIP_NAME="terraform-provider-genesyscloud_${VERSION}_linux_amd64.zip"
ZIP_PATH="$OUTPUT_DIR/$ZIP_NAME"

echo ""
echo "Creating zip archive..."
cd "$OUTPUT_DIR"
zip "$ZIP_NAME" terraform-provider-genesyscloud
cd - > /dev/null

ZIP_SIZE=$(stat -f%z "$ZIP_PATH" 2>/dev/null || stat -c%s "$ZIP_PATH")
echo "✅ Zip created: $ZIP_NAME ($(numfmt --to=iec-i --suffix=B $ZIP_SIZE || echo "$ZIP_SIZE bytes"))"

# Calculate checksums
echo ""
echo "Calculating checksums..."
SHA256=$(sha256sum "$ZIP_PATH" | awk '{print $1}')
echo "  SHA256: $SHA256"

# Calculate h1 hash (Terraform format)
H1_HASH=$(echo -n "$SHA256" | xxd -r -p | base64)
echo "  h1 hash: h1:$H1_HASH"

# Update mirror directory if it exists
if [ -d "$MIRROR_DIR" ]; then
    echo ""
    echo "Updating provider mirror..."
    mkdir -p "$MIRROR_DIR"
    
    # Copy zip file
    cp "$ZIP_PATH" "$MIRROR_DIR/"
    echo "  ✅ Copied $ZIP_NAME to mirror"
    
    # Update manifest JSON
    MANIFEST_FILE="$MIRROR_DIR/${VERSION}.json"
    cat > "$MANIFEST_FILE" <<EOF
{
  "archives": {
    "linux_amd64": {
      "hashes": [
        "h1:$H1_HASH"
      ],
      "url": "terraform-provider-genesyscloud_${VERSION}_linux_amd64.zip"
    }
  }
}
EOF
    echo "  ✅ Updated manifest: ${VERSION}.json"
    
    # Update index.json
    INDEX_FILE="$MIRROR_DIR/index.json"
    cat > "$INDEX_FILE" <<EOF
{
  "versions": {
    "$VERSION": {}
  }
}
EOF
    echo "  ✅ Updated index.json"
fi

echo ""
echo "======================================================================"
echo "  ✅ SUCCESS - Provider Built with Executable Permissions"
echo "======================================================================"
echo ""
echo "Summary:"
echo "  Binary: $OUTPUT_DIR/terraform-provider-genesyscloud"
echo "  Archive: $ZIP_NAME"
echo "  Size: $(numfmt --to=iec-i --suffix=B $ZIP_SIZE || echo "$ZIP_SIZE bytes")"
echo "  SHA256: $SHA256"
echo "  h1 hash: h1:$H1_HASH"
echo ""
echo "Next Steps:"
echo "  1. Test the zip file in a Linux environment"
echo "  2. Upload to your hosting location (Azure Blob, etc.)"
echo "  3. Run: terraform init"
echo ""
