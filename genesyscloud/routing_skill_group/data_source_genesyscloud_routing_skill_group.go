package routing_skill_group

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRoutingSkillGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillGroupsProxy(sdkConfig)
	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		skillGroupId, resp, retryable, err := proxy.getRoutingSkillGroupsIdByName(ctx, name)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error requesting routing skill group %s | error: %s", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No routing skill group found with name %s", name), resp))
		}
		d.SetId(skillGroupId)
		return nil
	})
}
