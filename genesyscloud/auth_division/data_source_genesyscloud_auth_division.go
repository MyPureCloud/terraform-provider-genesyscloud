package auth_division

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	dataSourceAuthDivisionCache *rc.DataSourceCache
)

func dataSourceAuthDivisionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	name := d.Get("name").(string)
	key := normaliseAuthDivisionName(name)

	if dataSourceAuthDivisionCache == nil {
		dataSourceAuthDivisionCache = rc.NewDataSourceCache(sdkConfig, hydrateAuthDivisionCacheFn, getDivisionIdByNameFn)
	}

	divisionId, err := rc.RetrieveId(dataSourceAuthDivisionCache, ResourceType, key, ctx)
	if err != nil {
		return err
	}
	d.SetId(divisionId)
	return nil
}

func normaliseAuthDivisionName(name string) string {
	return strings.ToLower(name)
}

func hydrateAuthDivisionCacheFn(c *rc.DataSourceCache, ctx context.Context) error {
	proxy := getAuthDivisionProxy(c.ClientConfig)

	log.Printf("hydrating cache for data source %s", ResourceType)

	allDivisions, resp, err := proxy.getAllAuthDivision(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to collect all auth divisions. Error: %s | Response: %s", err.Error(), resp.String())
	}

	if allDivisions == nil || len(*allDivisions) == 0 {
		return nil
	}

	for _, div := range *allDivisions {
		c.Cache[normaliseAuthDivisionName(*div.Name)] = *div.Id
	}

	log.Printf("cache hydration complete for data source %s", ResourceType)
	return nil
}

func getDivisionIdByNameFn(c *rc.DataSourceCache, name string, ctx context.Context) (string, diag.Diagnostics) {
	var (
		id    string
		proxy = getAuthDivisionProxy(c.ClientConfig)
	)

	diagErr := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		divisionId, resp, retryable, err := proxy.getAuthDivisionIdByName(ctx, name)
		if err != nil {
			errorDetails := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf(`Could not find division "%s" | error: %s`, name, err.Error()), resp)
			if !retryable {
				return retry.NonRetryableError(errorDetails)
			}
			return retry.RetryableError(errorDetails)
		}
		id = divisionId
		return nil
	})

	return id, diagErr
}
