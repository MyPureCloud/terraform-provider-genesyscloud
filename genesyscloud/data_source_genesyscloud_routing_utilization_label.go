package genesyscloud

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func dataSourceRoutingUtilizationLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Utilization Labels. Select a label by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceRoutingUtilizationLabelRead),
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
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		labels, resp, getErr := routingAPI.GetRoutingUtilizationLabels(1, 1, "", name)
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_utilization_label", fmt.Sprintf("Error requesting label %s | error: %s", name, getErr), resp))
		}

		if labels.Entities == nil || len(*labels.Entities) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_utilization_label", fmt.Sprintf("No labels found with name %s", name), resp))
		}

		label := (*labels.Entities)[0]
		d.SetId(*label.Id)
		return nil
	})
}
