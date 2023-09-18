package integration_credential

import (
	"context"
	"fmt"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceIntegrationCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationCredsProxy(sdkConfig)

	credName := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		credential, retryable, err := ip.getIntegrationCredByName(ctx, credName)
		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("failed to get integration credential: %s. %s", credential, err))
		}

		if retryable {
			return retry.RetryableError(fmt.Errorf("no integration credential found: %s", credName))
		}

		d.SetId(*credential.Id)
		return nil
	})
}
