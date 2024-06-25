package resource_exporter

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
OutboundCampaignAgentScriptResolver
Forces script_id to reference a data source for the Default Outbound Script
and the returns the script resource type "", the data source ID, and the config of the data source for the tfexporter package to add it to the export.
(We can't pass in the map and add the data source here because it causes a cyclic error between packages resource_exporter and tfexporter,
so instead we pass back all the details tfexporter needs to do it itself)
*/
func OutboundCampaignAgentScriptResolver(configMap map[string]interface{}, value any, sdkConfig *platformclientv2.Configuration) (dsType string, dsID string, dsConfig map[string]interface{}, resolve bool) {
	var (
		scriptDataSourceConfig = make(map[string]interface{})
		scriptDataSourceId     = strings.Replace(constants.DefaultOutboundScriptName, " ", "_", -1)
	)
	scriptId, _ := value.(string)
	if IsDefaultOutboundScript(scriptId, sdkConfig) {
		scriptDataSourceConfig["name"] = constants.DefaultOutboundScriptName

		configMap["script_id"] = fmt.Sprintf("${data.genesyscloud_script.%s.id}", scriptDataSourceId)

		return "genesyscloud_script", scriptDataSourceId, scriptDataSourceConfig, true
	}

	return "", "", nil, false
}

/*
IsDefaultOutboundScript
Takes a script ID and checks if the name of the script equals defaultOutboundScriptName.
If the operation fails, we will just log the error and allow the exporter to include the hard-coded GUID, as opposed to failing.
*/
func IsDefaultOutboundScript(scriptId string, sdkConfig *platformclientv2.Configuration) bool {
	if !isValidGuid(scriptId) {
		return false
	}

	apiInstance := platformclientv2.NewScriptsApiWithConfig(sdkConfig)

	log.Printf("reading published script %s", scriptId)
	data, _, err := apiInstance.GetScriptsPublishedScriptId(scriptId, "")
	if err != nil {
		log.Printf("failed to read script %s: %v", scriptId, err)
		return false
	}

	log.Printf("read published script %s %s", scriptId, *data.Name)
	return *data.Name == constants.DefaultOutboundScriptName
}

func isValidGuid(id string) bool {
	matched, err := regexp.MatchString("^[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}$", id)
	if err != nil {
		log.Printf("failed to validate format of GUID %s: %v", id, err)
		return false
	}
	return matched
}

/*
MemberGroupsResolver
The resource_genesyscloud_routing_queue object has the concept of bullseye ring with a member_groups attribute.
The routing team has overloaded the meaning of the member_groups so you can id and then define what "type" of id this is.
This causes problems with the exporter because our export process expects id to map to a specific resource.

This customer custom router will look at the member_group_type and resolve whether it is SKILLGROUP, GROUP type.  It will then
find the appropriate resource out of the exporters and build a reference appropriately.
*/
func MemberGroupsResolver(configMap map[string]interface{}, exporters map[string]*ResourceExporter, _ string) error {
	var (
		resourceID      string
		memberGroupType = configMap["member_group_type"].(string)
		memberGroupID   = configMap["member_group_id"].(string)
	)

	switch memberGroupType {
	case "SKILLGROUP":
		resourceID = "genesyscloud_routing_skill_group"
	case "GROUP":
		resourceID = "genesyscloud_group"
	case "TEAM":
		resourceID = "genesyscloud_team"
	default:
		return fmt.Errorf("the memberGroupType %s cannot be located. Can not resolve to a reference attribute", memberGroupType)
	}

	if exporter, ok := exporters[resourceID]; ok {
		memberGroupExport, ok := exporter.SanitizedResourceMap[memberGroupID]
		if !ok || memberGroupExport == nil {
			return fmt.Errorf("could not resolve member group %s to a resource of type %s", memberGroupID, resourceID)
		}
		exportId := memberGroupExport.Name
		configMap["member_group_id"] = fmt.Sprintf("${%s.%s.id}", resourceID, exportId)
	} else {
		return fmt.Errorf("unable to locate %s in the exporters array. Unable to resolve the ID for the member group resource", resourceID)
	}

	return nil
}

/*
RuleSetPropertyResolver
For resource_genesyscloud_outbound_ruleset, there is a property called properties which is a map of stings.
When exporting outbound rulesets, if one of the keys in the map is set to an empty string it will be ignored
by the export process. Example: properties = {"contact.Attempts" = ""}.

During the export process the value associated with the key is set to nil.
This custom exporter checks if a key has a value of nil and if it does sets it to an empty string so it is exported.
*/
func RuleSetPropertyResolver(configMap map[string]interface{}, exporters map[string]*ResourceExporter, resourceName string) error {
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
RuleSetSkillPropertyResolver
This property takes a key 'skills' with an array of skill ids wrapped into a string (Example: {'skills': '['skillIdHere']'} ).
This causes problems with the exporter because our export process expects id to map to a specific resource
and we have an array of attributes wrapped in a string.

This customer custom router will look at the skills array if present and resolve each string id find the appropriate resource out of the exporters and build a reference appropriately.
*/
func RuleSetSkillPropertyResolver(configMap map[string]interface{}, exporters map[string]*ResourceExporter, resourceName string) error {

	if exporter, ok := exporters["genesyscloud_routing_skill"]; ok {

		skillIDs, _ := configMap["skills"].(string)

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
				// DEVTOOLING-319: Outbound rulesets can reference skills that no longer exist. Plugin crash if we process a skill that doesn't exist in the skill map, making sure of its existence before proceeding.
				value, exists := exporter.SanitizedResourceMap[skillId]
				if exists {
					exportId = value.Name
					sanitisedSkillIds = append(sanitisedSkillIds, fmt.Sprintf("${genesyscloud_routing_skill.%s.id}", exportId))
				} else {
					log.Printf("Skill '%s' does not exist in the skill map.\n", skillId)
					sanitisedSkillIds = append(sanitisedSkillIds, fmt.Sprintf("skill_%s_not_found", skillId))
				}
			}

			jsonData, err := json.Marshal(sanitisedSkillIds)
			if err != nil {
				return fmt.Errorf("error converting sanitized skill ids array to JSON: %s", err)
			}
			configMap["skills"] = string(jsonData)
		}
	} else {
		return fmt.Errorf("unable to locate genesyscloud_routing_skill in the exporters array")
	}
	return nil
}

func FileContentHashResolver(configMap map[string]interface{}, filepath string) error {
	configMap["file_content_hash"] = fmt.Sprintf(`${filesha256(var.%s)}`, filepath)
	return nil
}

func CampaignStatusResolver(configMap map[string]interface{}, exporters map[string]*ResourceExporter, resourceName string) error {
	if configMap["campaign_status"] != "off" && configMap["campaign_status"] != "on" {
		configMap["campaign_status"] = "off"
	}

	return nil
}

func ReplyEmailAddressSelfReferenceRouteExporterResolver(configMap map[string]interface{}, exporters map[string]*ResourceExporter, resourceName string) error {

	routeId, _ := configMap["route_id"].(string)
	currentRouteReference := fmt.Sprintf("${genesyscloud_routing_email_route.%s.id}", resourceName)
	if routeId == currentRouteReference {
		configMap["self_reference_route"] = true
		configMap["route_id"] = nil
	}
	return nil
}

func ConditionValueResolver(configMap map[string]interface{}, exporters map[string]*ResourceExporter, resourceName string) error {
	if value := configMap["condition_value"]; value == nil {
		configMap["condition_value"] = 0
	}

	return nil
}
