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

func dataSourceQualityFormsSurvey() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud survey form. Select a form by name",
		ReadContext: ReadWithPooledClient(dataSourceQualityFormsSurveyRead),
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
	sdkConfig := m.(*ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			forms, _, getErr := qualityAPI.GetQualityFormsSurveys(pageSize, pageNum, "", "", "", "", name, "desc")

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting survey forms %s: %s", name, getErr))
			}

			if forms.Entities == nil || len(*forms.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No survey forms found with name %s", name))
			}

			d.SetId(*(*forms.Entities)[0].Id)
			return nil
		}
	})
}
