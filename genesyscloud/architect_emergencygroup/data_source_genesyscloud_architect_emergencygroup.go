package architect_emergencygroup

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceEmergencyGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*genesyscloud.ProviderMeta).ClientConfig
	ap := getArchitectEmergencyGroupProxy(sdkConfig)

	name := d.Get("name").(string)

	// Query emergency group by name. Retry in case search has not yet indexed the emergency group.
	return genesyscloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		emergencyGroups, _, getErr := ap.getArchitectEmergencyGroupIdByName(ctx, name)
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting emergency group %s: %s", name, getErr))
		}

		if emergencyGroups.Entities == nil || len(*emergencyGroups.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("No emergency groups found with name %s", name))
		}

		emergencyGroup := (*emergencyGroups.Entities)[0]
		d.SetId(*emergencyGroup.Id)
		return nil
	})
}
