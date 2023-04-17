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

func dataSourcePhone() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Phone. Select a phone by name",
		ReadContext: readWithPooledClient(dataSourcePhoneRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Phone name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourcePhoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			phone, _, getErr := edgesAPI.GetTelephonyProvidersEdgesPhones(pageNum, pageSize, "", "", "", "", "", "", "", "", "", "", name, "", "", nil, nil)

			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting phone %s: %s", name, getErr))
			}

			if phone.Entities == nil || len(*phone.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No phone found with name %s", name))
			}

			d.SetId(*(*phone.Entities)[0].Id)
			return nil
		}
	})
}
