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

func dataSourceEdgeGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Edge Group. Select an edge group by name",
		ReadContext: ReadWithPooledClient(dataSourceEdgeGroupRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Edge Group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"managed": {
				Description: "Return entities that are managed by Genesys Cloud.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func dataSourceEdgeGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)
	managed := d.Get("managed").(bool)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			edgeGroup, _, getErr := edgesAPI.GetTelephonyProvidersEdgesEdgegroups(pageSize, pageNum, name, "", managed)

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting edge group %s: %s", name, getErr))
			}

			if edgeGroup.Entities == nil || len(*edgeGroup.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No edge group found with name %s", name))
			}

			d.SetId(*(*edgeGroup.Entities)[0].Id)
			return nil
		}
	})
}
