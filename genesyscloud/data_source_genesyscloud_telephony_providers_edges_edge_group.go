package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v89/platformclientv2"
)

func dataSourceEdgeGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Edge Group. Select an edge group by name",
		ReadContext: readWithPooledClient(dataSourceEdgeGroupRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Edge Group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceEdgeGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			edgeGroup, _, getErr := edgesAPI.GetTelephonyProvidersEdgesEdgegroups(pageSize, pageNum, name, "", false)

			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting edge group %s: %s", name, getErr))
			}

			if edgeGroup.Entities == nil || len(*edgeGroup.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No edge group found with name %s", name))
			}

			d.SetId(*(*edgeGroup.Entities)[0].Id)
			return nil
		}
	})
}
