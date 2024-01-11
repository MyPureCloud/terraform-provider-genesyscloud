package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

func dataSourceRoutingUtilizationLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Utilization Labels. Select a label by name.",
		ReadContext: ReadWithPooledClient(dataSourceRoutingUtilizationLabelRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "Label name.",
				Type:         schema.TypeString,
				ValidateFunc: validation.StringDoesNotContainAny("*"),
				Required:     true,
			},
		},
	}
}

func dataSourceRoutingUtilizationLabelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		labels, _, getErr := routingAPI.GetRoutingUtilizationLabels(1, 1, "", name)
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting label %s: %s", name, getErr))
		}

		if labels.Entities == nil || len(*labels.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("No labels found with name %s", name))
		}

		label := (*labels.Entities)[0]
		d.SetId(*label.Id)
		return nil
	})
}
