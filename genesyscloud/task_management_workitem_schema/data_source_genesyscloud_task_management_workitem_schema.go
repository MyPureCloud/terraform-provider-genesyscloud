package task_management_workitem_schema

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
   The data_source_genesyscloud_task_management_workitem_schema.go contains the data source implementation
   for the resource.
*/

// dataSourceTaskManagementWorkitemSchemaRead retrieves by name the id in question
func dataSourceTaskManagementWorkitemSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementProxy(sdkConfig)

	name := d.Get("name").(string)

	// Query for workitem schemas by name. Retry in case new schema is not yet indexed by search.
	// As schema names are non-unique, fail in case of multiple results.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		schemas, retryable, resp, err := proxy.getTaskManagementWorkitemSchemasByName(ctx, name)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error getting workitem schema %s | error: %v", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no workitem schema found with name %s", name), resp))
		}

		if len(*schemas) > 1 {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("ambiguous workitem schema name: %s", name), resp))
		}

		schema := (*schemas)[0]
		d.SetId(*schema.Id)
		return nil
	})
}
