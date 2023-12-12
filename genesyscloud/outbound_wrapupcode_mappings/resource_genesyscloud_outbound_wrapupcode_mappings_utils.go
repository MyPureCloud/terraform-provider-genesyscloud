package outbound_wrapupcode_mappings

import (
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v116/platformclientv2"
)

// flattenOutboundWrapupCodeMappings maps objects and flags lists come back ordered differently than what is defined by the user in their config
// To avoid plan not empty errors, this function:
// checks that the maps/lists from the schema & sdk returned data are equivalent before returning the data in its original order.
//
// `wrapupCodeFilter` should contain the existing WUCs in the org, any mappings not in this list will be ignored.
// This is because deleted wrap up codes stil retain their mappings but there is no practical reason for processing
// them in CX as Code.
func flattenOutboundWrapupCodeMappings(d *schema.ResourceData, sdkWrapupcodemapping *platformclientv2.Wrapupcodemapping, wrapupCodeFilter *[]string) []interface{} {
	mappings := make([]interface{}, 0)
	schemaMappings := d.Get("mappings").([]interface{})

	// If read is called from export function, there's no need
	// to match orders. Just dump what's returned by the API
	if len(schemaMappings) == 0 {
		for sdkId, sdkFlags := range *sdkWrapupcodemapping.Mapping {
			if !lists.ItemInSlice(sdkId, *wrapupCodeFilter) {
				continue
			}

			currentMap := make(map[string]interface{}, 0)
			currentMap["wrapup_code_id"] = sdkId
			currentMap["flags"] = lists.StringListToInterfaceList(sdkFlags)

			mappings = append(mappings, currentMap)
		}
		return mappings
	}

	// flatten the wrapupcode mappings considering the order from the resource data
	for _, m := range schemaMappings {
		if mMap, ok := m.(map[string]interface{}); ok {
			var schemaFlags []string
			if flags, ok := mMap["flags"].([]interface{}); ok {
				schemaFlags = lists.InterfaceListToStrings(flags)
			}
			for sdkId, sdkFlags := range *sdkWrapupcodemapping.Mapping {
				if !lists.ItemInSlice(sdkId, *wrapupCodeFilter) {
					continue
				}

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
	if mappings := d.Get("mappings").([]interface{}); len(mappings) > 0 {
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
