package telephony_providers_edges_linebasesettings

// @team: Genesys Cloud Media
// @chat: #media-infrastructure-team
// @description: Telephony infrastructure and configuration management for Genesys Cloud Edge devices. Manages sites, phones, trunks, DIDs, and base settings for voice connectivity and call routing.

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

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
