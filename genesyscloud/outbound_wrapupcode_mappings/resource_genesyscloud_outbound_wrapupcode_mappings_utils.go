package outbound_wrapupcode_mappings

import (
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

// flattenOutboundWrapupCodeMappings maps objects and flags lists come back ordered differently than what is defined by the user in their config
// To avoid plan not empty errors, this function:
// checks that the maps/lists from the schema & sdk returned data are equivalent before returning the data in it's original order.
func flattenOutboundWrapupCodeMappings(d *schema.ResourceData, sdkWrapupcodemapping *platformclientv2.Wrapupcodemapping) []interface{} {
	mappings := make([]interface{}, 0)
	schemaMappings := d.Get("mappings").([]interface{})

	// If read is called from export function, placeholder field should not exist
	// In this case, dump whatever is returned from the API.
	if _, exists := d.GetOkExists("placeholder"); !exists {
		for sdkId, sdkFlags := range *sdkWrapupcodemapping.Mapping {
			currentMap := make(map[string]interface{}, 0)
			currentMap["wrapup_code_id"] = sdkId
			currentMap["flags"] = lists.StringListToInterfaceList(sdkFlags)
			mappings = append(mappings, currentMap)
		}
		return mappings
	}

	for _, m := range schemaMappings {
		if mMap, ok := m.(map[string]interface{}); ok {
			var schemaFlags []string
			if flags, ok := mMap["flags"].([]interface{}); ok {
				schemaFlags = lists.InterfaceListToStrings(flags)
			}
			for sdkId, sdkFlags := range *sdkWrapupcodemapping.Mapping {
				if mMap["wrapup_code_id"].(string) == sdkId {
					currentMap := make(map[string]interface{}, 0)
					currentMap["wrapup_code_id"] = sdkId
					if lists.AreEquivalent(schemaFlags, sdkFlags) {
						currentMap["flags"] = lists.StringListToInterfaceList(schemaFlags)
					} else {
						currentMap["flags"] = lists.StringListToInterfaceList(sdkFlags)
					}
					mappings = append(mappings, currentMap)
				}
			}
		}
	}
	return mappings
}

// buildWrapupCodeMappings builds the list of wrapupcode mappings from the schema object
func buildWrapupCodeMappings(d *schema.ResourceData) *map[string][]string {
	wrapupCodeMappings := make(map[string][]string, 0)
	if mappings := d.Get("mappings").([]interface{}); mappings != nil && len(mappings) > 0 {
		for _, m := range mappings {
			if mapping, ok := m.(map[string]interface{}); ok {
				id := mapping["wrapup_code_id"].(string)
				flags := lists.InterfaceListToStrings(mapping["flags"].([]interface{}))
				wrapupCodeMappings[id] = flags
			}
		}
	}
	return &wrapupCodeMappings
}
