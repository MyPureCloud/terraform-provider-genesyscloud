package tfexporter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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
type ExporterResourceFilter func(result resourceExporter.ResourceIDMetaMap, name string, filter []string) resourceExporter.ResourceIDMetaMap

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

func FilterResourceByName(result resourceExporter.ResourceIDMetaMap, name string, filter []string) resourceExporter.ResourceIDMetaMap {
	if lists.SubStringInSlice(fmt.Sprintf("%v::", name), filter) {
		names := make([]string, 0)
		for _, f := range filter {
			n := fmt.Sprintf("%v::", name)

			if strings.Contains(f, n) {
				names = append(names, strings.Replace(f, n, "", 1))
			}
		}

		newResult := make(resourceExporter.ResourceIDMetaMap)
		for _, name := range names {
			for k, v := range result {
				if v.Name == name {
					newResult[k] = v
				}
			}
		}
		return newResult
	}

	return result
}

func IncludeFilterResourceByRegex(result resourceExporter.ResourceIDMetaMap, name string, filter []string) resourceExporter.ResourceIDMetaMap {
	newFilters := make([]string, 0)
	for _, f := range filter {
		if strings.Contains(f, "::") && strings.Split(f, "::")[0] == name {
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
			match, _ := regexp.MatchString(pattern, result[k].Name)

			if match {
				newResourceMap[k] = result[k]
			}
		}
	}

	return newResourceMap
}

func ExcludeFilterResourceByRegex(result resourceExporter.ResourceIDMetaMap, name string, filter []string) resourceExporter.ResourceIDMetaMap {
	newFilters := make([]string, 0)
	for _, f := range filter {
		if strings.Contains(f, "::") && strings.Split(f, "::")[0] == name {
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

			match, _ := regexp.MatchString(pattern, result[k].Name)
			if !match {
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

func resolveReference(refSettings *resourceExporter.RefAttrSettings, refID string, exporters map[string]*resourceExporter.ResourceExporter, exportingState bool) string {
	if lists.ItemInSlice(refID, refSettings.AltValues) {
		// This is not actually a reference to another object. Keep the value
		return refID
	}

	if exporters[refSettings.RefType] != nil {
		// Get the sanitized name from the ID returned as a reference expression
		if idMetaMap := exporters[refSettings.RefType].SanitizedResourceMap; idMetaMap != nil {
			if meta := idMetaMap[refID]; meta != nil && meta.Name != "" {
				return fmt.Sprintf("${%s.%s.id}", refSettings.RefType, meta.Name)
			}
		}
	}

	if exportingState {
		// Don't remove unmatched IDs when exporting state. This will keep existing config in an org
		return refID
	}
	// No match found. Remove the value from the config since we do not have a reference to use
	return ""
}

// Correct exported e164 number e.g. +(1) 111-222-333 --> +1111222333
func sanitizeE164Number(number string) string {
	charactersToRemove := []string{" ", "-", "(", ")"}
	for _, c := range charactersToRemove {
		number = strings.Replace(number, c, "", -1)
	}
	return number
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
