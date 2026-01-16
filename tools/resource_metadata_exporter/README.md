# Resource Metadata Exporter

A standalone command-line tool for managing resource metadata in the Genesys Cloud Terraform Provider.

## Overview

The Resource Metadata Exporter is designed to help track and manage ownership information for the 120+ resources in the Genesys Cloud Terraform Provider. It provides functionality to:

- Discover and extract metadata from resource schema files
- Export metadata in multiple formats (Markdown, JSON, CSV)
- Validate metadata completeness across all resources
- Generate metadata templates for new resources

## Features

- **Discovery**: Automatically scan resource schema files for metadata annotations
- **Export**: Generate reports in Markdown, JSON, and CSV formats
- **Validation**: Ensure all resources have proper team ownership annotations
- **Templates**: Generate metadata templates for new resources
- **Standalone**: Completely independent tool with its own build system

## Installation

### Prerequisites

- Go 1.23.7 or later
- Make (for building)

### Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd terraform-provider-genesyscloud/tools/resource_metadata_exporter

# Build the tool
make build

# Or build for all platforms
make build-all
```

### Installation

```bash
# Install to system PATH
make install

# Or copy manually
cp build/resource-metadata-exporter /usr/local/bin/
```

## Usage

### Basic Commands

```bash
# Show help
resource-metadata-exporter --help

# Discover resources
resource-metadata-exporter discover --path ./genesyscloud

# Export metadata
resource-metadata-exporter export --format markdown --output metadata.md

# Validate metadata
resource-metadata-exporter validate --path ./genesyscloud

# Generate template
resource-metadata-exporter template --resource genesyscloud_new_resource --package new_package
```

### Command Reference

#### Discover Command

Scans the specified path for resource schema files and extracts metadata annotations.

```bash
resource-metadata-exporter discover [flags]

Flags:
  -p, --path string   Path to scan for resource schema files (default "./genesyscloud")
```

**Example:**
```bash
resource-metadata-exporter discover --path ./genesyscloud
```

#### Export Command

Exports discovered metadata in various formats.

```bash
resource-metadata-exporter export [flags]

Flags:
  -f, --format string   Output format (markdown, json, csv) (default "markdown")
  -o, --output string   Output file (defaults to stdout)
```

**Examples:**
```bash
# Export to Markdown file
resource-metadata-exporter export --format markdown --output metadata.md

# Export to JSON
resource-metadata-exporter export --format json --output metadata.json

# Export to CSV
resource-metadata-exporter export --format csv --output metadata.csv

# Export to stdout
resource-metadata-exporter export --format markdown
```

#### Validate Command

Validates that all resources have proper metadata annotations.

```bash
resource-metadata-exporter validate [flags]

Flags:
  -p, --path string   Path to scan for resource schema files (default "./genesyscloud")
```

**Example:**
```bash
resource-metadata-exporter validate --path ./genesyscloud
```

#### Template Command

Generates a metadata template that can be added to a resource schema file.

```bash
resource-metadata-exporter template [flags]

Required flags:
  -r, --resource string   Resource type name
  -p, --package string    Package name
```

**Example:**
```bash
resource-metadata-exporter template --resource genesyscloud_new_resource --package new_package
```

## Resource Annotation Guide

The tool supports two annotation formats for resource metadata. Here's how to annotate your resource schema files:

### Method 1: Comment-based Annotations (Recommended)

Add these comments to your schema files at the top, after the package declaration:

```go
package architect_flow

// @team: Platform Team
// @chat: #platform-team
// @description: Manages Genesys Cloud flows and their configurations

import (
    // ... existing imports
)
```

**Example: Annotating `genesyscloud/architect_flow/resource_genesyscloud_architect_flow_schema.go`**

```go
package architect_flow

// @team: Platform Team
// @chat: #platform-team
// @description: Manages Genesys Cloud flows and their configurations

import (
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const (
	ResourceType = "genesyscloud_flow"
)

// ... rest of the file remains unchanged
```

### Method 2: Build Tag Annotations

Use Go build tags for metadata (alternative approach):

```go
//go:build team=Platform chat=#platform-team

package architect_flow

import (
    // ... existing imports
)
```

**Example: Annotating with build tags**

```go
//go:build team=Platform chat=#platform-team

package architect_flow

import (
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const (
	ResourceType = "genesyscloud_flow"
)

// ... rest of the file remains unchanged
```

### Annotation Fields

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| `@team` | Yes | Team name responsible for the resource | `@team: Platform Team` |
| `@chat` | No | Team chat room for questions/updates | `@chat: #platform-team` |
| `@description` | No | Brief description of the resource | `@description: Manages Genesys Cloud flows` |

### Best Practices

1. **Place annotations at the top** of the schema file, after the package declaration
2. **Use consistent team names** across related resources
3. **Include chat rooms** for easy team communication
4. **Add descriptions** to help new developers understand the resource
5. **Use comment-based annotations** for better readability and IDE support

### Example: Complete Annotated Resource

Here's how a complete annotated resource schema file should look:

```go
package architect_flow

// @team: Platform Team
// @chat: #platform-team
// @description: Manages Genesys Cloud flows and their configurations

import (
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const (
	ResourceType = "genesyscloud_flow"
)

// SetRegistrar registers all resources, data sources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceArchitectFlow())
	l.RegisterResource(ResourceType, ResourceArchitectFlow())
	l.RegisterExporter(ResourceType, ArchitectFlowExporter())
}

// ... rest of the resource implementation
```

## Metadata Format

The tool supports two annotation formats for resource metadata:

### Comment-based Annotations

Add these comments to your schema files:

```go
// @team: Platform Team
// @chat: #platform-team
// @description: Manages Genesys Cloud flows and their configurations
```

### Build Tag Annotations

Use Go build tags for metadata:

```go
//go:build team=Platform chat=#platform-team
```

## Output Formats

### Markdown

Generates a human-readable table format:

```markdown
# Genesys Cloud Terraform Provider - Resource Metadata

| Resource Type | Package | Team | Chat Room | Description |
|---------------|---------|------|-----------|-------------|
| genesyscloud_flow | architect_flow | Platform Team | #platform-team | Manages Genesys Cloud flows |
| genesyscloud_queue | routing_queue | Routing Team | #routing-team | Manages routing queues |
```

### JSON

Structured data for programmatic consumption:

```json
[
  {
    "resource_type": "genesyscloud_flow",
    "package_name": "architect_flow",
    "team_name": "Platform Team",
    "team_chat_room": "#platform-team",
    "description": "Manages Genesys Cloud flows"
  }
]
```

### CSV

Spreadsheet-compatible format:

```csv
Resource Type,Package,Team,Chat Room,Description
genesyscloud_flow,architect_flow,Platform Team,#platform-team,Manages Genesys Cloud flows
genesyscloud_queue,routing_queue,Routing Team,#routing-team,Manages routing queues
```

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint
```

### Development Setup

```bash
# Setup development environment
make dev-setup

# Generate documentation
make docs
```

## Configuration

The tool can be configured using environment variables:

- `RME_PATH`: Default path for resource discovery
- `RME_OUTPUT_FORMAT`: Default output format
- `RME_OUTPUT_FILE`: Default output file

## Examples

### Complete Workflow

```bash
# 1. Discover resources
resource-metadata-exporter discover --path ./genesyscloud

# 2. Validate metadata
resource-metadata-exporter validate --path ./genesyscloud

# 3. Export report
resource-metadata-exporter export --format markdown --output resource-ownership.md

# 4. Generate template for new resource
resource-metadata-exporter template --resource genesyscloud_new_feature --package new_feature
```

### Integration with CI/CD

```bash
# Validate in CI pipeline
resource-metadata-exporter validate --path ./genesyscloud || exit 1

# Generate reports for documentation
resource-metadata-exporter export --format markdown --output docs/resource-ownership.md
```

## Troubleshooting

### Common Issues

1. **No resources found**: Ensure the path points to the correct directory containing resource schema files
2. **Validation failures**: Check that all resources have proper team annotations
3. **Build errors**: Ensure Go version is 1.23.7 or later

### Debug Mode

Run with verbose output:

```bash
resource-metadata-exporter discover --path ./genesyscloud -v
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This tool is part of the Genesys Cloud Terraform Provider project and follows the same license terms.

## Support

For issues and questions:

1. Check the troubleshooting section
2. Review the command help: `resource-metadata-exporter --help`
3. Open an issue in the main repository 