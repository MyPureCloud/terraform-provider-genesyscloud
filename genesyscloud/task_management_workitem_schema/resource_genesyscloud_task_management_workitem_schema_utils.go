package task_management_workitem_schema

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

const (
	TEXT       = "text"
	LONGTEXT   = "longtext"
	URL        = "url"
	IDENTIFIER = "identifier"
	ENUM       = "enum"
	DATE       = "date"
	DATETIME   = "datetime"
	INTEGER    = "integer"
	NUMBER     = "number"
	CHECKBOX   = "checkbox"
	TAG        = "tag"
)

type customField struct {
	title           string
	description     string
	varType         string
	additionalProps map[string]interface{}
}

// BuildSdkWorkitemSchema takes the resource data and builds the SDK platformclientv2.Dataschema
func BuildSdkWorkitemSchema(d *schema.ResourceData, version *int) (*platformclientv2.Dataschema, error) {
	// body for the creation/update of the schema
	dataSchema := &platformclientv2.Dataschema{
		Name:    platformclientv2.String(d.Get("name").(string)),
		Version: version,
		JsonSchema: &platformclientv2.Jsonschemadocument{
			Schema:      platformclientv2.String("http://json-schema.org/draft-04/schema#"),
			Title:       platformclientv2.String(d.Get("name").(string)),
			Description: platformclientv2.String(d.Get("description").(string)),
		},
		Enabled: platformclientv2.Bool(d.Get("enabled").(bool)),
	}

	// Custom attributes for the schema
	if d.Get("properties") != "" {
		var properties map[string]interface{}
		if err := json.Unmarshal([]byte(d.Get("properties").(string)), &properties); err != nil {
			return nil, err
		}

		dataSchema.JsonSchema.Properties = &properties
	}

	return dataSchema, nil
}

// GenerateWorkitemSchemaResourceBasic is a public util method to generate the simplest
// schema terraform resource for testing
func GenerateWorkitemSchemaResourceBasic(resourceId, name, description string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
	}
	`, resourceName, resourceId, name, description)
}

func GenerateWorkitemSchemaResource(resourceId, name, description, properties, enabledStr string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
		properties = %s
		enabled = %s
	}
	`, resourceName, resourceId, name, description, properties, enabledStr)
}
