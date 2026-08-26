package integration_config

import (
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

// buildIntegrationConfig converts Terraform ResourceData into SDK Integrationconfiguration struct
func buildIntegrationConfig(d *schema.ResourceData, currentVersion *int) *platformclientv2.Integrationconfiguration {
	config := &platformclientv2.Integrationconfiguration{
		Version: currentVersion,
	}

	if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		config.Name = &name
	}

	notes := d.Get("notes").(string)
	config.Notes = &notes

	// Properties - must be an object, never nil
	if v, ok := d.GetOk("properties"); ok {
		propertiesStr := v.(string)
		if len(propertiesStr) > 0 {
			var properties interface{}
			if err := json.Unmarshal([]byte(propertiesStr), &properties); err != nil {
				log.Printf("Failed to unmarshal properties JSON: %s", err)
			} else {
				config.Properties = &properties
			}
		}
	}
	if config.Properties == nil {
		var emptyObj interface{} = map[string]interface{}{}
		config.Properties = &emptyObj
	}

	// Advanced - must be an object, never nil
	if v, ok := d.GetOk("advanced"); ok {
		advancedStr := v.(string)
		if len(advancedStr) > 0 {
			var advanced interface{}
			if err := json.Unmarshal([]byte(advancedStr), &advanced); err != nil {
				log.Printf("Failed to unmarshal advanced JSON: %s", err)
			} else {
				config.Advanced = &advanced
			}
		}
	}
	if config.Advanced == nil {
		var emptyObj interface{} = map[string]interface{}{}
		config.Advanced = &emptyObj
	}

	// Credentials - must be an object, never nil
	if v, ok := d.GetOk("credentials"); ok {
		credentials := buildCredentials(v.(map[string]interface{}))
		config.Credentials = &credentials
	}
	if config.Credentials == nil {
		emptyCredentials := make(map[string]platformclientv2.Credentialinfo)
		config.Credentials = &emptyCredentials
	}

	return config
}

// buildCredentials converts Terraform map (credType → credId) into SDK map (credType → Credentialinfo)
func buildCredentials(credMap map[string]interface{}) map[string]platformclientv2.Credentialinfo {
	results := make(map[string]platformclientv2.Credentialinfo)
	for credType, credId := range credMap {
		id := credId.(string)
		results[credType] = platformclientv2.Credentialinfo{Id: &id}
	}
	return results
}

// flattenIntegrationConfig converts SDK Integrationconfiguration into Terraform state values and sets them on d
func flattenIntegrationConfig(d *schema.ResourceData, config *platformclientv2.Integrationconfiguration) {
	if config == nil {
		return
	}

	if config.Name != nil {
		_ = d.Set("name", *config.Name)
	}

	if config.Notes != nil {
		notes := *config.Notes
		if notes == "node_dynamodb_empty_string" {
			notes = ""
		}
		_ = d.Set("notes", notes)
	}

	if config.Properties != nil {
		propJSON, err := json.Marshal(*config.Properties)
		if err != nil {
			log.Printf("Failed to marshal integration config properties: %s", err)
		} else {
			_ = d.Set("properties", string(propJSON))
		}
	}

	if config.Advanced != nil {
		advJSON, err := json.Marshal(*config.Advanced)
		if err != nil {
			log.Printf("Failed to marshal integration config advanced: %s", err)
		} else {
			_ = d.Set("advanced", string(advJSON))
		}
	}

	if config.Credentials != nil {
		_ = d.Set("credentials", flattenCredentials(*config.Credentials))
	}
}

// flattenCredentials converts SDK map (credType → Credentialinfo) into Terraform map (credType → credId)
func flattenCredentials(credentials map[string]platformclientv2.Credentialinfo) map[string]interface{} {
	if len(credentials) == 0 {
		return nil
	}
	results := make(map[string]interface{})
	for credType, credInfo := range credentials {
		if credInfo.Id != nil {
			results[credType] = *credInfo.Id
		}
	}
	return results
}
