package exporter

import (
	"context"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v191/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportByType_NoResourceExporterRegistered(t *testing.T) {
	t.Setenv("MRMO_CXASCODE_INTEGRATION_ENABLED", "true")

	_, diags := ExportByType(context.Background(), ExportByTypeInput{
		ResourceType: "genesyscloud_conversations_messaging_settings_default",
	}, &platformclientv2.Configuration{})

	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "not supported for MRMO export")
}

func TestExport_NoResourceExporterRegistered(t *testing.T) {
	t.Setenv("MRMO_CXASCODE_INTEGRATION_ENABLED", "true")

	_, diags := Export(context.Background(), ExportInput{
		BaseExportInput: BaseExportInput{
			ResourceType: "genesyscloud_conversations_messaging_settings_default",
		},
		EntityId: "default",
	}, &platformclientv2.Configuration{})

	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "not supported for MRMO export")
}
