package external_contacts

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The data_source_genesyscloud_externalcontacts_contact.go contains the data source implementation
   for the resource.

   Note:  This code should contain no code for doing the actual lookup in Genesys Cloud.  Instead,
   it should be added to the _proxy.go file for the class using our proxy pattern.
*/

// dataSourceExternalContactsContactRead retrieves by search term the id in question
func dataSourceExternalContactsContactRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	ep := newExternalContactsContactsProxy(sdkConfig)

	search := d.Get("search").(string)
	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {

		contactId, retryable, err := ep.getExternalContactIdBySearch(ctx, search)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching exteral contact %s: %s", search, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No external contact found with search %s", search))
		}

		d.SetId(contactId)
		return nil
	})
}
