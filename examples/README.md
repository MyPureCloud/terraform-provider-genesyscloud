# Examples

This directory contains examples that demonstrate how to use the Genesys Cloud Terraform Provider. These examples appear in the official provider documentation and provide working examples for developers implementing Genesys Cloud resources. The examples are tested as part of the integration testing to ensure the examples consistently conform to the provider's schemas.

## Directory Structure

- **common/** - Common resources used across multiple examples (certificates, UUIDs, time)
- **data-sources/** - Examples for each data source
- **provider/** - Provider configuration examples
- **resources/** - Examples for each resource type. These are tested as part of integration testing to ensure conformity to the resource schemas defined by the provider.
- **genesyscloud/** - Additional examples for specific use cases

## Documentation Generation

The document generation tool looks for files in the following locations by default:

- **provider/provider.tf** - Example file for the provider index page
- **data-sources/<full data source name>/data-source.tf** - Example file for the named data source page
- **resources/<full resource name>/resource.tf** - Example file for the named resource page
- **resources/<full resource name>/apis.md** - Markdown file containing links to APIs in use by this resource

All other \*.tf files are ignored by the documentation tool. This allows examples to include additional files needed for testing without cluttering the documentation.

## Example File Structure

### Key Files

- **resource.tf** - The primary example file that appears in the documentation. It demonstrates a complete, working example with all commonly used attributes.
- **simplest_resource.tf** - (Optional) A minimal working example with only required attributes. Used for testing basic functionality and backward compatibility.
- **locals.tf** - Contains schema definitions, dependencies, and test constraints.
- **apis.md** - Contains links to the APIs used by the resource.

### Additional Files

Resources may include additional files specific to their functionality:

- Flow YAML files for flow resources
- Script JSON files for script resources
- Grammar files for grammar resources
- Audio files for prompt resources

## Understanding the locals.tf File

The `locals.tf` file is crucial for testing and serves several important purposes:

### 1. Dependencies Management

The `dependencies` block defines resources that need to be created together for testing:

```hcl
locals {
  dependencies = {
    resource = [
      "../other_resource/resource.tf"
    ]
  }
}
```

The key (e.g., `resource`) refers to the name of the file without extension that contains references to the dependencies. The value is an array of paths to the dependency files.

### 2. Working Directory Configuration

The `working_dir` block specifies paths for external files:

```hcl
locals {
  working_dir = {
    flows = "../flows",
    scripts = "../scripts"
  }
}
```

These paths are converted to absolute paths during testing to ensure files can be found regardless of where the test is run from.

### 3. Test Constraints

The `skip_if` block defines conditions for skipping tests based on environment:

```hcl
locals {
  skip_if = {
    products_missing_any = ["product_name"],
    not_in_domains = ["domain_name"],
    only_in_domains = ["domain_name"],
    products_existing_any = ["product_name"],
    products_existing_all = ["product_name"],
    products_missing_all = ["product_name"]
  }
}
```

This allows tests to be conditionally run based on:

- Available Genesys Cloud products in the test environment
- The domain of the test environment (e.g., TCA, mypurecloud.com, mypurecloud.ie)

### 4. Environment Variables

The `environment_vars` block sets environment variables needed for tests:

```hcl
locals {
  environment_vars = {
    FOO = "BAR"
  }
}
```

## Testing Framework

The examples directory includes a comprehensive testing framework with three main test types:

### 1. Complete Acceptance Tests (`TestAccExampleResourcesComplete`)

This test runs the full `resource.tf` example against a real Genesys Cloud environment:

- Creates all resources defined in the example
- Validates that resources were created correctly with all attributes
- Cleans up resources after testing
- Handles dependencies between resources

### 2. Plan-Only Tests (`TestUnitExampleResourcesPlanOnly`)

This test validates that the Terraform configurations are syntactically correct:

- Parses all example files
- Validates the configuration against the provider schema
- Doesn't create any real resources
- Useful for quick validation during development

### 3. Audit Tests (`TestAccExampleResourcesAudit`)

This test runs the `simplest_resource.tf` example (or falls back to `resource.tf`):

- Tests the minimal required configuration for each resource
- Ensures backward compatibility
- Validates basic functionality

### Testing Specific Resources

You can test specific resources by modifying the `TEST_SPECIFIC_RESOURCE_TYPES` variable in `docs_examples_test.go`:

```go
var TEST_SPECIFIC_RESOURCE_TYPES = []string{
    "genesyscloud_routing_queue",
    "genesyscloud_user"
}
```

### Running Tests

To run the tests:

```bash
# Run all example tests
go test -v ./examples

# Run a specific test
go test -v ./examples -run TestAccExampleResourcesComplete

# Run tests for a specific resource
go test -v ./examples -run TestAccExampleResourcesComplete/genesyscloud_routing_queue
```

## Developing New Examples

When creating new examples:

1. **Create the directory structure**:

   ```
   resources/genesyscloud_new_resource/
   ├── apis.md
   ├── locals.tf
   ├── resource.tf
   └── simplest_resource.tf (optional)
   ```

2. **Implement resource.tf**:

   - Include all commonly used attributes
   - Use descriptive names for resources
   - Include comments explaining non-obvious configurations
   - Ensure the example is complete and can be applied independently

3. **Implement simplest_resource.tf** (optional):

   - Include only required attributes
   - Keep it as minimal as possible while still being functional

4. **Define dependencies** in locals.tf:

   - List any other resources that need to be created first
   - Define working directories if needed
   - Add skip conditions if the resource requires specific products

5. **Document APIs** in apis.md:
   - List all APIs used by the resource
   - Include links to API documentation

## Dependency Resolution

The testing framework automatically resolves dependencies between examples:

1. It reads the `dependencies` block in `locals.tf`
2. Loads each dependency file
3. Recursively resolves dependencies of dependencies
4. Detects and handles cyclic dependencies
5. Ensures each file is processed only once

## Skip Constraints

Skip constraints allow tests to be conditionally run based on the environment:

- `products_missing_any`: Skip if any of these products are missing
- `products_missing_all`: Skip if all of these products are missing
- `products_existing_any`: Skip if any of these products exist
- `products_existing_all`: Skip if all of these products exist
- `not_in_domains`: Skip if not in these domains
- `only_in_domains`: Skip if in these domains

This ensures tests don't fail due to missing products or environment-specific constraints.

## Best Practices

- **Real world use cases**: Examples could represent real-world use cases.
- **Test thoroughly**: Ensure examples can be applied successfully.
- **Include comments**: Add comments to explain complex configurations.
- **Maintain dependencies**: Keep track of dependencies between resources.
- **Document APIs**: Keep API documentation up to date.
- **Use variables**: Use variables for values that might change.
- **Follow naming conventions**: Use consistent naming across examples.
- **Keep simplest_resource.tf minimal**: Include only required attributes.
- **Update examples when resources change**: Keep examples in sync with resource schemas.
