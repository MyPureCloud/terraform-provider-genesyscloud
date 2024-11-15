package journey_views

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceJourneyViewRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	journeyId, err := getJourneyByNameFn(name, ctx, m)
	if err != nil {
		return err
	}

	d.SetId(journeyId)
	return nil
}

// getJourneyByNameFn returns the journey id (blank if not found) and diag
func getJourneyByNameFn(name string, ctx context.Context, m interface{}) (string, diag.Diagnostics) {
	config := m.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneyViewProxy(config)
	journeyId := ""

	diag := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		foundId, resp, getErr := proxy.getJourneyViewByName(ctx, name)
		if getErr != nil {
			errMsg := util.BuildWithRetriesApiDiagnosticError(name, fmt.Sprintf("error requesting journey view %s | error %s", name, getErr), resp)
			return retry.NonRetryableError(errMsg)
		}

		journeyId = foundId
		return nil
	})

	return journeyId, diag
}
