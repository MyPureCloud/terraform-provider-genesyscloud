package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
	"time"
)

func dataSourceSite() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Sites. Select a site by name",
		ReadContext: readWithPooledClient(dataSourceSiteRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Site name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceSiteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 50
			sites, _, getErr := edgesAPI.GetTelephonyProvidersEdgesSites(pageSize, pageNum, "", "", name, "", false)
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting site %s: %s", name, getErr))
			}

			if sites.Entities == nil || len(*sites.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No sites found with name %s", name))
			}

			for _, site := range *sites.Entities {
				if site.Name != nil && *site.Name == name &&
					site.State != nil && *site.State != "deleted" {
					d.SetId(*site.Id)
					return nil
				}
			}
		}
	})
}
