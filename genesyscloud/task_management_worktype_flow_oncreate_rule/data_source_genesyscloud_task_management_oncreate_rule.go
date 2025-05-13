package task_management_worktype_flow_oncreate_rule

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
   The data_source_genesyscloud_task_management_oncreate_rule.go contains the data source implementation
   for the resource.
*/

// dataSourceTaskManagementOnCreateRuleRead retrieves by name the id in question
func dataSourceTaskManagementOnCreateRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementOnCreateRuleProxy(sdkConfig)

	name := d.Get("name").(string)
	typeId := d.Get("worktype_id").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		onCreateRuleId, retryable, resp, err := proxy.getTaskManagementOnCreateRuleIdByName(ctx, typeId, name)

		if err != nil {
			if retryable {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no task management oncreate rule found with name %s", name), resp))
			} else {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error searching task management oncreate rule %s | error: %s", name, err), resp))
			}
		}

		d.SetId(onCreateRuleId)
		return nil
	})
}
