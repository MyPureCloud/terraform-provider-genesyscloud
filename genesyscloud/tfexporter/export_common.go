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
type ExporterFilterType int64
type ExporterResourceTypeFilter func(exports map[string]*resourceExporter.ResourceExporter, filter []string) map[string]*resourceExporter.ResourceExporter
type ExporterResourceFilter func(resourceIdMetaMap resourceExporter.ResourceIDMetaMap, resourceType string, filter []string) resourceExporter.ResourceIDMetaMap

const (
	LegacyInclude ExporterFilterType = iota
	IncludeResources
	ExcludeResources
)

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

func FilterResourceByLabel(result resourceExporter.ResourceIDMetaMap, resourceType string, filters []string) resourceExporter.ResourceIDMetaMap {
	if lists.SubStringInSlice(fmt.Sprintf("%v::", resourceType), filters) {
		labels := make([]string, 0)
		for _, filter := range filters {
			resourceTypePrefix := fmt.Sprintf("%v::", resourceType)

			if strings.Contains(filter, resourceTypePrefix) {
				labels = append(labels, strings.Replace(filter, resourceTypePrefix, "", 1))
			}
		}

		newResult := make(resourceExporter.ResourceIDMetaMap)
		for _, label := range labels {
			for k, v := range result {
				if v.BlockLabel == label {
					newResult[k] = v
				}
			}
		}
		return newResult
	}

	return result
}

func FilterResourceById(result resourceExporter.ResourceIDMetaMap, resourceType string, filters []string) resourceExporter.ResourceIDMetaMap {
	if lists.SubStringInSlice(fmt.Sprintf("%v::", resourceType), filters) {
		resourceIds := make([]string, 0)
		for _, filter := range filters {
			resourceTypePrefix := fmt.Sprintf("%v::", resourceType)

			if strings.Contains(filter, resourceTypePrefix) {
				resourceIds = append(resourceIds, strings.Replace(filter, resourceTypePrefix, "", 1))
			}
		}
		newResult := make(resourceExporter.ResourceIDMetaMap)
		for _, resourceId := range resourceIds {
			for k, v := range result {
				if k == resourceId {
					newResult[k] = v
				}
			}
		}
		return newResult
	}

	return result
}

func IncludeFilterResourceByRegex(result resourceExporter.ResourceIDMetaMap, resourceType string, filters []string) resourceExporter.ResourceIDMetaMap {
	newFilters := make([]string, 0)
	for _, filter := range filters {
		if strings.Contains(filter, "::") && strings.Split(filter, "::")[0] == resourceType {
			i := strings.Index(filter, "::")
			regexStr := filter[i+2:]
			newFilters = append(newFilters, regexStr)
		}
	}

	newResourceMap := make(resourceExporter.ResourceIDMetaMap)

	if len(newFilters) == 0 {
		return result
	}

	sanitizer := resourceExporter.NewSanitizerProvider()

	for _, pattern := range newFilters {
		for k := range result {
			match, _ := regexp.MatchString(pattern, result[k].BlockLabel)

			// If filter label matches original label
			if match {
				newResourceMap[k] = result[k]
			}

			// If filter label matches sanitized label
			sanitizedMatch, _ := regexp.MatchString(pattern, sanitizer.S.SanitizeResourceBlockLabel(result[k].BlockLabel))
			if sanitizedMatch {
				newResourceMap[k] = result[k]
			}
		}
	}

	return newResourceMap
}

func ExcludeFilterResourceByRegex(result resourceExporter.ResourceIDMetaMap, resourceType string, filters []string) resourceExporter.ResourceIDMetaMap {
	newFilters := make([]string, 0)
	for _, filter := range filters {
		if strings.Contains(filter, "::") && strings.Split(filter, "::")[0] == resourceType {
			i := strings.Index(filter, "::")
			regexStr := filter[i+2:]
			newFilters = append(newFilters, regexStr)
		}
	}

	if len(newFilters) == 0 {
		return result
	}

	newResourceMap := make(resourceExporter.ResourceIDMetaMap)
	sanitizer := resourceExporter.NewSanitizerProvider()

	for k := range result {
		for _, pattern := range newFilters {

			// If filter label matches original label
			match, _ := regexp.MatchString(pattern, result[k].BlockLabel)
			if !match {
				newResourceMap[k] = result[k]
			} else {
				delete(newResourceMap, k)
				break
			}

			// If filter label matches sanitized label
			sanitizedMatch, _ := regexp.MatchString(pattern, sanitizer.S.SanitizeResourceBlockLabel(result[k].BlockLabel))
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
	return fmt.Sprintf("%s_%s_%s", attr.ResourceType, attr.ResourceLabel, attr.Name)
}
