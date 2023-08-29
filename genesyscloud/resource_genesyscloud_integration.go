package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

var (
	integrationConfigResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Integration name.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"notes": {
				Description: "Integration notes.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"properties": {
				Description:      "Integration config properties (JSON string).",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},
			"advanced": {
				Description:      "Integration advanced config (JSON string).",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},
			"credentials": {
				Description: "Credentials required for the integration. The required keys are indicated in the credentials property of the Integration Type.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
)

func getAllIntegrations(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		integrations, _, err := integrationAPI.GetIntegrations(pageSize, pageNum, "", nil, "", "")
		if err != nil {
			return nil, diag.Errorf("Failed to get page of integrations: %v", err)
		}

		if integrations.Entities == nil || len(*integrations.Entities) == 0 {
			break
		}

		for _, integration := range *integrations.Entities {
			resources[*integration.Id] = &resourceExporter.ResourceMeta{Name: *integration.Name}
		}
	}

	return resources, nil
}

func IntegrationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllIntegrations),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"config.credentials.*": {RefType: "genesyscloud_integration_credential"},
		},
		JsonEncodeAttributes: []string{"config.properties", "config.advanced"},
		EncodedRefAttrs: map[*resourceExporter.JsonEncodeRefAttr]*resourceExporter.RefAttrSettings{
			{Attr: "config.properties", NestedAttr: "groups"}: {RefType: "genesyscloud_group"},
		},
	}
}

func ResourceIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Integration",

		CreateContext: CreateWithPooledClient(createIntegration),
		ReadContext:   ReadWithPooledClient(readIntegration),
		UpdateContext: UpdateWithPooledClient(updateIntegration),
		DeleteContext: DeleteWithPooledClient(deleteIntegration),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"intended_state": {
				Description:  "Integration state (ENABLED | DISABLED | DELETED).",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DISABLED",
				ValidateFunc: validation.StringInSlice([]string{"ENABLED", "DISABLED", "DELETED"}, false),
			},
			"integration_type": {
				Description: "Integration type.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"config": {
				Description: "Integration config. Each integration type has different schema, use [GET /api/v2/integrations/types/{typeId}/configschemas/{configType}](https://developer.mypurecloud.com/api/rest/v2/integrations/#get-api-v2-integrations-types--typeId--configschemas--configType-) to check schema, then use the correct attribute names for properties.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        integrationConfigResource,
			},
		},
	}
}

func createIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	intendedState := d.Get("intended_state").(string)
	integrationType := d.Get("integration_type").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	createIntegration := platformclientv2.Createintegrationrequest{
		IntegrationType: &platformclientv2.Integrationtype{
			Id: &integrationType,
		},
	}

	integration, _, err := integrationAPI.PostIntegrations(createIntegration)

	if err != nil {
		return diag.Errorf("Failed to create integration : %s", err)
	}

	d.SetId(*integration.Id)

	//Update integration config separately
	diagErr, name := updateIntegrationConfig(d, integrationAPI)
	if diagErr != nil {
		return diagErr
	}

	// Set attributes that can only be modified in a patch
	if d.HasChange(
		"intended_state") {
		log.Printf("Updating additional attributes for integration %s", name)
		const pageSize = 25
		const pageNum = 1
		_, _, patchErr := integrationAPI.PatchIntegration(d.Id(), pageSize, pageNum, "", nil, "", "", platformclientv2.Integration{
			IntendedState: &intendedState,
		})

		if patchErr != nil {
			return diag.Errorf("Failed to update integration %s: %v", name, patchErr)
		}
	}

	log.Printf("Created integration %s %s", name, *integration.Id)
	return readIntegration(ctx, d, meta)
}

func readIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Reading integration %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		const pageSize = 100
		const pageNum = 1
		currentIntegration, resp, getErr := integrationAPI.GetIntegration(d.Id(), pageSize, pageNum, "", nil, "", "")
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read integration %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read integration %s: %s", d.Id(), getErr))
		}

		d.Set("integration_type", *currentIntegration.IntegrationType.Id)
		if currentIntegration.IntendedState != nil {
			d.Set("intended_state", *currentIntegration.IntendedState)
		} else {
			d.Set("intended_state", nil)
		}

		// Use returned ID to get current config, which contains complete configuration
		integrationConfig, _, err := integrationAPI.GetIntegrationConfigCurrent(*currentIntegration.Id)

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to read config of integration %s: %s", d.Id(), getErr))
		}

		d.Set("config", flattenIntegrationConfig(integrationConfig))

		log.Printf("Read integration %s %s", d.Id(), *currentIntegration.Name)

		return nil
	})
}

func updateIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	intendedState := d.Get("intended_state").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	diagErr, name := updateIntegrationConfig(d, integrationAPI)
	if diagErr != nil {
		return diagErr
	}

	if d.HasChange("intended_state") {

		log.Printf("Updating integration %s", name)
		const pageSize = 25
		const pageNum = 1
		_, _, patchErr := integrationAPI.PatchIntegration(d.Id(), pageSize, pageNum, "", nil, "", "", platformclientv2.Integration{
			IntendedState: &intendedState,
		})
		if patchErr != nil {
			return diag.Errorf("Failed to update integration %s: %s", name, patchErr)
		}
	}

	log.Printf("Updated integration %s %s", name, d.Id())
	return readIntegration(ctx, d, meta)
}

func deleteIntegration(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	_, _, err := integrationAPI.DeleteIntegration(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete the integration %s: %s", d.Id(), err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		const pageSize = 100
		const pageNum = 1
		_, resp, err := integrationAPI.GetIntegration(d.Id(), pageSize, pageNum, "", nil, "", "")
		if err != nil {
			if IsStatus404(resp) {
				// Integration deleted
				log.Printf("Deleted Integration %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting integration %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Integration %s still exists", d.Id()))
	})
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
