package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v67/platformclientv2"
)

func dataSourceQualityFormsSurvey() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud survey form. Select aform by name",
		ReadContext: readWithPooledClient(dataSourceQualityFormsSurveyRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Survey form policy name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceQualityFormsSurveyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			policy, _, getErr := qualityAPI.GetQualityFormsSurveys(pageSize, pageNum, "", "", "", "", name, "desc")

			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting media retention policy %s: %s", name, getErr))
			}

			if policy.Entities == nil || len(*policy.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No media retention policy found with name %s", name))
			}

			d.SetId(*(*policy.Entities)[0].Id)
			return nil
		}
	})
}
