package webdeployments_deployment

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDeploymentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*genesyscloud.ProviderMeta).ClientConfig
	wd := getWebDeploymentsProxy(sdkConfig)

	name := d.Get("name").(string)

	return genesyscloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		deployments, resp, err := wd.getWebDeployments(ctx)

		if err != nil && resp.StatusCode == http.StatusNotFound {
			return retry.RetryableError(fmt.Errorf("No web deployment record found %s: %s. Correlation id: %s", name, err, resp.CorrelationID))
		}

		if err != nil && resp.StatusCode != http.StatusNotFound {
			return retry.NonRetryableError(fmt.Errorf("Error retrieving web deployment %s: %s. Correlation id: %s", name, err, resp.CorrelationID))
		}

		for _, deployment := range *deployments.Entities {
			if name == *deployment.Name {
				d.SetId(*deployment.Id)
				return nil
			}
		}

		return retry.NonRetryableError(fmt.Errorf("No web deployment was found with the name %s", name))
	})
}
