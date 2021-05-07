package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

var (
	integrationConfigResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Integration config name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"notes": {
				Description: "Integration config notes.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"properties": {
				Description: "Integration config properties.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"advanced": {
				Description: "Integration advanced config.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"credentials": {
				Description: "Credentials required for the integration. The required keys are indicated in the credentials property of the Integration Type.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
)

func getAllIntegrations(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDNameMap, diag.Diagnostics) {
	resources := make(map[string]string)
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		integrations, _, err := integrationAPI.GetIntegrations(100, pageNum, "", nil, "", "")
		if err != nil {
			return nil, diag.Errorf("Failed to get page of integrations: %v", err)
		}

		if integrations.Entities == nil || len(*integrations.Entities) == 0 {
			break
		}

		for _, integration := range *integrations.Entities {
			resources[*integration.Id] = *integration.Name
		}
	}

	return resources, nil
}

func integrationExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllIntegrations),
		RefAttrs: map[string]*RefAttrSettings{
			// TODO: Since now it uses jsonencode, might need to change how export works
			"config.credentials.*": {RefType: "genesyscloud_integration_credentials"},
			"integration_type":     {},
		},
	}
}

func resourceIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Integration",

		CreateContext: createWithPooledClient(createIntegration),
		ReadContext:   readWithPooledClient(readIntegration),
		UpdateContext: updateWithPooledClient(updateIntegration),
		DeleteContext: deleteWithPooledClient(deleteIntegration),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Integration name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"intended_state": {
				Description: "Integration state.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "DISABLED",
			},
			"integration_type": {
				Description: "Integration type.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "Integration type id.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"name": {
							Description: "Integration type name.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"provider": {
							Description: "Integration type provider.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"description": {
							Description: "Integration type description.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"category": {
							Description: "Integration type category.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"config": {
				Description: "Integration config. Each integration type has different schema, use [GET /api/v2/integrations/types/{typeId}/configschemas/{configType}](https://developer.mypurecloud.com/api/rest/v2/integrations/#get-api-v2-integrations-types--typeId--configschemas--configType-) to check schema, then use the correct attribute names for properties.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem:        integrationConfigResource,
			},
		},
	}
}

func createIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	name := d.Get("name").(string)
	intendedState := d.Get("intended_state").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	integrationType := buildIntegrationType(d)

	if integrationType == nil {
		return diag.Errorf("Failed to generate integration type for integration %s", d.Id())
	}

	createIntegration := platformclientv2.Createintegrationrequest{
		Name:            &name,
		IntegrationType: integrationType,
	}

	log.Printf("Creating integration %s", name)
	integration, _, err := integrationAPI.PostIntegrations(createIntegration)

	if err != nil {
		return diag.Errorf("Failed to create integration %s: %s", name, err)
	}

	d.SetId(*integration.Id)

	// Set attributes that can only be modified in a patch
	if d.HasChange(
		"intended_state") {
		log.Printf("Updating additional attributes for integration %s", name)
		_, _, patchErr := integrationAPI.PatchIntegration(d.Id(), platformclientv2.Integration{
			IntendedState: &intendedState,
		}, 25, 1, "", nil, "", "")

		if patchErr != nil {
			return diag.Errorf("Failed to update integration %s: %v", name, patchErr)
		}
	}

	//Update integration config separately
	diagErr := updateIntegrationConfig(d, integrationAPI)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Created integration %s %s", name, *integration.Id)
	return readIntegration(ctx, d, meta)
}

func readIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Reading integration %s", d.Id())
	currentIntegration, resp, getErr := integrationAPI.GetIntegration(d.Id(), 100, 1, "", nil, "", "")

	if getErr != nil {
		if isStatus404(resp) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read integration %s: %s", d.Id(), getErr)
	}

	d.Set("name", *currentIntegration.Name)
	if currentIntegration.IntendedState != nil {
		d.Set("intended_state", *currentIntegration.IntendedState)
	} else {
		d.Set("intended_state", nil)
	}

	// Use returned ID to get current config, which contains complete configuration
	integrationConfig, _, err := integrationAPI.GetIntegrationConfigCurrent(*currentIntegration.Id)

	if err != nil {
		return diag.Errorf("Failed to read config of integration %s: %s", d.Id(), getErr)
	}

	d.Set("config", flattenIntegrationConfig(integrationConfig))

	// Use integration type Id to get complete type object
	typeName := *currentIntegration.IntegrationType.Id
	integrationType, _, err := integrationAPI.GetIntegrationsType(typeName)

	if err != nil {
		return diag.Errorf("Failed to read integration type of integration %s: %s", d.Id(), getErr)
	}

	d.Set("integration_type", flattenIntegrationType(integrationType))

	log.Printf("Read integration %s %s", d.Id(), *currentIntegration.Name)

	return nil
}

func updateIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	name := d.Get("name").(string)
	intendedState := d.Get("intended_state").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	if d.HasChange("intended_state") {

		log.Printf("Updating integration %s", name)

		diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			log.Printf("Updating integration %s", name)

			_, resp, patchErr := integrationAPI.PatchIntegration(d.Id(), platformclientv2.Integration{
				IntendedState: &intendedState,
			}, 25, 1, "", nil, "", "")
			if patchErr != nil {
				return resp, diag.Errorf("Failed to update integration %s: %s", name, patchErr)
			}
			return resp, nil
		})
		if diagErr != nil {
			return diagErr
		}
	}

	diagErr := updateIntegrationConfig(d, integrationAPI)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated integration %s %s", name, d.Id())
	return readIntegration(ctx, d, meta)
}

func deleteIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Deleting integration %s", name)
	return retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		_, resp, err := integrationAPI.DeleteIntegration(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete the integration %s: %s", name, err)
		}
		log.Printf("Deleted integration %s", name)
		return nil, nil
	})
}

func flattenIntegrationType(inteType *platformclientv2.Integrationtype) []interface{} {
	if inteType == nil {
		return nil
	}
	var (
		typeId          string
		typeName        string
		typeDescription string
		typeProvider    string
		typeCategory    string
	)

	if inteType.Id != nil {
		typeId = *inteType.Id
	}
	if inteType.Name != nil {
		typeName = *inteType.Name
	}
	if inteType.Description != nil {
		typeDescription = *inteType.Description
	}
	if inteType.Provider != nil {
		typeProvider = *inteType.Provider
	}
	if inteType.Category != nil {
		typeCategory = *inteType.Category
	}

	return []interface{}{map[string]interface{}{
		"id":          typeId,
		"name":        typeName,
		"description": typeDescription,
		"provider":    typeProvider,
		"category":    typeCategory,
	}}
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
		configCredentials string
	)

	if config.Name != nil {
		configName = *config.Name
	}
	if config.Notes != nil {
		configNotes = *config.Notes
	}
	if config.Properties != nil {
		propJSONStr, err := json.Marshal(*config.Properties)
		if err != nil {
			fmt.Errorf("Failed to marshal integration config properties. Error message: %s", err)
		} else {
			configProperties = string(propJSONStr)
		}
	}
	if config.Advanced != nil {
		advJSONStr, err := json.Marshal(*config.Advanced)
		if err != nil {
			fmt.Errorf("Failed to marshal integration config advanced properties. Error message: %s", err)
		} else {
			configAdvanced = string(advJSONStr)
		}
	}
	if config.Credentials != nil {
		//Only takes id for credential info
		credJSON := make(map[string]interface{})
		for k, v := range *config.Credentials {
			credJSON[k] = map[string]interface{}{
				"id": *v.Id,
			}
		}
		credJSONStr, err := json.Marshal(credJSON)
		if err != nil {
			fmt.Errorf("Failed to marshal integration config credential properties. Error message: %s", err)
		} else {
			configCredentials = string(credJSONStr)
		}
	}

	return []interface{}{map[string]interface{}{
		"name":        configName,
		"notes":       configNotes,
		"properties":  configProperties,
		"advanced":    configAdvanced,
		"credentials": configCredentials,
	}}
}

func buildIntegrationType(d *schema.ResourceData) *platformclientv2.Integrationtype {
	if typeInput := d.Get("integration_type").([]interface{}); typeInput != nil {
		var inteType platformclientv2.Integrationtype
		if len(typeInput) > 0 {
			inputMap := typeInput[0].(map[string]interface{})
			// Only set non-empty values.
			if id := inputMap["id"].(string); len(id) > 0 {
				inteType.Id = &id
			}
			if name := inputMap["name"].(string); len(name) > 0 {
				inteType.Name = &name
			}
			if description := inputMap["description"].(string); len(description) > 0 {
				inteType.Description = &description
			}
			if provider := inputMap["provider"].(string); len(provider) > 0 {
				inteType.Provider = &provider
			}
			if category := inputMap["category"].(string); len(category) > 0 {
				inteType.Category = &category
			}
		}
		return &inteType
	}
	return nil
}

func updateIntegrationConfig(d *schema.ResourceData, integrationAPI *platformclientv2.IntegrationsApi) diag.Diagnostics {
	if d.HasChange("config") {
		if configInput := d.Get("config").([]interface{}); configInput != nil {
			var config platformclientv2.Integrationconfiguration
			integrationConfig, _, err := integrationAPI.GetIntegrationConfigCurrent(d.Id())

			if err != nil {
				diag.Errorf("Failed to get the integration config for integration %s before updating its config: %s", d.Id(), err)
			}
			if len(configInput) > 0 {
				configMap := configInput[0].(map[string]interface{})

				name := configMap["name"].(string)
				config.Name = &name

				config.Version = integrationConfig.Version

				notes := configMap["notes"].(string)
				config.Notes = &notes

				propJSON := make(map[string]interface{})
				if properties := configMap["properties"].(string); len(properties) > 0 {
					if err := json.Unmarshal([]byte(properties), &propJSON); err != nil {
						return diag.Errorf("Failed to convert properties string to JSON for integration %s: %s", d.Id(), err)
					}
				}
				config.Properties = &propJSON

				advJSON := make(map[string]interface{})
				if advanced := configMap["advanced"].(string); len(advanced) > 0 {
					if err := json.Unmarshal([]byte(advanced), &advJSON); err != nil {
						return diag.Errorf("Failed to convert advanced property string to JSON for integration %s: %s", d.Id(), err)
					}
				}
				config.Advanced = &advJSON

				credentialMap := make(map[string]platformclientv2.Credentialinfo)
				if credentials := configMap["credentials"].(string); len(credentials) > 0 {
					credJSON := make(map[string]interface{})
					if err := json.Unmarshal([]byte(credentials), &credJSON); err != nil {
						return diag.Errorf("Failed to convert credentials property string to JSON for integration %s: %s", d.Id(), err)
					}
					for k, v := range credJSON {
						for _, value := range v.(map[string]interface{}) {
							credentialId := value.(string)
							credentialMap[k] = platformclientv2.Credentialinfo{Id: &credentialId}
						}
					}
				}
				config.Credentials = &credentialMap

			}

			diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
				_, resp, err := integrationAPI.PutIntegrationConfigCurrent(d.Id(), config)
				if err != nil {
					return resp, diag.Errorf("Failed to update config for integration %s: %s", d.Id(), err)
				}
				return nil, nil
			})
			if diagErr != nil {
				return diagErr
			}
		}
	}
	return nil
}
