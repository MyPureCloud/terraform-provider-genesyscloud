package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v103/platformclientv2"
)

func dataSourceOutboundContactListFilter() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound Contact List Filters. Select a contact list filter by name.",
		ReadContext: ReadWithPooledClient(dataSourceOutboundContactListFilterRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Contact List Filter name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceOutboundContactListFilterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *resource.RetryError {
		const pageNum = 1
		const pageSize = 100
		contactListFilters, _, getErr := outboundAPI.GetOutboundContactlistfilters(pageSize, pageNum, true, "", name, "", "", "")
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("error requesting contact list filter %s: %s", name, getErr))
		}
		if contactListFilters.Entities == nil || len(*contactListFilters.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("no contact list filters found with name %s", name))
		}
		contactListFilter := (*contactListFilters.Entities)[0]
		d.SetId(*contactListFilter.Id)
		return nil
	})
}
