package outbound_callanalysisresponseset

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getResponseSetFromResourceData(d *schema.ResourceData) platformclientv2.Responseset {
	sdkResponseSet := platformclientv2.Responseset{
		Name:                 platformclientv2.String(d.Get("name").(string)),
		BeepDetectionEnabled: platformclientv2.Bool(d.Get("beep_detection_enabled").(bool)),
	}

	responses := d.Get("responses").([]interface{})
	if responses != nil && len(responses) > 0 {
		sdkResponseSet.Responses = buildSdkOutboundCallAnalysisResponseSetReaction(responses)
	}

	return sdkResponseSet
}

func buildSdkOutboundCallAnalysisResponseSetReaction(responses []interface{}) *map[string]platformclientv2.Reaction {
	if len(responses) == 0 {
		return nil
	}
	sdkResponses := map[string]platformclientv2.Reaction{}
	if responsesMap, ok := responses[0].(map[string]interface{}); ok {
		types := []string{
			"callable_lineconnected",
			"callable_person",
			"callable_busy",
			"callable_noanswer",
			"callable_fax",
			"callable_disconnect",
			"callable_machine",
			"callable_sit",
			"uncallable_sit",
			"uncallable_notfound",
		}
		for _, t := range types {
			reactionSet := responsesMap[t].(*schema.Set).List()
			if len(reactionSet) == 0 {
				continue
			}
			if reactionMap, ok := reactionSet[0].(map[string]interface{}); ok {
				sdkKey := "disposition.classification." + strings.ReplaceAll(t, "_", ".")
				sdkResponses[sdkKey] = *buildSdkReaction(reactionMap)
			}
		}
	}
	return &sdkResponses
}

func buildSdkReaction(reactionMap map[string]interface{}) *platformclientv2.Reaction {
	var sdkReaction platformclientv2.Reaction

	resourcedata.BuildSDKStringValueIfNotNil(&sdkReaction.Name, reactionMap, "name")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkReaction.Data, reactionMap, "data")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkReaction.ReactionType, reactionMap, "reaction_type")

	return &sdkReaction
}

func flattenSdkOutboundCallAnalysisResponseSetReaction(responses *map[string]platformclientv2.Reaction) []interface{} {
	if responses == nil {
		return nil
	}
	responsesMap := make(map[string]interface{})
	for key, val := range *responses {
		schemaKey := strings.Replace(key, "disposition.classification.", "", -1)
		schemaKey = strings.Replace(schemaKey, ".", "_", -1)
		responsesMap[schemaKey] = flattenSdkReaction(val)
	}
	return []interface{}{responsesMap}
}

func flattenSdkReaction(sdkReaction platformclientv2.Reaction) *schema.Set {
	var (
		reactionMap = make(map[string]interface{})
		reactionSet = schema.NewSet(schema.HashResource(reactionResource), []interface{}{})
	)
	if sdkReaction.Name != nil {
		reactionMap["name"] = *sdkReaction.Name
	}
	if sdkReaction.Data != nil {
		reactionMap["data"] = *sdkReaction.Data
	}
	reactionMap["reaction_type"] = *sdkReaction.ReactionType
	reactionSet.Add(reactionMap)
	return reactionSet
}

func GenerateOutboundCallAnalysisResponseSetResource(resourceId string, name string, beepDetectionEnabled string, responsesBlock string) string {
	return fmt.Sprintf(`
resource "genesyscloud_outbound_callanalysisresponseset" "%s" {
	name                   = "%s"
	beep_detection_enabled = %s
	%s
}
`, resourceId, name, beepDetectionEnabled, responsesBlock)
}

func GenerateCarsResponsesBlock(nestedBlocks ...string) string {
	return fmt.Sprintf(`
	responses {
		%s
	}
`, strings.Join(nestedBlocks, "\n"))
}

func GenerateCarsResponse(identifier string, reactionType string, name string, data string) string {
	if name != "" {
		name = fmt.Sprintf(`name = "%s"`, name)
	}
	if data != "" {
		data = fmt.Sprintf(`data = "%s"`, data)
	}
	return fmt.Sprintf(`
		%s {
			reaction_type = "%s"
			%s
			%s
		}
`, identifier, reactionType, name, data)
}
