package telephony_providers_edges_trunk

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

func dataSourceTrunkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			trunks, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesTrunks(pageNum, pageSize, "", "", "", "", "")

			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error requesting trunk %s | error: %s", name, getErr), resp))
			}

			if trunks.Entities == nil || len(*trunks.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No trunk found with name %s", name), resp))
			}

			for _, trunk := range *trunks.Entities {
				if trunk.Name != nil && *trunk.Name == name &&
					trunk.State != nil && *trunk.State != "deleted" {
					d.SetId(*trunk.Id)
					return nil
				}
			}
		}
	})
}
