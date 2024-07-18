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
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func DataSourceLocation() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Location. Select a location by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceLocationRead),
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
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)

	exactSearchType := "EXACT"
	nameField := "name"
	nameStr := d.Get("name").(string)

	searchCriteria := platformclientv2.Locationsearchcriteria{
		VarType: &exactSearchType,
		Value:   &nameStr,
		Fields:  &[]string{nameField},
	}

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		locations, resp, getErr := locationsAPI.PostLocationsSearch(platformclientv2.Locationsearchrequest{
			Query: &[]platformclientv2.Locationsearchcriteria{searchCriteria},
		})
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_location", fmt.Sprintf("Error requesting location %s | error: %s", nameStr, getErr), resp))
		}

		if *locations.Total == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_location", fmt.Sprintf("No locations found with search criteria %v ", searchCriteria), resp))
		}

		// Select first location in the list
		location := (*locations.Results)[0]
		d.SetId(*location.Id)
		return nil
	})
}
