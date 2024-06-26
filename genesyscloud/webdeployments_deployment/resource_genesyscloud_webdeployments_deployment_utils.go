package webdeployments_deployment

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
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
