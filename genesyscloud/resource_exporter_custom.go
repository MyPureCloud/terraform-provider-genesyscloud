package genesyscloud

import (
	"fmt"
	"os"
	"path"
	"strings"
	"encoding/json"
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

func ArchitectPromptAudioResolver(promptId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}) error {
	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	audioDataList, err := getArchitectPromptAudioData(promptId, meta)
	if err != nil || len(audioDataList) == 0 {
		return err
	}

	for _, data := range audioDataList {
		if err := downloadExportFile(fullPath, data.FileName, data.MediaUri); err != nil {
			return err
		}
	}
	updateFilenamesInExportConfigMap(configMap, audioDataList, subDirectory)
	return nil
}

func ScriptResolver(scriptId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}) error {
	exportFileName := fmt.Sprintf("script-%s.json", scriptId)

	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	url, err := getScriptExportUrl(scriptId, meta)
	if err != nil {
		return err
	}

	if err := downloadExportFile(fullPath, exportFileName, url); err != nil {
		return err
	}

	// Update filepath field in configMap to point to exported script file
	configMap["filepath"] = path.Join(subDirectory, exportFileName)

	configMap["file_content_hash"] = fmt.Sprintf(`${filesha256("%s")}`, path.Join(subDirectory, exportFileName))

	return err
}
