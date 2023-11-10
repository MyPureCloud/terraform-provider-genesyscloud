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
		attr1 = textTypeField{
			title:       "custom_text_attribute",
			description: "custom_text_attribute description",
			varType:     TEXT,
			minLength:   1,
			maxLength:   50,
		}

		// "longtext" field
		attr2 = textTypeField{
			title:       "Custom_longtext_attribute",
			description: "Custom_longtext_attribute description",
			varType:     LONGTEXT,
			minLength:   1,
			maxLength:   50,
		}

		// "url" field
		attr3 = textTypeField{
			title:       "Custom_url_attribute",
			description: "Custom_url_attribute description",
			varType:     URL,
			minLength:   1,
			maxLength:   50,
		}

		// "identifier" field
		attr4 = textTypeField{
			title:       "Custom_identifier_attribute",
			description: "Custom_identifier_attribute description",
			varType:     IDENTIFIER,
			minLength:   1,
			maxLength:   50,
		}

		config = generateWorkitemSchemaResource(
			schemaResId,
			schemaName,
			schemaDescription,
			gcloud.GenerateJsonEncodedProperties(
				generateJsonSchemaProperty(attr1.title, attr1.description, attr1.varType,
					generateStrLengthFields(attr1.minLength, attr1.maxLength)),
				generateJsonSchemaProperty(attr2.title, attr2.description, attr2.varType,
					generateStrLengthFields(attr2.minLength, attr2.maxLength)),
				generateJsonSchemaProperty(attr3.title, attr3.description, attr3.varType,
					generateStrLengthFields(attr3.minLength, attr3.maxLength)),
				generateJsonSchemaProperty(attr4.title, attr4.description, attr4.varType,
					generateStrLengthFields(attr4.minLength, attr4.maxLength)),
			),
			gcloud.TrueValue,
		)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Default division
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "name", schemaName),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "description", schemaDescription),
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "enabled", gcloud.TrueValue),
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
