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

func dataSourceQualityFormsSurvey() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud survey form. Select a form by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceQualityFormsSurveyRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Survey form name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceQualityFormsSurveyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			forms, resp, getErr := qualityAPI.GetQualityFormsSurveys(pageSize, pageNum, "", "", "", "", name, "desc")

			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_quality_forms_survey", fmt.Sprintf("Error requesting survey forms %s | error: %s", name, getErr), resp))
			}

			if forms.Entities == nil || len(*forms.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_quality_forms_survey", fmt.Sprintf("No survey forms found with name %s", name), resp))
			}

			d.SetId(*(*forms.Entities)[0].Id)
			return nil
		}
	})
}
