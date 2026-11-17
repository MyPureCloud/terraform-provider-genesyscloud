package authorization_roles_settings

import (
	"context"
	"fmt"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
The data_source_genesyscloud_authorization_roles_settings.go contains the data source implementation for the resource.
*/

func dataSourceAuthorizationRolesSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthorizationRolesSettingsProxy(sdkConfig)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		settings, resp, err := proxy.getAuthorizationRolesSettings(ctx)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting authorization role settings | error: %s", err), resp))
		}

		d.SetId(authorizationRolesSettingsId)
		setAuthorizationRolesSettingsToResourceData(d, settings)
		return nil
	})
}
