package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func dataSourceFlowMilestone() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Flow Milestone. Select a milestone by name.",
		ReadContext: ReadWithPooledClient(dataSourceFlowMilestoneRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Milestone name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceFlowMilestoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageSize = 100
		for pageNum := 1; ; pageNum++ {
			milestone, _, getErr := archAPI.GetFlowsMilestones(pageNum, pageSize, "", "", nil, name, "", "", nil)

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting milestone %s: %s", name, getErr))
			}

			if milestone.Entities == nil || len(*milestone.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No milestone found with name %s", name))
			}

			d.SetId(*(*milestone.Entities)[0].Id)
			return nil
		}
	})
}
