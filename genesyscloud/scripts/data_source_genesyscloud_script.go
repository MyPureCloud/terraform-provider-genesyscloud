package scripts

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v154/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
DataSource for the Scripts resource
*/

// dataSourceScriptRead provides the main terraform code needed to read a script resource by name
func dataSourceScriptRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		sdkConfig    = m.(*provider.ProviderMeta).ClientConfig
		scriptsProxy = getScriptsProxy(sdkConfig)
		name         = d.Get("name").(string)

		apiResponse *platformclientv2.APIResponse
		err         error
		retryable   bool
		scriptId    string
	)

	// Query for scripts by name. Retry in case new script is not yet indexed by search.
	retryErr := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		scriptId, retryable, apiResponse, err = scriptsProxy.getScriptIdByName(ctx, name)
		if err != nil {
			if retryable {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		d.SetId(scriptId)
		return nil
	})

	if retryErr != nil {
		if apiResponse != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to find script with name %s: %s", name, err), apiResponse)
		}
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to find script with name %s: %s", name, err), err)
	}

	return nil
}
