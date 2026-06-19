package tfexporter

import (
	"context"
	"testing"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportByTypeForMrMo_NilExporterReturnsDiagnostic(t *testing.T) {
	g := &GenesysCloudResourceExporter{ctx: context.Background()}
	_, diags := g.ExportByTypeForMrMo("genesyscloud_tf_export", false, nil)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "resource exporter is nil")
}

func TestExportForMrMo_NilExporterReturnsDiagnostic(t *testing.T) {
	g := &GenesysCloudResourceExporter{ctx: context.Background()}
	_, diags := g.ExportForMrMo("genesyscloud_tf_export", "id", false, nil)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "resource exporter is nil")
}

func TestExportForMrMoByID_NilExporterReturnsDiagnostic(t *testing.T) {
	g := &GenesysCloudResourceExporter{ctx: context.Background()}
	_, diags := g.ExportForMrMoByID("genesyscloud_routing_queue", "id", false, nil)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "resource exporter is nil")
}

func TestBuildSanitizedResourceMapsForMrMo_NilExporterReturnsDiagnostic(t *testing.T) {
	g := &GenesysCloudResourceExporter{ctx: context.Background()}
	exporters := map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_example": nil,
	}
	diags := g.buildSanitizedResourceMapsForMrMo(exporters, nil, false)
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "resource exporter is nil")
}
