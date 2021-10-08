package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v55/platformclientv2"
	"time"
)

func dataSourceTrunkBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Trunk Base Settings. Select a trunk base settings by name",
		ReadContext: readWithPooledClient(dataSourceTrunkBaseSettingsRead),
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
	sdkConfig := m.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			trunkBaseSettings, _, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunkbasesettings(pageNum, 100, "", "", false, true, false, []string{"properties"}, name)

			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting trunk base settings %s: %s", name, getErr))
			}

			if trunkBaseSettings.Entities == nil || len(*trunkBaseSettings.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No trunkBaseSettings found with name %s", name))
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
