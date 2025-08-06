package business_rules_schema

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

const (
	BOOLEAN              = "boolean"
	DATE                 = "date"
	DATETIME             = "datetime"
	ENUM                 = "enum"
	INTEGER              = "integer"
	NUMBER               = "number"
	BUSINESS_RULES_QUEUE = "businessRulesQueue"
	STRING               = "string"
)

type customField struct {
	title           string
	description     string
	varType         string
	additionalProps map[string]interface{}
}

// BuildSdkBusinessRulesSchema takes the resource data and builds the SDK platformclientv2.Dataschema
func BuildSdkBusinessRulesSchema(d *schema.ResourceData, version *int) (*platformclientv2.Dataschema, error) {
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
	if propertiesStr, ok := d.Get("properties").(string); ok && propertiesStr != "" {
		var properties map[string]interface{}
		if err := json.Unmarshal([]byte(propertiesStr), &properties); err != nil {
			return nil, fmt.Errorf("failed to unmarshal properties JSON: %w", err)
		}
		dataSchema.JsonSchema.Properties = &properties
	}

	return dataSchema, nil
}

// GenerateBusinessRulesSchemaResourceBasic is a public util method to generate the simplest
// schema terraform resource for testing
func GenerateBusinessRulesSchemaResourceBasic(resourceLabel, name, description string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
	}
	`, ResourceType, resourceLabel, name, description)
}

func GenerateBusinessRulesSchemaResource(resourceLabel, name, description, properties, enabledStr string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
		properties = %s
		enabled = %s
	}
	`, ResourceType, resourceLabel, name, description, properties, enabledStr)
}
