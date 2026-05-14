package webdeployments_deployment

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

func alwaysDifferent(k, old, new string, d *schema.ResourceData) bool {
	return false
}

func validateDeploymentStatusChange(k, old, new string, d *schema.ResourceData) bool {
	// Deployments will begin in a pending status and may or may not make it to active (or error) by the time we retrieve their state,
	// so allow the status to change from pending to a less ephemeral status
	return old == "Pending"
}

func flattenConfiguration(configuration *platformclientv2.Webdeploymentconfigurationversionentityref) []interface{} {
	return []interface{}{map[string]interface{}{
		"id":      *configuration.Id,
		"version": *configuration.Version,
	}}
}

func flattenPushIntegrations(pushIntegrations *[]platformclientv2.Pushintegration) []interface{} {
	if pushIntegrations == nil || len(*pushIntegrations) == 0 {
		return nil
	}

	pushIntegrationList := make([]interface{}, 0)
	for _, integration := range *pushIntegrations {
		pushIntegrationMap := make(map[string]interface{})
		if integration.Id != nil {
			pushIntegrationMap["id"] = *integration.Id
		}
		if integration.Provider != nil {
			pushIntegrationMap["provider"] = *integration.Provider
		}
		pushIntegrationList = append(pushIntegrationList, pushIntegrationMap)
	}
	return pushIntegrationList
}

func buildPushIntegrations(d *schema.ResourceData) *[]platformclientv2.Pushintegration {
	if pushIntegrationsConfig, ok := d.GetOk("push_integrations"); ok {
		pushIntegrationsList := pushIntegrationsConfig.([]interface{})
		if len(pushIntegrationsList) == 0 {
			return nil
		}

		pushIntegrations := make([]platformclientv2.Pushintegration, 0)
		for _, integration := range pushIntegrationsList {
			integrationMap := integration.(map[string]interface{})
			pushIntegration := platformclientv2.Pushintegration{}

			if id, ok := integrationMap["id"].(string); ok && id != "" {
				pushIntegration.Id = &id
			}
			if provider, ok := integrationMap["provider"].(string); ok && provider != "" {
				pushIntegration.Provider = &provider
			}

			pushIntegrations = append(pushIntegrations, pushIntegration)
		}
		return &pushIntegrations
	}
	return nil
}

func buildSupportedContentReference(d *schema.ResourceData) *platformclientv2.Supportedcontentreference {
	if supportedContentId, ok := d.GetOk("supported_content_id"); ok {
		id := supportedContentId.(string)
		return &platformclientv2.Supportedcontentreference{
			Id: &id,
		}
	}
	return nil
}

func validAllowedDomainsSettings(d *schema.ResourceData) error {
	allowAllDomains := d.Get("allow_all_domains").(bool)
	_, allowedDomainsSet := d.GetOk("allowed_domains")

	if allowAllDomains && allowedDomainsSet {
		return errors.New("Allowed domains cannot be specified when all domains are allowed")
	}

	if !allowAllDomains && !allowedDomainsSet {
		return errors.New("Either allowed domains must be specified or all domains must be allowed")
	}

	return nil
}
