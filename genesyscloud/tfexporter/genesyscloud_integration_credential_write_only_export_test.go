package tfexporter

import (
	"context"
	"testing"

	integrationCred "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration_credential"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnitExportIntegrationCredentialWriteOnlyFields exercises the real export config-building path
// (buildResourceConfigMap -> instanceStateToMap -> sanitizeConfigMap) for a credential that was created
// with the write-only fields_wo argument. No API calls are made; the state is constructed directly.
//
// A write-only credential has fields_wo_version stored in state but never fields_wo (the secret is never
// persisted). This test documents and guards what the exporter writes for such a credential so that the
// generated config is valid on its own (it must not require fields_wo while never being able to provide it).
func TestUnitExportIntegrationCredentialWriteOnlyFields(t *testing.T) {
	const (
		resourceType  = integrationCred.ResourceType
		resourceLabel = "write_only_cred"
	)

	// Real resource schema and real exporter (carries UnResolvableAttributes for "fields").
	credResource := integrationCred.ResourceIntegrationCredential()
	credExporter := integrationCred.IntegrationCredentialExporter()

	testProvider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			resourceType: credResource,
		},
	}

	exporters := map[string]*resourceExporter.ResourceExporter{
		resourceType: credExporter,
	}

	ctx := context.Background()
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
		"export_deprecated":                  {Type: schema.TypeBool},
		"export_omit_unresolved_refs":        {Type: schema.TypeBool},
		"use_legacy_architect_flow_exporter": {Type: schema.TypeBool},
	}, map[string]interface{}{
		"export_format":                      "hcl",
		"split_files_by_resource":            false,
		"log_permission_errors":              false,
		"add_depends_on":                     false,
		"include_state_file":                 false,
		"version":                            "1.0.0",
		"provider_registry":                  "test",
		"export_dir_path":                    "/tmp/test",
		"ignore_cyclic_dependencies":         false,
		"export_computed":                    false,
		"export_deprecated":                  true,
		"export_omit_unresolved_refs":        false,
		"use_legacy_architect_flow_exporter": false,
	})

	g := NewThreadSafeGenesysCloudResourceExporter(d, ctx, nil, testProvider, &exporters)

	// State as it exists for a credential created via fields_wo:
	//   - name and credential_type_name are stored,
	//   - fields is an empty map (the provider sets it to {} so it is a known, non-secret value),
	//   - fields_wo is NOT present (write-only; never persisted),
	//   - fields_wo_version is stored.
	g.addResources([]resourceExporter.ResourceInfo{
		{
			State: &terraform.InstanceState{
				ID: "cred-id-1",
				Attributes: map[string]string{
					"name":                 "Integration-my integration",
					"credential_type_name": "basicAuth",
					"fields.%":             "0",
					"fields_wo_version":    "1",
				},
			},
			BlockLabel:    resourceLabel,
			OriginalLabel: resourceLabel,
			Type:          resourceType,
			CtyType:       credResource.CoreConfigSchema().ImpliedType(),
			BlockType:     "resource",
		},
	})

	diags := g.buildResourceConfigMap()
	require.False(t, diags.HasError(), "buildResourceConfigMap returned errors: %v", diags)

	resourceMaps := g.getResourceTypesMaps()
	require.Contains(t, resourceMaps, resourceType)
	require.Contains(t, resourceMaps[resourceType], resourceLabel)
	configMap := resourceMaps[resourceType][resourceLabel]

	// The secret itself must never appear.
	assert.Nil(t, configMap["fields_wo"], "fields_wo must never be exported (it is write-only)")

	// fields_wo_version must NOT be exported on its own. The schema marks fields_wo and
	// fields_wo_version as RequiredWith each other, so exporting fields_wo_version without
	// fields_wo would produce a config that fails validation. The exporter must drop it.
	assert.Nil(t, configMap["fields_wo_version"],
		"fields_wo_version must not be exported without fields_wo, otherwise the generated config fails RequiredWith validation")
}
