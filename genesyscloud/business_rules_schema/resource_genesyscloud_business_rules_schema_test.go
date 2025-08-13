package business_rules_schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The resource_genesyscloud_business_rules_schema_test.go contains all of the test cases for running the resource
tests for business_rules_schema.
*/

func TestAccResourceBusinessRulesSchema(t *testing.T) {
	t.Parallel()

	enabled, resp := businessRulesSchemaFtIsEnabled()
	if !enabled {
		t.Skipf("Skipping test as business rules schema is not configured: %s", resp.Status)
		return
	}

	var (
		schemaResourceLabel      = "tf_schema_1"
		schemaName               = "tf_schema_" + uuid.NewString()
		schemaDescription        = "created for CX as Code test case"
		updatedSchemaDescription = "updated description for CX as Code test case"

		// "boolean" field
		attr1 = customField{
			title:       "custom_boolean_attribute",
			description: "custom_boolean_attribute description",
			varType:     BOOLEAN,
		}

		// "date" field
		attr2 = customField{
			title:       "Custom_date_attribute",
			description: "Custom_date_attribute description",
			varType:     DATE,
		}

		// "datetime" field
		attr3 = customField{
			title:       "Custom_datetime_attribute",
			description: "Custom_datetime_attribute description",
			varType:     DATETIME,
		}

		// enum field
		attr4 = customField{
			title:       "Custom_enum_attribute",
			description: "Custom_enum_attribute description",
			varType:     ENUM,
			additionalProps: map[string]interface{}{
				"enum": []interface{}{"option_1", "option_2"},
				"_enumProperties": map[string]interface{}{
					"option_1": map[string]interface{}{
						"title": "Option 1",
					},
					"option_2": map[string]interface{}{
						"title": "Option 2",
					},
				},
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

		// "queue" field
		attr7 = customField{
			title:       "Custom_queue_attribute",
			description: "Custom_queue_attribute description",
			varType:     BUSINESS_RULES_QUEUE,
		}

		// "string" field
		attr8 = customField{
			title:       "custom_string_attribute",
			description: "custom_string_attribute description",
			varType:     STRING,
			additionalProps: map[string]interface{}{
				"minLength": 1,
				"maxLength": 100,
			},
		}

		customProperties = util.GenerateJsonEncodedProperties(
			generateJsonSchemaProperty(attr1.title, attr1.description, attr1.varType, ""),
			generateJsonSchemaProperty(attr2.title, attr2.description, attr2.varType, ""),
			generateJsonSchemaProperty(attr3.title, attr3.description, attr3.varType, ""),
			generateJsonSchemaProperty(attr4.title, attr4.description, attr4.varType, generateAdditionalProperties(attr4.additionalProps)),
			generateJsonSchemaProperty(attr5.title, attr5.description, attr5.varType, generateAdditionalProperties(attr5.additionalProps)),
			generateJsonSchemaProperty(attr6.title, attr6.description, attr6.varType, generateAdditionalProperties(attr6.additionalProps)),
			generateJsonSchemaProperty(attr7.title, attr7.description, attr7.varType, ""),
			generateJsonSchemaProperty(attr8.title, attr8.description, attr8.varType, generateAdditionalProperties(attr8.additionalProps)),
		)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Barebones schema. No custom fields
			{
				Config: GenerateBusinessRulesSchemaResourceBasic(
					schemaResourceLabel,
					schemaName,
					schemaDescription,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+schemaResourceLabel, "name", schemaName),
					resource.TestCheckResourceAttr(ResourceType+"."+schemaResourceLabel, "description", schemaDescription),
					resource.TestCheckResourceAttr(ResourceType+"."+schemaResourceLabel, "enabled", util.TrueValue),
				),
			},
			// Update with fields
			{
				Config: GenerateBusinessRulesSchemaResource(
					schemaResourceLabel,
					schemaName,
					schemaDescription,
					customProperties,
					util.TrueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+schemaResourceLabel, "name", schemaName),
					resource.TestCheckResourceAttr(ResourceType+"."+schemaResourceLabel, "description", schemaDescription),
					resource.TestCheckResourceAttr(ResourceType+"."+schemaResourceLabel, "enabled", util.TrueValue),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr1.title+"_"+attr1.varType, attr1),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr2.title+"_"+attr2.varType, attr2),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr3.title+"_"+attr3.varType, attr3),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr4.title+"_"+attr4.varType, attr4),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr5.title+"_"+attr5.varType, attr5),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr6.title+"_"+attr6.varType, attr6),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr7.title+"_"+attr7.varType, attr7),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr8.title+"_"+attr8.varType, attr8),
				),
			},
			// Update the description of the schema
			{
				Config: GenerateBusinessRulesSchemaResource(
					schemaResourceLabel,
					schemaName,
					updatedSchemaDescription,
					customProperties,
					util.TrueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+schemaResourceLabel, "name", schemaName),
					resource.TestCheckResourceAttr(ResourceType+"."+schemaResourceLabel, "description", updatedSchemaDescription),
					resource.TestCheckResourceAttr(ResourceType+"."+schemaResourceLabel, "enabled", util.TrueValue),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr1.title+"_"+attr1.varType, attr1),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr2.title+"_"+attr2.varType, attr2),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr3.title+"_"+attr3.varType, attr3),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr4.title+"_"+attr4.varType, attr4),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr5.title+"_"+attr5.varType, attr5),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr6.title+"_"+attr6.varType, attr6),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr7.title+"_"+attr7.varType, attr7),
					validateBusinessRulesSchemaField(ResourceType+"."+schemaResourceLabel, attr8.title+"_"+attr8.varType, attr8),
				),
			},
		},
		CheckDestroy: testVerifyBusinessRulesSchemaDestroyed,
	})
}

func testVerifyBusinessRulesSchemaDestroyed(state *terraform.State) error {
	businessRulesApi := platformclientv2.NewBusinessRulesApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_business_rules_schema" {
			continue
		}

		var successPayload map[string]interface{}
		_, resp, err := businessRulesApi.GetBusinessrulesSchema(rs.Primary.ID)
		if util.IsStatus404(resp) {
			continue // does not exist anymore so considered as deleted
		} else if err != nil {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}

		// Manually check for the 'deleted' property
		err = json.Unmarshal([]byte(resp.RawBody), &successPayload)
		if err != nil {
			return fmt.Errorf("error verifying if business rules schema %s is destroyed: %v", rs.Primary.ID, err)
		}
		if isDeleted, ok := successPayload["deleted"].(bool); ok && isDeleted {
			continue // business rules schema is 'deleted'
		}

		return fmt.Errorf("business rules schema (%s) still exists", rs.Primary.ID)
	}

	// Success. All business rules schemas destroyed
	return nil
}

func validateBusinessRulesSchemaField(resourcePath string, fieldName string, checkField customField) resource.TestCheckFunc {
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
						util.ValidateValueInJsonAttr(resourcePath, "properties", fieldName+"."+k, fmt.Sprint(elem)),
					)
				}
			default:
				additionalFieldsTest = append(additionalFieldsTest,
					util.ValidateValueInJsonAttr(resourcePath, "properties", fieldName+"."+k, fmt.Sprint(v)),
				)
			}
		}
	}
	additionalFieldsComposeTest := resource.ComposeTestCheckFunc(additionalFieldsTest...)

	return resource.ComposeTestCheckFunc(
		util.ValidateValueInJsonAttr(resourcePath, "properties", fieldName+".title", checkField.title),
		util.ValidateValueInJsonAttr(resourcePath, "properties", fieldName+".description", checkField.description),
		validateCustomFieldType(resourcePath, fieldName, checkField.varType),
		additionalFieldsComposeTest,
	)
}

// Validate the type of a custom field in the business rules schema
func validateCustomFieldType(resourcePath, fieldName, varType string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("Failed to find resource %s in state", resourcePath)
		}
		resourceLabel := resourceState.Primary.ID

		jsonAttr, ok := resourceState.Primary.Attributes["properties"]
		if !ok {
			return fmt.Errorf("No 'properties' found for %s in state", resourceLabel)
		}

		var jsonMap map[string]interface{}
		if err := json.Unmarshal([]byte(jsonAttr), &jsonMap); err != nil {
			return fmt.Errorf("error parsing JSON for %s in state: %v", resourceLabel, err)
		}

		typeRef, ok := (((jsonMap[fieldName].(map[string]interface{})["allOf"].([]interface{}))[0].(map[string]interface{}))["$ref"]).(string)
		if !ok {
			return fmt.Errorf("error trying to get type of custom field of schema %s", resourceLabel)
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
