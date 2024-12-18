package architect_schedules

import (
	"fmt"
	"regexp"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
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
