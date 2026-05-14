package architect_datatable

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func buildSdkDatatableSchema(d *schema.ResourceData) (*Jsonschemadocument, diag.Diagnostics) {
	// Hardcoded values the server expects in the JSON schema object
	var (
		schemaType           = "http://json-schema.org/draft-04/schema#"
		jsonType             = "object"
		additionalProperties interface{}
	)

	additionalProperties = false
	properties, err := buildSdkDatatableProperties(d)
	if err != nil {
		return nil, err
	}
	return &Jsonschemadocument{
		Schema:               &schemaType,
		VarType:              &jsonType,
		Required:             &[]string{"key"},
		Properties:           properties,
		AdditionalProperties: &additionalProperties,
	}, nil
}

func buildSdkDatatableProperties(d *schema.ResourceData) (*util.OrderedMap[Datatableproperty], diag.Diagnostics) {
	const propIdPrefix = "/properties/"
	if properties := d.Get("properties").([]interface{}); properties != nil {
		sdkProps := util.NewOrderedMap[Datatableproperty]()
		for i, property := range properties {
			propMap := property.(map[string]interface{})

			// Name and type are required
			propName := propMap["name"].(string)
			propType := propMap["type"].(string)
			propId := propIdPrefix + propName
			orderNum := i

			sdkProp := Datatableproperty{
				Id:           &propId,
				DisplayOrder: &orderNum,
				VarType:      &propType,
			}

			// Title is optional
			if propTitle, ok := propMap["title"]; ok {
				title := propTitle.(string)
				sdkProp.Title = &title
			}

			// Default is optional
			if propDefault, ok := propMap["default"]; ok {
				def := propDefault.(string)
				var defaultVal interface{}
				if def != "" {
					var err error
					// Convert default value to the appropriate type
					switch propType {
					case "boolean":
						defaultVal, err = strconv.ParseBool(def)
					case "string":
						defaultVal = def
					case "integer":
						defaultVal, err = strconv.Atoi(def)
					case "number":
						defaultVal, err = strconv.ParseFloat(def, 64)
					default:
						return nil, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Invalid type %s for Datatable property %s", propType, propName), fmt.Errorf("invalid type for Datatable property"))
					}
					if err != nil {
						return nil, diag.FromErr(err)
					}
				}
				if defaultVal != nil {
					sdkProp.Default = &defaultVal
				}
			}
			sdkProps.Set(propName, sdkProp)
		}
		return sdkProps, nil
	}
	return nil, nil
}

func flattenDatatableProperties(properties *util.OrderedMap[Datatableproperty]) []interface{} {
	type kv struct {
		Key   string
		Value Datatableproperty
	}

	// Build list preserving the OrderedMap key order (API response order) as the default
	var propList []kv
	hasDisplayOrder := false
	for _, key := range properties.Keys() {
		prop, _ := properties.Get(key)
		if prop.DisplayOrder != nil {
			hasDisplayOrder = true
		}
		propList = append(propList, kv{key, prop})
	}

	// If DisplayOrder is provided, sort by it; otherwise rely on the preserved JSON key order
	if hasDisplayOrder {
		sort.SliceStable(propList, func(i, j int) bool {
			di := propList[i].Value.DisplayOrder
			dj := propList[j].Value.DisplayOrder
			if di == nil {
				return false
			}
			if dj == nil {
				return true
			}
			return *di < *dj
		})
	}

	configProps := []interface{}{}
	for _, propKV := range propList {
		propMap := make(map[string]interface{})
		propMap["name"] = propKV.Key
		if propKV.Value.VarType != nil {
			propMap["type"] = *propKV.Value.VarType
		}
		if propKV.Value.Title != nil {
			propMap["title"] = *propKV.Value.Title
		}
		if propKV.Value.Default != nil {
			propMap["default"] = util.InterfaceToString(*propKV.Value.Default)
		}
		configProps = append(configProps, propMap)
	}
	return configProps
}
