package architect_schedules

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"regexp"
	"strings"
	"time"
)

const timeFormat = "2006-01-02T15:04:05.000000"

// Mapping rrule day abbreviations to full day names
var rruleDayMap = map[string]string{
	"MO": time.Monday.String(),
	"TU": time.Tuesday.String(),
	"WE": time.Wednesday.String(),
	"TH": time.Thursday.String(),
	"FR": time.Friday.String(),
	"SA": time.Saturday.String(),
	"SU": time.Sunday.String(),
}

func verifyStartDateConformsToRRule(dateTime time.Time, rrule string, scheduleName string) error {
	scheduleDays := getDaysFromRRule(rrule)
	if len(scheduleDays) == 0 {
		return nil
	}
	if !lists.SubStringInSlice(dateTime.Weekday().String(), scheduleDays) {
		return fmt.Errorf("invalid start date. %s is not specified in the rrule for schedule '%s'", dateTime.Weekday().String(), scheduleName)
	}
	return nil
}

// getDaysFromRRule parses the rrule to establish which weekdays are specified, if any.
func getDaysFromRRule(rrule string) []string {
	// Match BYDAY= followed by uppercase letters and commas
	re := regexp.MustCompile(`BYDAY=([A-Z,]+)(?:;|$)`)

	matches := re.FindStringSubmatch(rrule)
	if len(matches) < 2 {
		// No BYDAY parameter found
		return []string{}
	}

	// Split the matched days on commas
	days := strings.Split(matches[1], ",")

	// Convert abbreviations to full day names
	fullDays := make([]string, 0, len(days))
	for _, day := range days {
		if fullDay, ok := rruleDayMap[day]; ok {
			fullDays = append(fullDays, fullDay)
		}
	}

	return fullDays
}

// parseScheduleStartAndEndDateTimes parses and validates the start and end times from a Terraform resource data object.
// It extracts the "start" and "end" string values from the ResourceData, parses them according to the predefined timeFormat,
// and if a "rrule" is present, verifies that the start date conforms to the recurrence rule.
//
// Parameters:
//   - d: A pointer to schema.ResourceData containing the resource's state data
//
// Returns:
//   - start: A time.Time representing the parsed start datetime
//   - end: A time.Time representing the parsed end datetime
//   - err: An error if parsing fails or if the start date doesn't conform to the rrule
func parseScheduleStartAndEndDateTimes(d *schema.ResourceData) (start, end time.Time, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error in parseScheduleStartAndEndDateTimes: %w", err)
		}
	}()

	//The first parameter of the Parse() method specifies the date and time format/layout that should be used to interpret the second parameter.
	start, err = time.Parse(timeFormat, d.Get("start").(string))
	if err != nil {
		return
	}

	if rrule := d.Get("rrule").(string); rrule != "" {
		if err = verifyStartDateConformsToRRule(start, rrule, d.Get("name").(string)); err != nil {
			return
		}
	}

	end, err = time.Parse(timeFormat, d.Get("end").(string))
	return
}

func GenerateArchitectSchedulesResource(
	schedResourceLabel,
	name,
	divisionId,
	description,
	start,
	end,
	rrule string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		division_id = %s
		description = "%s"
		start = "%s"
		end = "%s"
		rrule = "%s"
	}
	`, ResourceType, schedResourceLabel, name, divisionId, description, start, end, rrule)
}
