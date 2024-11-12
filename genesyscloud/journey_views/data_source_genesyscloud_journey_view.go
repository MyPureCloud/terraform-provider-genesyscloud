package journey_view

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	dataSourceJourneyViewCache *rc.DataSourceCache
)

func dataSourceJourneyViewRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig

	key := d.Get("name").(string)

	if dataSourceJourneyViewCache == nil {
		log.Printf("Instantiating the %s data source cache object", resourceName)
		dataSourceJourneyViewCache = rc.NewDataSourceCache(sdkConfig, hydrateJourneyViewCacheFn, getJourneyByNameFn)
	}

	journeyId, err := rc.RetrieveId(dataSourceJourneyViewCache, resourceName, key, ctx)
	if err != nil {
		return err
	}

	d.SetId(journeyId)
	return nil
}

// hydrateJourneyViewCacheFn for hydrating the cache with Genesys Cloud journey views using the SDK
func hydrateJourneyViewCacheFn(c *rc.DataSourceCache, ctx context.Context) error {
	proxy := GetJourneyViewProxy(c.ClientConfig)

	log.Printf("Hydrating cache for data source %s", resourceName)

	allJourneys, resp, err := proxy.GetAllJourneyViews(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get journey views. Error: %s | API Response: %s", err.Error(), resp.String())
	}

	if allJourneys == nil || len(*allJourneys) == 0 {
		log.Printf("No journey views found. The cache will remain empty.")
		return nil
	}

	for _, journey := range *allJourneysQueues {
		c.Cache[*journey.Name] = *journey.Id
	}

	log.Printf("Cache hydration complete for data source %s", resourceName)
	return nil
}

// getJourneyByNameFn returns the journey id (blank if not found) and diag
func getJourneyByNameFn(c *rc.DataSourceCache, name string, ctx context.Context) (string, diag.Diagnostics) {
	proxy := GetJourneyViewProxy(c.ClientConfig)
	journeyId := ""

	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		journeyID, resp, retryable, getErr := proxy.getJourneyViewByName(ctx, name)
		if getErr != nil {
			errMsg := util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting journey view %s | error %s", name, getErr), resp)
			if !retryable {
				return retry.NonRetryableError(errMsg)
			}
			return retry.RetryableError(errMsg)
		}

		journeyId = journeyID
		return nil
	})

	return journeyId, diag
}
