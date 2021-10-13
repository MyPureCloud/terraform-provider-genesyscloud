package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v55/platformclientv2"
	"time"
)

func dataSourceDidPool() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud DID pools. Select a DID pool by name",
		ReadContext: readWithPooledClient(dataSourceDidPoolRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "DID pool name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceDidPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	telephonyAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	didPoolName := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			didPools, _, getErr := telephonyAPI.GetTelephonyProvidersEdgesDidpools(100, pageNum, "", nil)

			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting DID pools %s: %s", didPoolName, getErr))
			}

			if didPools.Entities == nil || len(*didPools.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No DID pools found with id %s", didPoolName))
			}

			for _, didPool := range *didPools.Entities {
				if didPool.Name != nil && *didPool.Name == didPoolName &&
					didPool.State != nil && *didPool.State != "deleted" {
					d.SetId(*didPool.Id)
					return nil
				}
			}
		}
	})
}