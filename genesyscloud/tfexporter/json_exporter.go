package tfexporter

import (
	"fmt"
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
