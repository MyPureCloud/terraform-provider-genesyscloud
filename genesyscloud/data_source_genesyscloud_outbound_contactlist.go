package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
	"time"
)

func dataSourceOutboundContactList() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound Contact Lists. Select a contact list by name.",
		ReadContext: readWithPooledClient(dataSourceOutboundContactListRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Contact List name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceOutboundContactListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		const pageNum = 1
		const pageSize = 100
		contactLists, _, getErr := outboundAPI.GetOutboundContactlists(false, false, pageSize, pageNum, true, "", name, []string{""}, []string{""}, "", "")
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("error requesting contact list %s: %s", name, getErr))
		}
		if contactLists.Entities == nil || len(*contactLists.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("no contact lists found with name %s", name))
		}
		contactList := (*contactLists.Entities)[0]
		d.SetId(*contactList.Id)
		return nil
	})
}
