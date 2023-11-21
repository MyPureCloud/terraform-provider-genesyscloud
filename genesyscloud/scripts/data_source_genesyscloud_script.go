package scripts

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
DataSource for the Scripts resource
*/

// dataSourceScriptRead provides the main terraform code needed to read a script resource by name
func dataSourceScriptRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	name := d.Get("name").(string)

	// Query for scripts by name. Retry in case new script is not yet indexed by search.
	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		scriptId, retryable, err := scriptsProxy.getScriptIdByName(ctx, name)
		if err != nil {
			if retryable {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		d.SetId(scriptId)
		return nil
	})
}
