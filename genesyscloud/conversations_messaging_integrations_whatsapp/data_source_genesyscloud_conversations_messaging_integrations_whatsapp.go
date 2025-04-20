package conversations_messaging_integrations_whatsapp

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

/*
   The data_source_genesyscloud_conversations_messaging_integrations_whatsapp.go contains the data source implementation
   for the resource.
*/

var (
	dataSourceIntegrationWhatsappCache *rc.DataSourceCache
)

// dataSourceConversationsMessagingIntegrationsWhatsappRead retrieves by name the id in question
func dataSourceConversationsMessagingIntegrationsWhatsappRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig

	key := d.Get("name").(string)

	if dataSourceIntegrationWhatsappCache == nil {
		log.Printf("Instantiating the %s data source cache object", ResourceType)
		dataSourceIntegrationWhatsappCache = rc.NewDataSourceCache(sdkConfig, hydrateIntegrationsWhatsappCacheFn, getWhatsappByNameFn)
	}

	whatsappId, err := rc.RetrieveId(dataSourceIntegrationWhatsappCache, ResourceType, key, ctx)
	if err != nil {
		return err
	}

	d.SetId(whatsappId)
	return nil
}

// hydrateIntegrationsWhatsapp for hydrating the cache with Genesys Cloud whatsapp integrations using the SDK
func hydrateIntegrationsWhatsappCacheFn(c *rc.DataSourceCache, ctx context.Context) error {
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(c.ClientConfig)

	log.Printf("Hydrating cache for data source %s", ResourceType)

	whatsappIntegrations, resp, err := proxy.getAllConversationsMessagingIntegrationsWhatsapp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get whatsapp integrations. Error: %s | API Response: %s", err, resp)
	}

	if whatsappIntegrations != nil || len(*whatsappIntegrations) != 0 {
		log.Printf("no integrations returned. Cache will remain empty")
		return nil
	}

	for _, whatsapp := range *whatsappIntegrations {
		c.Cache[*whatsapp.Name] = *whatsapp.Id
	}

	log.Printf("cache hydration completed for data source %s", ResourceType)

	return nil
}

// getWhatsappByNameFn returns the whatsapp id (blank if not found) and diag
func getWhatsappByNameFn(c *rc.DataSourceCache, name string, ctx context.Context) (string, diag.Diagnostics) {
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(c.ClientConfig)

	whatsappId := ""

	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		whatsappID, retryable, resp, err := proxy.getConversationsMessagingIntegrationsWhatsappIdByName(ctx, name)
		if err != nil {
			errMsg := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No whatsapp integration found with name %s : %s", name, err), resp)
			if !retryable {
				return retry.NonRetryableError(errMsg)
			}
			return retry.RetryableError(errMsg)
		}

		whatsappId = whatsappID
		return nil
	})

	return whatsappId, diag

}
