package tfexporter

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/files"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const resourceJSONFileExt = "tf.json"

type resourceJSONMaps map[string]util.JsonMap

type JsonExporter struct {
	resourceTypesJSONMaps map[string]resourceJSONMaps
	dataSourceTypesMaps   map[string]resourceJSONMaps
	unresolvedAttrs       []unresolvableAttributeInfo
	providerSource        string
	version               string
	dirPath               string
	splitFilesByResource  bool
}

func NewJsonExporter(resourceTypesJSONMaps map[string]resourceJSONMaps, dataSourceTypesMaps map[string]resourceJSONMaps, unresolvedAttrs []unresolvableAttributeInfo, providerSource string, version string, dirPath string, splitFilesByResource bool) *JsonExporter {
	jsonExporter := &JsonExporter{
		resourceTypesJSONMaps: resourceTypesJSONMaps,
		dataSourceTypesMaps:   dataSourceTypesMaps,
		unresolvedAttrs:       unresolvedAttrs,
		providerSource:        providerSource,
		version:               version,
		dirPath:               dirPath,
		splitFilesByResource:  splitFilesByResource,
	}
	return jsonExporter
}

/*
This file contains all of the functions used to generate the JSON export.
*/
func (j *JsonExporter) exportJSONConfig() diag.Diagnostics {
	providerJsonMap := createProviderJsonMap(j.providerSource, j.version)
	variablesJsonMap := createVariablesJsonMap(j.unresolvedAttrs)

	if j.splitFilesByResource {
		// Provider file
		terraformRoot := map[string]interface{}{
			"terraform": providerJsonMap,
		}
		providerJSONFilePath := filepath.Join(j.dirPath, defaultTfJSONProviderFile)
		if providerJSONFilePath == "" {
			return diag.Errorf("Failed to create file path %s", providerJSONFilePath)
		}
		if diagErr := writeConfig(terraformRoot, providerJSONFilePath); diagErr != nil {
			return diagErr
		}

		// Variables file
		variablesRoot := map[string]interface{}{
			"variable": variablesJsonMap,
		}
		variablesJSONFilePath := filepath.Join(j.dirPath, defaultTfJSONVariablesFile)
		if variablesJSONFilePath == "" {
			return diag.Errorf("Failed to create file path %s", variablesJSONFilePath)
		}
		if diagErr := writeConfig(variablesRoot, variablesJSONFilePath); diagErr != nil {
			return diagErr
		}

		// Resource files
		for resType, resJsonMap := range j.resourceTypesJSONMaps {
			resourceRoot := map[string]interface{}{
				"resource": util.JsonMap{
					resType: resJsonMap,
				},
			}

			resourceJSONFilePath := filepath.Join(j.dirPath, fmt.Sprintf("%s.%s", resType, resourceJSONFileExt))
			if resourceJSONFilePath == "" {
				return diag.Errorf("Failed to create file path %s", resourceJSONFilePath)
			}
			if diagErr := writeConfig(resourceRoot, resourceJSONFilePath); diagErr != nil {
				return diagErr
			}
		}

		// DataSource files
		for resType, resJsonMap := range j.dataSourceTypesMaps {
			resourceRoot := map[string]interface{}{
				"data": util.JsonMap{
					resType: resJsonMap,
				},
			}

			resourceJSONFilePath := filepath.Join(j.dirPath, fmt.Sprintf("%s.%s", resType, resourceJSONFileExt))
			if resourceJSONFilePath == "" {
				return diag.Errorf("Failed to create file path %s", resourceJSONFilePath)
			}
			if diagErr := writeConfig(resourceRoot, resourceJSONFilePath); diagErr != nil {
				return diagErr
			}
		}

	} else {
		// Single file export
		rootJSONObject := util.JsonMap{
			"resource":  j.resourceTypesJSONMaps,
			"terraform": providerJsonMap,
			"data":      j.dataSourceTypesMaps,
		}

		if len(variablesJsonMap) > 0 {
			rootJSONObject["variable"] = variablesJsonMap
		}

		jsonFilePath := filepath.Join(j.dirPath, defaultTfJSONFile)
		if jsonFilePath == "" {
			return diag.Errorf("Failed to create file path %s", jsonFilePath)
		}

		writeConfig(rootJSONObject, jsonFilePath)
	}

	// Optional tfvars file creation for unresolved attributes
	if len(j.unresolvedAttrs) > 0 {
		tfVars := make(map[string]interface{})
		for _, attr := range j.unresolvedAttrs {
			key := createUnresolvedAttrKey(attr)
			tfVars[key] = make(util.JsonMap)
			tfVars[key] = determineVarValue(attr.Schema)
		}

		tfVarsFilePath := filepath.Join(j.dirPath, defaultTfVarsFile)
		if tfVarsFilePath == "" {
			return diag.Errorf("Failed to create tfvars file path %s", tfVarsFilePath)
		}
		if err := writeTfVars(tfVars, tfVarsFilePath); err != nil {
			return err
		}
	}

	return nil
}

func createProviderJsonMap(providerSource string, version string) util.JsonMap {
	return util.JsonMap{
		"required_providers": util.JsonMap{
			"genesyscloud": util.JsonMap{
				"source":  providerSource,
				"version": version,
			},
		},
	}
}

func createVariablesJsonMap(unresolvedAttrs []unresolvableAttributeInfo) map[string]util.JsonMap {
	variable := make(map[string]util.JsonMap)
	for _, attr := range unresolvedAttrs {
		key := createUnresolvedAttrKey(attr)
		variable[key] = make(util.JsonMap)
		variable[key]["description"] = attr.Schema.Description
		if variable[key]["description"] == "" {
			variable[key]["description"] = fmt.Sprintf("%s value for resource %s of type %s", attr.Name, attr.ResourceName, attr.ResourceType)
		}

		variable[key]["sensitive"] = attr.Schema.Sensitive
		if attr.Schema.Default != nil {
			variable[key]["default"] = attr.Schema.Default
		}

		variable[key]["type"] = determineVarType(attr.Schema)
	}

	return variable
}

func getDecodedData(jsonString string, currAttr string) (string, error) {
	var jsonVar interface{}
	err := json.Unmarshal([]byte(jsonString), &jsonVar)
	if err != nil {
		return "", err
	}

	formattedJson, err := json.MarshalIndent(jsonVar, "", "\t")
	if err != nil {
		return "", err
	}

	// replace : with = as is expected syntax in a jsonencode object
	decodedJson := strings.Replace(string(formattedJson), "\": ", "\" = ", -1)
	// fix indentation
	numOfIndents := strings.Count(currAttr, ".") + 1
	spaces := ""
	for i := 0; i < numOfIndents; i++ {
		spaces = spaces + "  "
	}
	decodedJson = strings.Replace(decodedJson, "\t", fmt.Sprintf("\t%v", spaces), -1)
	// add extra space before the final character (either ']' or '}')
	decodedJson = fmt.Sprintf("%v%v%v", decodedJson[:len(decodedJson)-1], spaces, decodedJson[len(decodedJson)-1:])
	decodedJson = fmt.Sprintf("jsonencode(%v)", decodedJson)
	return decodedJson, nil
}

func (g *GenesysCloudResourceExporter) resolveRefAttributesInJsonString(currAttr string, currVal string, exporter *resourceExporter.ResourceExporter, exporters map[string]*resourceExporter.ResourceExporter, exportingState bool) (string, error) {
	var jsonData interface{}
	err := json.Unmarshal([]byte(currVal), &jsonData)
	if err != nil {
		return "", err
	}

	nestedAttrs, _ := exporter.ContainsNestedRefAttrs(currAttr)
	for _, value := range nestedAttrs {
		refSettings := exporter.GetNestedRefAttrSettings(value)
		if data, ok := jsonData.(map[string]interface{}); ok {
			switch data[value].(type) {
			case string:
				data[value] = g.resolveReference(refSettings, data[value].(string), exporters, exportingState)
			case []interface{}:
				array := data[value].([]interface{})
				for k, v := range array {
					array[k] = g.resolveReference(refSettings, v.(string), exporters, exportingState)
				}
				data[value] = array
			}
			jsonData = data
		}
	}
	jsonDataMarshalled, err := json.Marshal(jsonData)
	if err != nil {
		return "", err
	}
	return string(jsonDataMarshalled), nil
}

func determineVarType(s *schema.Schema) string {
	var varType string
	switch s.Type {
	case schema.TypeMap:
		if elem, ok := s.Elem.(*schema.Schema); ok {
			varType = fmt.Sprintf("map(%s)", determineVarType(elem))
		} else {
			varType = "map"
		}
	case schema.TypeBool:
		varType = "bool"
	case schema.TypeString:
		varType = "string"
	case schema.TypeList:
		fallthrough
	case schema.TypeSet:
		if elem, ok := s.Elem.(*schema.Schema); ok {
			varType = fmt.Sprintf("list(%s)", determineVarType(elem))
		} else {
			if properties, ok := s.Elem.(*schema.Resource); ok {
				propPairs := ""
				for k, v := range properties.Schema {
					propPairs = fmt.Sprintf("%s%v = %v\n", propPairs, k, determineVarType(v))
				}
				varType = fmt.Sprintf("object({%s})", propPairs)
			} else {
				varType = "object({})"
			}
		}
	case schema.TypeInt:
		fallthrough
	case schema.TypeFloat:
		varType = "number"
	}

	return varType
}

func writeConfig(jsonMap map[string]interface{}, path string) diag.Diagnostics {
	dataJSONBytes, err := json.MarshalIndent(jsonMap, "", "  ")
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("Writing export config file to %s", path)
	if err := files.WriteToFile(postProcessJsonBytes(dataJSONBytes), path); err != nil {
		return err
	}
	return nil
}

func postProcessJsonBytes(resource []byte) []byte {
	resourceStr := string(resource)
	resourceStr = correctDependsOn(resourceStr, false)
	return []byte(resourceStr)
}
