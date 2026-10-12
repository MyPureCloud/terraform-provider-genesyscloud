package integration_credential

// @team: Integration Services Indy
// @chat: #genesys-cloud-integrations
// @pm: Richard Schott
// @jira: INTINDY
// @description: Manages integrations with third-party services and systems. Provides the foundation for connecting Genesys Cloud to external APIs, enabling data exchange and workflow automation across platforms.

import (
	"fmt"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	featureToggles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
resource_genesyscloud_integration_credential_schema.go should hold four types of functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the integration_credential resource.
3.  The datasource schema definitions for the integration_credential datasource.
4.  The resource exporter configuration for the integration_credential exporter.
*/
const ResourceType = "genesyscloud_integration_credential"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceIntegrationCredential())
	l.RegisterResource(ResourceType, ResourceIntegrationCredential())
	l.RegisterExporter(ResourceType, IntegrationCredentialExporter())
}

// ResourceIntegrationCredential registers the genesyscloud_integration_credential resource with Terraform
func ResourceIntegrationCredential() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Credential",

		CreateContext: provider.CreateWithPooledClient(createCredential),
		ReadContext:   provider.ReadWithPooledClient(readCredential),
		UpdateContext: provider.UpdateWithPooledClient(updateCredential),
		DeleteContext: provider.DeleteWithPooledClient(deleteCredential),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Credential name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"credential_type_name": {
				Description: "Credential type name. Use [GET /api/v2/integrations/credentials/types](https://developer.genesys.cloud/api/rest/v2/integrations/#get-api-v2-integrations-credentials-types) to see the list of available integration credential types. ",
				Type:        schema.TypeString,
				Required:    true,
			},
			"fields": {
				Description: "Credential fields. Different credential types require different fields. Missing any correct required fields will result API request failure. Use [GET /api/v2/integrations/credentials/types](https://developer.genesys.cloud/api/rest/v2/integrations/#get-api-v2-integrations-credentials-types) to check out the specific credential type schema to find out what fields are required. ",
				Type:        schema.TypeMap,
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// IntegrationCredentialExporter returns the resourceExporter object used to hold the genesyscloud_integration_credential exporter's config
func IntegrationCredentialExporter() *resourceExporter.ResourceExporter {
	exporter := &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllCredentials),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
		UnResolvableAttributes: map[string]*schema.Schema{
			"fields": ResourceIntegrationCredential().Schema["fields"],
		},
	}

	// When the standalone integration config toggle is enabled, resolve the GUID in the
	// credential name field to a Terraform reference. This is safe because with the toggle ON,
	// the config is a separate resource and the cycle is broken.
	// e.g., "Integration-502f8452-a4fd-43f1-9edd-2d0b947115a7" → "Integration-${genesyscloud_integration.Example.id}"
	if featureToggles.ICToggleExists() {
		exporter.CustomAttributeResolver = map[string]*resourceExporter.RefAttrCustomResolver{
			"name": {
				ResolverFunc: resolveCredentialNameGUID,
			},
		}
	}

	return exporter
}

// resolveCredentialNameGUID resolves the integration GUID embedded in credential names
// (format: "Integration-<GUID>" or "*Integration-<GUID>") to a Terraform reference.
func resolveCredentialNameGUID(configMap map[string]interface{}, exporters map[string]*resourceExporter.ResourceExporter, resourceLabel string) error {
	name, ok := configMap["name"].(string)
	if !ok || name == "" {
		return nil
	}

	// Find "Integration-" pattern anywhere in the name
	idx := strings.Index(name, "Integration-")
	if idx == -1 {
		return nil
	}

	// Extract the GUID after "Integration-"
	guid := name[idx+len("Integration-"):]
	if guid == "" {
		return nil
	}

	// Look up the GUID in the integration exporter's resource map
	integrationExporter, exists := exporters["genesyscloud_integration"]
	if !exists {
		return nil
	}

	// Find the integration resource by its ID (the GUID is the map key)
	resourceMeta, exists := integrationExporter.SanitizedResourceMap[guid]
	if !exists || resourceMeta == nil {
		return nil
	}

	// Replace the GUID portion with a Terraform reference expression
	// e.g., "Integration-<GUID>" → "Integration-${genesyscloud_integration.Example.id}"
	prefix := name[:idx]
	configMap["name"] = fmt.Sprintf("%sIntegration-${genesyscloud_integration.%s.id}", prefix, resourceMeta.BlockLabel)
	return nil
}

// DataSourceIntegrationCredential registers the genesyscloud_integration_credential data source
func DataSourceIntegrationCredential() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud integration credential. Select an integration credential by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceIntegrationCredentialRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the integration credential",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
