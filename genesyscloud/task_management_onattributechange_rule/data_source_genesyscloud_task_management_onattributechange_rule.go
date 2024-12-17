package task_management_onattributechange_rule

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
   The data_source_genesyscloud_task_management_onattributechange_rule.go contains the data source implementation
   for the resource.
*/

// dataSourceTaskManagementOnAttributeChangeRuleRead retrieves by name the id in question
func dataSourceTaskManagementOnAttributeChangeRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetTaskManagementOnAttributeChangeRuleProxy(sdkConfig)

	name := d.Get("name").(string)
	typeId := d.Get("worktype_id").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		onAttributeChangeRuleId, retryable, resp, err := proxy.getTaskManagementOnAttributeChangeRuleIdByName(ctx, typeId, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error searching task management onattributechange rule %s | error: %s", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no task management onattributechange rule found with name %s", name), resp))
		}

		d.SetId(typeId + "/" + onAttributeChangeRuleId)
		return nil
	})
}
