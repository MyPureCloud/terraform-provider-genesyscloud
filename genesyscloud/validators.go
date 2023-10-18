package genesyscloud

import (
	"fmt"
	"regexp"
	"time"

	"strings"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	files "terraform-provider-genesyscloud/genesyscloud/util/files"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nyaruka/phonenumbers"
)

func ValidatePhoneNumber(number interface{}, _ cty.Path) diag.Diagnostics {
	if numberStr, ok := number.(string); ok {
		phoneNumber, err := phonenumbers.Parse(numberStr, "US")
		if err != nil {
			return diag.Errorf("Failed to validate phone number %s: %s", numberStr, err)
		}

		formattedNum := phonenumbers.Format(phoneNumber, phonenumbers.E164)
		if formattedNum != numberStr {
			return diag.Errorf("Failed to parse number in an E.164 format.  Passed %s and expected: %s", numberStr, formattedNum)
		}
		return nil
	}
	return diag.Errorf("Phone number %v is not a string", number)
}

// Validates a phone extension pool
func validateExtensionPool(number interface{}, _ cty.Path) diag.Diagnostics {
	if numberStr, ok := number.(string); ok {

		re := regexp.MustCompile(`^\d{3,9}$`)
		// check if the string matches the regular expression
		if !re.MatchString(numberStr) {
			return diag.Errorf("The extension provided %q must between 3-9 characters long and made up of all integer values\n", numberStr)
		}

		return nil
	}
	return diag.Errorf("Extension provided %v is not a string", number)
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
func ValidateDateTime(date interface{}, _ cty.Path) diag.Diagnostics {
	if dateStr, ok := date.(string); ok {
		_, err := time.Parse("2006-01-02T15:04Z", dateStr)
		if err != nil {
			return diag.Errorf("Failed to parse date %s: %s", dateStr, err)
		}
		return nil
	}
	return diag.Errorf("Date %v is not a string", date)
}

// Validates a country code is in format ISO 3166-1 alpha-2
func ValidateCountryCode(code interface{}, _ cty.Path) diag.Diagnostics {
	countryCode := code.(string)
	if len(countryCode) == 2 {
		return nil
	} else if countryCode == "country-code-1" {
		return nil
	}
	return diag.Errorf("Country code %v is not of format ISO 3166-1 alpha-2", code)
}

// Validates a date string is in format hh:mm:ss
func ValidateTime(time interface{}, _ cty.Path) diag.Diagnostics {
	timeStr := time.(string)
	if len(timeStr) > 9 {
		timeStr = timeStr[:8]
	}
	if valid, _ := regexp.MatchString("^(0?[0-9]|1?[0-9]|2[0-4]):([0-5][0-9]):([0-5][0-9])", timeStr); valid {
		return nil
	}

	return diag.Errorf("Time %v is not a valid time", time)
}

// Validates a date string is in format hh:mm
func ValidateTimeHHMM(time interface{}, _ cty.Path) diag.Diagnostics {
	timeStr := time.(string)
	if timeStr == "" {
		return nil
	}

	if valid, _ := regexp.MatchString("^(0?[0-9]|1?[0-9]|2[0-4]):([0-5][0-9])", timeStr); valid {
		return nil
	}

	return diag.Errorf("Time %v is not a valid time, must use format HH:mm", time)
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
func ValidatePath(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	if v == "" {
		errors = append(errors, fmt.Errorf("empty file path specified"))
		return warnings, errors
	}

	_, file, err := files.DownloadOrOpenFile(v)
	if err != nil {
		errors = append(errors, err)
	}
	if file != nil {
		defer file.Close()
	}

	return warnings, errors
}

// Validate a response asset filename matches the criteria outlined in the description
func validateResponseAssetName(name interface{}, _ cty.Path) diag.Diagnostics {
	if nameStr, ok := name.(string); ok {
		matched, err := regexp.MatchString("^[^\\.][^\\`\\\\{\\^\\}\\% \"\\>\\<\\[\\]\\#\\~|]+[^/]$", nameStr)
		if err != nil {
			return diag.Errorf("Error applying regular expression against filename: %v", err)
		}
		if !matched {
			return diag.Errorf("Invalid filename. It must not start with a dot and not end with a forward slash. Whitespace and the following characters are not allowed: \\{^}%s]\">[~<#|", "%`")
		}
		return nil
	}
	return diag.Errorf("filename %v is not a string", name)
}

func ValidateSubStringInSlice(valid []string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return warnings, errors
		}

		for _, b := range valid {
			if strings.Contains(v, b) {
				return warnings, errors
			}
		}

		if !lists.ItemInSlice(v, valid) || !lists.SubStringInSlice(v, valid) {
			errors = append(errors, fmt.Errorf("string %s not in slice", v))
			return warnings, errors
		}

		if !lists.SubStringInSlice(v, valid) {
			errors = append(errors, fmt.Errorf("substring %s not in slice", v))
			return warnings, errors
		}

		return warnings, errors
	}
}
