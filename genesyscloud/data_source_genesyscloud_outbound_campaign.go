package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v102/platformclientv2"
)

func dataSourceOutboundCampaign() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Outbound Campaign. Select a Outbound Campaign by name.`,

		ReadContext: readWithPooledClient(dataSourceOutboundCampaignRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Outbound Campaign name.`,
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func dataSourceOutboundCampaignRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			sdkcampaignentitylisting, _, getErr := outboundApi.GetOutboundCampaigns(pageSize, pageNum, "", "", []string{}, "", "", "", "", "", []string{}, "", "")
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("error requesting Outbound Campaign %s: %s", name, getErr))
			}

			if sdkcampaignentitylisting.Entities == nil || len(*sdkcampaignentitylisting.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("no Outbound Campaign found with name %s", name))
			}

			for _, entity := range *sdkcampaignentitylisting.Entities {
				if entity.Name != nil && *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
