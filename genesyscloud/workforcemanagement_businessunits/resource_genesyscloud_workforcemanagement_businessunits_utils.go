package workforcemanagement_businessunits

import (
	"fmt"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

/*
The resource_genesyscloud_workforcemanagement_businessunits_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getCreateWorkforcemanagementBusinessUnitRequestFromResourceData maps data from schema ResourceData object to a *platformclientv2.Createbusinessunitrequest
func getCreateWorkforcemanagementBusinessUnitRequestFromResourceData(d *schema.ResourceData) platformclientv2.Createbusinessunitrequest {
	divisionId := d.Get("division_id").(string)
	return platformclientv2.Createbusinessunitrequest{
		Name:       platformclientv2.String(d.Get("name").(string)),
		Settings:   buildCreateBusinessUnitSettingsRequest(d.Get("settings").([]interface{})),
		DivisionId: &divisionId,
	}
}

// getUpdateWorkforcemanagementBusinessUnitRequestFromResourceData maps data from schema ResourceData object to a *platformclientv2.Updatebusinessunitrequest
func getUpdateWorkforcemanagementBusinessUnitRequestFromResourceData(d *schema.ResourceData) platformclientv2.Updatebusinessunitrequest {
	divisionId := d.Get("division_id").(string)

	return platformclientv2.Updatebusinessunitrequest{
		Name:       platformclientv2.String(d.Get("name").(string)),
		Settings:   buildUpdateBusinessUnitSettingsRequest(d.Get("settings").([]interface{})),
		DivisionId: &divisionId,
	}
}

// buildBuShortTermForecastingSettings maps an []interface{} into a Genesys Cloud *platformclientv2.Bushorttermforecastingsettings
func buildBuShortTermForecastingSettings(buShortTermForecastingSettings []interface{}) *platformclientv2.Bushorttermforecastingsettings {
	buShortTermForecastingSettingsSlice := make([]platformclientv2.Bushorttermforecastingsettings, 0)
	for _, buShortTermForecastingSettings := range buShortTermForecastingSettings {
		var sdkBuShortTermForecastingSettings platformclientv2.Bushorttermforecastingsettings
		buShortTermForecastingSettingsMap, ok := buShortTermForecastingSettings.(map[string]interface{})
		if !ok {
			continue
		}

		sdkBuShortTermForecastingSettings.DefaultHistoryWeeks = platformclientv2.Int(buShortTermForecastingSettingsMap["default_history_weeks"].(int))

		buShortTermForecastingSettingsSlice = append(buShortTermForecastingSettingsSlice, sdkBuShortTermForecastingSettings)
	}

	return &buShortTermForecastingSettingsSlice[0]
}

// buildSchedulerMessageTypeSeverities maps an []interface{} into a Genesys Cloud *[]platformclientv2.Schedulermessagetypeseverity
func buildSchedulerMessageTypeSeverities(schedulerMessageTypeSeverities []interface{}) *[]platformclientv2.Schedulermessagetypeseverity {
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
	wfmServiceGoalImpactsSlice := make([]platformclientv2.Wfmservicegoalimpact, 0)
	for _, wfmServiceGoalImpact := range wfmServiceGoalImpacts {
		var sdkWfmServiceGoalImpact platformclientv2.Wfmservicegoalimpact
		wfmServiceGoalImpactsMap, ok := wfmServiceGoalImpact.(map[string]interface{})
		if !ok {
			continue
		}

		increaseByPercent := wfmServiceGoalImpactsMap["increase_by_percent"].(float64)
		sdkWfmServiceGoalImpact.IncreaseByPercent = &increaseByPercent
		decreaseByPercent := wfmServiceGoalImpactsMap["decrease_by_percent"].(float64)
		sdkWfmServiceGoalImpact.DecreaseByPercent = &decreaseByPercent

		wfmServiceGoalImpactsSlice = append(wfmServiceGoalImpactsSlice, sdkWfmServiceGoalImpact)
	}

	if len(wfmServiceGoalImpactsSlice) == 0 {
		return nil
	}
	return &wfmServiceGoalImpactsSlice[0]
}

// buildWfmServiceGoalImpactSettings maps an []interface{} into a Genesys Cloud *[]platformclientv2.Wfmservicegoalimpactsettings
func buildWfmServiceGoalImpactSettings(wfmServiceGoalImpactSettings []interface{}) *platformclientv2.Wfmservicegoalimpactsettings {
	wfmServiceGoalImpactSettingsSlice := make([]platformclientv2.Wfmservicegoalimpactsettings, 0)
	for _, wfmServiceGoalImpactSettings := range wfmServiceGoalImpactSettings {
		var sdkWfmServiceGoalImpactSettings platformclientv2.Wfmservicegoalimpactsettings
		wfmServiceGoalImpactSettingsMap, ok := wfmServiceGoalImpactSettings.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWfmServiceGoalImpactSettings.ServiceLevel, wfmServiceGoalImpactSettingsMap, "service_level", buildWfmServiceGoalImpact)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWfmServiceGoalImpactSettings.AverageSpeedOfAnswer, wfmServiceGoalImpactSettingsMap, "average_speed_of_answer", buildWfmServiceGoalImpact)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkWfmServiceGoalImpactSettings.AbandonRate, wfmServiceGoalImpactSettingsMap, "abandon_rate", buildWfmServiceGoalImpact)

		wfmServiceGoalImpactSettingsSlice = append(wfmServiceGoalImpactSettingsSlice, sdkWfmServiceGoalImpactSettings)
	}

	return &wfmServiceGoalImpactSettingsSlice[0] // TODO is this right
}

// buildBuSchedulingSettingsResponses maps an []interface{} into a Genesys Cloud *[]platformclientv2.Buschedulingsettingsrequest
func buildBuSchedulingSettings(buSchedulingSettingsResponses []interface{}) *platformclientv2.Buschedulingsettingsrequest {
	buSchedulingSettingsResponsesSlice := make([]platformclientv2.Buschedulingsettingsrequest, 0)
	for _, buSchedulingSettingsResponse := range buSchedulingSettingsResponses {
		var sdkBuSchedulingSettingsResponse platformclientv2.Buschedulingsettingsrequest
		buSchedulingSettingsResponsesMap, ok := buSchedulingSettingsResponse.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBuSchedulingSettingsResponse.MessageSeverities, buSchedulingSettingsResponsesMap, "message_severities", buildSchedulerMessageTypeSeverities)

		// Handle sync_time_off_properties - it's a []interface{} in the schema but needs to be converted to *platformclientv2.Setwrappersynctimeoffproperty
		if syncTimeOffProps, ok := buSchedulingSettingsResponsesMap["sync_time_off_properties"].([]interface{}); ok && len(syncTimeOffProps) > 0 {
			syncTimeOffPropertiesList := make([]string, 0)
			for _, prop := range syncTimeOffProps {
				syncTimeOffPropertiesList = append(syncTimeOffPropertiesList, prop.(string))
			}
			sdkBuSchedulingSettingsResponse.SyncTimeOffProperties = &platformclientv2.Setwrappersynctimeoffproperty{
				Values: &syncTimeOffPropertiesList,
			}
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBuSchedulingSettingsResponse.ServiceGoalImpact, buSchedulingSettingsResponsesMap, "service_goal_impact", buildWfmServiceGoalImpactSettings)
		if allowWorkPlanPerMinuteGranularity, ok := buSchedulingSettingsResponsesMap["allow_work_plan_per_minute_granularity"].(bool); ok {
			sdkBuSchedulingSettingsResponse.AllowWorkPlanPerMinuteGranularity = platformclientv2.Bool(allowWorkPlanPerMinuteGranularity)
		}

		buSchedulingSettingsResponsesSlice = append(buSchedulingSettingsResponsesSlice, sdkBuSchedulingSettingsResponse)
	}

	return &buSchedulingSettingsResponsesSlice[0]
}

// buildCreateBusinessUnitSettingsRequest maps an interface{} into a Genesys Cloud *platformclientv2.Createbusinessunitsettingsrequest
func buildCreateBusinessUnitSettingsRequest(businessUnitSettingsResponses []interface{}) *platformclientv2.Createbusinessunitsettingsrequest {
	businessUnitSettingsResponsesSlice := make([]platformclientv2.Createbusinessunitsettingsrequest, 0)
	for _, businessUnitSettingsResponse := range businessUnitSettingsResponses {
		var sdkBusinessUnitSettingsResponse platformclientv2.Createbusinessunitsettingsrequest
		businessUnitSettingsResponsesMap, ok := businessUnitSettingsResponse.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkBusinessUnitSettingsResponse.StartDayOfWeek, businessUnitSettingsResponsesMap, "start_day_of_week")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkBusinessUnitSettingsResponse.TimeZone, businessUnitSettingsResponsesMap, "time_zone")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBusinessUnitSettingsResponse.ShortTermForecasting, businessUnitSettingsResponsesMap, "short_term_forecasting", buildBuShortTermForecastingSettings)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBusinessUnitSettingsResponse.Scheduling, businessUnitSettingsResponsesMap, "scheduling", buildBuSchedulingSettings)

		businessUnitSettingsResponsesSlice = append(businessUnitSettingsResponsesSlice, sdkBusinessUnitSettingsResponse)
	}

	return &businessUnitSettingsResponsesSlice[0]
}

// buildCreateBusinessUnitSettingsRequest maps an interface{} into a Genesys Cloud *platformclientv2.Createbusinessunitsettingsrequest
func buildUpdateBusinessUnitSettingsRequest(businessUnitSettingsResponses []interface{}) *platformclientv2.Updatebusinessunitsettingsrequest {
	businessUnitSettingsResponsesSlice := make([]platformclientv2.Updatebusinessunitsettingsrequest, 0)
	for _, businessUnitSettingsResponse := range businessUnitSettingsResponses {
		var sdkBusinessUnitSettingsResponse platformclientv2.Updatebusinessunitsettingsrequest
		businessUnitSettingsResponsesMap, ok := businessUnitSettingsResponse.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkBusinessUnitSettingsResponse.StartDayOfWeek, businessUnitSettingsResponsesMap, "start_day_of_week")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkBusinessUnitSettingsResponse.TimeZone, businessUnitSettingsResponsesMap, "time_zone")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBusinessUnitSettingsResponse.ShortTermForecasting, businessUnitSettingsResponsesMap, "short_term_forecasting", buildBuShortTermForecastingSettings)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBusinessUnitSettingsResponse.Scheduling, businessUnitSettingsResponsesMap, "scheduling", buildBuSchedulingSettings)

		businessUnitSettingsResponsesSlice = append(businessUnitSettingsResponsesSlice, sdkBusinessUnitSettingsResponse)
	}

	return &businessUnitSettingsResponsesSlice[0]
}

// flattenBuShortTermForecastingSettings maps a Genesys Cloud *[]platformclientv2.Bushorttermforecastingsettings into a []interface{}
func flattenBuShortTermForecastingSettings(buShortTermForecastingSettings *platformclientv2.Bushorttermforecastingsettings) []interface{} {
	var buShortTermForecastingSettingsList []interface{}
	buShortTermForecastingSettingsMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(buShortTermForecastingSettingsMap, "default_history_weeks", buShortTermForecastingSettings.DefaultHistoryWeeks)
	buShortTermForecastingSettingsList = append(buShortTermForecastingSettingsList, buShortTermForecastingSettingsMap)

	return buShortTermForecastingSettingsList
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

	var wfmVersionedEntityMetadataList []interface{}
	wfmVersionedEntityMetadataMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(wfmVersionedEntityMetadataMap, "version", wfmVersionedEntityMetadata.Version)
	if wfmVersionedEntityMetadata.ModifiedBy != nil {
		wfmVersionedEntityMetadataMap["modified_by_id"] = *wfmVersionedEntityMetadata.ModifiedBy.Id
	}
	resourcedata.SetMapValueIfNotNil(wfmVersionedEntityMetadataMap, "date_modified", wfmVersionedEntityMetadata.DateModified)
	if wfmVersionedEntityMetadata.CreatedBy != nil {
		wfmVersionedEntityMetadataMap["created_by_id"] = *wfmVersionedEntityMetadata.CreatedBy.Id
	}
	resourcedata.SetMapValueIfNotNil(wfmVersionedEntityMetadataMap, "date_created", wfmVersionedEntityMetadata.DateCreated)

	wfmVersionedEntityMetadataList = append(wfmVersionedEntityMetadataList, wfmVersionedEntityMetadataMap)

	return wfmVersionedEntityMetadataList
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
func GenerateWorkforcemanagementBusinessUnitResource(resourceLabel string, name string, divisionId string, settings string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		division_id = %s
		%s
	}
	`, ResourceName, resourceLabel, name, divisionId, settings)
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
func GenerateWorkforcemanagementBusinessUnitShortTermForecasting(defaultHistoryWeeks int) string {
	return fmt.Sprintf(`short_term_forecasting {
		default_history_weeks = %d
	}
	`, defaultHistoryWeeks)
}

// GenerateWorkforcemanagementBusinessUnitScheduling generates a scheduling block for testing
func GenerateWorkforcemanagementBusinessUnitScheduling(messageSeverities string, syncTimeOffProperties string, serviceGoalImpact string, allowWorkPlanPerMinuteGranularity string) string {
	return fmt.Sprintf(`scheduling {
		%s
		%s
		%s
		%s
	}
	`, messageSeverities, syncTimeOffProperties, serviceGoalImpact, allowWorkPlanPerMinuteGranularity)
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
func GenerateWorkforcemanagementBusinessUnitServiceGoalImpactValue(increaseByPercent float64, decreaseByPercent float64) string {
	return fmt.Sprintf(`increase_by_percent = %.2f
		decrease_by_percent = %.2f
	`, increaseByPercent, decreaseByPercent)
}
