package external_contacts

import (
	"context"
	"fmt"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceExternalContactsContactRead retrieves by search term the id in question
func dataSourceExternalContactsContactRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	ep := newExternalContactsContactsProxy(sdkConfig)

	search := d.Get("search").(string)
	return gcloud.WithRetries(ctx, 15*time.Second, func() *resource.RetryError {

		contactId, retryable, err := ep.GetExternalContactIdBySearch(ctx, search)

		if err != nil && !retryable {
			return resource.NonRetryableError(fmt.Errorf("Error searching exteral contact %s: %s", search, err))
		}

		if retryable {
			return resource.RetryableError(fmt.Errorf("No external contact found with search %s", search))
		}

		d.SetId(contactId)
		return nil
	})
}
