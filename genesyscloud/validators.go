package genesyscloud

import (
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/ronanwatkins/terraform-plugin-sdk/v2/diag"
	"github.com/nyaruka/phonenumbers"
)

func validatePhoneNumber(number interface{}, _ cty.Path) diag.Diagnostics {
	if numberStr, ok := number.(string); ok {
		_, err := phonenumbers.Parse(numberStr, "US")
		if err != nil {
			return diag.Errorf("Failed to validate phone number %s: %s", numberStr, err)
		}
		return nil
	}
	return diag.Errorf("Phone number %v is not a string", number)
}

// Validates a date string is in the format yyyy-MM-dd
func validateDate(date interface{}, _ cty.Path) diag.Diagnostics {
	if dateStr, ok := date.(string); ok {
		_, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return diag.Errorf("Failed to parse date %s: %s", dateStr, err)
		}
		return nil
	}
	return diag.Errorf("Date %v is not a string", date)
}

// Validates a date string is in the format 2006-01-02T15:04:05.000000
func validateLocalDateTimes(date interface{}, _ cty.Path) diag.Diagnostics {
	if dateStr, ok := date.(string); ok {
		_, err := time.Parse("2006-01-02T15:04:05.000000", dateStr)
		if err != nil {
			return diag.Errorf("Failed to parse date %s: %s", dateStr, err)
		}
		return nil
	}
	return diag.Errorf("Date %v is not a string", date)
}
