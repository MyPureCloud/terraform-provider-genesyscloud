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

func dataSourceJourneyActionTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Action Template. Select a journey action template by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceJourneyActionTemplateRead),
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
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	var response *platformclientv2.APIResponse

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		pageCount := 1 // Needed because of broken journey common paging
		for pageNum := 1; pageNum <= pageCount; pageNum++ {
			const pageSize = 100
			journeyActionTemplates, resp, getErr := journeyApi.GetJourneyActiontemplates(pageNum, pageSize, "", "", "", nil, "")
			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_action_template", fmt.Sprintf("failed to get page of journey action template: %v", getErr), resp))
			}

			response = resp

			if journeyActionTemplates.Entities == nil || len(*journeyActionTemplates.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_action_template", fmt.Sprintf("no journey action template found with name %s", name), resp))
			}

			for _, actionTemplate := range *journeyActionTemplates.Entities {
				if actionTemplate.Name != nil && *actionTemplate.Name == name {
					d.SetId(*actionTemplate.Id)
					return nil
				}
			}

			pageCount = *journeyActionTemplates.PageCount
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_journey_action_template", fmt.Sprintf("no journey action template found with name %s", name), response))
	})
}
