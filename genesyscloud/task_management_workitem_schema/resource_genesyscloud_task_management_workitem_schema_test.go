package task_management_workitem_schema

import (
	"fmt"
	"strconv"
	"strings"
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
		schemaTitle       = "tf_schema_" + uuid.NewString()
		schemaDescription = "created for CX as Code test case"

		// custom attribute 1
		attr1Title       = "custom_attribute"
		attr1Description = "custom_attribute description"
		attr1Type        = TEXT
		attr1Min         = 1
		attr1Max         = 50

		config = generateWorkitemSchemaResource(
			schemaResId,
			gcloud.GenerateJsonEncodedProperties(
				generateJsonSchema(strconv.Quote(schemaTitle), strconv.Quote(schemaDescription),
					generateJsonSchemaProperty(attr1Title, attr1Description, attr1Type,
						generateStrLengthFields(attr1Min, attr1Max)),
				),
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
					resource.TestCheckResourceAttr(resourceName+"."+schemaResId, "name", schemaTitle),
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

		workbin, resp, err := taskMgmtApi.GetTaskmanagementWorkitemsSchema(rs.Primary.ID)
		if workbin != nil {
			return fmt.Errorf("task management workitem schema (%s) still exists", rs.Primary.ID)
		} else if gcloud.IsStatus404(resp) {
			// Workitem schema not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All workitem schemas destroyed
	return nil
}

func generateJsonSchema(title string, description string, properties ...string) string {
	return fmt.Sprintf(`"$schema": "http://json-schema.org/draft-04/schema#",
		"title": %s,
		"description": %s,
		"properties": {
			%s
		}
	`, title, description, strings.Join(properties, "\n"))
}

// json schema properties
func generateJsonSchemaProperty(title, description, coreType, otherFields string) string {
	return fmt.Sprintf(`"%s_%s": {
		"allOf": [
          {
            "$ref": "#/definitions/%s"
          }
        ],
        "title": "%s",
        "description": "%s",
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

func generateWorkitemSchemaResource(resourceId string, json_schema string, enabledStr string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		json_schema = %s
		enabled = %s
	}
	`, resourceName, resourceId, json_schema, enabledStr)
}
