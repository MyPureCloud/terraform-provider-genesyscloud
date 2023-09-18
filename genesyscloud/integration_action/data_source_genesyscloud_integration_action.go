package integration_action

import (
	"context"
	"fmt"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceIntegrationActionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	actionName := d.Get("name").(string)

	// Query for integration actions by name. Retry in case new action is not yet indexed by search.
	// As action names are non-unique, fail in case of multiple results.
	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		actions, err := iap.getIntegrationActionsByName(ctx, actionName)

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("error requesting data action %s: %s", actionName, err))
		}

		if len(*actions) == 0 {
			return retry.RetryableError(fmt.Errorf("no data actions found with name %s", actionName))
		}

		if len(*actions) > 1 {
			return retry.NonRetryableError(fmt.Errorf("ambiguous data action name: %s", actionName))
		}

		action := (*actions)[0]
		d.SetId(*action.Id)
		return nil
	})
}
