package util

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
)

var (
	sdkConfig *platformclientv2.Configuration
	err       error
	// Map of SDK media type name to schema media type name
	utilizationMediaTypes = map[string]string{
		"call":     "call",
		"callback": "callback",
		"chat":     "chat",
		"email":    "email",
		"message":  "message",
	}
)

type MediaUtilization struct {
	MaximumCapacity         int32    `json:"maximumCapacity"`
	InterruptableMediaTypes []string `json:"interruptableMediaTypes"`
	IncludeNonAcd           bool     `json:"includeNonAcd"`
}

type LabelUtilization struct {
	MaximumCapacity      int32    `json:"maximumCapacity"`
	InterruptingLabelIds []string `json:"interruptingLabelIds"`
}

func GetUtilizationMediaTypes() map[string]string {
	return utilizationMediaTypes
}

func GetSdkUtilizationTypes() []string {
	types := make([]string, 0, len(utilizationMediaTypes))
	for t := range utilizationMediaTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func BuildHeaderParams(routingAPI *platformclientv2.RoutingApi) map[string]string {
	headerParams := make(map[string]string)

	for key := range routingAPI.Configuration.DefaultHeader {
		headerParams[key] = routingAPI.Configuration.DefaultHeader[key]
	}

	headerParams["Authorization"] = "Bearer " + routingAPI.Configuration.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	return headerParams
}

func FlattenUtilizationSetting(settings MediaUtilization) []interface{} {
	settingsMap := make(map[string]interface{})

	settingsMap["maximum_capacity"] = settings.MaximumCapacity
	settingsMap["include_non_acd"] = settings.IncludeNonAcd
	if settings.InterruptableMediaTypes != nil {
		settingsMap["interruptible_media_types"] = lists.StringListToSet(settings.InterruptableMediaTypes)
	}

	return []interface{}{settingsMap}
}

func FilterAndFlattenLabelUtilizations(labelUtilizations map[string]LabelUtilization, originalLabelUtilizations []interface{}) []interface{} {
	flattenedLabelUtilizations := make([]interface{}, 0)

	for _, originalLabelUtilization := range originalLabelUtilizations {
		originalLabelId := (originalLabelUtilization.(map[string]interface{}))["label_id"].(string)

		for currentLabelId, currentLabelUtilization := range labelUtilizations {
			if currentLabelId == originalLabelId {
				flattenedLabelUtilizations = append(flattenedLabelUtilizations, FlattenLabelUtilization(currentLabelId, currentLabelUtilization))
				delete(labelUtilizations, currentLabelId)
				break
			}
		}
	}

	return flattenedLabelUtilizations
}

func FlattenLabelUtilization(labelId string, labelUtilization LabelUtilization) map[string]interface{} {
	utilizationMap := make(map[string]interface{})

	utilizationMap["label_id"] = labelId
	utilizationMap["maximum_capacity"] = labelUtilization.MaximumCapacity
	if labelUtilization.InterruptingLabelIds != nil {
		utilizationMap["interrupting_label_ids"] = lists.StringListToSet(labelUtilization.InterruptingLabelIds)
	}

	return utilizationMap
}

func BuildSdkMediaUtilizations(d *schema.ResourceData) *map[string]platformclientv2.Mediautilization {
	settings := make(map[string]platformclientv2.Mediautilization)

	for sdkType, schemaType := range GetUtilizationMediaTypes() {
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

func BuildLabelUtilizationsRequest(labelUtilizations []interface{}) map[string]LabelUtilization {
	request := make(map[string]LabelUtilization)
	for _, labelUtilization := range labelUtilizations {
		labelUtilizationMap := labelUtilization.(map[string]interface{})
		interruptingLabelIds := lists.SetToStringList(labelUtilizationMap["interrupting_label_ids"].(*schema.Set))

		request[labelUtilizationMap["label_id"].(string)] = LabelUtilization{
			MaximumCapacity:      int32(labelUtilizationMap["maximum_capacity"].(int)),
			InterruptingLabelIds: *interruptingLabelIds,
		}
	}
	return request
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

func GenerateLabelUtilization(
	labelResource string,
	maxCapacity string,
	interruptingLabelResourceNames ...string) string {

	interruptingLabelResources := make([]string, 0)
	for _, resourceName := range interruptingLabelResourceNames {
		interruptingLabelResources = append(interruptingLabelResources, "genesyscloud_routing_utilization_label."+resourceName+".id")
	}

	return fmt.Sprintf(`label_utilizations {
		label_id = genesyscloud_routing_utilization_label.%s.id
		maximum_capacity = %s
		interrupting_label_ids = [%s]
	}
	`, labelResource, maxCapacity, strings.Join(interruptingLabelResources, ","))
}

func GenerateRoutingUtilizationLabelResource(resourceID string, name string, dependsOnResource string) string {
	dependsOn := ""

	if dependsOnResource != "" {
		dependsOn = fmt.Sprintf("depends_on=[genesyscloud_routing_utilization_label.%s]", dependsOnResource)
	}

	return fmt.Sprintf(`resource "genesyscloud_routing_utilization_label" "%s" {
		name = "%s"
		%s
	}
	`, resourceID, name, dependsOn)
}

func CheckIfLabelsAreEnabled() error { // remove once the feature is globally enabled
	if sdkConfig == nil {
		if sdkConfig, err = provider.AuthorizeSdk(); err != nil {
			log.Fatal(err)
		}
	}
	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	_, resp, _ := api.GetRoutingUtilizationLabels(100, 1, "", "")
	if resp.StatusCode == 501 {
		return fmt.Errorf("feature is not yet implemented in this org.")
	}
	return nil
}
