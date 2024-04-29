package telephony

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

func DataSourceTrunkBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Trunk Base Settings. Select a trunk base settings by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceTrunkBaseSettingsRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Trunk Base Settings name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceTrunkBaseSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			trunkBaseSettings, resp, getErr := getTelephonyProvidersEdgesTrunkbasesettings(sdkConfig, pageNum, pageSize, name)

			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error requesting trunk base settings %s | error: %s", name, getErr), resp))
			}

			if trunkBaseSettings.Entities == nil || len(*trunkBaseSettings.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No trunkBaseSettings found with name %s", name), resp))
			}

			for _, trunkBaseSetting := range *trunkBaseSettings.Entities {
				if trunkBaseSetting.Name != nil && *trunkBaseSetting.Name == name &&
					trunkBaseSetting.State != nil && *trunkBaseSetting.State != "deleted" {
					d.SetId(*trunkBaseSetting.Id)
					return nil
				}
			}
		}
	})
}
