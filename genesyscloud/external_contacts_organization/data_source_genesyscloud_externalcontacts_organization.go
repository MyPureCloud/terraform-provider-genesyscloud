package external_contacts_organization

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func dataSourceExternalContactsOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newExternalContactsOrganizationProxy(sdkConfig)

	name := d.Get("name").(string)

	if externalOrganization := rc.GetCacheItem(proxy.externalOrganizationCache, name); externalOrganization != nil {
		d.SetId(*externalOrganization.Id)
		return nil
	}

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		externalOrganizationId, retryable, response, err := proxy.getExternalContactsOrganizationIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceType, fmt.Sprintf("Error searching exteral organization %s | error: %s", name, err), response))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceType, fmt.Sprintf("No organizations found with the provided name %s", name), response))
		}

		d.SetId(externalOrganizationId)
		return nil
	})
}
