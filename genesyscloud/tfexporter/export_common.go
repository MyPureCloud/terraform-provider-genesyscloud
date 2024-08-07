package tfexporter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	defaultTfJSONFile          = "genesyscloud.tf.json"
	defaultTfHCLFile           = "genesyscloud.tf"
	defaultTfHCLProviderFile   = "provider.tf"
	defaultTfJSONProviderFile  = "provider.tf.json"
	defaultTfHCLVariablesFile  = "variables.tf"
	defaultTfJSONVariablesFile = "variables.tf.json"
	defaultTfVarsFile          = "terraform.tfvars"
	defaultTfStateFile         = "terraform.tfstate"
)

// Common Exporter interface to abstract away whether we are using HCL or JSON as our exporter
type Exporter func() diag.Diagnostics
type ExporterFilterHandler int64
type ExporterResourceTypeFilter func(exports map[string]*resourceExporter.ResourceExporter, filter []string) map[string]*resourceExporter.ResourceExporter
type ExporterResourceNameFilter func(result resourceExporter.ResourceIDMetaMap, name string, filter []string) resourceExporter.ResourceIDMetaMap

// ExporterAdvancedFilters allows defining a set of filters to export
type ExporterAdvancedFilters struct {
	IncludeTypes []string
	ExcludeTypes []string
	IncludeNames []string
	ExcludeNames []string
}

const (
	LegacyFilterInclude ExporterFilterHandler = iota
	FilterIncludeResources
	FilterExcludeResources
	FilterAdvancedResources
)

// Returns two grouped lists: one of any resources with a name, one with only resource types
func GroupFilterResourcesByTypeOrName(filterResources []string) ([]string, []string) {
	if len(filterResources) > 0 {
		var resourceWithNames []string
		var resourceTypeOnly []string
		for _, filter := range filterResources {
			if strings.Contains(filter, "::") {
				resourceWithNames = append(resourceWithNames, filter)
			} else {
				resourceTypeOnly = append(resourceTypeOnly, filter)
			}
		}
		return resourceTypeOnly, resourceWithNames
	}
	return nil, nil
}

// Returns map of ResourceExporters that includes any resources by name passed into the filter
func IncludeFilterByResourceType(exports map[string]*resourceExporter.ResourceExporter, filter []string) map[string]*resourceExporter.ResourceExporter {
	if len(filter) > 0 {
		for resType := range exports {
			if !lists.ItemInSlice(resType, formatFilter(filter)) {
				delete(exports, resType)
			}
		}
	}

	return exports
}

// Returns map of ResourceExporters that excludes any resources by name passed into the filter
func ExcludeFilterByResourceType(exports map[string]*resourceExporter.ResourceExporter, filter []string) map[string]*resourceExporter.ResourceExporter {
	if len(filter) > 0 {
		for resType := range exports {
			for _, f := range filter {
				if resType == f {
					delete(exports, resType)
				}
			}
		}
	}
	return exports
}

// Returns ResourceIDMetaMap filtered by list of strings with names
func FilterResourceByName(result resourceExporter.ResourceIDMetaMap, resourceName string, filter []string) resourceExporter.ResourceIDMetaMap {
	if lists.SubStringInSlice(fmt.Sprintf("%v::", resourceName), filter) {
		names := make([]string, 0)
		for _, f := range filter {
			n := fmt.Sprintf("%v::", resourceName)

			if strings.Contains(f, n) {
				names = append(names, strings.Replace(f, n, "", 1))
			}
		}

		newResult := make(resourceExporter.ResourceIDMetaMap)
		for _, name := range names {
			for k, v := range result {
				if v.LabelName == name {
					newResult[k] = v
				}
			}
		}
		return newResult
	}

	return result
}

// Returns ResourceIDMetaMap filtered by list of strings by ID
func FilterResourceById(result resourceExporter.ResourceIDMetaMap, resourceName string, filter []string) resourceExporter.ResourceIDMetaMap {
	if lists.SubStringInSlice(fmt.Sprintf("%v::", resourceName), filter) {
		names := make([]string, 0)
		for _, f := range filter {
			n := fmt.Sprintf("%v::", resourceName)

			if strings.Contains(f, n) {
				names = append(names, strings.Replace(f, n, "", 1))
			}
		}
		newResult := make(resourceExporter.ResourceIDMetaMap)
		for _, name := range names {
			for k, v := range result {
				if k == name {
					newResult[k] = v
				}
			}
		}
		return newResult
	}

	return result
}

// Returns a ResourceIdMetaMap that includes any resources instances whose name (either as returned from the API or sanitized) matches the Regexp filter
func IncludeFilterResourceByRegex(result resourceExporter.ResourceIDMetaMap, resourceName string, filter []string) resourceExporter.ResourceIDMetaMap {
	newFilters := make([]string, 0)
	for _, f := range filter {
		if strings.Contains(f, "::") && strings.Split(f, "::")[0] == resourceName {
			i := strings.Index(f, "::")
			regexStr := f[i+2:]
			newFilters = append(newFilters, regexStr)
		}
	}

	newResourceMap := make(resourceExporter.ResourceIDMetaMap)

	if len(newFilters) == 0 {
		return result
	}

	for _, pattern := range newFilters {
		for k := range result {

			// If name matches label name
			originalNameMatch, _ := regexp.MatchString(pattern, result[k].ResourceName)
			if originalNameMatch {
				newResourceMap[k] = result[k]
			}

			// If name matches label name
			labelMatch, _ := regexp.MatchString(pattern, result[k].LabelName)
			if labelMatch {
				newResourceMap[k] = result[k]
			}

			// If name matches sanitized name
			sanitizedMatch, _ := regexp.MatchString(pattern, result[k].SanitizedLabelName)
			if sanitizedMatch {
				newResourceMap[k] = result[k]
			}
		}
	}

	return newResourceMap
}

// Returns a ResourceIdMetaMap that excludes any resources instances whose name (either as returned from the API or sanitized) matches the Regexp filter
func ExcludeFilterResourceByRegex(result resourceExporter.ResourceIDMetaMap, resourceName string, filter []string) resourceExporter.ResourceIDMetaMap {
	newFilters := make([]string, 0)
	for _, f := range filter {
		if strings.Contains(f, "::") && strings.Split(f, "::")[0] == resourceName {
			i := strings.Index(f, "::")
			regexStr := f[i+2:]
			newFilters = append(newFilters, regexStr)
		}
	}

	if len(newFilters) == 0 {
		return result
	}

	newResourceMap := make(resourceExporter.ResourceIDMetaMap)

	for k := range result {
		for _, pattern := range newFilters {

			// If name matches original name
			originalNameMatch, _ := regexp.MatchString(pattern, result[k].ResourceName)
			if !originalNameMatch {
				newResourceMap[k] = result[k]
			} else {
				delete(newResourceMap, k)
				break
			}

			// If name matches label name
			labelMatch, _ := regexp.MatchString(pattern, result[k].LabelName)
			if !labelMatch {
				newResourceMap[k] = result[k]
			} else {
				delete(newResourceMap, k)
				break
			}

			// If name matches sanitized name
			sanitizedMatch, _ := regexp.MatchString(pattern, result[k].SanitizedLabelName)
			if !sanitizedMatch {
				newResourceMap[k] = result[k]
			} else {
				delete(newResourceMap, k)
				break
			}
		}
	}
	return newResourceMap
}

/*
This file is used to hold common methods that are used across the exporter.  They do not have strong affinity to any one particular export process (e.g. HCL or JSON).
*/
func determineVarValue(s *schema.Schema) interface{} {
	if s.Default != nil {
		if m, ok := s.Default.(map[string]string); ok {
			stringMap := make(map[string]interface{})
			for k, v := range m {
				stringMap[k] = v
			}
			return stringMap
		}

		return s.Default
	}

	switch s.Type {
	case schema.TypeString:
		return ""
	case schema.TypeInt:
		return 0
	case schema.TypeFloat:
		return 0.0
	case schema.TypeBool:
		return false
	default:
		if properties, ok := s.Elem.(*schema.Resource); ok {
			propertyMap := make(map[string]interface{})
			for k, v := range properties.Schema {
				propertyMap[k] = determineVarValue(v)
			}
			return propertyMap
		}
	}

	return nil
}

// Correct exported e164 number e.g. +(1) 111-222-333 --> +1111222333
func sanitizeE164Number(number string) string {
	charactersToRemove := []string{" ", "-", "(", ")"}
	for _, c := range charactersToRemove {
		number = strings.Replace(number, c, "", -1)
	}
	return number
}
func sanitizeRrule(input string) string {
	attributeRegex := map[string]*regexp.Regexp{
		"INTERVAL":   regexp.MustCompile(`INTERVAL=([1-9][0-9]*|0?[1-9][0-9]*);`),
		"BYMONTH":    regexp.MustCompile(`BYMONTH=(0?[1-9]|1[0-2]);`),
		"BYMONTHDAY": regexp.MustCompile(`BYMONTHDAY=(0?[1-9]|[1-2][0-9]|3[0-1])$`),
	}

	// Iterate over attributes and modify the input string
	for attributeName, regex := range attributeRegex {
		input = regex.ReplaceAllStringFunc(input, func(match string) string {
			return removeTrailingZeros(match, attributeName)
		})
	}
	return input
}

func removeTrailingZeros(match, attributeName string) string {
	pattern := `=(\d{1,2})`
	re := regexp.MustCompile(pattern)
	outputText := re.ReplaceAllStringFunc(match, func(match string) string {
		numericPart := match[1:]
		numericPart = fmt.Sprintf("%d", parseInt(numericPart))
		return "=" + numericPart
	})
	return outputText
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

// Get a string path to the target export file
func getFilePath(d *schema.ResourceData, filename string) (string, diag.Diagnostics) {
	directory, diagErr := getDirPath(d)
	if diagErr != nil {
		return "", diagErr
	}

	path := filepath.Join(directory, filename)
	if path == "" {
		return "", diag.Errorf("Failed to create file path with directory %s", directory)
	}
	return path, nil
}

// Get a string path to the target export directory
func getDirPath(d *schema.ResourceData) (string, diag.Diagnostics) {
	directory := d.Get("directory").(string)
	if strings.HasPrefix(directory, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", diag.Errorf("Failed to evaluate home directory: %v", err)
		}
		directory = strings.Replace(directory, "~", homeDir, 1)
	}
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return "", diag.FromErr(err)
	}

	return directory, nil
}

// Checks if a directory path is empty
func isDirEmpty(path string) (bool, diag.Diagnostics) {
	f, err := os.Open(path)
	if err != nil {
		return false, diag.FromErr(err)
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, diag.FromErr(err)
}

func createUnresolvedAttrKey(attr unresolvableAttributeInfo) string {
	return fmt.Sprintf("%s_%s_%s", attr.ResourceType, attr.ResourceName, attr.Name)
}
