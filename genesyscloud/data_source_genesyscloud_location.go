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

func DataSourceLocation() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Location. Select a location by name.",
		ReadContext: ReadWithPooledClient(dataSourceLocationRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Location name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceLocationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)

	exactSearchType := "EXACT"
	nameField := "name"
	nameStr := d.Get("name").(string)

	searchCriteria := platformclientv2.Locationsearchcriteria{
		VarType: &exactSearchType,
		Value:   &nameStr,
		Fields:  &[]string{nameField},
	}

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		locations, _, getErr := locationsAPI.PostLocationsSearch(platformclientv2.Locationsearchrequest{
			Query: &[]platformclientv2.Locationsearchcriteria{searchCriteria},
		})
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting location %s: %s", nameStr, getErr))
		}

		if *locations.Total == 0 {
			return retry.RetryableError(fmt.Errorf("No locations found with search criteria %v ", searchCriteria))
		}

		// Select first location in the list
		location := (*locations.Results)[0]
		d.SetId(*location.Id)
		return nil
	})
}
