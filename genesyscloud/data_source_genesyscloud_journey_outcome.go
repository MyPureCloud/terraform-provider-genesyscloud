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

func dataSourceJourneyOutcome() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Journey Outcome. Select a journey outcome by name",
		ReadContext: ReadWithPooledClient(dataSourceJourneyOutcomeRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Journey Outcome name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceJourneyOutcomeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		pageCount := 1 // Needed because of broken journey common paging
		for pageNum := 1; pageNum <= pageCount; pageNum++ {
			const pageSize = 100
			journeyOutcomes, _, getErr := journeyApi.GetJourneyOutcomes(pageNum, pageSize, "", nil, nil, "")
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("failed to get page of journey outcomes: %v", getErr))
			}

			if journeyOutcomes.Entities == nil || len(*journeyOutcomes.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no journey outcome found with name %s", name))
			}

			for _, journeyOutcome := range *journeyOutcomes.Entities {
				if journeyOutcome.DisplayName != nil && *journeyOutcome.DisplayName == name {
					d.SetId(*journeyOutcome.Id)
					return nil
				}
			}

			pageCount = *journeyOutcomes.PageCount
		}
		return retry.RetryableError(fmt.Errorf("no journey outcome found with name %s", name))
	})
}
