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
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
   The data_source_genesyscloud_auth_role.go contains the data source implementation
   for the resource.
*/

// dataSourceAuthRoleRead retrieves by name the id in question
func DataSourceAuthRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query role by name. Retry in case search has not yet indexed the role.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageSize = 100
		const pageNum = 1
		roles, proxyResponse, getErr := authAPI.GetAuthorizationRoles(pageSize, pageNum, "", nil, "", "", name, nil, nil, false, nil)
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error requesting role %s | error: %s", name, getErr), proxyResponse))
		}

		if roles.Entities == nil || len(*roles.Entities) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No authorization roles found with name %s", name), proxyResponse))
		}
		role := (*roles.Entities)[0]
		d.SetId(*role.Id)
		return nil
	})
}
