package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v45/platformclientv2"
)

func dataSourceRoutingLanguage() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Languages. Select a language by name.",
		ReadContext: readWithPooledClient(dataSourceRoutingLanguageRead),
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
	sdkConfig := m.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Find first non-deleted language by name. Retry in case new language is not yet indexed by search
	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			languages, _, getErr := routingAPI.GetRoutingLanguages(50, pageNum, "", name, nil)
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting language %s: %s", name, getErr))
			}

			if languages.Entities == nil || len(*languages.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No routing languages found with name %s", name))
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
