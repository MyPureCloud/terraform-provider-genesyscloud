package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func DataSourceAuthRole() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Roles. Select a role by name.",
		ReadContext: ReadWithPooledClient(DataSourceAuthRoleRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Role name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func DataSourceAuthRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query role by name. Retry in case search has not yet indexed the role.
	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageSize = 100
		const pageNum = 1
		roles, _, getErr := authAPI.GetAuthorizationRoles(pageSize, pageNum, "", nil, "", "", name, nil, nil, false, nil)
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting role %s: %s", name, getErr))
		}

		if roles.Entities == nil || len(*roles.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("No authorization roles found with name %s", name))
		}

		role := (*roles.Entities)[0]
		d.SetId(*role.Id)
		return nil
	})
}
