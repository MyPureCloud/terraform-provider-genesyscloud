package genesyscloud

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
)

func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Users. Select a user by email or name.",
		ReadContext: provider.ReadWithPooledClient(DataSourceUserRead),
		Schema: map[string]*schema.Schema{
			"email": {
				Description: "User email.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "User name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func DataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	exactSearchType := "EXACT"
	sortOrderAsc := "ASC"
	emailField := "email"
	nameField := "name"

	searchCriteria := platformclientv2.Usersearchcriteria{
		VarType: &exactSearchType,
	}
	if email, ok := d.GetOk("email"); ok {
		emailStr := email.(string)
		searchCriteria.Fields = &[]string{emailField}
		searchCriteria.Value = &emailStr
	} else if name, ok := d.GetOk("name"); ok {
		nameStr := name.(string)
		searchCriteria.Fields = &[]string{nameField}
		searchCriteria.Value = &nameStr
	} else {
		return util.BuildDiagnosticError("genesyscloud_user", fmt.Sprintf("No user search field specified"), fmt.Errorf("no user search field specified"))
	}

	// Retry in case user is not yet indexed
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		users, resp, getErr := usersAPI.PostUsersSearch(platformclientv2.Usersearchrequest{
			SortBy:    &emailField,
			SortOrder: &sortOrderAsc,
			Query:     &[]platformclientv2.Usersearchcriteria{searchCriteria},
		})
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_user", fmt.Sprintf("Error requesting users: %s", getErr), resp))
		}

		if users.Results == nil || len(*users.Results) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_user", fmt.Sprintf("No users found with search criteria %v", searchCriteria), resp))
		}

		// Select first user in the list
		user := (*users.Results)[0]
		d.SetId(*user.Id)
		return nil
	})
}
