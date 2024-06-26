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

func dataSourceJourneyActionMap() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Action Map. Select a journey action map by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceJourneyActionMapRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Journey Action Map name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceJourneyActionMapRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	var response *platformclientv2.APIResponse

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		pageCount := 1 // Needed because of broken journey common paging
		for pageNum := 1; pageNum <= pageCount; pageNum++ {
			const pageSize = 100
			journeyActionMaps, resp, getErr := journeyApi.GetJourneyActionmaps(pageNum, pageSize, "", "", "", nil, nil, "")
			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_action_map", fmt.Sprintf("failed to get page of journey action maps: %v", getErr), resp))
			}
			response = resp

			if journeyActionMaps.Entities == nil || len(*journeyActionMaps.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_action_map", fmt.Sprintf("no journey action map found with name %s", name), resp))
			}

			for _, actionMap := range *journeyActionMaps.Entities {
				if actionMap.DisplayName != nil && *actionMap.DisplayName == name {
					d.SetId(*actionMap.Id)
					return nil
				}
			}

			pageCount = *journeyActionMaps.PageCount
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_action_map", fmt.Sprintf("no journey action map found with name %s", name), response))
	})
}
