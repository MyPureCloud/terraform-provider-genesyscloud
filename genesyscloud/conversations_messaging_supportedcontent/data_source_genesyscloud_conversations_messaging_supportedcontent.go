package conversations_messaging_supportedcontent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

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

	contentId, err := rc.RetrieveId(dataSourceSupportedContentCache, ResourceType, key, ctx)
	if err != nil {
		return err
	}

	d.SetId(contentId)
	return nil
}

var (
	dataSourceSupportedContentCache *rc.DataSourceCache
)

func hydrateSupportedContentCacheFn(c *rc.DataSourceCache, ctx context.Context) error {
	log.Printf("hydrating cache for data source " + ResourceType)
	proxy := getSupportedContentProxy(c.ClientConfig)

	supportedContents, resp, getErr := proxy.getAllSupportedContent(ctx)
	if getErr != nil {
		return fmt.Errorf("failed to get supported content: %v %v", getErr, resp)
	}

	if supportedContents == nil || len(*supportedContents) == 0 {
		return nil
	}

	for _, supportedContent := range *supportedContents {
		c.Cache[*supportedContent.Name] = *supportedContent.Id
	}
	log.Printf("cache hydration completed for data source " + ResourceType)
	return nil
}

func getSupportedContentIdByName(c *rc.DataSourceCache, searchName string, ctx context.Context) (string, diag.Diagnostics) {
	proxy := getSupportedContentProxy(c.ClientConfig)
	contentId := ""
	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		supportedContentId, retryable, resp, err := proxy.getSupportedContentIdByName(ctx, searchName)

		if err != nil && !retryable {
			retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching supported content %s: %s", searchName, err), resp))
		}

		if retryable {
			retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No supported content found with name %s", searchName), resp))
		}

		contentId = supportedContentId
		return nil
	})

	return contentId, diag
}
