package integration_credential

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesyscloud_integration_credential_schema.go should hold four types of functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the integration_credential resource.
3.  The datasource schema definitions for the integration_credential datasource.
4.  The resource exporter configuration for the integration_credential exporter.
*/
const resourceName = "genesyscloud_integration_credential"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceIntegrationCredential())
	l.RegisterResource(resourceName, ResourceIntegrationCredential())
	l.RegisterExporter(resourceName, IntegrationCredentialExporter())
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
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllCredentials),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No Reference
		UnResolvableAttributes: map[string]*schema.Schema{
			"fields": ResourceIntegrationCredential().Schema["fields"],
		},
	}
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
