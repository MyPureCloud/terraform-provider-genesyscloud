package business_rules_schema

import (
	"encoding/json"
	"fmt"

	"log"
	"net/http"
	"net/url"

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

func businessRulesSchemaFtIsEnabled() (bool, *http.Response) {
	clientConfig := platformclientv2.GetDefaultConfiguration()
	client := &http.Client{}
	baseURL := clientConfig.BasePath + "/api/v2/businessrules/schemas"

	u, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+clientConfig.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, resp
	}

	return false, resp
}
