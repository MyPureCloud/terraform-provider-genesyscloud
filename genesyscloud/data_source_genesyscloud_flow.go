package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/ronanwatkins/terraform-plugin-sdk/v2/diag"
	"github.com/ronanwatkins/terraform-plugin-sdk/v2/helper/resource"
	"github.com/ronanwatkins/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func dataSourceFlow() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Flows. Select a flow by name.",
		ReadContext: readWithPooledClient(dataSourceFlowRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Flow name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceFlowRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query flow by name. Retry in case search has not yet indexed the flow.
	return withRetries(ctx, 5*time.Second, func() *resource.RetryError {
		const pageNum = 1
		const pageSize = 10
		flows, _, getErr := archAPI.GetFlows(nil, pageNum, pageSize, "", "", nil, name, "", "", "", "", "", "", "", false, false, "", "", nil)
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("Error requesting flow %s: %s", name, getErr))
		}

		if flows.Entities == nil || len(*flows.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("No flows found with name %s", name))
		}

		flow := (*flows.Entities)[0]
		d.SetId(*flow.Id)
		return nil
	})
}
