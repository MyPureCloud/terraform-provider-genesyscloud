package responsemanagement_responseasset

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"
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
	regInstance.RegisterDataSource(resourceName, DataSourceResponseManagementResponseAsset())
	regInstance.RegisterExporter(resourceName, ExporterResponseManagementResponseAsset())
}

// ResourceResponsemanagementResponseasset registers the genesyscloud_responsemanagement_responseasset resource with Terraform
func ResourceResponseManagementResponseAsset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud responsemanagement response asset`,

		CreateContext: provider.CreateWithPooledClient(createRespManagementRespAsset),
		ReadContext:   provider.ReadWithPooledClient(readRespManagementRespAsset),
		UpdateContext: provider.UpdateWithPooledClient(updateRespManagementRespAsset),
		DeleteContext: provider.DeleteWithPooledClient(deleteRespManagementRespAsset),
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
				ValidateDiagFunc: validators.ValidateResponseAssetName,
			},
			`division_id`: {
				Description: `Division to associate to this asset. Can only be used with this division.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			"file_content_hash": {
				Description: "Hash value of the response asset file content. Used to detect changes.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func DataSourceResponseManagementResponseAsset() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Response Management Response Assets. Select a response asset by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceResponseManagementResponseAssetRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Response asset name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func ExporterResponseManagementResponseAsset() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllResponseAssets),
		CustomFileWriter: resourceExporter.CustomFileWriterSettings{
			RetrieveAndWriteFilesFunc: responsemanagementResponseassetResolver,
			SubDirectory:              "response_assets",
		},
	}
}
