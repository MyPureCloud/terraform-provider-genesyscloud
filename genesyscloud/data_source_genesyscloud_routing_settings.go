package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

func dataSourceRoutingSettings() *schema.Resource {
	return &schema.Resource{
		Description:   "An organization's routing settings",
		ReadContext:   provider.ReadWithPooledClient(dataSourceRoutingSettingsRead),
		SchemaVersion: 1,
		Schema:        ResourceRoutingSettings().Schema,
	}
}

func dataSourceRoutingSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	settings, resp, getErr := routingAPI.GetRoutingSettings()
	if getErr != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_settings", fmt.Sprintf("Error requesting routing settings error: %s", getErr), resp)
	}

	d.SetId("datasource-settings")
	if settings.ResetAgentScoreOnPresenceChange != nil {
		d.Set("reset_agent_on_presence_change", *settings.ResetAgentScoreOnPresenceChange)
	}

	if diagErr := readRoutingSettingsContactCenter(d, routingAPI); diagErr != nil {
		return util.BuildDiagnosticError("genesyscloud_routing_settings", fmt.Sprintf("Error reading routing settings contact center"), fmt.Errorf("%v", diagErr))
	}

	if diagErr := readRoutingSettingsTranscription(d, routingAPI); diagErr != nil {
		return util.BuildDiagnosticError("genesyscloud_routing_settings", fmt.Sprintf("Error reading routing settings transcription"), fmt.Errorf("%v", diagErr))
	}

	return nil
}
