package telephony_providers_edges_linebasesettings

import (
	"context"
	"fmt"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

func dataSourceLineBaseSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 50
			lineBaseSettings, _, getErr := edgesAPI.GetTelephonyProvidersEdgesLinebasesettings(pageNum, pageSize, "", "", nil)
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting line base settings %s: %s", name, getErr))
			}

			if lineBaseSettings.Entities == nil || len(*lineBaseSettings.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No lineBaseSettings found with name %s", name))
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
