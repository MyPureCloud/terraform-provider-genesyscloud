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

func dataSourceResponsemanagementLibrary() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Responsemanagement Library. Select a Responsemanagement Library by name.`,

		ReadContext: ReadWithPooledClient(dataSourceResponsemanagementLibraryRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Responsemanagement Library name.`,
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func dataSourceResponsemanagementLibraryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			sdklibraryentitylisting, _, getErr := responseManagementApi.GetResponsemanagementLibraries(pageNum, pageSize, "", "")
			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting Responsemanagement Library %s: %s", name, getErr))
			}

			if sdklibraryentitylisting.Entities == nil || len(*sdklibraryentitylisting.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No Responsemanagement Library found with name %s", name))
			}

			for _, entity := range *sdklibraryentitylisting.Entities {
				if entity.Name != nil && *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
