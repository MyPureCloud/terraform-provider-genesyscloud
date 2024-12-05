package external_contacts_organization

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

var (
	dataSourceOrganizationCache *rc.DataSourceCache
)

func dataSourceExternalContactsOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newExternalContactsOrganizationProxy(sdkConfig)

	name := d.Get("name").(string)

	if dataSourceOrganizationCache == nil {
		log.Printf("Instantiating the %s data source cache object", ResourceType)
		dataSourceOrganizationCache = rc.NewDataSourceCache(sdkConfig, hydrateOrganizationCacheFn, getOrganizationByNameFn)
	}
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		externalOrganizationId, retryable, response, err := proxy.getExternalContactsOrganizationIdByName(ctx, name)

		if err != nil {
			if retryable {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No organizations found with the provided name %s", name), response))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching exteral organization %s | error: %s", name, err), response))

		}

		d.SetId(externalOrganizationId)
		return nil
	})
}

func hydrateOrganizationCacheFn(c *rc.DataSourceCache, ctx context.Context) error {
	proxy := getExternalContactsOrganizationProxy(c.ClientConfig)

	log.Printf("Hydrating cache for data source %s", ResourceType)

	allExternalOrganization, resp, err := proxy.getAllExternalContactsOrganization(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get external organization. Error: %s | API Response: %s", err.Error(), resp.String())
	}

	if allExternalOrganization == nil || len(*allExternalOrganization) == 0 {
		log.Printf("no external organization found. The cache will remain empty.")
		return nil
	}

	for _, organization := range *allExternalOrganization {
		c.Cache[*organization.Name] = *organization.Id
	}

	log.Printf("Cache hydration complete for data source %s", ResourceType)
	return nil
}

// returns the organization id (blank if not found) and diag
func getOrganizationByNameFn(c *rc.DataSourceCache, name string, ctx context.Context) (string, diag.Diagnostics) {
	proxy := getExternalContactsOrganizationProxy(c.ClientConfig)
	organizationId := ""

	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		organizationID, retryable, response, err := proxy.getExternalContactsOrganizationIdByName(ctx, name)
		if err != nil {
			errMsg := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting organization %s | error %s", name, err), response)
			if !retryable {
				return retry.NonRetryableError(errMsg)
			}
			return retry.RetryableError(errMsg)
		}

		organizationId = organizationID
		return nil
	})

	return organizationId, diag
}
