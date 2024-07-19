package validators

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"strings"

	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	files "terraform-provider-genesyscloud/genesyscloud/util/files"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ValidatePhoneNumber(number interface{}, _ cty.Path) diag.Diagnostics {
	if numberStr, ok := number.(string); ok {

		formattedNum, err := util.FormatAsE164Number(numberStr)
		if err != nil {
			return err
		}
		if formattedNum != numberStr {
			return diag.Errorf("Failed to parse number in an E.164 format.  Passed %s and expected: %s", numberStr, formattedNum)
		}
		return nil
	}
	return diag.Errorf("Phone number %v is not a string", number)
}

// ValidateRrule validates rrule attribute
func ValidateRrule(rrule interface{}, _ cty.Path) diag.Diagnostics {
	if input, ok := rrule.(string); ok {
		// FREQ Attribute validation
		freqRegex := regexp.MustCompile(`FREQ=([A-Z]+)`)
		if match := freqRegex.FindStringSubmatch(input); strings.Contains(input, "FREQ=") && match == nil {
			return diag.Errorf("Invalid FREQ attribute. Should consist of uppercase letters.")
		}
		// INTERVAL Attribute validation
		intervalRegex := regexp.MustCompile(`INTERVAL=([1-9][0-9]*)`)
		if match := intervalRegex.FindStringSubmatch(input); strings.Contains(input, "INTERVAL=") && match == nil {
			return diag.Errorf("Invalid INTERVAL attribute. Should be a positive integer greater than 0 without leading zeros.")
		}

		// rrule is split and stored in array using ';' as delimiter
		// array is iterated over and variables are assigned if they exist
		// This allows for the values for BYMONTH and BYMONTHDAY to be split, parsed and checked that they are within the valid range
		rRuleAttributes := strings.Split(input, ";")
		for _, value := range rRuleAttributes {
			// BYMONTH Attribute validation
			if strings.Contains(value, "BYMONTH=") {
				byMonth := value
				byMonthString := strings.Split(byMonth, "=")[1]
				byMonthValues := strings.Split(byMonthString, ",")

				for _, month := range byMonthValues {
					byMonthValue, err := strconv.Atoi(month)
					if err != nil {
						return diag.Errorf("Failed to validate BYMONTH. [Error: %v]", err)
					}
					if byMonthValue <= 0 || byMonthValue > 12 {
						return diag.Errorf("Invalid BYMONTH attribute. Should be a valid month (1-12) without leading zeros for single-digit months.")
					}
				}
			}

			// BYMONTHDAY Attribute validation
			if strings.Contains(value, "BYMONTHDAY=") {
				byMonthDay := value
				byMonthDayString := strings.Split(byMonthDay, "=")[1]
				byMonthDayValues := strings.Split(byMonthDayString, ",")

				for _, day := range byMonthDayValues {
					byMonthDayValue, err := strconv.Atoi(day)
					if err != nil {
						return diag.Errorf("Failed to validate BYMONTHDAY. [Error: %v]", err)
					}

					if byMonthDayValue <= 0 || byMonthDayValue > 31 {
						return diag.Errorf("Invalid BYMONTHDAY attribute. Should be a valid day of the month (1-31) without leading zeros for single-digit days.")
					}
				}
			}
		}
		return nil
	}
	return diag.Errorf("Provided rrule %v is not in string format", rrule)
}

// ValidateExtensionPool validates a phone extension pool
func ValidateExtensionPool(number interface{}, _ cty.Path) diag.Diagnostics {
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

// ValidateDate validates a date string is in the format yyyy-MM-dd
func ValidateDate(date interface{}, _ cty.Path) diag.Diagnostics {
	if dateStr, ok := date.(string); ok {
		_, err := time.Parse(resourcedata.DateParseFormat, dateStr)
		if err != nil {
			return diag.Errorf("Failed to parse date %s: %s", dateStr, err)
		}
		return nil
	}
	return diag.Errorf("Date %v is not a string", date)
}

// ValidateDateTime validates a date string is in the format 2006-01-02T15:04Z
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

// ValidateCountryCode validates a country code is in format ISO 3166-1 alpha-2
func ValidateCountryCode(code interface{}, _ cty.Path) diag.Diagnostics {
	countryCode := code.(string)
	if len(countryCode) == 2 {
		return nil
	} else if countryCode == "country-code-1" {
		return nil
	}
	return diag.Errorf("Country code %v is not of format ISO 3166-1 alpha-2", code)
}

// ValidateTime validates a date string is in format hh:mm:ss
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

// ValidateTimeHHMM validates a date string is in format hh:mm
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

// ValidateLocalDateTimes validates a date string is in the format 2006-01-02T15:04:05.000000
func ValidateLocalDateTimes(date interface{}, _ cty.Path) diag.Diagnostics {
	if dateStr, ok := date.(string); ok {
		_, err := time.Parse(resourcedata.TimeParseFormat, dateStr)
		if err != nil {
			return diag.Errorf("Failed to parse date %s: %s", dateStr, err)
		}
		return nil
	}
	return diag.Errorf("Date %v is not a string", date)
}

// ValidatePath validates a file path or URL
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

// ValidateResponseAssetName validate a response asset filename matches the criteria outlined in the description
func ValidateResponseAssetName(name interface{}, _ cty.Path) diag.Diagnostics {
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

// ValidateHexColor validates if a string matches '#FFFFFF' RGB color representation.
func ValidateHexColor(color interface{}, _ cty.Path) diag.Diagnostics {
	if colorStr, ok := color.(string); ok {
		matched, err := regexp.MatchString("^#([A-Fa-f0-9]{6})$", colorStr)
		if err != nil {
			return diag.Errorf("Error applying regular expression against color: %v", err)
		}
		if !matched {
			return diag.Errorf("Invalid color. It must be in the format #FFFFFF")
		}
		return nil
	}
	return diag.Errorf("Color %v is not a string", color)
}

// ValidateLanguageCode validates that a valid language code that Genesys Cloud supports is passed.
func ValidateLanguageCode(lang interface{}, _ cty.Path) diag.Diagnostics {
	langCodeList := []string{"en-US", "en-UK", "en-AU", "en-CA", "en-HK", "en-IN", "en-IE", "en-NZ", "en-PH", "en-SG", "en-ZA", "de-DE", "de-AT", "de-CH", "es-AR", "es-CO", "es-MX", "es-US", "es-ES", "fr-FR", "fr-BE", "fr-CA", "fr-CH", "pt-BR", "pt-PT", "nl-NL", "nl-BE", "it-IT", "ca-ES", "tr-TR", "sv-SE", "fi-FI", "nb-NO", "da-DK", "ja-JP", "ar-AE", "zh-CN", "zh-TW", "zh-HK", "ko-KR", "pl-PL", "hi-IN", "th-TH", "hu-HU", "vi-VN", "uk-UA"}
	if langCode, ok := lang.(string); ok {
		if lists.ItemInSlice(langCode, langCodeList) {
			return nil
		}
		return diag.Errorf("Language code %s not found in language code list %v", langCode, langCodeList)
	}
	return diag.Errorf("Language code %v is not a string", lang)
}
