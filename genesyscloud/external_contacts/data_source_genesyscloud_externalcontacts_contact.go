package external_contacts

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

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
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	ep := newExternalContactsContactsProxy(sdkConfig)

	search := d.Get("search").(string)
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {

		contactId, retryable, resp, err := ep.getExternalContactIdBySearch(ctx, search)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error searching exteral contact %s | error: %s", search, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No external contact found with search %s", search), resp))
		}

		d.SetId(contactId)
		return nil
	})
}
