package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func dataSourceQualityFormsEvaluations() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Evaluation Forms. Select an evaluations form by name",
		ReadContext: ReadWithPooledClient(dataSourceQualityFormsEvaluationsRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Evaluation Form name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceQualityFormsEvaluationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			form, _, getErr := qualityAPI.GetQualityForms(pageSize, pageNum, "", "", "", "", name, "")

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting evaluation form %s: %s", name, getErr))
			}

			if form.Entities == nil || len(*form.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No evaluation form found with name %s", name))
			}

			d.SetId(*(*form.Entities)[0].Id)
			return nil
		}
	})
}
