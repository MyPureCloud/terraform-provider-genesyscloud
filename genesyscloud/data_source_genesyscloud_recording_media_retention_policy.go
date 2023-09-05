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

func dataSourceRecordingMediaRetentionPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud media retention policy. Select a policy by name",
		ReadContext: ReadWithPooledClient(dataSourceRecordingMediaRetentionPolicyRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Media retention policy name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceRecordingMediaRetentionPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	recordingAPI := platformclientv2.NewRecordingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			policy, _, getErr := recordingAPI.GetRecordingMediaretentionpolicies(pageSize, pageNum, "", nil, "", "", name, true, false, false, 0)

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting media retention policy %s: %s", name, getErr))
			}

			if policy.Entities == nil || len(*policy.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No media retention policy found with name %s", name))
			}

			d.SetId(*(*policy.Entities)[0].Id)
			return nil
		}
	})
}
