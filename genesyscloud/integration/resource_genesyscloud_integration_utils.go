package integration

import (
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

// getIntegrationFromResourceData maps data from schema ResourceData object to a platformclientv2.Integration
func getIntegrationFromResourceData(d *schema.ResourceData) platformclientv2.Integration {
	intendedState := d.Get("intended_state").(string)
	integrationType := d.Get("integration_type").(string)

	return platformclientv2.Integration{
		IntegrationType: &platformclientv2.Integrationtype{
			Id: &integrationType,
		},
		IntendedState: &intendedState,
	}
}

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

func updateIntegrationConfig(d *schema.ResourceData, integrationAPI *platformclientv2.IntegrationsApi) (diag.Diagnostics, string) {
	if d.HasChange("config") {
		if configInput := d.Get("config").([]interface{}); configInput != nil {

			integrationConfig, _, err := integrationAPI.GetIntegrationConfigCurrent(d.Id())
			if err != nil {
				return diag.Errorf("Failed to get the integration config for integration %s before updating its config: %s", d.Id(), err), ""
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
						return diag.Errorf("Failed to convert properties string to JSON for integration %s: %s", d.Id(), err), name
					}
				}

				if advanced := configMap["advanced"].(string); len(advanced) > 0 {
					if err := json.Unmarshal([]byte(advanced), &advJSON); err != nil {
						return diag.Errorf("Failed to convert advanced property string to JSON for integration %s: %s", d.Id(), err), name
					}
				}

				credential = buildConfigCredentials(configMap["credentials"].(map[string]interface{}))
			}

			diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {

				// Get latest config version
				integrationConfig, resp, err := integrationAPI.GetIntegrationConfigCurrent(d.Id())
				if err != nil {
					return resp, diag.Errorf("Failed to get the integration config for integration %s before updating its config: %s", d.Id(), err)
				}

				_, resp, err = integrationAPI.PutIntegrationConfigCurrent(d.Id(), platformclientv2.Integrationconfiguration{
					Name:        &name,
					Notes:       &notes,
					Version:     integrationConfig.Version,
					Properties:  &propJSON,
					Advanced:    &advJSON,
					Credentials: &credential,
				})
				if err != nil {
					return resp, diag.Errorf("Failed to update config for integration %s: %s", d.Id(), err)
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
