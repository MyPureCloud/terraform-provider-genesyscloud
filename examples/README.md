# Examples

This directory contains examples that are mostly used for documentation, but can also be run/tested manually via the Terraform CLI.

## Documentation Generation

The document generation tool looks for files in the following locations by default. All other \*.tf files besides the ones mentioned below are ignored by the documentation tool. This is useful for creating examples that can run and/or are testable even if some parts are not relevant for the documentation.

- **provider/provider.tf** example file for the provider index page
- **data-sources/<full data source name>/data-source.tf** example file for the named data source page
- **resources/<full resource name>/resource.tf** example file for the named resource page
- **resources/<full resource name>/apis.md** markdown file containing links to APIs in use by this resource

## Example File Structure

### Key Files

- **resource.tf** - This is the primary example file that gets compiled into the generated documentation. It should demonstrate a complete, working example of the resource with all commonly used attributes.
- **simplest_resource.tf** - (Optional) A minimal working example of the resource with only required attributes. Used for testing basic functionality.
- **locals.tf** - Contains schema definitions and dependencies for testing the examples.
- **apis.md** - Contains links to the APIs used by the resource.

### Understanding the locals.tf File

The `locals.tf` file serves several important purposes:

1. **Dependencies Management**: The `dependencies` key/value defines dependencies between resources that need to be created together for testing. The `resource` key/value is a reference to the name of the file that contains the references to the dependencies (i.e., `resource.tf`).

   ```hcl
   locals {
     dependencies = {
       resource = [
         "../other_resource/resource.tf"
       ]
     }
   }
   ```

2. **Working Directory Configuration**: Specifies paths for working directories. This is important for providing references to external files (such as flows, scripts, etc).

3. **Test Constraints**: Defines conditions for skipping tests based on environment.

   ```hcl
   locals {
     skip_if = {
       products_missing_any = ["product_name"]
       not_in_domains = ["domain_name"]
     }
   }
   ```

4. **Environment Variables**: Sets environment variables needed for tests.
   ```hcl
   locals {
    environment_vars = {
      FOO = "BAR"
    }
   }
   ```

## Testing Examples

The examples directory includes test files that validate the examples:

1. **docs_examples_test.go**: Tests the examples that are used in documentation.
2. **examples_test.go**: Contains unit tests for the example loading functionality.

## Resource vs. Simplest Resource

- **resource.tf** - Comprehensive example showing most features of a resource. This is what appears in the documentation.
- **simplest_resource.tf** - Minimal working example with only required attributes. This is useful for:
  - Testing basic functionality
  - Providing a starting point for users
  - Ensuring backward compatibility

## Maintaining Examples

When updating resources:

1. Always update the corresponding `resource.tf` file to reflect any changes in the resource schema.
2. Ensure the example in `resource.tf` is complete and demonstrates best practices.
3. Update `simplest_resource.tf` if the minimum required attributes have changed.
4. Update `locals.tf` if dependencies have changed.
5. Update `apis.md` if the APIs used by the resource have changed.

## Testing Framework

The examples are automatically tested to ensure they conform to the resource schemas and can be successfully applied. The testing framework:

1. Loads each example with its dependencies
2. Validates the configuration against the provider schema
3. Applies the configuration to create real resources
4. Verifies the resources were created correctly
5. Destroys the resources

This ensures that all examples in the documentation are accurate and working.
