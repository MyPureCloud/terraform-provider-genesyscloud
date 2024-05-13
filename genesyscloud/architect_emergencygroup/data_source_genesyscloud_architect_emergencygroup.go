package architect_emergencygroup

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceEmergencyGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	ap := getArchitectEmergencyGroupProxy(sdkConfig)

	name := d.Get("name").(string)

	// Query emergency group by name. Retry in case search has not yet indexed the emergency group.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		emergencyGroups, resp, getErr := ap.getArchitectEmergencyGroupIdByName(ctx, name)
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error requesting emergency group %s | error: %s", name, getErr), resp))
		}

		if emergencyGroups.Entities == nil || len(*emergencyGroups.Entities) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No emergency groups found with name %s", name), resp))
		}

		emergencyGroup := (*emergencyGroups.Entities)[0]
		d.SetId(*emergencyGroup.Id)
		return nil
	})
}
