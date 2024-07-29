package conversations_messaging_supportedcontent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

/*
   The data_source_genesyscloud_conversations_messaging_supportedcontent.go contains the data source implementation
   for the resource.
*/

// dataSourceSupportedContentRead retrieves by name the id in question
func dataSourceSupportedContentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig

	key := ""

	key = d.Get("name").(string)

	if dataSourceSupportedContentCache == nil {
		dataSourceSupportedContentCache = rc.NewDataSourceCache(sdkConfig, hydrateSupportedContentCacheFn, getSupportedContentIdByName)
	}

	contentId, err := rc.RetrieveId(dataSourceSupportedContentCache, resourceName, key, ctx)
	if err != nil {
		return err
	}

	d.SetId(contentId)
	return nil
}

var (
	dataSourceSupportedContentCache *rc.DataSourceCache
)

func hydrateSupportedContentCacheFn(c *rc.DataSourceCache) error {
	log.Printf("hydrating cache for data source " + resourceName)
	supportContentApi := platformclientv2.NewConversationsApiWithConfig(c.ClientConfig)

	const pageSize = 100

	supportedContents, _, err := supportContentApi.GetConversationsMessagingSupportedcontent(pageSize, 1)
	if err != nil {
		return fmt.Errorf("Failed to get supported content: %v", err)
	}
	if supportedContents.Entities == nil || len(*supportedContents.Entities) == 0 {
		return nil
	}

	for _, supportedContent := range *supportedContents.Entities {
		c.Cache[*supportedContent.Name] = *supportedContent.Id
	}

	for pageNum := 2; pageNum <= *supportedContents.PageCount; pageNum++ {
		supportedContents, _, err := supportContentApi.GetConversationsMessagingSupportedcontent(pageSize, pageNum)

		log.Printf("hydrating cache for data source genesyscloud_conversations_messaging_supportedcontent with page number: %v", pageNum)
		if err != nil {
			return fmt.Errorf("Failed to get supported content: %v", err)
		}

		if supportedContents.Entities == nil || len(*supportedContents.Entities) == 0 {
			break
		}

		// Add ids to cache
		for _, supportedContent := range *supportedContents.Entities {
			c.Cache[*supportedContent.Name] = *supportedContent.Id
		}
	}
	log.Printf("cache hydration completed for data source " + resourceName)

	return nil
}

func getSupportedContentIdByName(c *rc.DataSourceCache, searchName string, ctx context.Context) (string, diag.Diagnostics) {
	proxy := getSupportedContentProxy(c.ClientConfig)
	contentId := ""
	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		supportedContentId, retryable, resp, err := proxy.getSupportedContentIdByName(ctx, searchName)

		if err != nil && !retryable {
			retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error searching supported content %s: %s", searchName, err), resp))
		}

		if retryable {
			retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No supported content found with name %s", searchName), resp))
		}

		contentId = supportedContentId
		return nil
	})

	return contentId, diag
}
