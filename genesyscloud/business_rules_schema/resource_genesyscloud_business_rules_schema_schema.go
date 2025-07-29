package business_rules_schema

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesyscloud_business_rules_schema_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the business_rules_schema resource.
3.  The datasource schema definitions for the business_rules_schema datasource.
4.  The resource exporter configuration for the business_rules_schema exporter.
*/
const ResourceType = "genesyscloud_business_rules_schema"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceBusinessRulesSchema())
	regInstance.RegisterDataSource(ResourceType, DataSourceBusinessRulesSchema())
	regInstance.RegisterExporter(ResourceType, BusinessRulesSchemaExporter())
}

// ResourceBusinessRulesSchema registers the genesyscloud_business_rules_schema resource with Terraform
func ResourceBusinessRulesSchema() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud business rules schema`,

		CreateContext: provider.CreateWithPooledClient(createBusinessRulesSchema),
		ReadContext:   provider.ReadWithPooledClient(readBusinessRulesSchema),
		UpdateContext: provider.UpdateWithPooledClient(updateBusinessRulesSchema),
		DeleteContext: provider.DeleteWithPooledClient(deleteBusinessRulesSchema),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "The name of the Business Rules Schema",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"description": {
				Description: "The description of the Business Rules Schema",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"properties": {
				Description:      "The properties for the JSON Schema document.",
				Optional:         true,
				Type:             schema.TypeString,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			"enabled": {
				Description: `The schema's enabled/disabled status. A disabled schema cannot be assigned to any other entities, but the data on those entities from the schema still exists.`,
				Optional:    true,
				Default:     true,
				Type:        schema.TypeBool,
			},
			"version": {
				Description: `The version number of the Business Rules Schema. The version number is incremented each time the schema is modified.`,
				Computed:    true,
				Type:        schema.TypeFloat,
			},
		},
	}
}

// BusinessRulesSchemaExporter returns the resourceExporter object used to hold the genesyscloud_business_rules_schema exporter's config
func BusinessRulesSchemaExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc:     provider.GetAllWithPooledClient(getAllBusinessRulesSchemas),
		RefAttrs:             map[string]*resourceExporter.RefAttrSettings{},
		JsonEncodeAttributes: []string{"properties"},
	}
}

// DataSourceBusinessRulesSchema registers the genesyscloud_business_rules_schema data source
func DataSourceBusinessRulesSchema() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud business rules schema data source. Select a business rules schema by its name.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceBusinessRulesSchemaRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `business rules schema name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
