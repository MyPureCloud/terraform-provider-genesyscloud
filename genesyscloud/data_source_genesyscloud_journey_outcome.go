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
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func dataSourceJourneyOutcome() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Journey Outcome. Select a journey outcome by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceJourneyOutcomeRead),
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
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	var response *platformclientv2.APIResponse

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		pageCount := 1 // Needed because of broken journey common paging
		for pageNum := 1; pageNum <= pageCount; pageNum++ {
			const pageSize = 100
			journeyOutcomes, resp, getErr := journeyApi.GetJourneyOutcomes(pageNum, pageSize, "", nil, nil, "")
			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("failed to get page of journey outcomes: %v", getErr), resp))
			}

			response = resp

			if journeyOutcomes.Entities == nil || len(*journeyOutcomes.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("no journey outcome found with name %s", name), resp))
			}

			for _, journeyOutcome := range *journeyOutcomes.Entities {
				if journeyOutcome.DisplayName != nil && *journeyOutcome.DisplayName == name {
					d.SetId(*journeyOutcome.Id)
					return nil
				}
			}

			pageCount = *journeyOutcomes.PageCount
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_outcome", fmt.Sprintf("no journey outcome found with name %s", name), response))
	})
}
