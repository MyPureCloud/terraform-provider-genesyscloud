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

func dataSourceRoutingLanguage() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Languages. Select a language by name.",
		ReadContext: ReadWithPooledClient(dataSourceRoutingLanguageRead),
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
	sdkConfig := m.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Find first non-deleted language by name. Retry in case new language is not yet indexed by search
	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 50
			languages, _, getErr := routingAPI.GetRoutingLanguages(pageSize, pageNum, "", name, nil)
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting language %s: %s", name, getErr))
			}

			if languages.Entities == nil || len(*languages.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No routing languages found with name %s", name))
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
