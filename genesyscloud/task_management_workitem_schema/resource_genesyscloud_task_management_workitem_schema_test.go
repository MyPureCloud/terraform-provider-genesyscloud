package task_management_workitem_schema

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
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

		// TODO: enum and tag type fields

		customProperties = gcloud.GenerateJsonEncodedProperties(
			generateJsonSchemaProperty(attr1.title, attr1.description, attr1.varType,
				generateStrLengthFields(attr1.additionalProps["minLength"].(int), attr1.additionalProps["maxLength"].(int))),
			generateJsonSchemaProperty(attr2.title, attr2.description, attr2.varType,
				generateStrLengthFields(attr2.additionalProps["minLength"].(int), attr2.additionalProps["maxLength"].(int))),
			generateJsonSchemaProperty(attr3.title, attr3.description, attr3.varType,
				generateStrLengthFields(attr3.additionalProps["minLength"].(int), attr3.additionalProps["maxLength"].(int))),
			generateJsonSchemaProperty(attr4.title, attr4.description, attr4.varType,
				generateStrLengthFields(attr4.additionalProps["minLength"].(int), attr4.additionalProps["maxLength"].(int))),

			generateJsonSchemaProperty(attr5.title, attr5.description, attr5.varType,
				generateNumLimitsFields(attr5.additionalProps["minimum"].(int), attr5.additionalProps["maximum"].(int))),
			generateJsonSchemaProperty(attr6.title, attr6.description, attr6.varType,
				generateNumLimitsFields(attr6.additionalProps["minimum"].(int), attr6.additionalProps["maximum"].(int))),

			generateJsonSchemaProperty(attr7.title, attr7.description, attr7.varType, ""),
			generateJsonSchemaProperty(attr8.title, attr8.description, attr8.varType, ""),
			generateJsonSchemaProperty(attr9.title, attr9.description, attr9.varType, ""),
		)

		config1 = generateWorkitemSchemaResource(
			schemaResId,
			schemaName,
			schemaDescription,
			customProperties,
			gcloud.TrueValue,
		)

		config2 = generateWorkitemSchemaResource(
			schemaResId,
			schemaName,
			schemaDescription,
			customProperties,
			gcloud.FalseValue,
		)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "name", schemaName),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "description", schemaDescription),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "enabled", gcloud.TrueValue),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr1),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr2),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr3),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr4),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr5),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr6),
				),
			},
			// Disable the schema
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "name", schemaName),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "description", schemaDescription),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "enabled", gcloud.FalseValue),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr1),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr2),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr3),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr4),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr5),
					validateWorkitemSchemaField(resourceName+"."+schemaResId, attr6),
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
		if gcloud.IsStatus404(resp) {
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

func validateWorkitemSchemaField(resourceName string, field customField) resource.TestCheckFunc {
	fullFieldName := field.title + "_" + field.varType

	additionalFieldsTest := []resource.TestCheckFunc{}
	if field.additionalProps != nil {
		for k, v := range field.additionalProps {
			additionalFieldsTest = append(additionalFieldsTest,
				gcloud.ValidateValueInJsonAttr(resourceName, "properties", fullFieldName+"."+k, fmt.Sprint(v)),
			)
		}
	}
	additionalFieldsComposeTest := resource.ComposeTestCheckFunc(additionalFieldsTest...)

	return resource.ComposeTestCheckFunc(
		gcloud.ValidateValueInJsonAttr(resourceName, "properties", fullFieldName+".title", field.title),
		gcloud.ValidateValueInJsonAttr(resourceName, "properties", fullFieldName+".description", field.description),
		validateCustomFieldType(resourceName, fullFieldName, field.varType),
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
        "title" = "%s",
        "description" = "%s",
        %s
	}
	`, title, coreType, coreType, title, description, otherFields)
}

func generateStrLengthFields(min int, max int) string {
	return fmt.Sprintf(`"minLength": %v,
		"maxLength": %v
		`, min, max)
}

func generateNumLimitsFields(min int, max int) string {
	return fmt.Sprintf(`"minimum": %v,
		"maximum": %v
		`, min, max)
}

func generateWorkitemSchemaResource(resourceId, name, description, properties, enabledStr string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
		properties = %s
		enabled = %s
	}
	`, resourceName, resourceId, name, description, properties, enabledStr)
}

func generateWorkitemSchemaResourceBasic(resourceId, name, description string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
	}
	`, resourceName, resourceId, name, description)
}
