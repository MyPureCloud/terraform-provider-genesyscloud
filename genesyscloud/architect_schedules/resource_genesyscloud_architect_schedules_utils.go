package architect_schedules

import (
	"errors"
	"fmt"
	"time"
)

const timeFormat = "2006-01-02T15:04:05.000000"

func verifyDateTimeIsWeekday(dateTime time.Time) error {
	if dateTime.Weekday() == time.Saturday || dateTime.Weekday() == time.Sunday {
		return errors.New("schedule start date cannot be a Saturday or Sunday")
	}
	return nil
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
