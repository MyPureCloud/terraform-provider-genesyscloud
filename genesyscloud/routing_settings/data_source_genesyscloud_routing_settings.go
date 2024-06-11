package routing_settings

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

func dataSourceRoutingSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSettingsProxy(sdkConfig)

	settings, resp, getErr := proxy.getRoutingSettings(ctx)
	if getErr != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Error requesting routing settings error: %s", getErr), resp)
	}

	d.SetId("datasource-settings")
	resourcedata.SetNillableValue(d, "reset_agent_on_presence_change", settings.ResetAgentScoreOnPresenceChange)

	if diagErr := readRoutingSettingsContactCenter(ctx, d, proxy); diagErr != nil {
		return util.BuildDiagnosticError(resourceName, "Error reading routing settings contact center", fmt.Errorf("%v", diagErr))
	}

	if diagErr := readRoutingSettingsTranscription(ctx, d, proxy); diagErr != nil {
		return util.BuildDiagnosticError(resourceName, "Error reading routing settings transcription", fmt.Errorf("%v", diagErr))
	}
	return nil
}
