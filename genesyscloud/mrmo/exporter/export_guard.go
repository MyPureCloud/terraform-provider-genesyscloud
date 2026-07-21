package exporter

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// diagUnsupportedResourceTypeForMRMOExport is returned when Export/ExportByType is
// asked to handle a resource type that has no ResourceExporter in the provider registry.
func diagUnsupportedResourceTypeForMRMOExport(resourceType string) diag.Diagnostics {
	return diag.Errorf(
		"resource type %q is not supported for MRMO export (no ResourceExporter registered in the provider)",
		resourceType,
	)
}
