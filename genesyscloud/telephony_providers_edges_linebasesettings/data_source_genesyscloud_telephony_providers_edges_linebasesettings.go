package telephony_providers_edges_linebasesettings

import (
	"context"
	"fmt"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func dataSourceLineBaseSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 50
			lineBaseSettings, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesLinebasesettings(pageNum, pageSize, "", "", nil)
			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting line base settings %s | error: %s", name, getErr), resp))
			}

			if lineBaseSettings.Entities == nil || len(*lineBaseSettings.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No lineBaseSettings found with name %s", name), resp))
			}

			for _, lineBaseSetting := range *lineBaseSettings.Entities {
				if lineBaseSetting.Name != nil && *lineBaseSetting.Name == name &&
					lineBaseSetting.State != nil && *lineBaseSetting.State != "deleted" {
					d.SetId(*lineBaseSetting.Id)
					return nil
				}
			}
		}
	})
}
