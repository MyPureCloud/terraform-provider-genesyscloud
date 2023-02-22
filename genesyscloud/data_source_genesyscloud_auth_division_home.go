package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

func dataSourceAuthDivisionHome() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Divisions. Get the Home division",
		ReadContext: readWithPooledClient(dataSourceAuthDivisionHomeRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Home division name.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "Home division description.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceAuthDivisionHomeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	// Query home division
	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		division, _, getErr := authAPI.GetAuthorizationDivisionsHome()
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("Error requesting division: %s", getErr))
		}

		d.SetId(*division.Id)
		d.Set("name", *division.Name)
		if division.Description != nil {
			d.Set("description", *division.Description)
		}

		return nil
	})
}
