package teams_resource

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

/*
   The data_source_genesyscloud_teams_resource.go contains the data source implementation
   for the resource.
*/

// dataSourceTeamsResourceRead retrieves by name the id in question
func dataSourceTeamsResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newTeamsResourceProxy(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		teamId, retryable, err := proxy.getTeamsResourceIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching teams resource %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No teams resource found with name %s", name))
		}

		d.SetId(teamId)
		return nil
	})
}
