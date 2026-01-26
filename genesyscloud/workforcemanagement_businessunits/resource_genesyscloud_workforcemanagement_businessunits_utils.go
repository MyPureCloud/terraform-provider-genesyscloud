package workforcemanagement_businessunits

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"strings"
)

/*
The resource_genesyscloud_workforcemanagement_businessunits_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getCreateWorkforcemanagementBusinessUnitRequestFromResourceData maps data from schema ResourceData object to a *platformclientv2.Createbusinessunitrequest
func getCreateWorkforcemanagementBusinessUnitRequestFromResourceData(d *schema.ResourceData) platformclientv2.Createbusinessunitrequest {
	return platformclientv2.Createbusinessunitrequest{
		Name:       platformclientv2.String(d.Get("name").(string)),
		Settings:   buildCreateBusinessUnitSettingsRequest(d.Get("settings").([]interface{})),
		DivisionId: resourcedata.GetNonZeroPointer[string](d, "division_id"),
	}
}

// getUpdateWorkforcemanagementBusinessUnitRequestFromResourceData maps data from schema ResourceData object to a *platformclientv2.Updatebusinessunitrequest
func getUpdateWorkforcemanagementBusinessUnitRequestFromResourceData(d *schema.ResourceData) platformclientv2.Updatebusinessunitrequest {
	return platformclientv2.Updatebusinessunitrequest{
		Name:       platformclientv2.String(d.Get("name").(string)),
		Settings:   buildUpdateBusinessUnitSettingsRequest(d.Get("settings").([]interface{})),
		DivisionId: resourcedata.GetNonZeroPointer[string](d, "division_id"),
	}
}

// buildBuShortTermForecastingSettings maps an []interface{} into a Genesys Cloud *platformclientv2.Bushorttermforecastingsettings
func buildBuShortTermForecastingSettings(buShortTermForecastingSettings []interface{}) *platformclientv2.Bushorttermforecastingsettings {
	if len(buShortTermForecastingSettings) == 0 {
		return nil
	}

	buShortTermForecastingSettingsMap, ok := buShortTermForecastingSettings[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkBuShortTermForecastingSettings platformclientv2.Bushorttermforecastingsettings
	sdkBuShortTermForecastingSettings.DefaultHistoryWeeks = platformclientv2.Int(buShortTermForecastingSettingsMap["default_history_weeks"].(int))

	return &sdkBuShortTermForecastingSettings
}

// buildSchedulerMessageTypeSeverities maps an []interface{} into a Genesys Cloud *[]platformclientv2.Schedulermessagetypeseverity
func buildSchedulerMessageTypeSeverities(schedulerMessageTypeSeverities []interface{}) *[]platformclientv2.Schedulermessagetypeseverity {
	if len(schedulerMessageTypeSeverities) == 0 {
		return nil
	}

	schedulerMessageTypeSeveritysSlice := make([]platformclientv2.Schedulermessagetypeseverity, 0)
	for _, schedulerMessageTypeSeverity := range schedulerMessageTypeSeverities {
		var sdkSchedulerMessageTypeSeverity platformclientv2.Schedulermessagetypeseverity
		schedulerMessageTypeSeveritysMap, ok := schedulerMessageTypeSeverity.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkSchedulerMessageTypeSeverity.VarType, schedulerMessageTypeSeveritysMap, "type")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSchedulerMessageTypeSeverity.Severity, schedulerMessageTypeSeveritysMap, "severity")

		schedulerMessageTypeSeveritysSlice = append(schedulerMessageTypeSeveritysSlice, sdkSchedulerMessageTypeSeverity)
	}

	return &schedulerMessageTypeSeveritysSlice
}

// buildWfmServiceGoalImpacts maps an []interface{} into a Genesys Cloud *[]platformclientv2.Wfmservicegoalimpact
func buildWfmServiceGoalImpact(wfmServiceGoalImpacts []interface{}) *platformclientv2.Wfmservicegoalimpact {
	if len(wfmServiceGoalImpacts) == 0 {
		return nil
	}

	wfmServiceGoalImpactsMap, ok := wfmServiceGoalImpacts[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkWfmServiceGoalImpact platformclientv2.Wfmservicegoalimpact
	increaseByPercent := wfmServiceGoalImpactsMap["increase_by_percent"].(float64)
	sdkWfmServiceGoalImpact.IncreaseByPercent = &increaseByPercent
	decreaseByPercent := wfmServiceGoalImpactsMap["decrease_by_percent"].(float64)
	sdkWfmServiceGoalImpact.DecreaseByPercent = &decreaseByPercent

	return &sdkWfmServiceGoalImpact
}

// buildWfmServiceGoalImpactSettings maps an []interface{} into a Genesys Cloud *[]platformclientv2.Wfmservicegoalimpactsettings
func buildWfmServiceGoalImpactSettings(wfmServiceGoalImpactSettings []interface{}) *platformclientv2.Wfmservicegoalimpactsettings {
	if len(wfmServiceGoalImpactSettings) == 0 {
		return nil
	}

	wfmServiceGoalImpactSettingsMap, ok := wfmServiceGoalImpactSettings[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkWfmServiceGoalImpactSettings platformclientv2.Wfmservicegoalimpactsettings
	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWfmServiceGoalImpactSettings.ServiceLevel, wfmServiceGoalImpactSettingsMap, "service_level", buildWfmServiceGoalImpact)
	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWfmServiceGoalImpactSettings.AverageSpeedOfAnswer, wfmServiceGoalImpactSettingsMap, "average_speed_of_answer", buildWfmServiceGoalImpact)
	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWfmServiceGoalImpactSettings.AbandonRate, wfmServiceGoalImpactSettingsMap, "abandon_rate", buildWfmServiceGoalImpact)

	return &sdkWfmServiceGoalImpactSettings
}

// buildBuSchedulingSettingsResponses maps an []interface{} into a Genesys Cloud *[]platformclientv2.Buschedulingsettingsrequest
func buildBuSchedulingSettings(buSchedulingSettings []interface{}) *platformclientv2.Buschedulingsettingsrequest {
	if len(buSchedulingSettings) == 0 {
		return nil
	}

	buSchedulingSettingsResponsesMap, ok := buSchedulingSettings[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkBuSchedulingSettingsRequest platformclientv2.Buschedulingsettingsrequest
	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBuSchedulingSettingsRequest.MessageSeverities, buSchedulingSettingsResponsesMap, "message_severities", buildSchedulerMessageTypeSeverities)

	// Handle sync_time_off_properties - it's a []interface{} in the schema but needs to be converted to *platformclientv2.Setwrappersynctimeoffproperty
	if syncTimeOffProps, ok := buSchedulingSettingsResponsesMap["sync_time_off_properties"].([]interface{}); ok && len(syncTimeOffProps) > 0 {
		syncTimeOffPropertiesList := make([]string, 0)
		for _, prop := range syncTimeOffProps {
			syncTimeOffPropertiesList = append(syncTimeOffPropertiesList, prop.(string))
		}
		sdkBuSchedulingSettingsRequest.SyncTimeOffProperties = &platformclientv2.Setwrappersynctimeoffproperty{
			Values: &syncTimeOffPropertiesList,
		}
	}

	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBuSchedulingSettingsRequest.ServiceGoalImpact, buSchedulingSettingsResponsesMap, "service_goal_impact", buildWfmServiceGoalImpactSettings)
	if allowWorkPlanPerMinuteGranularity, ok := buSchedulingSettingsResponsesMap["allow_work_plan_per_minute_granularity"].(bool); ok {
		sdkBuSchedulingSettingsRequest.AllowWorkPlanPerMinuteGranularity = platformclientv2.Bool(allowWorkPlanPerMinuteGranularity)
	}

	return &sdkBuSchedulingSettingsRequest
}

func buildWfmVersionedEntityMetadata(metadata []interface{}) *platformclientv2.Wfmversionedentitymetadata {
	if len(metadata) == 0 {
		return nil
	}

	metadataMap, ok := metadata[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Wfmversionedentitymetadata{
		Version: platformclientv2.Int(metadataMap["version"].(int)),
	}
}

// buildCreateBusinessUnitSettingsRequest maps an interface{} into a Genesys Cloud *platformclientv2.Createbusinessunitsettingsrequest
func buildCreateBusinessUnitSettingsRequest(businessUnitSettingsRequests []interface{}) *platformclientv2.Createbusinessunitsettingsrequest {
	if len(businessUnitSettingsRequests) == 0 {
		return nil
	}

	businessUnitSettingsResponsesMap, ok := businessUnitSettingsRequests[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkBusinessUnitSettingsRequest platformclientv2.Createbusinessunitsettingsrequest

	resourcedata.BuildSDKStringValueIfNotNil(&sdkBusinessUnitSettingsRequest.StartDayOfWeek, businessUnitSettingsResponsesMap, "start_day_of_week")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkBusinessUnitSettingsRequest.TimeZone, businessUnitSettingsResponsesMap, "time_zone")
	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBusinessUnitSettingsRequest.ShortTermForecasting, businessUnitSettingsResponsesMap, "short_term_forecasting", buildBuShortTermForecastingSettings)
	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBusinessUnitSettingsRequest.Scheduling, businessUnitSettingsResponsesMap, "scheduling", buildBuSchedulingSettings)

	return &sdkBusinessUnitSettingsRequest
}

// buildUpdateBusinessUnitSettingsRequest maps an interface{} into a Genesys Cloud *platformclientv2.Updatebusinessunitsettingsrequest
func buildUpdateBusinessUnitSettingsRequest(businessUnitSettingsRequests []interface{}) *platformclientv2.Updatebusinessunitsettingsrequest {
	if len(businessUnitSettingsRequests) == 0 {
		return nil
	}

	businessUnitSettingsResponsesMap, ok := businessUnitSettingsRequests[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkBusinessUnitSettingsRequest platformclientv2.Updatebusinessunitsettingsrequest
	resourcedata.BuildSDKStringValueIfNotNil(&sdkBusinessUnitSettingsRequest.StartDayOfWeek, businessUnitSettingsResponsesMap, "start_day_of_week")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkBusinessUnitSettingsRequest.TimeZone, businessUnitSettingsResponsesMap, "time_zone")
	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBusinessUnitSettingsRequest.ShortTermForecasting, businessUnitSettingsResponsesMap, "short_term_forecasting", buildBuShortTermForecastingSettings)
	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBusinessUnitSettingsRequest.Scheduling, businessUnitSettingsResponsesMap, "scheduling", buildBuSchedulingSettings)
	resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBusinessUnitSettingsRequest.Metadata, businessUnitSettingsResponsesMap, "metadata", buildWfmVersionedEntityMetadata)

	return &sdkBusinessUnitSettingsRequest
}

// flattenBuShortTermForecastingSettings maps a Genesys Cloud *[]platformclientv2.Bushorttermforecastingsettings into a []interface{}
func flattenBuShortTermForecastingSettings(buShortTermForecastingSettings *platformclientv2.Bushorttermforecastingsettings) []interface{} {
	buShortTermForecastingSettingsMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(buShortTermForecastingSettingsMap, "default_history_weeks", buShortTermForecastingSettings.DefaultHistoryWeeks)

	return []interface{}{buShortTermForecastingSettingsMap}
}

// flattenSchedulerMessageTypeSeverities maps a Genesys Cloud *[]platformclientv2.Schedulermessagetypeseverity into a []interface{}
func flattenSchedulerMessageTypeSeverities(schedulerMessageTypeSeveritys *[]platformclientv2.Schedulermessagetypeseverity) []interface{} {
	if len(*schedulerMessageTypeSeveritys) == 0 {
		return nil
	}

	var schedulerMessageTypeSeverityList []interface{}
	for _, schedulerMessageTypeSeverity := range *schedulerMessageTypeSeveritys {
		schedulerMessageTypeSeverityMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(schedulerMessageTypeSeverityMap, "type", schedulerMessageTypeSeverity.VarType)
		resourcedata.SetMapValueIfNotNil(schedulerMessageTypeSeverityMap, "severity", schedulerMessageTypeSeverity.Severity)

		schedulerMessageTypeSeverityList = append(schedulerMessageTypeSeverityList, schedulerMessageTypeSeverityMap)
	}

	return schedulerMessageTypeSeverityList
}

// flattenWfmServiceGoalImpacts maps a Genesys Cloud *[]platformclientv2.Wfmservicegoalimpact into a []interface{}
func flattenWfmServiceGoalImpact(wfmServiceGoalImpact *platformclientv2.Wfmservicegoalimpact) []interface{} {
	var wfmServiceGoalImpactList []interface{}

	wfmServiceGoalImpactMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(wfmServiceGoalImpactMap, "increase_by_percent", wfmServiceGoalImpact.IncreaseByPercent)
	resourcedata.SetMapValueIfNotNil(wfmServiceGoalImpactMap, "decrease_by_percent", wfmServiceGoalImpact.DecreaseByPercent)
	wfmServiceGoalImpactList = append(wfmServiceGoalImpactList, wfmServiceGoalImpactMap)

	return wfmServiceGoalImpactList
}

// flattenWfmServiceGoalImpactSettings maps a Genesys Cloud *[]platformclientv2.Wfmservicegoalimpactsettings into a []interface{}
func flattenWfmServiceGoalImpactSettings(wfmServiceGoalImpactSettings *platformclientv2.Wfmservicegoalimpactsettings) []interface{} {
	var wfmServiceGoalImpactSettingsList []interface{}
	wfmServiceGoalImpactSettingsMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(wfmServiceGoalImpactSettingsMap, "service_level", wfmServiceGoalImpactSettings.ServiceLevel, flattenWfmServiceGoalImpact)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(wfmServiceGoalImpactSettingsMap, "average_speed_of_answer", wfmServiceGoalImpactSettings.AverageSpeedOfAnswer, flattenWfmServiceGoalImpact)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(wfmServiceGoalImpactSettingsMap, "abandon_rate", wfmServiceGoalImpactSettings.AbandonRate, flattenWfmServiceGoalImpact)

	wfmServiceGoalImpactSettingsList = append(wfmServiceGoalImpactSettingsList, wfmServiceGoalImpactSettingsMap)
	return wfmServiceGoalImpactSettingsList
}

// flattenBuSchedulingSettingsResponse maps a Genesys Cloud *[]platformclientv2.Buschedulingsettingsresponse into a []interface{}
func flattenBuSchedulingSettingsResponse(buSchedulingSettingsResponse *platformclientv2.Buschedulingsettingsresponse) []interface{} {
	buSchedulingSettingsResponseMap := make(map[string]interface{})

	var buSchedulingSettingsResponseList []interface{}
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(buSchedulingSettingsResponseMap, "message_severities", buSchedulingSettingsResponse.MessageSeverities, flattenSchedulerMessageTypeSeverities)
	resourcedata.SetMapStringArrayValueIfNotNil(buSchedulingSettingsResponseMap, "sync_time_off_properties", buSchedulingSettingsResponse.SyncTimeOffProperties)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(buSchedulingSettingsResponseMap, "service_goal_impact", buSchedulingSettingsResponse.ServiceGoalImpact, flattenWfmServiceGoalImpactSettings)
	resourcedata.SetMapValueIfNotNil(buSchedulingSettingsResponseMap, "allow_work_plan_per_minute_granularity", buSchedulingSettingsResponse.AllowWorkPlanPerMinuteGranularity)
	buSchedulingSettingsResponseList = append(buSchedulingSettingsResponseList, buSchedulingSettingsResponseMap)

	return buSchedulingSettingsResponseList
}

// flattenWfmVersionedEntityMetadata maps a Genesys Cloud *[]platformclientv2.Wfmversionedentitymetadata into a []interface{}
func flattenWfmVersionedEntityMetadata(wfmVersionedEntityMetadata *platformclientv2.Wfmversionedentitymetadata) []interface{} {
	if wfmVersionedEntityMetadata == nil {
		return nil
	}
	wfmVersionedEntityMetadataMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(wfmVersionedEntityMetadataMap, "version", wfmVersionedEntityMetadata.Version)

	return []interface{}{wfmVersionedEntityMetadataMap}
}

// flattenBusinessUnitSettingsResponse maps a Genesys Cloud *platformclientv2.Businessunitsettingsresponse into a []interface{}
func flattenBusinessUnitSettingsResponse(businessUnitSettingsResponse *platformclientv2.Businessunitsettingsresponse) []interface{} {
	if businessUnitSettingsResponse == nil {
		return nil
	}

	businessUnitSettingsResponseMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(businessUnitSettingsResponseMap, "start_day_of_week", businessUnitSettingsResponse.StartDayOfWeek)
	resourcedata.SetMapValueIfNotNil(businessUnitSettingsResponseMap, "time_zone", businessUnitSettingsResponse.TimeZone)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(businessUnitSettingsResponseMap, "short_term_forecasting", businessUnitSettingsResponse.ShortTermForecasting, flattenBuShortTermForecastingSettings)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(businessUnitSettingsResponseMap, "scheduling", businessUnitSettingsResponse.Scheduling, flattenBuSchedulingSettingsResponse)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(businessUnitSettingsResponseMap, "metadata", businessUnitSettingsResponse.Metadata, flattenWfmVersionedEntityMetadata)

	return []interface{}{businessUnitSettingsResponseMap}
}

// GenerateWorkforcemanagementBusinessUnitResource generates a terraform resource string for testing
func GenerateWorkforcemanagementBusinessUnitResource(resourceLabel string, name string, settings string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		%s
	}
	`, ResourceType, resourceLabel, name, settings)
}

// GenerateWorkforcemanagementBusinessUnitSettings generates a terraform settings block for testing
func GenerateWorkforcemanagementBusinessUnitSettings(startDayOfWeek string, timeZone string, shortTermForecasting string, scheduling string) string {
	return fmt.Sprintf(`settings {
		start_day_of_week = "%s"
		time_zone = "%s"
		%s
		%s
	}
	`, startDayOfWeek, timeZone, shortTermForecasting, scheduling)
}

// GenerateWorkforcemanagementBusinessUnitShortTermForecasting generates a short_term_forecasting block for testing
func GenerateWorkforcemanagementBusinessUnitShortTermForecasting(defaultHistoryWeeks string) string {
	return fmt.Sprintf(`short_term_forecasting {
		default_history_weeks = %s
	}
	`, defaultHistoryWeeks)
}

// GenerateWorkforcemanagementBusinessUnitScheduling generates a scheduling block for testing
func GenerateWorkforcemanagementBusinessUnitScheduling(messageSeverities string, syncTimeOffProperties []string, serviceGoalImpact string, allowWorkPlanPerMinuteGranularity string) string {
	return fmt.Sprintf(`scheduling {
		%s
		sync_time_off_properties = [ %s ]
		%s
		allow_work_plan_per_minute_granularity = %s
	}
	`, messageSeverities, strings.Join(syncTimeOffProperties, ", "), serviceGoalImpact, allowWorkPlanPerMinuteGranularity)
}

// GenerateWorkforcemanagementBusinessUnitMessageSeverities generates message_severities block for testing
func GenerateWorkforcemanagementBusinessUnitMessageSeverities(messageType string, severity string) string {
	return fmt.Sprintf(`message_severities {
		type = "%s"
		severity = "%s"
	}
	`, messageType, severity)
}

// GenerateWorkforcemanagementBusinessUnitServiceGoalImpact generates service_goal_impact block for testing
func GenerateWorkforcemanagementBusinessUnitServiceGoalImpact(serviceLevel string, averageSpeedOfAnswer string, abandonRate string) string {
	return fmt.Sprintf(`service_goal_impact {
		service_level {
			%s
		}
		average_speed_of_answer {
			%s
		}
		abandon_rate {
			%s
		}
	}
	`, serviceLevel, averageSpeedOfAnswer, abandonRate)
}

// GenerateWorkforcemanagementBusinessUnitServiceGoalImpactValue generates service goal impact value block for testing
func GenerateWorkforcemanagementBusinessUnitServiceGoalImpactValue(increaseByPercent, decreaseByPercent string) string {
	return fmt.Sprintf(`increase_by_percent = %s
		decrease_by_percent = %s
	`, increaseByPercent, decreaseByPercent)
}
