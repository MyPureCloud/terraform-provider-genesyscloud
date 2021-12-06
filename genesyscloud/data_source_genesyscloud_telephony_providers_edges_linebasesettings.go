package genesyscloud

import (
	"context"
	"fmt"
	"github.com/ronanwatkins/terraform-plugin-sdk/v2/diag"
	"github.com/ronanwatkins/terraform-plugin-sdk/v2/helper/resource"
	"github.com/ronanwatkins/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
	"time"
)

func dataSourceLineBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Line Base Settings. Select a line base settings by name",
		ReadContext: readWithPooledClient(dataSourceLineBaseSettingsRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Line Base Settings name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceLineBaseSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 50
			lineBaseSettings, _, getErr := edgesAPI.GetTelephonyProvidersEdgesLinebasesettings(pageNum, pageSize, "", "", nil)
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting line base settings %s: %s", name, getErr))
			}

			if lineBaseSettings.Entities == nil || len(*lineBaseSettings.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No lineBaseSettings found with name %s", name))
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
