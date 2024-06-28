package task_management_workitem_schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_task_management_workitem_schema_test.go contains all of the test cases for running the resource
tests for task_management_workitem_schema.
*/

func TestAccResourceTaskManagementWorkitemSchema(t *testing.T) {
	t.Parallel()
	var (
		schemaResId       = "tf_schema_1"
		schemaName        = "tf_schema_" + uuid.NewString()
		schemaDescription = "created for CX as Code test case"

		// "text" field
		attr1 = customField{
			title:       "custom_text_attribute",
			description: "custom_text_attribute description",
			varType:     TEXT,
			additionalProps: map[string]interface{}{
				"minLength": 1,
				"maxLength": 100,
			},
		}

		// "longtext" field
		attr2 = customField{
			title:       "Custom_longtext_attribute",
			description: "Custom_longtext_attribute description",
			varType:     LONGTEXT,
			additionalProps: map[string]interface{}{
				"minLength": 1,
				"maxLength": 1000,
			},
		}

		// "url" field
		attr3 = customField{
			title:       "Custom_url_attribute",
			description: "Custom_url_attribute description",
			varType:     URL,
			additionalProps: map[string]interface{}{
				"minLength": 1,
				"maxLength": 200,
			},
		}

		// "identifier" field
		attr4 = customField{
			title:       "Custom_identifier_attribute",
			description: "Custom_identifier_attribute description",
			varType:     IDENTIFIER,
			additionalProps: map[string]interface{}{
				"minLength": 1,
				"maxLength": 100,
			},
		}

		// "integer" field
		attr5 = customField{
			title:       "Custom_int_attribute",
			description: "Custom_int_attribute description",
			varType:     INTEGER,
			additionalProps: map[string]interface{}{
				"minimum": -100,
				"maximum": 100,
			},
		}

		// "number" field
		attr6 = customField{
			title:       "Custom_number_attribute",
			description: "Custom_number_attribute description",
			varType:     NUMBER,
			additionalProps: map[string]interface{}{
				"minimum": -100,
				"maximum": 100,
			},
		}

		// "date" field
		attr7 = customField{
			title:       "Custom_date_attribute",
			description: "Custom_date_attribute description",
			varType:     DATE,
		}

		// "datetime" field
		attr8 = customField{
			title:       "Custom_datetime_attribute",
			description: "Custom_datetime_attribute description",
			varType:     DATETIME,
		}

		// "checkbox" field
		attr9 = customField{
			title:       "Custom_checkbox_attribute",
			description: "Custom_checkbox_attribute description",
			varType:     CHECKBOX,
		}

		// enum field
		attr10 = customField{
			title:       "Custom_enum_attribute",
			description: "Custom_enum_attribute description",
			varType:     ENUM,
			additionalProps: map[string]interface{}{
				"enum": []interface{}{"option_1", "option_2"},
				"_enumProperties": map[string]interface{}{
					"option_1": map[string]interface{}{
						"title":     "Option 1",
						"_disabled": false,
					},
					"option_2": map[string]interface{}{
						"title":     "Option 2",
						"_disabled": false,
					},
				},
			},
		}

		// tag field
		attr11 = customField{
			title:       "Custom_tag_attribute",
			description: "Custom_tag_attribute description",
			varType:     TAG,
			additionalProps: map[string]interface{}{
				"items": map[string]interface{}{
					"minLength": 1,
					"maxLength": 100,
				},
				"minItems":    0,
				"maxItems":    10,
				"uniqueItems": true,
			},
		}

		customProperties = util.GenerateJsonEncodedProperties(
			generateJsonSchemaProperty(attr1.title, attr1.description, attr1.varType,
				generateAdditionalProperties(attr1.additionalProps)),
			generateJsonSchemaProperty(attr2.title, attr2.description, attr2.varType,
				generateAdditionalProperties(attr2.additionalProps)),
			generateJsonSchemaProperty(attr3.title, attr3.description, attr3.varType,
				generateAdditionalProperties(attr3.additionalProps)),
			generateJsonSchemaProperty(attr4.title, attr4.description, attr4.varType,
				generateAdditionalProperties(attr4.additionalProps)),

			generateJsonSchemaProperty(attr5.title, attr5.description, attr5.varType,
				generateAdditionalProperties(attr5.additionalProps)),
			generateJsonSchemaProperty(attr6.title, attr6.description, attr6.varType,
				generateAdditionalProperties(attr6.additionalProps)),

			generateJsonSchemaProperty(attr7.title, attr7.description, attr7.varType, ""),
			generateJsonSchemaProperty(attr8.title, attr8.description, attr8.varType, ""),
			generateJsonSchemaProperty(attr9.title, attr9.description, attr9.varType, ""),

			generateJsonSchemaProperty(attr10.title, attr10.description, attr10.varType,
				generateAdditionalProperties(attr10.additionalProps)),
			generateJsonSchemaProperty(attr11.title, attr11.description, attr11.varType,
				generateAdditionalProperties(attr11.additionalProps)),
		)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Barebones schema. No custom fields
			{
				Config: GenerateWorkitemSchemaResourceBasic(
					schemaResId,
					schemaName,
					schemaDescription,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "name", schemaName),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "description", schemaDescription),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "enabled", util.TrueValue),
				),
			},
			// Update with fields
			{
				Config: GenerateWorkitemSchemaResource(
					schemaResId,
					schemaName,
					schemaDescription,
					customProperties,
					util.TrueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "name", schemaName),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "description", schemaDescription),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "enabled", util.TrueValue),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr1.title+"_"+attr1.varType, attr1),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr2.title+"_"+attr2.varType, attr2),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr3.title+"_"+attr3.varType, attr3),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr4.title+"_"+attr4.varType, attr4),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr5.title+"_"+attr5.varType, attr5),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr6.title+"_"+attr6.varType, attr6),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr7.title+"_"+attr7.varType, attr7),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr8.title+"_"+attr8.varType, attr8),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr9.title+"_"+attr9.varType, attr9),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr10.title+"_"+attr10.varType, attr10),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr11.title+"_"+attr11.varType, attr11),
				),
			},
			// Disable the schema
			{
				Config: GenerateWorkitemSchemaResource(
					schemaResId,
					schemaName,
					schemaDescription,
					customProperties,
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "name", schemaName),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "description", schemaDescription),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "enabled", util.FalseValue),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr1.title+"_"+attr1.varType, attr1),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr2.title+"_"+attr2.varType, attr2),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr3.title+"_"+attr3.varType, attr3),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr4.title+"_"+attr4.varType, attr4),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr5.title+"_"+attr5.varType, attr5),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr6.title+"_"+attr6.varType, attr6),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr7.title+"_"+attr7.varType, attr7),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr8.title+"_"+attr8.varType, attr8),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr9.title+"_"+attr9.varType, attr9),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr10.title+"_"+attr10.varType, attr10),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr11.title+"_"+attr11.varType, attr11),
				),
			},
		},
		CheckDestroy: testVerifyTaskManagementWorkitemSchemaDestroyed,
	})
}

func testVerifyTaskManagementWorkitemSchemaDestroyed(state *terraform.State) error {
	taskMgmtApi := platformclientv2.NewTaskManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_task_management_workitem_schema" {
			continue
		}

		var successPayload map[string]interface{}
		_, resp, err := taskMgmtApi.GetTaskmanagementWorkitemsSchema(rs.Primary.ID)
		if util.IsStatus404(resp) {
			continue // does not exist anymore so considered as deleted
		} else if err != nil {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}

		// Manually check for the 'deleted' property
		err = json.Unmarshal([]byte(resp.RawBody), &successPayload)
		if err != nil {
			return fmt.Errorf("error verifying if workitem schems %s is destroyed: %v", rs.Primary.ID, err)
		}
		if isDeleted, ok := successPayload["deleted"].(bool); ok && isDeleted {
			continue // workitem schema is 'deleted'
		}

		return fmt.Errorf("task management workitem schema (%s) still exists", rs.Primary.ID)
	}

	// Success. All workitem schemas destroyed
	return nil
}

func validateWorkitemSchemaField(resourceName string, fieldName string, checkField customField) resource.TestCheckFunc {
	// Tests for the dynamic properties of the custom field
	additionalFieldsTest := []resource.TestCheckFunc{}
	if checkField.additionalProps != nil {
		for k, v := range checkField.additionalProps {
			vv := reflect.ValueOf(v)
			switch vv.Kind() {
			case reflect.Slice:
				// If slice, do a test for each element
				for _, elem := range v.([]interface{}) {
					additionalFieldsTest = append(additionalFieldsTest,
						util.ValidateValueInJsonAttr(resourceName, "properties", fieldName+"."+k, fmt.Sprint(elem)),
					)
				}
			default:
				additionalFieldsTest = append(additionalFieldsTest,
					util.ValidateValueInJsonAttr(resourceName, "properties", fieldName+"."+k, fmt.Sprint(v)),
				)
			}
		}
	}
	additionalFieldsComposeTest := resource.ComposeTestCheckFunc(additionalFieldsTest...)

	return resource.ComposeTestCheckFunc(
		util.ValidateValueInJsonAttr(resourceName, "properties", fieldName+".title", checkField.title),
		util.ValidateValueInJsonAttr(resourceName, "properties", fieldName+".description", checkField.description),
		validateCustomFieldType(resourceName, fieldName, checkField.varType),
		additionalFieldsComposeTest,
	)
}

// Validate the type of a custom field in the workitem schema
func validateCustomFieldType(resourceName, fieldName, varType string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Failed to find resource %s in state", resourceName)
		}
		resourceID := resourceState.Primary.ID

		jsonAttr, ok := resourceState.Primary.Attributes["properties"]
		if !ok {
			return fmt.Errorf("No 'properties' found for %s in state", resourceID)
		}

		var jsonMap map[string]interface{}
		if err := json.Unmarshal([]byte(jsonAttr), &jsonMap); err != nil {
			return fmt.Errorf("error parsing JSON for %s in state: %v", resourceID, err)
		}

		typeRef, ok := (((jsonMap[fieldName].(map[string]interface{})["allOf"].([]interface{}))[0].(map[string]interface{}))["$ref"]).(string)
		if !ok {
			return fmt.Errorf("error trying to get type of custom field of schema %s", resourceID)
		}

		if typeRef == "#/definitions/"+varType {
			return nil
		}

		return fmt.Errorf(`actual type "%s" does not match expected: "%s"`, typeRef, varType)
	}
}

// json schema properties
func generateJsonSchemaProperty(title, description, coreType, otherFields string) string {
	return fmt.Sprintf(`"%s_%s" = {
		"allOf" = [
          {
            "$ref" = "#/definitions/%s"
          }
        ],
        "title" = "%s"
        "description" = "%s"
        %s
	}
	`, title, coreType, coreType, title, description, otherFields)
}

func generateAdditionalProperties(props map[string]interface{}) string {
	ret := ""
	for k, v := range props {
		vv := reflect.ValueOf(v)
		switch vv.Kind() {
		case reflect.String:
			ret += util.GenerateMapProperty(k, strconv.Quote(v.(string)))
		case reflect.Map:
			ret += util.GenerateMapAttr(k, generateAdditionalProperties(v.(map[string]interface{})))
		case reflect.Slice:
			ret += util.GenerateJsonArrayPropertyEnquote(k, lists.InterfaceListToStrings(v.([]interface{}))...)
		case reflect.Int:
			fallthrough
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			fallthrough
		case reflect.Bool:
			ret += fmt.Sprintf("\"%s\" = %v \n", k, v)
		}
		ret += "\n"
	}
	return ret
}
