package telephony_providers_edges_did_pool

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"
)

const resourceName = "genesyscloud_telephony_providers_edges_did_pool"

// SetRegistrar registers all resources, data sources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceDidPool())
	l.RegisterResource(resourceName, ResourceTelephonyDidPool())
	l.RegisterExporter(resourceName, TelephonyDidPoolExporter())
}

// TelephonyDidPoolExporter returns the resourceExporter object used to hold the genesyscloud_telephony_providers_edges_did_pool exporter's config
func TelephonyDidPoolExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllDidPools),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

// ResourceTelephonyDidPool registers the genesyscloud_telephony_providers_edges_did_pool resource with Terraform
func ResourceTelephonyDidPool() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud DID Pool",

		CreateContext: provider.CreateWithPooledClient(createDidPool),
		ReadContext:   provider.ReadWithPooledClient(readDidPool),
		UpdateContext: provider.UpdateWithPooledClient(updateDidPool),
		DeleteContext: provider.DeleteWithPooledClient(deleteDidPool),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"start_phone_number": {
				Description:      "Starting phone number of the DID Pool range. Phone number must be in a E.164 number format. Changing the start_phone_number attribute will cause the did_pool object to be dropped and recreated with a new ID.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidatePhoneNumber,
			},
			"end_phone_number": {
				Description:      "Ending phone number of the DID Pool range.  Phone number must be in an E.164 number format. Changing the end_phone_number attribute will cause the did_pool object to be dropped and recreated with a new ID.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidatePhoneNumber,
			},
			"description": {
				Description: "DID Pool description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"comments": {
				Description: "Comments for the DID Pool.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"pool_provider": {
				Description:  "Provider (PURE_CLOUD | PURE_CLOUD_VOICE).",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"PURE_CLOUD", "PURE_CLOUD_VOICE"}, false),
			},
		},
	}
}

// DataSourceDidPool registers the genesyscloud_telephony_providers_edges_did_pool data source
func DataSourceDidPool() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud DID pool. Select a DID pool by starting phone number and ending phone number",
		ReadContext: provider.ReadWithPooledClient(dataSourceDidPoolRead),
		Schema: map[string]*schema.Schema{
			"start_phone_number": {
				Description:      "Starting phone number of the DID Pool range. Must be in an E.164 number format.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidatePhoneNumber,
			},
			"end_phone_number": {
				Description:      "Ending phone number of the DID Pool range.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidatePhoneNumber,
			},
		},
	}
}
