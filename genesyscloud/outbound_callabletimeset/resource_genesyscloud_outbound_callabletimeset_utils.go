package outbound_callabletimeset

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_outbound_callabletimeset_utils.go file contains helper methods to marshal and unmarshal data into formats consumable by Terraform and/or Genesys Cloud
*/

func getOutboundCallableTimesetFromResourceData(d *schema.ResourceData) platformclientv2.Callabletimeset {
	name := d.Get("name").(string)

	return platformclientv2.Callabletimeset{
		Name:          &name,
		CallableTimes: buildCallableTimes(d.Get("callable_times").(*schema.Set)),
	}
}

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

func buildCampaignTimeslots(campaigntimeslot *schema.Set) *[]platformclientv2.Campaigntimeslot {
	if campaigntimeslot == nil {
		return nil
	}
	sdkCampaigntimeslotSlice := make([]platformclientv2.Campaigntimeslot, 0)
	campaigntimeslotList := campaigntimeslot.List()
	for _, configcampaigntimeslot := range campaigntimeslotList {
		var sdkCampaigntimeslot platformclientv2.Campaigntimeslot

		campaigntimeslotMap := configcampaigntimeslot.(map[string]interface{})
		if startTime := campaigntimeslotMap["start_time"].(string); startTime != "" {
			sdkCampaigntimeslot.StartTime = &startTime
		}
		if stopTime := campaigntimeslotMap["stop_time"].(string); stopTime != "" {
			sdkCampaigntimeslot.StopTime = &stopTime
		}
		sdkCampaigntimeslot.Day = platformclientv2.Int(campaigntimeslotMap["day"].(int))

		sdkCampaigntimeslotSlice = append(sdkCampaigntimeslotSlice, sdkCampaigntimeslot)
	}
	return &sdkCampaigntimeslotSlice
}

func buildCallableTimes(callabletime *schema.Set) *[]platformclientv2.Callabletime {
	if callabletime == nil {
		return nil
	}
	sdkCallabletimeSlice := make([]platformclientv2.Callabletime, 0)
	callabletimeList := callabletime.List()
	for _, configcallabletime := range callabletimeList {
		var sdkCallabletime platformclientv2.Callabletime
		callabletimeMap := configcallabletime.(map[string]interface{})
		if timeSlots := callabletimeMap["time_slots"]; timeSlots != nil {
			sdkCallabletime.TimeSlots = buildCampaignTimeslots(timeSlots.(*schema.Set))
		}
		if timeZoneId := callabletimeMap["time_zone_id"].(string); timeZoneId != "" {
			sdkCallabletime.TimeZoneId = &timeZoneId
		}

		sdkCallabletimeSlice = append(sdkCallabletimeSlice, sdkCallabletime)
	}
	return &sdkCallabletimeSlice
}

func flattenCampaignTimeslots(campaigntimeslots []platformclientv2.Campaigntimeslot) *schema.Set {
	if len(campaigntimeslots) == 0 {
		return nil
	}

	campaigntimeslotSet := schema.NewSet(schema.HashResource(campaignTimeslotResource), []interface{}{})
	for _, campaigntimeslot := range campaigntimeslots {
		campaigntimeslotMap := make(map[string]interface{})

		if campaigntimeslot.StartTime != nil {
			campaigntimeslotMap["start_time"] = *campaigntimeslot.StartTime
		}
		if campaigntimeslot.StopTime != nil {
			campaigntimeslotMap["stop_time"] = *campaigntimeslot.StopTime
		}
		if campaigntimeslot.Day != nil {
			campaigntimeslotMap["day"] = *campaigntimeslot.Day
		}

		campaigntimeslotSet.Add(campaigntimeslotMap)
	}
	return campaigntimeslotSet
}

func flattenCallableTimes(callabletimes []platformclientv2.Callabletime) *schema.Set {
	if len(callabletimes) == 0 {
		return nil
	}

	callabletimeSet := schema.NewSet(schema.HashResource(timeSlotResource), []interface{}{})
	for _, callabletime := range callabletimes {
		callabletimeMap := make(map[string]interface{})

		if callabletime.TimeSlots != nil {
			callabletimeMap["time_slots"] = flattenCampaignTimeslots(*callabletime.TimeSlots)
		}
		if callabletime.TimeZoneId != nil {
			callabletimeMap["time_zone_id"] = *callabletime.TimeZoneId
		}

		callabletimeSet.Add(callabletimeMap)
	}
	return callabletimeSet
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
