

## Introduction 

The Genesys Cloud CX as Code exporter is a big piece of code that can export all of the configuration for supported Genesys Cloud resources.  Originally this code was written as one massive export file, but has been slowly been refactored into multiple files.  These files include:

* **resource_genesyscloud_tf_export.go** - This file contains all of the Terraform Schema definitions and method needed for the CX as Code exported to function as a Terraform resource.

* **genesyscloud_resource_exporter.go** - This file contains all of the logic to carry out the flow of an export.  The code in this file is used for the execution and coordination of a Genesys Cloud export.

* **json_exporter.go** - This file contains all of the logic needed to export Genesys Cloud objects into a terraform-compliant JSON file.

* **hcl_exporter.go** - This file contains all of the logic needed to export Genesys Cloud objects into a terraform-compliant HCL file.

* **tftstate_exporter.go** - This file contains all of the logic to write a tfstate file for the exported Genesys Cloud objects.

* **export_common.go** - This file contains functions that are used across multiple exporters.


## Deprecating a Resource

When a resource's API endpoints are removed or the resource is being sunset, follow these steps to ensure the exporter handles it correctly:

1. **Set `DeprecationMessage` on the `schema.Resource`** in the resource's schema file:
   ```go
   func ResourceExample() *schema.Resource {
       return &schema.Resource{
           DeprecationMessage: "This resource is being removed. See <link>",
           // ...
       }
   }
   ```

2. **Remove the `RegisterExporter` call** from the resource's `SetRegistrar` function. If the API is gone, there is nothing to export, and leaving the exporter registered will cause 410/404 errors during export. Keep `RegisterResource` and `RegisterDataSource` so that customers can still manage existing state entries (e.g. `terraform destroy`, `terraform state rm`).
   ```go
   func SetRegistrar(regInstance registrar.Registrar) {
       regInstance.RegisterResource(ResourceType, ResourceExample())
       regInstance.RegisterDataSource(ResourceType, DataSourceExample())
       // Exporter intentionally not registered — API endpoints have been removed.
   }
   ```

3. **The `export_deprecated` flag** in the exporter will also skip resource types that have a `DeprecationMessage` set when a customer configures `export_deprecated = false`. However, since the default is `true`, unregistering the exporter (step 2) is the primary safeguard to prevent export failures for all customers regardless of their configuration.
