package integration

import (
	"context"
	"fmt"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceIntegrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationsProxy(sdkConfig)

	integrationName := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		integration, retryable, err := ip.getIntegrationByName(ctx, integrationName)
		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("failed to get page of integrations: %s. %s", integrationName, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("failed to get integration %s", integrationName))
		}

		d.SetId(*integration.Id)
		return nil
	})
}
