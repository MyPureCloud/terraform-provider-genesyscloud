package external_contacts_external_source

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"log"
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

	name := d.Get("name").(string)

	if dataSourceExternalSourceCache == nil {
		log.Printf("Instantiating the %s data source cache object", ResourceType)
		dataSourceExternalSourceCache = rc.NewDataSourceCache(sdkConfig, hydrateExternalSourceCacheFn, getExternalSourceByNameFn)
	}

	externalSourceId, err := rc.RetrieveId(dataSourceExternalSourceCache, ResourceType, name, ctx)
	if err != nil {
		return err
	}

	d.SetId(externalSourceId)
	return nil
}

func hydrateExternalSourceCacheFn(c *rc.DataSourceCache, ctx context.Context) error {
	proxy := getExternalContactsExternalSourceProxy(c.ClientConfig)

	log.Printf("Hydrating cache for data source %s", ResourceType)

	allExternalSources, resp, err := proxy.getAllExternalContactsExternalSources(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get external sources. Error: %s | API Response: %s", err.Error(), resp.String())
	}

	if allExternalSources == nil || len(*allExternalSources) == 0 {
		log.Printf("no external sources found. The cache will remain empty.")
		return nil
	}

	for _, externalSource := range *allExternalSources {
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
