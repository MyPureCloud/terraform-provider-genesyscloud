package telephony_providers_edges_extension_pool

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ResourceName = "genesyscloud_telephony_providers_edges_extension_pool"
)

func ResourceTelephonyExtensionPool() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Extension Pool",
		CreateContext: provider.CreateWithPooledClient(createExtensionPool),
		ReadContext:   provider.ReadWithPooledClient(readExtensionPool),
		UpdateContext: provider.UpdateWithPooledClient(updateExtensionPool),
		DeleteContext: provider.DeleteWithPooledClient(deleteExtensionPool),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"start_number": {
				Description:      "Starting phone number of the Extension Pool range. Changing the start_number attribute will cause the extension object to be dropped and recreated with a new ID.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validators.ValidateExtensionPool,
			},
			"end_number": {
				Description:      "Ending phone number of the Extension Pool range. Changing the end_number attribute will cause the extension object to be dropped and recreated with a new ID.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validators.ValidateExtensionPool,
			},
			"description": {
				Description: "Extension Pool description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func DataSourceExtensionPool() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Extension pool. Select an Extension pool by starting number and ending number",
		ReadContext: provider.ReadWithPooledClient(dataSourceExtensionPoolRead),
		Schema: map[string]*schema.Schema{
			"start_number": {
				Description:      "Starting number of the Extension Pool range.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateExtensionPool,
			},
			"end_number": {
				Description:      "Ending number of the Extension Pool range.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateExtensionPool,
			},
		},
	}
}

func TelephonyExtensionPoolExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllExtensionPools),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceName, DataSourceExtensionPool())
	l.RegisterResource(ResourceName, ResourceTelephonyExtensionPool())
	l.RegisterExporter(ResourceName, TelephonyExtensionPoolExporter())
}
