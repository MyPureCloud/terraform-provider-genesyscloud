package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func dataSourceStation() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Stations. Select a station by name.",
		ReadContext: ReadWithPooledClient(dataSourceStationRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Station name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func dataSourceStationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	stationsAPI := platformclientv2.NewStationsApiWithConfig(sdkConfig)

	stationName := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageSize = 50
		const pageNum = 1
		stations, _, getErr := stationsAPI.GetStations(pageSize, pageNum, "", stationName, "", "", "", "")
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting station %s", getErr))
		}

		if stations.Entities == nil || len(*stations.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("No stations found"))
		}

		d.SetId(*(*stations.Entities)[0].Id)
		return nil
	})
}
