package tfexporter

import (
	"encoding/json"
	"fmt"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func exportJSONConfig(
	resourceTypeJSONMaps map[string]map[string]gcloud.JsonMap,
	unresolvedAttrs []unresolvableAttributeInfo,
	providerSource,
	version,
	filePath,
	tfVarsFilePath string) diag.Diagnostics {
	rootJSONObject := gcloud.JsonMap{
		"resource": resourceTypeJSONMaps,
		"terraform": gcloud.JsonMap{
			"required_providers": gcloud.JsonMap{
				"genesyscloud": gcloud.JsonMap{
					"source":  providerSource,
					"version": version,
				},
			},
		},
	}

	if len(unresolvedAttrs) > 0 {
		tfVars := make(map[string]interface{})
		variable := make(map[string]gcloud.JsonMap)
		for _, attr := range unresolvedAttrs {
			key := fmt.Sprintf("%s_%s_%s", attr.ResourceType, attr.ResourceName, attr.Name)
			variable[key] = make(gcloud.JsonMap)
			tfVars[key] = make(gcloud.JsonMap)
			variable[key]["description"] = attr.Schema.Description
			if variable[key]["description"] == "" {
				variable[key]["description"] = fmt.Sprintf("%s value for resource %s of type %s", attr.Name, attr.ResourceName, attr.ResourceType)
			}

			variable[key]["sensitive"] = attr.Schema.Sensitive
			if attr.Schema.Default != nil {
				variable[key]["default"] = attr.Schema.Default
			}

			tfVars[key] = determineVarValue(attr.Schema)

			variable[key]["type"] = determineVarType(attr.Schema)
		}
		rootJSONObject["variable"] = variable
		if err := writeTfVars(tfVars, tfVarsFilePath); err != nil {
			return err
		}
	}

	return writeConfig(rootJSONObject, filePath)
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

func resolveRefAttributesInJsonString(currAttr string, currVal string, exporter *gcloud.ResourceExporter, exporters map[string]*gcloud.ResourceExporter, exportingState bool) (string, error) {
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
				data[value] = resolveReference(refSettings, data[value].(string), exporters, exportingState)
			case []interface{}:
				array := data[value].([]interface{})
				for k, v := range array {
					array[k] = resolveReference(refSettings, v.(string), exporters, exportingState)
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
