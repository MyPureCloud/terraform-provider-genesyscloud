package recording_media_retention_policy

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The data_source_genesyscloud_recording_media_retention_policy.go contains the data source implementation
   for the resource.

   Note:  This code should contain no code for doing the actual lookup in Genesys Cloud.  Instead,
   it should be added to the _proxy.go file for the class using our proxy pattern.
*/

func dataSourceRecordingMediaRetentionPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	pp := getPolicyProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		policy, retryable, resp, err := pp.getPolicyByName(ctx, name)
		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting media retention policy %s: %s %v", name, err, resp))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("no media retention policy found with name %s", name))
		}

		d.SetId(*policy.Id)
		return nil
	})
}
