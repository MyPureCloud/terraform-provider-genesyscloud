package outbound

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func dataSourceOutboundMessagingcampaign() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Outbound Messaging Campaign. Select a Outbound Messaging Campaign by name.`,

		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundMessagingcampaignRead),
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
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			sdkMessagingcampaignEntityListing, resp, getErr := outboundApi.GetOutboundMessagingcampaigns(pageSize, pageNum, "", "", "", "", []string{}, "", "", []string{})
			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting Outbound Messaging Campaign %s | error: %s", name, getErr), resp))
			}

			if sdkMessagingcampaignEntityListing.Entities == nil || len(*sdkMessagingcampaignEntityListing.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no Outbound Messaging Campaign found with name %s", name), resp))
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
