package speechandtextanalytics_settings

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
The resource_genesyscloud_speechandtextanalytics_settings_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getSpeechAndTextAnalyticsSettingsFromResourceData maps data from a schema ResourceData object to a
// platformclientv2.Speechtextanalyticssettingsrequest. This is the PUT (full replace) request body,
// so every writable field is sent.
func getSpeechAndTextAnalyticsSettingsFromResourceData(d *schema.ResourceData) platformclientv2.Speechtextanalyticssettingsrequest {
	return platformclientv2.Speechtextanalyticssettingsrequest{
		DefaultProgramId:     resourcedata.GetNillableValue[string](d, "default_program_id"),
		ExpectedDialects:     lists.BuildSdkStringListFromInterfaceArray(d, "expected_dialects"),
		TextAnalyticsEnabled: platformclientv2.Bool(d.Get("text_analytics_enabled").(bool)),
		AgentEmpathyEnabled:  platformclientv2.Bool(d.Get("agent_empathy_enabled").(bool)),
	}
}

// setSpeechAndTextAnalyticsSettingsToResourceData maps data from a platformclientv2.Speechtextanalyticssettingsresponse
// into a schema ResourceData object. Shared by both the resource read and the data source read.
func setSpeechAndTextAnalyticsSettingsToResourceData(d *schema.ResourceData, settings *platformclientv2.Speechtextanalyticssettingsresponse) {
	// The response returns the default program as an object; flatten it down to its ID for the schema.
	if settings.DefaultProgram != nil {
		resourcedata.SetNillableValue(d, "default_program_id", settings.DefaultProgram.Id)
	} else {
		_ = d.Set("default_program_id", nil)
	}
	resourcedata.SetNillableValue(d, "expected_dialects", settings.ExpectedDialects)
	resourcedata.SetNillableValue(d, "text_analytics_enabled", settings.TextAnalyticsEnabled)
	resourcedata.SetNillableValue(d, "agent_empathy_enabled", settings.AgentEmpathyEnabled)
}
