package outbound_contact_list

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"
)

func DataSourceOutboundContactList() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound Contact Lists. Select a contact list by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundContactListRead),
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
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		contactLists, _, getErr := outboundAPI.GetOutboundContactlists(false, false, pageSize, pageNum, true, "", name, []string{""}, []string{""}, "", "")
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("error requesting contact list %s: %s", name, getErr))
		}
		if contactLists.Entities == nil || len(*contactLists.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("no contact lists found with name %s", name))
		}
		contactList := (*contactLists.Entities)[0]
		d.SetId(*contactList.Id)
		return nil
	})
}
