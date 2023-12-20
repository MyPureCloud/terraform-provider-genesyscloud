package telephony_providers_edges_extension_pool

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceTelephonyExtensionPool() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Extension Pool",
		CreateContext: gcloud.CreateWithPooledClient(createExtensionPool),
		ReadContext:   gcloud.ReadWithPooledClient(readExtensionPool),
		UpdateContext: gcloud.UpdateWithPooledClient(updateExtensionPool),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteExtensionPool),
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
				ValidateDiagFunc: gcloud.ValidateExtensionPool,
			},
			"end_number": {
				Description:      "Ending phone number of the Extension Pool range. Changing the end_number attribute will cause the extension object to be dropped and recreated with a new ID.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: gcloud.ValidateExtensionPool,
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
		ReadContext: gcloud.ReadWithPooledClient(dataSourceExtensionPoolRead),
		Schema: map[string]*schema.Schema{
			"start_number": {
				Description:      "Starting number of the Extension Pool range.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: gcloud.ValidateExtensionPool,
			},
			"end_number": {
				Description:      "Ending number of the Extension Pool range.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: gcloud.ValidateExtensionPool,
			},
		},
	}
}

func TelephonyExtensionPoolExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllExtensionPools),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource("genesyscloud_telephony_providers_edges_extension_pool", DataSourceExtensionPool())
	l.RegisterResource("genesyscloud_telephony_providers_edges_extension_pool", ResourceTelephonyExtensionPool())
	l.RegisterExporter("genesyscloud_telephony_providers_edges_extension_pool", TelephonyExtensionPoolExporter())
}
