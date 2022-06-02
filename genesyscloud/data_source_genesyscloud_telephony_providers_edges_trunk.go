package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v72/platformclientv2"
)

func dataSourceTrunk() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Trunk. Select a trunk by name",
		ReadContext: readWithPooledClient(dataSourceTrunkRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Trunk name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceTrunkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			trunks, _, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunks(pageNum, pageSize, "", "", "", "", "")

			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting trunk %s: %s", name, getErr))
			}

			if trunks.Entities == nil || len(*trunks.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No trunk found with name %s", name))
			}

			for _, trunk := range *trunks.Entities {
				if trunk.Name != nil && *trunk.Name == name &&
					trunk.State != nil && *trunk.State != "deleted" {
					d.SetId(*trunk.Id)
					return nil
				}
			}
		}
	})
}
