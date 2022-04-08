package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v67/platformclientv2"
)

func dataSourceDid() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud DID. The identifier is the E-164 phone number.",
		ReadContext: readWithPooledClient(dataSourceDidRead),
		Schema: map[string]*schema.Schema{
			"phone_number": {
				Description:      "Phone number for the DID.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validatePhoneNumber,
			},
		},
	}
}

func dataSourceDidRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	telephonyAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	didPhoneNumber := d.Get("phone_number").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			dids, _, getErr := telephonyAPI.GetTelephonyProvidersEdgesDids(100, pageNum, "", "", didPhoneNumber, "", "", nil)

			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("error requesting list of DIDs: %s", getErr))
			}

			if dids.Entities == nil || len(*dids.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("no DIDs found"))
			}

			for _, did := range *dids.Entities {
				if *did.PhoneNumber == didPhoneNumber {
					d.SetId(*did.Id)
					return nil
				}
			}

		}
	})

}
