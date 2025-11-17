package ai_studio_summary_setting

import (
	"fmt"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

/*
The resource_genesyscloud_ai_studio_summary_setting_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getAiStudioSummarySettingFromResourceData maps data from schema ResourceData object to a platformclientv2.Summarysetting
func getAiStudioSummarySettingFromResourceData(d *schema.ResourceData) platformclientv2.Summarysetting {
	return platformclientv2.Summarysetting{
		Name:               platformclientv2.String(d.Get("name").(string)),
		Language:           platformclientv2.String(d.Get("language").(string)),
		SummaryType:        platformclientv2.String(d.Get("summary_type").(string)),
		Format:             platformclientv2.String(d.Get("format").(string)),
		MaskPII:            buildSummarySettingPIIs(d.Get("mask_p_i_i").([]interface{})),
		ParticipantLabels:  buildSummarySettingParticipantLabelss(d.Get("participant_labels").([]interface{})),
		PredefinedInsights: lists.BuildSdkStringListFromInterfaceArray(d, "predefined_insights"),
		CustomEntities:     buildSummarySettingCustomEntitys(d.Get("custom_entities").([]interface{})),
		SettingType:        platformclientv2.String(d.Get("setting_type").(string)),
		Prompt:             platformclientv2.String(d.Get("prompt").(string)),
	}
}

// buildSummarySettingPIIs maps an []interface{} into a Genesys Cloud *[]platformclientv2.Summarysettingpii
func buildSummarySettingPIIs(summarySettingPIIs []interface{}) *platformclientv2.Summarysettingpii {
	if len(summarySettingPIIs) == 0 {
		return nil
	}

	summarySettingPIIsMap, ok := summarySettingPIIs[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Summarysettingpii{
		All: resourcedata.GetNillableValueFromMap[bool](summarySettingPIIsMap, "all", false),
	}
}

// buildSummarySettingParticipantLabelss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Summarysettingparticipantlabels
func buildSummarySettingParticipantLabelss(summarySettingParticipantLabelss []interface{}) *platformclientv2.Summarysettingparticipantlabels {
	if len(summarySettingParticipantLabelss) == 0 {
		return nil
	}

	summarySettingParticipantLabelssMap, ok := summarySettingParticipantLabelss[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Summarysettingparticipantlabels{
		Internal: resourcedata.GetNillableValueFromMap[string](summarySettingParticipantLabelssMap, "internal", false),
		External: resourcedata.GetNillableValueFromMap[string](summarySettingParticipantLabelssMap, "external", false),
	}
}

// buildSummarySettingCustomEntitys maps an []interface{} into a Genesys Cloud *[]platformclientv2.Summarysettingcustomentity
func buildSummarySettingCustomEntitys(summarySettingCustomEntitys []interface{}) *[]platformclientv2.Summarysettingcustomentity {
	summarySettingCustomEntitysSlice := make([]platformclientv2.Summarysettingcustomentity, 0)
	for _, summarySettingCustomEntity := range summarySettingCustomEntitys {
		var sdkSummarySettingCustomEntity platformclientv2.Summarysettingcustomentity
		summarySettingCustomEntitysMap, ok := summarySettingCustomEntity.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkSummarySettingCustomEntity.Label, summarySettingCustomEntitysMap, "label")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSummarySettingCustomEntity.Description, summarySettingCustomEntitysMap, "description")

		summarySettingCustomEntitysSlice = append(summarySettingCustomEntitysSlice, sdkSummarySettingCustomEntity)
	}

	return &summarySettingCustomEntitysSlice
}

// flattenSummarySettingPIIs maps a Genesys Cloud *[]platformclientv2.Summarysettingpii into a []interface{}
func flattenSummarySettingPIIs(summarySettingPIIs *[]platformclientv2.Summarysettingpii) []interface{} {
	if len(*summarySettingPIIs) == 0 {
		return nil
	}

	var summarySettingPIIList []interface{}
	for _, summarySettingPII := range *summarySettingPIIs {
		summarySettingPIIMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(summarySettingPIIMap, "all", summarySettingPII.All)

		summarySettingPIIList = append(summarySettingPIIList, summarySettingPIIMap)
	}

	return summarySettingPIIList
}

// flattenSummarySettingParticipantLabelss maps a Genesys Cloud *[]platformclientv2.Summarysettingparticipantlabels into a []interface{}
func flattenSummarySettingParticipantLabelss(summarySettingParticipantLabelss *[]platformclientv2.Summarysettingparticipantlabels) []interface{} {
	if len(*summarySettingParticipantLabelss) == 0 {
		return nil
	}

	var summarySettingParticipantLabelsList []interface{}
	for _, summarySettingParticipantLabels := range *summarySettingParticipantLabelss {
		summarySettingParticipantLabelsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(summarySettingParticipantLabelsMap, "internal", summarySettingParticipantLabels.Internal)
		resourcedata.SetMapValueIfNotNil(summarySettingParticipantLabelsMap, "external", summarySettingParticipantLabels.External)

		summarySettingParticipantLabelsList = append(summarySettingParticipantLabelsList, summarySettingParticipantLabelsMap)
	}

	return summarySettingParticipantLabelsList
}

// flattenSummarySettingCustomEntitys maps a Genesys Cloud *[]platformclientv2.Summarysettingcustomentity into a []interface{}
func flattenSummarySettingCustomEntitys(summarySettingCustomEntitys *[]platformclientv2.Summarysettingcustomentity) []interface{} {
	if len(*summarySettingCustomEntitys) == 0 {
		return nil
	}

	var summarySettingCustomEntityList []interface{}
	for _, summarySettingCustomEntity := range *summarySettingCustomEntitys {
		summarySettingCustomEntityMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(summarySettingCustomEntityMap, "label", summarySettingCustomEntity.Label)
		resourcedata.SetMapValueIfNotNil(summarySettingCustomEntityMap, "description", summarySettingCustomEntity.Description)

		summarySettingCustomEntityList = append(summarySettingCustomEntityList, summarySettingCustomEntityMap)
	}

	return summarySettingCustomEntityList
}

func GenerateBasicSummarySettingResource(resourceLabel string, name string, language string, extras ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_ai_studio_summary_setting" "%s" {
		name = "%s"
		language = "%s"
		%s
	}
	`, resourceLabel, name, language, strings.Join(extras, "\n"))
}
