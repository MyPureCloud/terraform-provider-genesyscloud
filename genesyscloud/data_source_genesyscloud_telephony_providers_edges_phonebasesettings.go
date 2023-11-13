package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func dataSourcePhoneBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Phone Base Settings. Select a phone base settings by name",
		ReadContext: ReadWithPooledClient(dataSourcePhoneBaseSettingsRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Phone Base Settings name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourcePhoneBaseSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 50
			phoneBaseSettings, _, getErr := edgesAPI.GetTelephonyProvidersEdgesPhonebasesettings(pageSize, pageNum, "", "", nil, name)
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting phone base settings %s: %s", name, getErr))
			}

			if phoneBaseSettings.Entities == nil || len(*phoneBaseSettings.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No phoneBaseSettings found with name %s", name))
			}

			for _, phoneBaseSetting := range *phoneBaseSettings.Entities {
				if phoneBaseSetting.Name != nil && *phoneBaseSetting.Name == name &&
					phoneBaseSetting.State != nil && *phoneBaseSetting.State != "deleted" {
					d.SetId(*phoneBaseSetting.Id)
					return nil
				}
			}
		}
	})
}
