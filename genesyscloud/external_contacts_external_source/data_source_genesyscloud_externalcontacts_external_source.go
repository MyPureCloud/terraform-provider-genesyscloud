package external_contacts_external_source

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
	dataSourceExternalSourceCache *rc.DataSourceCache
)

func dataSourceExternalContactsExternalSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newExternalContactsExternalSourceProxy(sdkConfig)

	name := d.Get("name").(string)

	if dataSourceExternalSourceCache == nil {
		log.Printf("Instantiating the %s data source cache object", ResourceType)
		dataSourceExternalSourceCache = rc.NewDataSourceCache(sdkConfig, hydrateExternalSourceCacheFn, getExternalSourceByNameFn)
	}
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		externalSourceId, retryable, response, err := proxy.getExternalContactsExternalSourceIdByName(ctx, name)

		if err != nil {
			if retryable {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No external sources found with the provided name %s", name), response))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching exteral source %s | error: %s", name, err), response))

		}

		d.SetId(externalSourceId)
		return nil
	})
}

func hydrateExternalSourceCacheFn(c *rc.DataSourceCache, ctx context.Context) error {
	proxy := getExternalContactsExternalSourceProxy(c.ClientConfig)

	log.Printf("Hydrating cache for data source %s", ResourceType)

	allExternalSource, resp, err := proxy.getAllExternalContactsExternalSources(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get external source. Error: %s | API Response: %s", err.Error(), resp.String())
	}

	if allExternalSource == nil || len(*allExternalSource) == 0 {
		log.Printf("no external source found. The cache will remain empty.")
		return nil
	}

	for _, externalSource := range *allExternalSource {
		c.Cache[*externalSource.Name] = *externalSource.Id
	}

	log.Printf("Cache hydration complete for data source %s", ResourceType)
	return nil
}

// returns the external source id (blank if not found) and diag
func getExternalSourceByNameFn(c *rc.DataSourceCache, name string, ctx context.Context) (string, diag.Diagnostics) {
	proxy := getExternalContactsExternalSourceProxy(c.ClientConfig)
	externalSourceId := ""

	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		externalSourceID, retryable, response, err := proxy.getExternalContactsExternalSourceIdByName(ctx, name)
		if err != nil {
			errMsg := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting external source %s | error %s", name, err), response)
			if !retryable {
				return retry.NonRetryableError(errMsg)
			}
			return retry.RetryableError(errMsg)
		}

		externalSourceId = externalSourceID
		return nil
	})

	return externalSourceId, diag
}
