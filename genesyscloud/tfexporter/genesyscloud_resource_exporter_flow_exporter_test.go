package tfexporter

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	architectFlow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnitRetrieveExportersUsesNewFlowExporterOnRebuild covers DEVTOOLING-1776. With
// use_legacy_architect_flow_exporter = false, every call to retrieveExporters, including the
// rebuilds done during dependency resolution (rebuildExports), must return the new flow exporter.
// If a rebuild returns the legacy exporter, dependent flows get an unused filepath variable.
// The flow exporters used here are the real ones registered by initTestResources (TestMain).
func TestUnitRetrieveExportersUsesNewFlowExporterOnRebuild(t *testing.T) {
	tests := []struct {
		name      string
		useLegacy bool
	}{
		{name: "new flow exporter", useLegacy: false},
		{name: "legacy flow exporter", useLegacy: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"export_format":                      {Type: schema.TypeString},
				"split_files_by_resource":            {Type: schema.TypeBool},
				"log_permission_errors":              {Type: schema.TypeBool},
				"add_depends_on":                     {Type: schema.TypeBool},
				"include_state_file":                 {Type: schema.TypeBool},
				"version":                            {Type: schema.TypeString},
				"provider_registry":                  {Type: schema.TypeString},
				"export_dir_path":                    {Type: schema.TypeString},
				"ignore_cyclic_dependencies":         {Type: schema.TypeBool},
				"export_computed":                    {Type: schema.TypeBool},
				"export_omit_unresolved_refs":        {Type: schema.TypeBool},
				"max_concurrent_operations":          {Type: schema.TypeInt},
				"use_legacy_architect_flow_exporter": {Type: schema.TypeBool},
				"exclude_attributes":                 {Type: schema.TypeList, Elem: &schema.Schema{Type: schema.TypeString}},
			}, map[string]interface{}{
				"export_format":                      "hcl",
				"use_legacy_architect_flow_exporter": tt.useLegacy,
			})
			exporters := make(map[string]*resourceExporter.ResourceExporter)
			g := NewThreadSafeGenesysCloudResourceExporter(d, context.Background(), nil, &schema.Provider{}, &exporters)

			assertFlowExporter := func(stage string) {
				flowExporter := (*g.exporters)[architectFlow.ResourceType]
				require.NotNil(t, flowExporter, "%s: flow exporter missing", stage)
				_, hasUnresolvableFilepath := flowExporter.UnResolvableAttributes["filepath"]
				if tt.useLegacy {
					assert.Nil(t, flowExporter.CustomFileWriter.RetrieveAndWriteFilesFunc, "%s: expected legacy flow exporter", stage)
					assert.True(t, hasUnresolvableFilepath, "%s: legacy flow exporter should mark filepath unresolvable", stage)
				} else {
					assert.NotNil(t, flowExporter.CustomFileWriter.RetrieveAndWriteFilesFunc, "%s: expected new flow exporter", stage)
					assert.False(t, hasUnresolvableFilepath, "%s: new flow exporter must not mark filepath unresolvable", stage)
				}
			}

			// Initial retrieval, as done at the start of Export().
			require.False(t, g.retrieveExporters().HasError())
			assertFlowExporter("initial retrieval")

			// Rebuild with a by-ID filter, as done by exportDependentResources/chainDependencies -> rebuildExports.
			g.reAssignFilters()
			filterList := []string{architectFlow.ResourceType + "::00000000-0000-0000-0000-000000000000"}
			g.filterList = &filterList
			require.False(t, g.retrieveExporters().HasError())
			assertFlowExporter("dependency resolution rebuild")
		})
	}
}
