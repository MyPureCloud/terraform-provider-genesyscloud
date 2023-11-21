package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTrunkBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Trunk Base Settings. Select a trunk base settings by name",
		ReadContext: ReadWithPooledClient(dataSourceTrunkBaseSettingsRead),
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
	sdkConfig := m.(*ProviderMeta).ClientConfig

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			trunkBaseSettings, _, getErr := getTelephonyProvidersEdgesTrunkbasesettings(sdkConfig, pageNum, pageSize, name)

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting trunk base settings %s: %s", name, getErr))
			}

			if trunkBaseSettings.Entities == nil || len(*trunkBaseSettings.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No trunkBaseSettings found with name %s", name))
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
