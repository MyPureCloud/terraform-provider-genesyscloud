package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v98/platformclientv2"
)

func dataSourcePhoneBaseSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Phone Base Settings. Select a phone base settings by name",
		ReadContext: readWithPooledClient(dataSourcePhoneBaseSettingsRead),
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

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 50
			phoneBaseSettings, _, getErr := edgesAPI.GetTelephonyProvidersEdgesPhonebasesettings(pageSize, pageNum, "", "", nil, name)
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting phone base settings %s: %s", name, getErr))
			}

			if phoneBaseSettings.Entities == nil || len(*phoneBaseSettings.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No phoneBaseSettings found with name %s", name))
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
