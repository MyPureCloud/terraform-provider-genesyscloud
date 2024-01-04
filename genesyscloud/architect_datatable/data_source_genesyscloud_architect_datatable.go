package architect_datatable

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

func DataSourceArchitectDatatableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*genesyscloud.ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query architect architect_datatable by name. Retry in case search has not yet indexed the architect architect_datatable.
	return genesyscloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		datatables, _, getErr := archAPI.GetFlowsDatatables("", pageNum, pageSize, "", "", nil, name)
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting architect architect_datatable %s: %s", name, getErr))
		}

		if datatables.Entities == nil || len(*datatables.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("No architect architect_datatable found with name %s", name))
		}

		datatable := (*datatables.Entities)[0]
		d.SetId(*datatable.Id)
		return nil
	})
}
