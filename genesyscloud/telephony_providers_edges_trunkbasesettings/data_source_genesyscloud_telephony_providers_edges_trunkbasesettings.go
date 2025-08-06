package telephony_providers_edges_trunkbasesettings

import (
	"context"
	"fmt"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTrunkBaseSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		sdkConfig = m.(*provider.ProviderMeta).ClientConfig
		proxy     = getTrunkBaseSettingProxy(sdkConfig)

		name          = d.Get("name").(string)
		notFoundError = fmt.Errorf("no trunkbase settings found with name '%s'", name)
		response      *platformclientv2.APIResponse
	)

	retryErr := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		trunkBaseSettings, resp, getErr := proxy.GetAllTrunkBaseSettingWithName(ctx, name)
		response = resp

		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("error requesting trunk base setting with name '%s' | error: %s", name, getErr))
		}

		if trunkBaseSettings == nil || len(*trunkBaseSettings) == 0 {
			return retry.RetryableError(notFoundError)
		}

		for _, trunkBaseSetting := range *trunkBaseSettings {
			if trunkBaseSetting.Name == nil || *trunkBaseSetting.Name != name {
				continue
			}
			if trunkBaseSetting.State != nil && *trunkBaseSetting.State != "deleted" {
				d.SetId(*trunkBaseSetting.Id)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("found trunkbase setting with name '%s', but could not verify that state is not 'deleted'", name))
		}

		return retry.RetryableError(notFoundError)
	})

	if retryErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to read trunkbase setting with name '%s' | errror: %v", name, retryErr), response)
	}
	return nil
}
