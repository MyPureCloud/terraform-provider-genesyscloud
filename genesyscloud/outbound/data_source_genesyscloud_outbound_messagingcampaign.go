package outbound

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func dataSourceOutboundMessagingcampaign() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Outbound Messaging Campaign. Select a Outbound Messaging Campaign by name.`,

		ReadContext: gcloud.ReadWithPooledClient(dataSourceOutboundMessagingcampaignRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Outbound Messaging Campaign name.`,
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func dataSourceOutboundMessagingcampaignRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			sdkMessagingcampaignEntityListing, _, getErr := outboundApi.GetOutboundMessagingcampaigns(pageSize, pageNum, "", "", "", "", []string{}, "", "", []string{})
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("error requesting Outbound Messaging Campaign %s: %s", name, getErr))
			}

			if sdkMessagingcampaignEntityListing.Entities == nil || len(*sdkMessagingcampaignEntityListing.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no Outbound Messaging Campaign found with name %s", name))
			}

			for _, entity := range *sdkMessagingcampaignEntityListing.Entities {
				if entity.Name != nil && *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
