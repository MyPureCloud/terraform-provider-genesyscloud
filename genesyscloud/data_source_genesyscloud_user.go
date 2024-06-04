package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"net/mail"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
)

func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		Description:        "Data source for Genesys Cloud Users. Select a user by email or name.",
		ReadWithoutTimeout: provider.ReadWithPooledClient(DataSourceUserRead),
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

var (
	dataSourceUserCache *rc.DataSourceCache
)

func DataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	key := ""

	if email, ok := d.GetOk("email"); ok {
		key = email.(string)

	} else if name, ok := d.GetOk("name"); ok {
		key = name.(string)

	} else {
		return util.BuildDiagnosticError("genesyscloud_user", "no user search field specified", nil)
	}
	log.Printf(fmt.Sprintf("\n\nkey for cache is %s", key))

	if dataSourceUserCache == nil {
		dataSourceUserCache = rc.NewDataSourceCache(sdkConfig, hydrateUserCacheFn, getUserByNameFn)
	}

	userId, err := rc.RetrieveId(dataSourceUserCache, "genesyscloud_user", key, ctx)
	if err != nil {
		return err
	}

	d.SetId(userId)
	return nil
}

func hydrateUserCacheFn(c *rc.DataSourceCache) error {
	log.Printf("hydrating cache for data source genesyscloud_user")

	usersAPI := platformclientv2.NewUsersApiWithConfig(c.ClientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		log.Printf("calling cache for data source genesyscloud_user")

		userEntityList, _, getErr := usersAPI.GetUsers(pageSize, pageNum, nil, nil, "", nil, "", "")
		log.Printf("calling cache for data source genesyscloud_user %v", pageNum)
		if getErr != nil {
			return fmt.Errorf("failed to get page of users: %v", getErr)
		}

		if userEntityList.Entities == nil || len(*userEntityList.Entities) == 0 {
			break
		}

		// Add ids to cache
		for _, user := range *userEntityList.Entities {
			c.Cache[*user.Name] = *user.Id
			c.Cache[*user.Email] = *user.Id

		}
	}

	log.Printf("cache hydration completed for data source genesyscloud_user")

	return nil
}

func getUserByNameFn(c *rc.DataSourceCache, searchField string, ctx context.Context) (string, diag.Diagnostics) {
	userId := ""
	usersAPI := platformclientv2.NewUsersApiWithConfig(c.ClientConfig)

	exactSearchType := "EXACT"
	sortOrderAsc := "ASC"
	emailField := "email"

	searchCriteria := platformclientv2.Usersearchcriteria{
		VarType: &exactSearchType,
	}
	searchFieldValue, searchFieldType := emailorNameDisambiguation(searchField)
	searchCriteria.Fields = &[]string{searchFieldType}
	searchCriteria.Value = &searchFieldValue

	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
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
		userId = *(*users.Results)[0].Id
		return nil
	})
	return userId, diag

}

func emailorNameDisambiguation(searchField string) (string, string) {
	emailField := "email"
	nameField := "name"
	_, err := mail.ParseAddress(searchField)
	if err == nil {
		return searchField, emailField
	}
	return searchField, nameField
}
