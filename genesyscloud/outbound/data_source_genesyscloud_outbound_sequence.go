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

func dataSourceOutboundSequence() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Outbound Sequence. Select a Outbound Sequence by name.`,

		ReadContext: gcloud.ReadWithPooledClient(dataSourceOutboundSequenceRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Outbound Sequence name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceOutboundSequenceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			sdkcampaignsequenceentitylisting, _, getErr := outboundApi.GetOutboundSequences(pageSize, pageNum, true, "", "", "", "")
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting Outbound Sequence %s: %s", name, getErr))
			}

			if sdkcampaignsequenceentitylisting.Entities == nil || len(*sdkcampaignsequenceentitylisting.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No Outbound Sequence found with name %s", name))
			}

			for _, entity := range *sdkcampaignsequenceentitylisting.Entities {
				if entity.Name != nil && *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
