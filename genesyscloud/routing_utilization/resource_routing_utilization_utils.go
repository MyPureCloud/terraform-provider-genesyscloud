package routing_utilization

import (
	"fmt"
	"sort"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func BuildSdkMediaUtilizations(d *schema.ResourceData) *map[string]platformclientv2.Mediautilization {
	settings := make(map[string]platformclientv2.Mediautilization)

	for sdkType, schemaType := range UtilizationMediaTypes {
		mediaSettings := d.Get(schemaType).([]interface{})
		if mediaSettings != nil && len(mediaSettings) > 0 {
			settings[sdkType] = BuildSdkMediaUtilization(mediaSettings)
		}
	}

	return &settings
}

func BuildSdkMediaUtilization(settings []interface{}) platformclientv2.Mediautilization {
	settingsMap := settings[0].(map[string]interface{})

	maxCapacity := settingsMap["maximum_capacity"].(int)
	includeNonAcd := settingsMap["include_non_acd"].(bool)

	// Optional
	interruptableMediaTypes := &[]string{}
	if types, ok := settingsMap["interruptible_media_types"]; ok {
		interruptableMediaTypes = lists.SetToStringList(types.(*schema.Set))
	}

	return platformclientv2.Mediautilization{
		MaximumCapacity:         &maxCapacity,
		IncludeNonAcd:           &includeNonAcd,
		InterruptableMediaTypes: interruptableMediaTypes,
	}
}

func BuildSdkLabelUtilizations(labelUtilizations []interface{}) *map[string]platformclientv2.Labelutilizationrequest {
	request := make(map[string]platformclientv2.Labelutilizationrequest)

	for _, labelUtilization := range labelUtilizations {
		labelUtilizationMap := labelUtilization.(map[string]interface{})
		maxCapacity := labelUtilizationMap["maximum_capacity"].(int)
		interruptingLabelIds := lists.SetToStringList(labelUtilizationMap["interrupting_label_ids"].(*schema.Set))

		request[labelUtilizationMap["label_id"].(string)] = platformclientv2.Labelutilizationrequest{
			MaximumCapacity:      &maxCapacity,
			InterruptingLabelIds: interruptingLabelIds,
		}
	}

	return &request
}

func FlattenMediaUtilization(mediaUtilization platformclientv2.Mediautilization) []interface{} {
	settingsMap := make(map[string]interface{})

	settingsMap["maximum_capacity"] = mediaUtilization.MaximumCapacity
	settingsMap["include_non_acd"] = mediaUtilization.IncludeNonAcd
	if mediaUtilization.InterruptableMediaTypes != nil {
		settingsMap["interruptible_media_types"] = lists.StringListToSet(*mediaUtilization.InterruptableMediaTypes)
	}

	return []interface{}{settingsMap}
}

func FilterAndFlattenLabelUtilizations(labelUtilizations map[string]platformclientv2.Labelutilizationresponse, originalLabelUtilizations []interface{}) []interface{} {
	flattenedLabelUtilizations := make([]interface{}, 0)

	for _, originalLabelUtilization := range originalLabelUtilizations {
		originalLabelId := (originalLabelUtilization.(map[string]interface{}))["label_id"].(string)

		for currentLabelId, currentLabelUtilization := range labelUtilizations {
			if currentLabelId == originalLabelId {
				flattenedLabelUtilizations = append(flattenedLabelUtilizations, flattenLabelUtilization(currentLabelId, currentLabelUtilization))
				delete(labelUtilizations, currentLabelId)
				break
			}
		}
	}

	return flattenedLabelUtilizations
}

func flattenLabelUtilization(labelId string, labelUtilization platformclientv2.Labelutilizationresponse) map[string]interface{} {
	utilizationMap := make(map[string]interface{})

	utilizationMap["label_id"] = labelId
	utilizationMap["maximum_capacity"] = labelUtilization.MaximumCapacity
	if labelUtilization.InterruptingLabelIds != nil {
		utilizationMap["interrupting_label_ids"] = lists.StringListToSet(*labelUtilization.InterruptingLabelIds)
	}

	return utilizationMap
}

func GenerateRoutingUtilMediaType(
	mediaType string,
	maxCapacity string,
	includeNonAcd string,
	interruptTypes ...string) string {
	return fmt.Sprintf(`%s {
		maximum_capacity = %s
		include_non_acd = %s
		interruptible_media_types = [%s]
	}
	`, mediaType, maxCapacity, includeNonAcd, strings.Join(interruptTypes, ","))
}

func getSdkUtilizationTypes() []string {
	types := make([]string, 0, len(UtilizationMediaTypes))
	for t := range UtilizationMediaTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}
