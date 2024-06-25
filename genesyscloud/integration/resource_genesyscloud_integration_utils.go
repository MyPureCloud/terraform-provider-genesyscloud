package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_integration_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.

Note:  Look for opportunities to minimize boilerplate code using functions and Generics
*/

// flattenIntegrationConfig converts a platformclientv2.Integrationconfiguration into a map and then into single-element array for consumption by Terraform
func flattenIntegrationConfig(config *platformclientv2.Integrationconfiguration) []interface{} {
	if config == nil {
		return nil
	}
	var (
		configName        string
		configNotes       string
		configProperties  string
		configAdvanced    string
		configCredentials map[string]interface{}
	)

	if config.Name != nil {
		configName = *config.Name
	}
	if config.Notes != nil {
		if *config.Notes == "node_dynamodb_empty_string" {
			*config.Notes = ""
		}
		configNotes = *config.Notes
	}
	if config.Properties != nil {
		propJSONStr, err := json.Marshal(*config.Properties)
		if err != nil {
			log.Printf("Failed to marshal integration config properties. Error message: %s", err)
		} else {
			configProperties = string(propJSONStr)
		}
	}
	if config.Advanced != nil {
		advJSONStr, err := json.Marshal(*config.Advanced)
		if err != nil {
			log.Printf("Failed to marshal integration config advanced properties. Error message: %s", err)
		} else {
			configAdvanced = string(advJSONStr)
		}
	}
	if config.Credentials != nil {
		configCredentials = flattenConfigCredentials(*config.Credentials)
	}

	return []interface{}{map[string]interface{}{
		"name":        configName,
		"notes":       configNotes,
		"properties":  configProperties,
		"advanced":    configAdvanced,
		"credentials": configCredentials,
	}}
}

// flattenConfigCredentials converts a map of platformclientv2.Credentialinfo into a map of only the credential IDs for consumption by Terraform
func flattenConfigCredentials(credentials map[string]platformclientv2.Credentialinfo) map[string]interface{} {
	if len(credentials) == 0 {
		return nil
	}

	results := make(map[string]interface{})
	for k, v := range credentials {
		results[k] = *v.Id
	}
	return results
}

// updateIntegrationConfigFromResourceData takes the integrationsProxy to update updates the config of an integration
// Returns a diag error and the name of the integration
func updateIntegrationConfigFromResourceData(ctx context.Context, d *schema.ResourceData, p *integrationsProxy) (diag.Diagnostics, string) {
	if d.HasChange("config") {
		if configInput := d.Get("config").([]interface{}); configInput != nil {

			integrationConfig, resp, err := p.getIntegrationConfig(ctx, d.Id())
			if err != nil {
				return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get the integration config for integration %s before updating its config error: %s", d.Id(), err), resp), ""
			}

			name := *integrationConfig.Name
			notes := *integrationConfig.Notes
			propJSON := *integrationConfig.Properties
			advJSON := *integrationConfig.Advanced
			credential := *integrationConfig.Credentials

			if len(configInput) > 0 {
				configMap := configInput[0].(map[string]interface{})

				if configMap["name"].(string) != "" {
					name = configMap["name"].(string)
				}

				notes = configMap["notes"].(string)

				if properties := configMap["properties"].(string); len(properties) > 0 {
					if err := json.Unmarshal([]byte(properties), &propJSON); err != nil {
						return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to convert properties string to JSON for integration %s", d.Id()), err), name
					}
				}

				if advanced := configMap["advanced"].(string); len(advanced) > 0 {
					if err := json.Unmarshal([]byte(advanced), &advJSON); err != nil {
						return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to convert advanced property string to JSON for integration %s", d.Id()), err), name
					}
				}

				credential = buildConfigCredentials(configMap["credentials"].(map[string]interface{}))
			}

			diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {

				// Get latest config version
				integrationConfig, resp, err := p.getIntegrationConfig(ctx, d.Id())
				if err != nil {
					return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get the integration config for integration %s before updating its config. error: %s", d.Id(), err), resp)
				}

				_, resp, err = p.updateIntegrationConfig(ctx, d.Id(), &platformclientv2.Integrationconfiguration{
					Name:        &name,
					Notes:       &notes,
					Version:     integrationConfig.Version,
					Properties:  &propJSON,
					Advanced:    &advJSON,
					Credentials: &credential,
				})
				if err != nil {
					return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update config for integration %s error: %s", d.Id(), err), resp)
				}
				return nil, nil
			})
			if diagErr != nil {
				return diagErr, ""
			}
		}
	}
	return nil, ""
}

// buildConfigCredentials takes a map of credential IDs and turns it into a map of platformclientv2.Credentialinfo
func buildConfigCredentials(credentials map[string]interface{}) map[string]platformclientv2.Credentialinfo {
	results := make(map[string]platformclientv2.Credentialinfo)
	if len(credentials) > 0 {
		for k, v := range credentials {
			credID := v.(string)
			results[k] = platformclientv2.Credentialinfo{Id: &credID}
		}
		return results
	}
	return results
}

// GenerateIntegrationResource builds the terraform string for creating an integration
func GenerateIntegrationResource(resourceID string, intendedState string, integrationType string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_integration" "%s" {
        intended_state = %s
        integration_type = %s
        %s
	}
	`, resourceID, intendedState, integrationType, strings.Join(attrs, "\n"))
}

func GenerateIntegrationConfig(name string, notes string, cred string, props string, adv string) string {
	return fmt.Sprintf(`config {
        name = %s
        notes = %s
        credentials = {
            %s
        }
        properties = %s
        advanced = %s
	}
	`, name, notes, cred, props, adv)
}
