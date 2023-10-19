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

func DataSourceArchitectScheduleGroups() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Schedule Groups. Select a schedule group by name.",
		ReadContext: ReadWithPooledClient(dataSourceScheduleGroupRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Schedule Group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceScheduleGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query schedule group by name. Retry in case search has not yet indexed the schedule group.
	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		scheduleGroups, _, getErr := archAPI.GetArchitectSchedulegroups(pageNum, pageSize, "", "", name, "", nil)
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting schedule group %s: %s", name, getErr))
		}

		if scheduleGroups.Entities == nil || len(*scheduleGroups.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("No schedule groups found with name %s", name))
		}

		scheduleGroup := (*scheduleGroups.Entities)[0]
		d.SetId(*scheduleGroup.Id)
		return nil
	})
}
