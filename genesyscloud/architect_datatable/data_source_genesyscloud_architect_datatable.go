package architect_datatable

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

func DataSourceArchitectDatatableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query architect architect_datatable by name. Retry in case search has not yet indexed the architect architect_datatable.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		datatables, resp, getErr := archAPI.GetFlowsDatatables("", pageNum, pageSize, "", "", nil, name)
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error requesting architect architect_datatable %s | error: %s", name, getErr), resp))
		}

		if datatables.Entities == nil || len(*datatables.Entities) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No architect architect_datatable found with name %s", name), resp))
		}

		datatable := (*datatables.Entities)[0]
		d.SetId(*datatable.Id)
		return nil
	})
}
