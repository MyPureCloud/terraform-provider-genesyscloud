package telephony_provider_edges_trunkbasesettings

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

func dataSourceTrunkBaseSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getTrunkBaseSettingProxy(sdkConfig)

	name := d.Get("name").(string)
	return util.WithRetries(ctx, 1*time.Second, func() *retry.RetryError {
		trunkBaseSettings, resp, getErr := proxy.GetAllTrunkBaseSettingWithName(ctx, name)

		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting trunk base setting with name %s | error: %s", name, getErr), resp))
		}

		for _, trunkBaseSetting := range *trunkBaseSettings {
			if trunkBaseSetting.Name != nil && *trunkBaseSetting.Name == name &&
				trunkBaseSetting.State != nil && *trunkBaseSetting.State != "deleted" {
				d.SetId(*trunkBaseSetting.Id)

				return nil
			}
		}

		return nil
	})
}
