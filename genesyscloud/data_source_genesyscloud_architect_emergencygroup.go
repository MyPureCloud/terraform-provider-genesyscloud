package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v74/platformclientv2"
)

func dataSourceArchitectEmergencyGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Emergency Groups. Select an emergency group by name.",
		ReadContext: readWithPooledClient(dataSourceEmergencyGroupRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Emergency Group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceEmergencyGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*providerMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query emergency group by name. Retry in case search has not yet indexed the emergency group.
	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		const pageNum = 1
		const pageSize = 100
		emergencyGroups, _, getErr := archAPI.GetArchitectEmergencygroups(pageNum, pageSize, "", "", name)
		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("Error requesting emergency group %s: %s", name, getErr))
		}

		if emergencyGroups.Entities == nil || len(*emergencyGroups.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("No emergency groups found with name %s", name))
		}

		emergencyGroup := (*emergencyGroups.Entities)[0]
		d.SetId(*emergencyGroup.Id)
		return nil
	})
}
