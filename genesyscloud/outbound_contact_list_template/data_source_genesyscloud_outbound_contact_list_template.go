package outbound_contact_list_template

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

func DataSourceOutboundContactListTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound Contact Lists Templates. Select a contact list template by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundContactListTemplateRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Contact List Template name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceOutboundContactListTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		contactListTemplates, resp, getErr := outboundAPI.GetOutboundContactlisttemplates(pageSize, pageNum, true, "", name, "", "")
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting contact list template %s | error: %s", name, getErr), resp))
		}
		if contactListTemplates.Entities == nil || len(*contactListTemplates.Entities) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no contact list templates found with name %s", name), resp))
		}
		contactListTemplate := (*contactListTemplates.Entities)[0]
		d.SetId(*contactListTemplate.Id)
		return nil
	})
}
