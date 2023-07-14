package tfexporter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	defaultTfJSONFile  = "genesyscloud.tf.json"
	defaultTfHCLFile   = "genesyscloud.tf"
	defaultTfVarsFile  = "terraform.tfvars"
	defaultTfStateFile = "terraform.tfstate"
)

// Common Exporter interface to abstract away whether we are using HCL or JSON as our exporter
type Exporter func() diag.Diagnostics
type ExporterFilterType int64
type ExporterResourceTypeFilter func(exports map[string]*gcloud.ResourceExporter, filter []string) map[string]*gcloud.ResourceExporter
type ExporterResourceFilter func(result gcloud.ResourceIDMetaMap, name string, filter []string) gcloud.ResourceIDMetaMap

const (
	LegacyInclude ExporterFilterType = iota
	IncludeResources
	ExcludeResources
)

func IncludeFilterByResourceType(exports map[string]*gcloud.ResourceExporter, filter []string) map[string]*gcloud.ResourceExporter {
	if len(filter) > 0 {
		for resType := range exports {
			if !gcloud.StringInSlice(resType, formatFilter(filter)) {
				delete(exports, resType)
			}
		}
	}

	return exports
}

func ExcludeFilterByResourceType(exports map[string]*gcloud.ResourceExporter, filter []string) map[string]*gcloud.ResourceExporter {
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

func FilterResourceByName(result gcloud.ResourceIDMetaMap, name string, filter []string) gcloud.ResourceIDMetaMap {
	if gcloud.SubStringInSlice(fmt.Sprintf("%v::", name), filter) {
		names := make([]string, 0)
		for _, f := range filter {
			n := fmt.Sprintf("%v::", name)

			if strings.Contains(f, n) {
				names = append(names, strings.Replace(f, n, "", 1))
			}
		}

		newResult := make(gcloud.ResourceIDMetaMap)
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

func IncludeFilterResourceByRegex(result gcloud.ResourceIDMetaMap, name string, filter []string) gcloud.ResourceIDMetaMap {
	newFilters := make([]string, 0)
	for _, f := range filter {
		if strings.Contains(f, "::") {
			i := strings.Index(f, "::")
			regexStr := f[i+2:]
			newFilters = append(newFilters, regexStr)
		}
	}

	newResourceMap := make(gcloud.ResourceIDMetaMap)

	if len(newFilters) == 0 {
		return result
	}

	for _, pattern := range newFilters {
		for k, _ := range result {
			match, _ := regexp.MatchString(pattern, result[k].Name)

			if match {
				newResourceMap[k] = result[k]
			}
		}
	}

	return newResourceMap
}

func ExcludeFilterResourceByRegex(result gcloud.ResourceIDMetaMap, name string, filter []string) gcloud.ResourceIDMetaMap {

	newFilters := make([]string, 0)
	for _, f := range filter {
		if strings.Contains(f, "::") {
			i := strings.Index(f, "::")
			regexStr := f[i+2:]
			newFilters = append(newFilters, regexStr)
		}
	}

	if len(newFilters) == 0 {
		return result
	}

	newResourceMap := make(gcloud.ResourceIDMetaMap)

	for k, _ := range result {
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

func resolveReference(refSettings *gcloud.RefAttrSettings, refID string, exporters map[string]*gcloud.ResourceExporter, exportingState bool) string {
	if gcloud.StringInSlice(refID, refSettings.AltValues) {
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

func getFilePath(d *schema.ResourceData, filename string) (string, diag.Diagnostics) {
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

	path := filepath.Join(directory, filename)
	if path == "" {
		return "", diag.Errorf("Failed to create file path with directory %s", directory)
	}
	return path, nil
}
