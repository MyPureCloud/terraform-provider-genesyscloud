package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
)

func dataSourceAuthRole() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Roles. Select a role by name.",
		ReadContext: readWithPooledClient(dataSourceAuthRoleRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Role name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceAuthRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query role by name. Retry in case search has not yet indexed the role.
	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		const pageSize = 100
		const pageNum = 1
		roles, _, getErr := authAPI.GetAuthorizationRoles(pageSize, pageNum, "", nil, "", "", name, nil, nil, false, nil)
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("Error requesting role %s: %s", name, getErr))
		}

		if roles.Entities == nil || len(*roles.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("No authorization roles found with name %s", name))
		}

		role := (*roles.Entities)[0]
		d.SetId(*role.Id)
		return nil
	})
}
