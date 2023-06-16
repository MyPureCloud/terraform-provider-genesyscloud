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

func dataSourceOutboundCallAnalysisResponseSet() *schema.Resource {
	return &schema.Resource{
		Description: "",
		ReadContext: readWithPooledClient(dataSourceOutboundCallAnalysisReponseSetRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Data source for Genesys Cloud Outbound Call Analysis Response Sets. Select a response set by name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceOutboundCallAnalysisReponseSetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		const pageNum = 1
		const pageSize = 100
		responseSets, _, getErr := outboundAPI.GetOutboundCallanalysisresponsesets(pageSize, pageNum, true, "", name, "", "")
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("error requesting call analysis response set %s: %s", name, getErr))
		}
		if responseSets.Entities == nil || len(*responseSets.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("no call analysis response sets found with name %s", name))
		}
		responseSet := (*responseSets.Entities)[0]
		d.SetId(*responseSet.Id)
		return nil
	})
}
