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

func dataSourceFlowOutcome() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Flow Outcome. Select an outcome by name.",
		ReadContext: ReadWithPooledClient(dataSourceFlowOutcomeRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Outcome name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceFlowOutcomeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageSize = 100
		for pageNum := 1; ; pageNum++ {
			outcomes, _, getErr := archAPI.GetFlowsOutcomes(pageNum, pageSize, "", "", nil, name, "", "", nil)

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting outcomes %s: %s", name, getErr))
			}

			if outcomes.Entities == nil || len(*outcomes.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No outcomes found with name %s", name))
			}

			d.SetId(*(*outcomes.Entities)[0].Id)
			return nil
		}
	})
}
