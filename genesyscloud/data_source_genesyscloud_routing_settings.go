package genesyscloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func dataSourceRoutingSettings() *schema.Resource {
	return &schema.Resource{
		Description:   "An organization's routing settings",
		ReadContext:   ReadWithPooledClient(dataSourceRoutingSettingsRead),
		SchemaVersion: 1,
		Schema:        ResourceRoutingSettings().Schema,
	}
}

func dataSourceRoutingSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	settings, _, getErr := routingAPI.GetRoutingSettings()
	if getErr != nil {
		return diag.Errorf("Error requesting routing settings: %s", getErr)
	}

	d.SetId("datasource-settings")
	if settings.ResetAgentScoreOnPresenceChange != nil {
		d.Set("reset_agent_on_presence_change", *settings.ResetAgentScoreOnPresenceChange)
	}

	if diagErr := readRoutingSettingsContactCenter(d, routingAPI); diagErr != nil {
		return diag.Errorf("%v", diagErr)
	}

	if diagErr := readRoutingSettingsTranscription(d, routingAPI); diagErr != nil {
		return diag.Errorf("%v", diagErr)
	}

	return nil
}
