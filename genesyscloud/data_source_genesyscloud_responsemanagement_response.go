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

func dataSourceResponsemanagementResponse() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Responsemanagement Response. Select a Responsemanagement Response by name.`,

		ReadContext: ReadWithPooledClient(dataSourceResponsemanagementResponseRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Responsemanagement Response name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"library_id": {
				Description: `ID of the library that contains the response.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceResponsemanagementResponseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	name := d.Get("name").(string)
	library := d.Get("library_id").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			sdkresponseentitylisting, _, getErr := responseManagementApi.GetResponsemanagementResponses(library, pageNum, pageSize, "")
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting Responsemanagement Response %s: %s", name, getErr))
			}

			if sdkresponseentitylisting.Entities == nil || len(*sdkresponseentitylisting.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No Responsemanagement Response found with name %s", name))
			}

			for _, entity := range *sdkresponseentitylisting.Entities {
				if entity.Name != nil && *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
