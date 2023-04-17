package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v99/platformclientv2"
)

func dataSourceArchitectDatatable() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Architect Datatables. Select an architect datatable by name.",
		ReadContext: readWithPooledClient(dataSourceArchitectDatatableRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Datatable name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceArchitectDatatableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query architect datatable by name. Retry in case search has not yet indexed the architect datatable.
	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		const pageNum = 1
		const pageSize = 100
		datatables, _, getErr := archAPI.GetFlowsDatatables("", pageNum, pageSize, "", "", nil, name)
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("Error requesting architect datatable %s: %s", name, getErr))
		}

		if datatables.Entities == nil || len(*datatables.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("No architect datatable found with name %s", name))
		}

		datatable := (*datatables.Entities)[0]
		d.SetId(*datatable.Id)
		return nil
	})
}
