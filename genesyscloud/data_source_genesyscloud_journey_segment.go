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

func dataSourceJourneySegment() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Journey Segment. Select a journey segment by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceJourneySegmentRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Journey Segment name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceJourneySegmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	var response *platformclientv2.APIResponse

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		pageCount := 1 // Needed because of broken journey common paging
		for pageNum := 1; pageNum <= pageCount; pageNum++ {
			const pageSize = 100
			journeySegments, resp, getErr := journeyApi.GetJourneySegments("", pageSize, pageNum, true, nil, nil, "")
			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("failed to get page of journey segments: %v", getErr), resp))
			}

			response = resp

			if journeySegments.Entities == nil || len(*journeySegments.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("no journey segment found with name %s", name), resp))
			}

			for _, journeySegment := range *journeySegments.Entities {
				if journeySegment.DisplayName != nil && *journeySegment.DisplayName == name {
					d.SetId(*journeySegment.Id)
					return nil
				}
			}

			pageCount = *journeySegments.PageCount
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_segment", fmt.Sprintf("no journey segment found with name %s", name), response))
	})
}
