package location

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func dataSourceLocationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getLocationProxy(sdkConfig)

	exactSearchType := "EXACT"
	nameField := "name"
	nameStr := d.Get("name").(string)

	searchCriteria := platformclientv2.Locationsearchcriteria{
		VarType: &exactSearchType,
		Value:   &nameStr,
		Fields:  &[]string{nameField},
	}

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		location, resp, getErr := proxy.getLocationBySearch(ctx, &platformclientv2.Locationsearchrequest{
			Query: &[]platformclientv2.Locationsearchcriteria{searchCriteria},
		})
		if getErr != nil {
			if strings.Contains(getErr.Error(), "404") {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting location %s | error: %s", nameStr, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting location %s | error: %s", nameStr, getErr), resp))
		}

		d.SetId(*location.Id)
		return nil
	})
}
