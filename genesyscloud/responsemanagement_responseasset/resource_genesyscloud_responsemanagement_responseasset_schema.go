package responsemanagement_responseasset

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	gcloud "terraform-provider-genesyscloud/genesyscloud"

	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_responsemanagement_responseasset_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the responsemanagement_responseasset resource.
3.  The datasource schema definitions for the responsemanagement_responseasset datasource.
4.  The resource exporter configuration for the responsemanagement_responseasset exporter.
*/
const resourceName = "genesyscloud_responsemanagement_responseasset"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceResponseManagementResponseAsset())
	regInstance.RegisterDataSource(resourceName, DataSourceResponseManagamentResponseAsset())
}

// ResourceResponsemanagementResponseasset registers the genesyscloud_responsemanagement_responseasset resource with Terraform
func ResourceResponseManagementResponseAsset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud responsemanagement response asset`,

		CreateContext: gcloud.CreateWithPooledClient(createRespManagementRespAsset),
		ReadContext:   gcloud.ReadWithPooledClient(readRespManagementRespAsset),
		UpdateContext: gcloud.UpdateWithPooledClient(updateRespManagementRespAsset),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteRespManagementRespAsset),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`filename`: {
				Description:      "Name of the file to upload. Changing the name attribute will cause the existing response asset to be dropped and recreated with a new ID. It must not start with a dot and not end with a forward slash. Whitespace and the following characters are not allowed: \\{^}%`]\">[~<#|",
				Required:         true,
				ForceNew:         true,
				Type:             schema.TypeString,
				ValidateDiagFunc: gcloud.ValidateResponseAssetName,
			},
			`division_id`: {
				Description: `Division to associate to this asset. Can only be used with this division.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func DataSourceResponseManagamentResponseAsset() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Response Management Response Assets. Select a response asset by name.",
		ReadContext: gcloud.ReadWithPooledClient(dataSourceResponseManagamentResponseAssetRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Response asset name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
