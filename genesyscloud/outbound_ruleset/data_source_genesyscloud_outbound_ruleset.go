package outbound_ruleset

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v102/platformclientv2"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

func dataSourceOutboundRuleset() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Outbound Ruleset. Select an Outbound Ruleset by name.`,

		ReadContext: gcloud.ReadWithPooledClient(dataSourceOutboundRulesetRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Outbound Ruleset name.`,
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func dataSourceOutboundRulesetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			sdkrulesetentitylisting, _, getErr := getOutboundRulesets(pageSize, pageNum, false, outboundApi)
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting Outbound Ruleset %s: %s", name, getErr))
			}

			if sdkrulesetentitylisting.Entities == nil || len(*sdkrulesetentitylisting.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No Outbound Ruleset found with name %s", name))
			}

			for _, entity := range *sdkrulesetentitylisting.Entities {
				if entity.Name != nil && *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
