package genesyscloud

import (
	"fmt"
)

/*
The resource_genesyscloud_routing_queue object has the concept of bullseye ring with a member_groups attribute.
The routing team has overloaded the meaning of the member_groups so you can id and then define what "type" of id this is.
This causes problems with the exporter because our export process expects id to map to a specific resource.

This customer custom router will look at the member_group_type and resolve whether it is SKILLGROUP, GROUP type.  It will then
find the appropriate resource out of the exporters and build a reference appropriately.
*/
func MemberGroupsResolver(configMap map[string]interface{}, exporters map[string]*ResourceExporter) error {

	var exporter *ResourceExporter
	memberGroupType := configMap["member_group_type"]
	memberGroupID := configMap["member_group_id"].(string)

	switch memberGroupType {
	case "SKILLGROUP":
		exporter = exporters["genesyscloud_routing_skill_group"]
		exportId := (*exporter.SanitizedResourceMap[memberGroupID]).Name

		configMap["member_group_id"] = fmt.Sprintf("${genesyscloud_routing_skill_group.%s.id}", exportId)
	case "GROUP":
		exporter = exporters["genesyscloud_group"]
		exportId := (*exporter.SanitizedResourceMap[memberGroupID]).Name

		configMap["member_group_id"] = fmt.Sprintf("${genesyscloud_group.%s.id}", exportId)
	default:
		fmt.Printf("The memberGroupType %s cannot be located. Can not resolve to a reference attribute", memberGroupType)
	}

	return nil
}
