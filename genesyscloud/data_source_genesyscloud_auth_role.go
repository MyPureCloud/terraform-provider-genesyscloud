package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

func dataSourceAuthRole() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Roles. Select a role by name.",
		ReadContext: dataSourceAuthRoleRead,
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
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(GetSdkClient())

	name := d.Get("name").(string)

	// Query role by name. Retry in case search has not yet indexed the role.
	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		roles, _, getErr := authAPI.GetAuthorizationRoles(1, 1, "", nil, "", "", name, nil, nil, false, nil)
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
