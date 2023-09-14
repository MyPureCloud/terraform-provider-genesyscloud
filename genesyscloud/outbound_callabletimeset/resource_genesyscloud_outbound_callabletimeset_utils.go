package outbound_callabletimeset

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

/*
The resource_genesyscloud_outbound_callabletimeset_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getOutboundCallabletimesetFromResourceData maps data from schema ResourceData object to a platformclientv2.Callabletimeset
func getOutboundCallabletimesetFromResourceData(d *schema.ResourceData) platformclientv2.Callabletimeset {
	name := d.Get("name").(string)

	return platformclientv2.Callabletimeset{
		Name:          &name,
		CallableTimes: buildCallableTimes(d.Get("callable_times").([]interface{})),
	}
}

// buildCampaignTimeSlots maps an []interface{} into a Genesys Cloud *[]platformclientv2.Campaigntimeslot
func buildCampaignTimeSlots(campaignTimeSlots []interface{}) *[]platformclientv2.Campaigntimeslot {
	campaignTimeSlotsSlice := make([]platformclientv2.Campaigntimeslot, 0)
	for _, campaignTimeSlot := range campaignTimeSlots {
		var sdkCampaignTimeSlot platformclientv2.Campaigntimeslot
		campaignTimeSlotsMap, ok := campaignTimeSlot.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkCampaignTimeSlot.StartTime, campaignTimeSlotsMap, "start_time")

		resourcedata.BuildSDKStringValueIfNotNil(&sdkCampaignTimeSlot.StopTime, campaignTimeSlotsMap, "stop_time")

		sdkCampaignTimeSlot.Day = platformclientv2.Int(campaignTimeSlotsMap["day"].(int))

		campaignTimeSlotsSlice = append(campaignTimeSlotsSlice, sdkCampaignTimeSlot)
	}

	return &campaignTimeSlotsSlice
}

// flattenCampaignTimeSlots maps a Genesys Cloud *[]platformclientv2.Campaigntimeslot into a []interface{}
func flattenCampaignTimeSlots(campaignTimeSlots *[]platformclientv2.Campaigntimeslot) []interface{} {
	if len(*campaignTimeSlots) == 0 {
		return nil
	}

	var campaignTimeSlotList []interface{}
	for _, campaignTimeSlot := range *campaignTimeSlots {
		campaignTimeSlotMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(campaignTimeSlotMap, "start_time", campaignTimeSlot.StartTime)

		resourcedata.SetMapValueIfNotNil(campaignTimeSlotMap, "stop_time", campaignTimeSlot.StopTime)

		resourcedata.SetMapValueIfNotNil(campaignTimeSlotMap, "day", campaignTimeSlot.Day)

		campaignTimeSlotList = append(campaignTimeSlotList, campaignTimeSlotMap)
	}

	return campaignTimeSlotList
}

// buildCallableTimes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Callabletime
func buildCallableTimes(callableTimes []interface{}) *[]platformclientv2.Callabletime {
	callableTimesSlice := make([]platformclientv2.Callabletime, 0)
	for _, callableTime := range callableTimes {
		var sdkCallableTime platformclientv2.Callabletime
		callableTimesMap, ok := callableTime.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkCallableTime.TimeZoneId, callableTimesMap, "time_zone_id")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkCallableTime.TimeSlots, callableTimesMap, "time_slots", buildCampaignTimeSlots)

		callableTimesSlice = append(callableTimesSlice, sdkCallableTime)
	}

	return &callableTimesSlice
}

// flattenCallableTimes maps a Genesys Cloud *[]platformclientv2.Callabletime into a []interface{}
func flattenCallableTimes(callableTimes *[]platformclientv2.Callabletime) []interface{} {
	if len(*callableTimes) == 0 {
		return nil
	}

	trimTime(callableTimes)
	var callableTimeList []interface{}
	for _, callableTime := range *callableTimes {
		callableTimeMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(callableTimeMap, "time_slots", callableTime.TimeSlots, flattenCampaignTimeSlots)

		resourcedata.SetMapValueIfNotNil(callableTimeMap, "time_zone_id", callableTime.TimeZoneId)

		callableTimeList = append(callableTimeList, callableTimeMap)
	}

	return callableTimeList
}

// This function will remove the milliseconds from the callable times
func trimTime(values *[]platformclientv2.Callabletime) {
	for _, value := range *values {
		for _, slot := range *value.TimeSlots {
			startTime := *slot.StartTime
			*slot.StartTime = startTime[:8]

			stopTime := *slot.StopTime
			*slot.StopTime = stopTime[:8]
		}
	}
}

func GenerateOutboundCallabletimeset(
	resourceId string,
	name string,
	nestedBlocks ...string) string {

	return fmt.Sprintf(`
		resource "genesyscloud_outbound_callabletimeset" "%s"{
			name = "%s"
			%s
		}
		`, resourceId, name, strings.Join(nestedBlocks, "\n"),
	)
}

func GenerateCallableTimesBlock(
	timeZoneID string,
	attrs ...string) string {
	return fmt.Sprintf(`
		callable_times {
			time_zone_id = "%s"
			%s
		}
	`, timeZoneID, strings.Join(attrs, "\n"))
}

func GenerateTimeSlotsBlock(
	startTime string,
	stopTime string,
	day string) string {
	return fmt.Sprintf(`
		time_slots {
			start_time = "%s"
			stop_time = "%s"
			day = %s
		}
	`, startTime, stopTime, day)
}
