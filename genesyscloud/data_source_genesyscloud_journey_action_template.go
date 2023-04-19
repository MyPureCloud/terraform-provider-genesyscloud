package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v99/platformclientv2"
)

func dataSourceJourneyActionTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Action Template. Select a journey action template by name",
		ReadContext: readWithPooledClient(dataSourceJourneyActionTemplateRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Journey Action Template name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceJourneyActionTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		pageCount := 1 // Needed because of broken journey common paging
		for pageNum := 1; pageNum <= pageCount; pageNum++ {
			const pageSize = 100
			journeyActionTemplates, _, getErr := journeyApi.GetJourneyActiontemplates(pageNum, pageSize, "", "", "", nil, "")
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("failed to get page of journey action template: %v", getErr))
			}

			if journeyActionTemplates.Entities == nil || len(*journeyActionTemplates.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("no journey action template found with name %s", name))
			}

			for _, actionTemplate := range *journeyActionTemplates.Entities {
				if actionTemplate.Name != nil && *actionTemplate.Name == name {
					d.SetId(*actionTemplate.Id)
					return nil
				}
			}

			pageCount = *journeyActionTemplates.PageCount
		}
		return resource.RetryableError(fmt.Errorf("no journey action template found with name %s", name))
	})
}
