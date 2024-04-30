package genesyscloud

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"
)

func dataSourceAuthDivision() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Divisions. Select a division by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceAuthDivisionRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Division name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceAuthDivisionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query division by name. Retry in case search has not yet indexed the division.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageSize = 100
		const pageNum = 1
		divisions, _, getErr := authAPI.GetAuthorizationDivisions(pageSize, pageNum, "", nil, "", "", false, nil, name)
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting division %s: %s", name, getErr))
		}

		if divisions.Entities == nil || len(*divisions.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("No authorization divisions found with name %s", name))
		}

		for _, division := range *divisions.Entities {
			if *division.Name == name {
				d.SetId(*division.Id)
				return nil
			}
		}

		return retry.RetryableError(fmt.Errorf("No division with name %s found", name))
	})
}
