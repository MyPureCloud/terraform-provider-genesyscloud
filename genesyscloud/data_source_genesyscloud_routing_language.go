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
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
)

func dataSourceRoutingLanguage() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Languages. Select a language by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceRoutingLanguageRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Language name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceRoutingLanguageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Find first non-deleted language by name. Retry in case new language is not yet indexed by search
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 50
			languages, resp, getErr := routingAPI.GetRoutingLanguages(pageSize, pageNum, "", name, nil)
			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_language", fmt.Sprintf("Error requesting language %s | error: %s", name, getErr), resp))
			}

			if languages.Entities == nil || len(*languages.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_language", fmt.Sprintf("No routing languages found with name %s", name), resp))
			}

			for _, language := range *languages.Entities {
				if language.Name != nil && *language.Name == name &&
					language.State != nil && *language.State != "deleted" {
					d.SetId(*language.Id)
					return nil
				}
			}
		}
	})
}
