package validators

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"regexp"
	"strconv"
	"time"

	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ValidatePhoneNumber(number interface{}, _ cty.Path) diag.Diagnostics {
	if feature_toggles.BcpModeEnabledExists() {
		return nil
	}
	if numberStr, ok := number.(string); ok {
		utilE164 := util.NewUtilE164Service()
		validNum, err := utilE164.IsValidE164Number(numberStr)
		if err != nil {
			return err
		}
		if !validNum {
			return diag.Errorf("Failed to validate number is in an E.164 format: %s", numberStr)
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
	// amazonq-ignore-next-line
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

	_, file, err := files.DownloadOrOpenFile(context.Background(), v, true)
	if err != nil {
		return warnings, append(errors, err)
	}
	if file != nil {
		defer file.Close()
	}

	return warnings, errors
}

type ValidateCSVOptions struct {
	RequiredColumns []string
	SampleSize      int
	MaxRowCount     int64
	SkipInterval    int // How often to sample after initial sampling
}

func ValidateCSVFormatWithConfig(filepath string, opts ValidateCSVOptions) error {

	const (
		maxSkipInterval     = 1000000 // Maximum allowed skip interval
		defaultSkipInterval = 1000    // Default skip interval
		maxSampleSize       = 100000  // Maximum allowed sample size
	)
	// Validate configuration
	if opts.SkipInterval < 0 {
		return fmt.Errorf("skip interval must be non-negative, got %d", opts.SkipInterval)
	}
	if opts.SkipInterval > maxSkipInterval {
		return fmt.Errorf("skip interval too large, maximum allowed is %d", maxSkipInterval)
	}
	if opts.SampleSize < 0 {
		return fmt.Errorf("sample size must be non-negative, got %d", opts.SampleSize)
	}
	if opts.SampleSize > maxSampleSize {
		return fmt.Errorf("sample size too large, maximum allowed is %d", maxSampleSize)
	}

	// Open the file
	_, fileHandler, err := files.DownloadOrOpenFile(context.Background(), filepath, true)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer fileHandler.Close()

	reader := csv.NewReader(fileHandler)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = 0

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Validate required columns if specified
	if len(opts.RequiredColumns) > 0 {
		headerMap := make(map[string]bool)
		for _, header := range headers {
			headerMap[header] = true
		}

		requiredColumnsNotFound := []string{}
		for _, required := range opts.RequiredColumns {
			if !headerMap[required] {
				requiredColumnsNotFound = append(requiredColumnsNotFound, required)
			}
		}

		if len(requiredColumnsNotFound) > 0 {
			return fmt.Errorf("CSV file is missing required columns: %v", requiredColumnsNotFound)
		}
	}

	expectedFields := len(headers)

	skipInterval := opts.SkipInterval
	if skipInterval == 0 {
		skipInterval = defaultSkipInterval
	}

	var rowCount uint64 = 1 // Start at 1 since we already read header
	skipIntervalU64 := uint64(skipInterval)

	for {
		// Check for uint64 overflow
		if rowCount == math.MaxUint64 {
			return fmt.Errorf("file exceeds maximum supported row count")
		}

		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading line %d: %w", rowCount, err)
		}

		if opts.MaxRowCount > 0 && rowCount > uint64(opts.MaxRowCount) {
			return fmt.Errorf("CSV file exceeds maximum allowed rows of %d", opts.MaxRowCount)
		}

		// Validate sampled rows
		if rowCount <= uint64(opts.SampleSize) || rowCount%skipIntervalU64 == 0 {
			if len(row) != expectedFields {
				return fmt.Errorf("line %d has %d fields, expected %d", rowCount, len(row), expectedFields)
			}
		} else {
			reader.FieldsPerRecord = -1
		}

		rowCount++
	}

	return nil
}

// ValidateResponseAssetName validate a response asset filename matches the criteria outlined in the description
func ValidateResponseAssetName(name interface{}, _ cty.Path) diag.Diagnostics {
	if nameStr, ok := name.(string); ok {
		matched, err := regexp.MatchString("^[^\\.]([^\\`\\\\{\\^\\}\\% \"\\>\\<\\[\\]\\#\\~|]|\\s)+[^/]$", nameStr)
		if err != nil {
			return diag.Errorf("Error applying regular expression against filename: %v", err)
		}
		if !matched {
			return diag.Errorf("Invalid filename. It must not start with a dot and not end with a forward slash. The following characters are not allowed: \\{^}%s]\">[~<#|", "%`")
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

// Function factory that returns a custom diff function
// Note: supportS3 lets us know if the resource is prepared to handle S3 paths (e.g. architect_flow). Once all resources support S3 paths, we can remove this parameter.
func ValidateFileContentHashChanged(filepathAttr, hashAttr string, supportS3 bool) customdiff.ResourceConditionFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta any) bool {
		filepath := d.Get(filepathAttr).(string)

		newHash, err := files.HashFileContent(ctx, filepath, supportS3)
		if err != nil {
			log.Printf("Error calculating file content hash: %v", err)
			return false
		}

		// Get the current hash value
		oldHash := d.Get(hashAttr).(string)

		// Return true if the hashes are different
		return oldHash != newHash
	}
}

// ValidateCSVColumns returns a CustomizeDiffFunction that validates if a CSV file
// contains the required columns. It takes the names of the attributes that contain
// the file path and the column names.
func ValidateCSVWithColumns(filePathAttr string, columnNamesAttr string) schema.CustomizeDiffFunc {

	// This function ensures that the contacts file is a CSV file and that it includes the columns defined on the resource
	return func(ctx context.Context, d *schema.ResourceDiff, _ interface{}) error {
		if !d.HasChange(filePathAttr) || !d.HasChange(columnNamesAttr) {
			return nil
		}

		filepath := d.Get(filePathAttr).(string)
		if filepath == "" {
			return nil
		}

		columnNamesRaw := d.Get(columnNamesAttr).([]interface{})
		requiredColumns := make([]string, len(columnNamesRaw))
		for i, v := range columnNamesRaw {
			requiredColumns[i] = v.(string)
		}

		validatorOpts := ValidateCSVOptions{
			RequiredColumns: requiredColumns,
		}

		err := ValidateCSVFormatWithConfig(filepath, validatorOpts)
		if err != nil {
			return fmt.Errorf("failed to validate contacts file: %s", err)
		}
		return nil
	}
}

// ValidateStringInMap returns a SchemaValidateDiagFunc that validates if a string
// included in a map is in a list of acceptable values
func ValidateStringInMap(valid []string, ignoreCase bool) schema.SchemaValidateDiagFunc {
	// Create the regular expression pattern
	pattern := strings.Join(valid, "|")
	if ignoreCase {
		pattern = fmt.Sprintf(`(?i)^(%s)$`, pattern)
	} else {
		pattern = fmt.Sprintf(`^(%s)$`, pattern)
	}

	return validation.MapKeyMatch(
		regexp.MustCompile(pattern),
		fmt.Sprintf(`expected key to be one of ["%s"], got`, strings.Join(valid, `", "`)),
	)
}
