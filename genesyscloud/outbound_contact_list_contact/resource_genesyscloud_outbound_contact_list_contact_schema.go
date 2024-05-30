package outbound_contact_list_contact

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
)

func ResourceOutboundContactListContact() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Contact List Contact`,

		CreateContext: provider.CreateWithPooledClient(nil),
		ReadContext:   provider.ReadWithPooledClient(nil),
		UpdateContext: provider.UpdateWithPooledClient(nil),
		DeleteContext: provider.DeleteWithPooledClient(nil),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "",
				Optional:    true,
				Type:        schema.TypeString,
			},
			"contact_list_id": {
				Description: "",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
}
