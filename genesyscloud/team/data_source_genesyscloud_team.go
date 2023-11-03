package team

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
   The data_source_genesyscloud_team.go contains the data source implementation
   for the resource.
*/

// dataSourceTeamRead retrieves by name the id in question
func dataSourceTeamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newTeamProxy(sdkConfig)

	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		teamId, retryable, err := proxy.getTeamIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error searching team %s: %s", name, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("No team found with name %s", name))
		}

		d.SetId(teamId)
		return nil
	})
}
