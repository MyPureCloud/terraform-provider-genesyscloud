package resource_exporter

import (
	"encoding/json"
	"fmt"
	"strings"
)

/*
The resource_genesyscloud_routing_queue object has the concept of bullseye ring with a member_groups attribute.
The routing team has overloaded the meaning of the member_groups so you can id and then define what "type" of id this is.
This causes problems with the exporter because our export process expects id to map to a specific resource.

This customer custom router will look at the member_group_type and resolve whether it is SKILLGROUP, GROUP type.  It will then
find the appropriate resource out of the exporters and build a reference appropriately.
*/
func MemberGroupsResolver(configMap map[string]interface{}, exporters map[string]*ResourceExporter) error {

	memberGroupType := configMap["member_group_type"]
	memberGroupID := configMap["member_group_id"].(string)

	switch memberGroupType {
	case "SKILLGROUP":
		if exporter, ok := exporters["genesyscloud_routing_skill_group"]; ok {
			exportId := (*exporter.SanitizedResourceMap[memberGroupID]).Name
			configMap["member_group_id"] = fmt.Sprintf("${genesyscloud_routing_skill_group.%s.id}", exportId)
		} else {
			return fmt.Errorf("unable to locate genesyscloud_routing_skill_group in the exporters array. Unable to resolve the ID for the member group resource")
		}

	case "GROUP":
		if exporter, ok := exporters["genesyscloud_group"]; ok {
			exportId := (*exporter.SanitizedResourceMap[memberGroupID]).Name
			configMap["member_group_id"] = fmt.Sprintf("${genesyscloud_group.%s.id}", exportId)
		} else {
			return fmt.Errorf("unable to locate genesyscloud_group in the exporters array. Unable to resolve the ID for the member group resource")
		}
	default:
		return fmt.Errorf("the memberGroupType %s cannot be located. Can not resolve to a reference attribute", memberGroupType)
	}

	return nil
}

/*
For resource_genesyscloud_outbound_ruleset, there is a property called properties which is a map of stings.
When exporting outbound rulesets, if one of the keys in the map is set to an empty string it will be ignored
by the export process. Example: properties = {"contact.Attempts" = ""}.

During the export process the value associated with the key is set to nil.
This custom exporter checks if a key has a value of nil and if it does sets it to an empty string so it is exported.
*/
func RuleSetPropertyResolver(configMap map[string]interface{}, exporters map[string]*ResourceExporter) error {
	if properties, ok := configMap["properties"].(map[string]interface{}); ok {
		for key, value := range properties {
			if value == nil {
				properties[key] = ""
			}
		}
	}

	return nil
}

/*
This property takes a key 'skills' with an array of skill ids wrapped into a string (Example: {'skills': '['skillIdHere']'} ).
This causes problems with the exporter because our export process expects id to map to a specific resource
and we have an array of attributes wrapped in a string.

This customer custom router will look at the skills array if present and resolve each string id find the appropriate resource out of the exporters and build a reference appropriately.
*/
func RuleSetSkillPropertyResolver(configMap map[string]interface{}, exporters map[string]*ResourceExporter) error {

	if exporter, ok := exporters["genesyscloud_routing_skill"]; ok {
		skillIDs := configMap["skills"].(string)

		if len(skillIDs) == 0 {
			return nil
		} else {
			sanitisedSkillIds := []string{}
			skillIDs = skillIDs[1 : len(skillIDs)-1]
			skillIdList := strings.Split(skillIDs, ",")
			exportId := ""

			// Trim the double quotes from each element in the array
			for i := 0; i < len(skillIdList); i++ {
				skillIdList[i] = strings.Trim(skillIdList[i], "\"")
			}

			for _, skillId := range skillIdList {
				exportId = (*exporter.SanitizedResourceMap[skillId]).Name
				sanitisedSkillIds = append(sanitisedSkillIds, fmt.Sprintf("${genesyscloud_routing_skill.%s.id}", exportId))
			}

			jsonData, _ := json.Marshal(sanitisedSkillIds)
			configMap["skills"] = string(jsonData)
		}
	} else {
		return fmt.Errorf("unable to locate genesyscloud_routing_skill in the exporters array.")
	}
	return nil
}

func FileContentHashResolver(configMap map[string]interface{}, filepath string) error {
	configMap["file_content_hash"] = fmt.Sprintf(`${filesha256(var.%s)}`, filepath)
	return nil
}
