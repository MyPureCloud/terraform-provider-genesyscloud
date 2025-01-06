package auth_role

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

/*
   The data_source_genesyscloud_auth_role.go contains the data source implementation
   for the resource.
*/

// DataSourceAuthRoleRead retrieves by name the id in question
func DataSourceAuthRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		sdkConfig = m.(*provider.ProviderMeta).ClientConfig
		proxy     = getAuthRoleProxy(sdkConfig)

		name = d.Get("name").(string)

		response *platformclientv2.APIResponse
		id       string
	)

	diagErr := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		roleId, retryable, resp, err := proxy.getAuthRoleIdByName(ctx, name)
		if err != nil {
			response = resp
			if retryable {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		id = roleId
		return nil
	})

	if diagErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("%v", diagErr), response)
	}

	d.SetId(id)
	return nil
}
