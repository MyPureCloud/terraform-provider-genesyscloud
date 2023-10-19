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

func dataSourceJourneyActionMap() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Action Map. Select a journey action map by name",
		ReadContext: ReadWithPooledClient(dataSourceJourneyActionMapRead),
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
	sdkConfig := m.(*ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		pageCount := 1 // Needed because of broken journey common paging
		for pageNum := 1; pageNum <= pageCount; pageNum++ {
			const pageSize = 100
			journeyActionMaps, _, getErr := journeyApi.GetJourneyActionmaps(pageNum, pageSize, "", "", "", nil, nil, "")
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("failed to get page of journey action maps: %v", getErr))
			}

			if journeyActionMaps.Entities == nil || len(*journeyActionMaps.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no journey action map found with name %s", name))
			}

			for _, actionMap := range *journeyActionMaps.Entities {
				if actionMap.DisplayName != nil && *actionMap.DisplayName == name {
					d.SetId(*actionMap.Id)
					return nil
				}
			}

			pageCount = *journeyActionMaps.PageCount
		}
		return retry.RetryableError(fmt.Errorf("no journey action map found with name %s", name))
	})
}
