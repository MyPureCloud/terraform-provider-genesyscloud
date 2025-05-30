package architect_schedulegroups

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
   The data_source_genesyscloud_architect_schedulegroups.go contains the data source implementation
   for the resource.
*/

// dataSourceArchitectSchedulegroupsRead retrieves by name the id in question
func dataSourceArchitectSchedulegroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newArchitectSchedulegroupsProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		scheduleGroupId, retryable, proxyResponse, err := proxy.getArchitectSchedulegroupsIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error searching architect schedulegroups %s | error: %s", name, err), proxyResponse))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no architect schedulegroups found with name %s", name), proxyResponse))
		}

		d.SetId(scheduleGroupId)
		return nil
	})
}
