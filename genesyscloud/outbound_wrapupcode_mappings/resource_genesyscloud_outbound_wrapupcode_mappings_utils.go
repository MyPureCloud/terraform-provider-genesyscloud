package outbound_wrapupcode_mappings

import (
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// flattenOutboundWrapupCodeMappings maps a Genesys Cloud Wrapupcodemapping to a schema.Set
// We allow preexisting WUC mappings to exist in Genesys Cloud so we filter by the defined WUC ids
// in the CX as Code configuration
//
// `wrapupCodeFilter` should contain the existing WUCs in the org, any mappings not in this list will be ignored.
// This is because deleted wrap up codes stil retain their mappings but there is no practical reason for processing
// them in CX as Code.
func flattenOutboundWrapupCodeMappings(d *schema.ResourceData, sdkWrapupcodemapping *platformclientv2.Wrapupcodemapping, wrapupCodeFilter *[]string) *schema.Set {
	mappings := schema.NewSet(schema.HashResource(mappingResource), []interface{}{})
	schemaMappings := d.Get("mappings").(*schema.Set)
	schemaMappingsList := schemaMappings.List()
	forExport := false

	if _, ok := d.GetOk("placeholder"); !ok {
		forExport = true
	}

	for sdkId, sdkFlags := range *sdkWrapupcodemapping.Mapping {
		// If this is for export, we export all valid wuc mappings.
		// ie no need to check if it's defined in the tf config file.
		configuredMapping := false
		if !forExport {
			for _, sMap := range schemaMappingsList {
				sMapI := sMap.(map[string]interface{})
				if sMapI["wrapup_code_id"] == sdkId {
					configuredMapping = true
					break
				}
			}
		}
		if (!forExport && !configuredMapping) || !lists.ItemInSlice(sdkId, *wrapupCodeFilter) {
			continue
		}

		setSdkFlags := schema.NewSet(schema.HashSchema(flagsSchema), []interface{}{})
		for _, f := range sdkFlags {
			setSdkFlags.Add(f)
		}

		currentMap := make(map[string]interface{}, 0)
		currentMap["wrapup_code_id"] = sdkId
		currentMap["flags"] = setSdkFlags

		mappings.Add(currentMap)
	}

	return mappings
}

// buildWrapupCodeMappings builds the list of wrapupcode mappings from the schema object
func buildWrapupCodeMappings(d *schema.ResourceData) *map[string][]string {
	wrapupCodeMappings := make(map[string][]string, 0)
	if mappings := d.Get("mappings").(*schema.Set); mappings.Len() > 0 {
		for _, m := range mappings.List() {
			if mapping, ok := m.(map[string]interface{}); ok {
				id := mapping["wrapup_code_id"].(string)
				flags := lists.InterfaceListToStrings(mapping["flags"].(*schema.Set).List())
				wrapupCodeMappings[id] = flags
			}
		}
	}
	return &wrapupCodeMappings
}
