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

func dataSourceExternalContactsContact() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud external contacts. Select a contact by any string search.",
		ReadContext: readWithPooledClient(dataSourceExternalContactsContactRead),
		Schema: map[string]*schema.Schema{
			"search": {
				Description: "The search string for the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func dataSourceExternalContactsContactRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewExternalContactsApiWithConfig(sdkConfig)

	search := d.Get("search").(string)

	// Query architect datatable by name. Retry in case search has not yet indexed the architect datatable.
	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		const pageNum = 1
		const pageSize = 100
		contacts, _, getErr := archAPI.GetExternalcontactsContacts(pageSize, pageNum, search, "", nil)
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("Error searching exteral contact %s: %s", search, getErr))
		}

		if contacts.Entities == nil || len(*contacts.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("No external contact found with search %s", search))
		}

		contact := (*contacts.Entities)[0]
		d.SetId(*contact.Id)
		return nil
	})
}
