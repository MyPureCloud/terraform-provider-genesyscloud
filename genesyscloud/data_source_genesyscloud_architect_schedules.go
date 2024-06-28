package genesyscloud

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func DataSourceSchedule() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Schedule. Select a schedule by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceScheduleRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Schedule name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceScheduleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageSize = 100
		for pageNum := 1; ; pageNum++ {
			schedule, resp, getErr := archAPI.GetArchitectSchedules(pageNum, pageSize, "", "", name, nil)

			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("Error requesting schedule %s | error: %s", name, getErr), resp))
			}

			if schedule.Entities == nil || len(*schedule.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_architect_schedules", fmt.Sprintf("No schedule found with name %s", name), resp))
			}

			d.SetId(*(*schedule.Entities)[0].Id)
			return nil
		}
	})
}
