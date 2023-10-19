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

func dataSourceEmployeeperformanceExternalmetricsDefinition() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Employeeperformance Externalmetrics Definition. Select a Employeeperformance Externalmetrics Definition by name.`,

		ReadContext: ReadWithPooledClient(dataSourceEmployeeperformanceExternalmetricsDefinitionRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Employeeperformance Externalmetrics Definition name.`,
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func dataSourceEmployeeperformanceExternalmetricsDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	gamificationApi := platformclientv2.NewGamificationApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			sdkexternalmetricdefinitionlisting, _, getErr := gamificationApi.GetEmployeeperformanceExternalmetricsDefinitions(pageSize, pageNum)
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting Employeeperformance Externalmetrics Definition %s: %s", name, getErr))
			}

			if sdkexternalmetricdefinitionlisting.Entities == nil || len(*sdkexternalmetricdefinitionlisting.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No Employeeperformance Externalmetrics Definition found with name %s", name))
			}

			for _, entity := range *sdkexternalmetricdefinitionlisting.Entities {
				if entity.Name != nil && *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
