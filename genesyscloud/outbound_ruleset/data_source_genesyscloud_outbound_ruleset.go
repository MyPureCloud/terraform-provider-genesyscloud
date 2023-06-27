package outbound_ruleset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

func DataSourceOutboundRuleset() *schema.Resource {
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
	outboundAPIProxy.ConfigureProxyApiInstance(sdkConfig)
	name := d.Get("name").(string)
	return outboundAPIProxy.ReadOutboundRulesetsData(ctx, outboundAPIProxy,d, name)	
}
