package user

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

var (
	dataSourceUserCache *rc.DataSourceCache
)

func DataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	key := ""

	if email, ok := d.GetOk("email"); ok {
		key = email.(string)

	}
	if name, ok := d.GetOk("name"); ok {
		key = name.(string)

	}
	if d.Get("name").(string) == "" && d.Get("email").(string) == "" {
		return util.BuildDiagnosticError(resourceName, "no user search field specified", nil)
	}

	if dataSourceUserCache == nil {
		dataSourceUserCache = rc.NewDataSourceCache(sdkConfig, hydrateUserCache, getUserByName)
	}

	userId, err := rc.RetrieveId(dataSourceUserCache, resourceName, key, ctx)
	if err != nil {
		return err
	}

	d.SetId(userId)
	return nil
}

func hydrateUserCache(c *rc.DataSourceCache, ctx context.Context) error {
	log.Printf("hydrating cache for data source %s", resourceName)
	proxy := getUserProxy(c.ClientConfig)
	const pageSize = 100
	users, response, err := proxy.hydrateUserCache(ctx, pageSize, 1)
	if err != nil {
		return fmt.Errorf("failed to get first page of users: %v %v", err, response)
	}

	if users.Entities == nil || len(*users.Entities) == 0 {
		return nil
	}

	for _, user := range *users.Entities {
		c.Cache[*user.Name] = *user.Id
		c.Cache[*user.Email] = *user.Id

	}

	for pageNum := 2; pageNum <= *users.PageCount; pageNum++ {

		users, response, err := proxy.hydrateUserCache(ctx, pageSize, pageNum)

		log.Printf("hydrating cache for data source %s with page number: %v", resourceName, pageNum)
		if err != nil {
			return fmt.Errorf("failed to get page of users: %v %v", err, response)
		}
		if users.Entities == nil || len(*users.Entities) == 0 {
			break
		}
		// Add ids to cache
		for _, user := range *users.Entities {
			c.Cache[*user.Name] = *user.Id
			c.Cache[*user.Email] = *user.Id

		}
	}
	log.Printf("cache hydration completed for data source %s", resourceName)
	return nil
}

func getUserByName(c *rc.DataSourceCache, searchField string, ctx context.Context) (string, diag.Diagnostics) {
	log.Printf("getUserByName for data source %s", resourceName)
	proxy := getUserProxy(c.ClientConfig)
	userId := ""
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
		users, resp, getErr := proxy.getUserByName(ctx, platformclientv2.Usersearchrequest{
			SortBy:    &emailField,
			SortOrder: &sortOrderAsc,
			Query:     &[]platformclientv2.Usersearchcriteria{searchCriteria},
		})
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error requesting users: %s", getErr), resp))
		}

		if users.Results == nil || len(*users.Results) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No users found with search criteria %v", searchCriteria), resp))
		}

		// Select first user in the list
		userId = *(*users.Results)[0].Id
		return nil
	})

	log.Printf("getUserByName completed for data source %s", resourceName)
	return userId, diag
}
