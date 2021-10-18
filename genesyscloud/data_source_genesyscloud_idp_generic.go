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

func dataSourceIdpGeneric() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Identity Provider. Select an identity provider by name",
		ReadContext: readWithPooledClient(dataSourceIdpGenericRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description:      "Name of the provider.",
				Type:             schema.TypeString,
				Required:         true,
			},
		},
	}
}

func dataSourceIdpGenericRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig

	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	idpName := d.Get("name").(string)

	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			identityProviders, _, getErr := idpAPI.GetIdentityproviders()

			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("error requesting list of identity providers: %s", getErr))
			}

			if identityProviders == nil  {
				return resource.RetryableError(fmt.Errorf("no identity providers found"))
			}

			for _, provider := range *identityProviders.Entities {
				fmt.Println(provider.Name)
				if provider.Name != nil && *provider.Name == idpName &&
					provider.Disabled != nil && *provider.Disabled == false {
					d.SetId(*provider.Id)
					return nil
				}
			}

		}
	})

}
