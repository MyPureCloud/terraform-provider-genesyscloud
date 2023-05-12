

## Introduction 

The Genesys Cloud CX as Code exporter is a big piece of code that can export all of the configuration for supported Genesys Cloud resources.  Originally this code was written as one massive export file, but has been slowly been refactored into multiple files.  These files include:

* **resource_genesyscloud_tf_export.go** - This file contains all of the Terraform Schema definitions and method needed for the CX as Code exported to function as a Terraform resource.

* **genesyscloud_resource_exporter.go** - This file contains all of the logic to carry out the flow of an export.  The code in this file is used for the execution and coordination of a Genesys Cloud export.

* **json_exporter.go** - This file contains all of the logic needed to export Genesys Cloud objects into a terraform-compliant JSON file.

* **hcl_exporter.go** - This file contains all of the logic needed to export Genesys Cloud objects into a terraform-compliant HCL file.

* **tftstate_exporter.go** - This file contains all of the logic to write a tfstate file for the exported Genesys Cloud objects.

* **export_common.go** - This file contains functions that are used across multiple exporters.

