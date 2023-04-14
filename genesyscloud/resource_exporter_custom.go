package genesyscloud

import (
	"fmt"
	"os"
	"path"
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
		if err := downloadAudioFile(fullPath, data.FileName, data.MediaUri); err != nil {
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

	if err := downloadAudioFile(fullPath, exportFileName, url); err != nil {
		return err
	}

	filePath := path.Join(subDirectory, exportFileName)
	// Update filepath field in configMap to point to exported script file
	configMap["filepath"] = filePath

	// TODO : Figure out if we can get the " to stop being dumped as \" in the export file
	//fileContentHash := `${filesha256("` + filePath + `")}`
	//configMap["file_content_hash"] = fileContentHash

	return err
}
