package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v91/platformclientv2"
)

func dataSourceOutboundDncList() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound DNC Lists. Select a DNC list by name.",
		ReadContext: readWithPooledClient(dataSourceOutboundDncListRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "DNC List name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceOutboundDncListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		const pageNum = 1
		const pageSize = 100
		dncLists, _, getErr := outboundAPI.GetOutboundDnclists(false, false, pageSize, pageNum, true, "", name, "", []string{}, "", "")
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("error requesting dnc lists %s: %s", name, getErr))
		}
		if dncLists.Entities == nil || len(*dncLists.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("no dnc lists found with name %s", name))
		}
		dncList := (*dncLists.Entities)[0]
		d.SetId(*dncList.Id)
		return nil
	})
}
