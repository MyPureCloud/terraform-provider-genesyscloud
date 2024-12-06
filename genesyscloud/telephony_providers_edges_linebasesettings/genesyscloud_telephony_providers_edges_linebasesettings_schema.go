package telephony_providers_edges_linebasesettings

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_telephony_providers_edges_linebasesettings"

func DataSourceLineBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Line Base Settings. Select a line base settings by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceLineBaseSettingsRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Line Base Settings name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource("genesyscloud_telephony_providers_edges_linebasesettings", DataSourceLineBaseSettings())
}
