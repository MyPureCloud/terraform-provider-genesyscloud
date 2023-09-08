package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func DataSourceSite() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Sites. Select a site by name",
		ReadContext: ReadWithPooledClient(dataSourceSiteRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Site name.",
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

func dataSourceSiteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)
	managed := d.Get("managed").(bool)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 50
			sites, _, getErr := edgesAPI.GetTelephonyProvidersEdgesSites(pageSize, pageNum, "", "", name, "", managed)
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting site %s: %s", name, getErr))
			}

			if sites.Entities == nil || len(*sites.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No sites found with name %s", name))
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
