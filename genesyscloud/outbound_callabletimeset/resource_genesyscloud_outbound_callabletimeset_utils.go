package outbound_callabletimeset

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
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
		if name := callabletimeMap["name"].(string); name != "" {
			sdkCallabletime.Name = &name
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

// flattenCallableTimes converts the API callable times to the schema set.
//
// `existing` is the previous state, or the planned values right after a create or update (nil when there is none,
// for example on import or export). It decides whether the `name` the API returned is written to state.
//
// When a callable time is created without a `name`, the API fills one in (the time zone id). `name` is Computed, so
// Terraform accepts that. The provider's consistency check does not: callable_times is a set, and the identity of each
// element includes `name`, so the element that was read back (with a name) does not match the planned one (without),
// and the apply fails after the check retries. To avoid that, the API's name is only written to state when the
// matching planned or previous element already had a name. Otherwise it is left empty.
func flattenCallableTimes(callabletimes []platformclientv2.Callabletime, existing *schema.Set) *schema.Set {
	if len(callabletimes) == 0 {
		return nil
	}

	var existingElements []map[string]interface{}
	if existing != nil {
		for _, e := range existing.List() {
			if m, ok := e.(map[string]interface{}); ok {
				existingElements = append(existingElements, m)
			}
		}
	}
	used := make([]bool, len(existingElements))

	callabletimeSet := schema.NewSet(schema.HashResource(timeSlotResource), []interface{}{})
	for _, callabletime := range callabletimes {
		callabletimeMap := make(map[string]interface{})

		var timeSlots *schema.Set
		if callabletime.TimeSlots != nil {
			timeSlots = flattenCampaignTimeslots(*callabletime.TimeSlots)
			callabletimeMap["time_slots"] = timeSlots
		}
		timeZoneId := ""
		if callabletime.TimeZoneId != nil {
			timeZoneId = *callabletime.TimeZoneId
			callabletimeMap["time_zone_id"] = timeZoneId
		}
		if callabletime.Name != nil {
			if keepApiCallableTimeName(existingElements, used, timeZoneId, timeSlots, *callabletime.Name) {
				callabletimeMap["name"] = *callabletime.Name
			}
		}

		callabletimeSet.Add(callabletimeMap)
	}
	return callabletimeSet
}

// keepApiCallableTimeName reports whether the name returned by the API should be written to state.
// The API callable time is matched to a planned or previous element with the same time zone and time slots. If that
// element has no name, the name was filled in by the API and is not written. If it has a name, or there is no matching
// element at all (import, export, or the callable time changed outside Terraform), the API's name is kept.
// Each existing element is matched at most once, so duplicates are paired one to one.
func keepApiCallableTimeName(existing []map[string]interface{}, used []bool, timeZoneId string, timeSlots *schema.Set, apiName string) bool {
	match := -1
	for i, candidate := range existing {
		if used[i] {
			continue
		}
		if existingZone, _ := candidate["time_zone_id"].(string); existingZone != timeZoneId {
			continue
		}
		existingSlots, _ := candidate["time_slots"].(*schema.Set)
		if !sameTimeSlots(existingSlots, timeSlots) {
			continue
		}
		if match == -1 {
			match = i
		}
		// Prefer the element whose name is the one the API returned, so duplicates are paired correctly.
		if existingName, _ := candidate["name"].(string); existingName == apiName {
			match = i
			break
		}
	}
	if match == -1 {
		return true
	}
	used[match] = true
	plannedName, _ := existing[match]["name"].(string)
	return plannedName != ""
}

func sameTimeSlots(a, b *schema.Set) bool {
	aLen, bLen := 0, 0
	if a != nil {
		aLen = a.Len()
	}
	if b != nil {
		bLen = b.Len()
	}
	if aLen == 0 && bLen == 0 {
		return true
	}
	if aLen != bLen {
		return false
	}
	return a.Equal(b)
}

func GenerateOutboundCallabletimeset(
	resourceLabel string,
	name string,
	nestedBlocks ...string) string {

	return fmt.Sprintf(`
		resource "genesyscloud_outbound_callabletimeset" "%s"{
			name = "%s"
			%s
		}
		`, resourceLabel, name, strings.Join(nestedBlocks, "\n"),
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
