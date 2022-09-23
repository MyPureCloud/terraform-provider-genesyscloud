package genesyscloud

import (
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
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
		_, err := time.Parse(resourcedata.DateParseFormat, dateStr)
		if err != nil {
			return diag.Errorf("Failed to parse date %s: %s", dateStr, err)
		}
		return nil
	}
	return diag.Errorf("Date %v is not a string", date)
}

// Validates a date string is in the format 2006-01-02T15:04Z
func validateDateTime(date interface{}, _ cty.Path) diag.Diagnostics {
	if dateStr, ok := date.(string); ok {
		_, err := time.Parse("2006-01-02T15:04Z", dateStr)
		if err != nil {
			return diag.Errorf("Failed to parse date %s: %s", dateStr, err)
		}
		return nil
	}
	return diag.Errorf("Date %v is not a string", date)
}

func validateTime(time interface{}, _ cty.Path) diag.Diagnostics {
	timeStr := time.(string)
	if len(timeStr) > 9 {
		timeStr = timeStr[:8]
	}
	if valid, _ := regexp.MatchString("^(0?[0-9]|1?[0-9]|2[0-4]):([0-5][0-9]):([0-5][0-9])", timeStr); valid {
		return nil
	}

	return diag.Errorf("Time %v is not a valid time", time)
}

// Validates a date string is in the format 2006-01-02T15:04:05.000000
func validateLocalDateTimes(date interface{}, _ cty.Path) diag.Diagnostics {
	if dateStr, ok := date.(string); ok {
		_, err := time.Parse(resourcedata.TimeParseFormat, dateStr)
		if err != nil {
			return diag.Errorf("Failed to parse date %s: %s", dateStr, err)
		}
		return nil
	}
	return diag.Errorf("Date %v is not a string", date)
}

// Validates a file path or URL
func validatePath(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	if v == "" {
		errors = append(errors, fmt.Errorf("empty file path specified"))
		return warnings, errors
	}

	_, file, err := downloadOrOpenFile(v)
	if err != nil {
		errors = append(errors, err)
	}
	if file != nil {
		defer file.Close()
	}

	return warnings, errors
}
