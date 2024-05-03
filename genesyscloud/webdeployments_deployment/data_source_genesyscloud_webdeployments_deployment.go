package webdeployments_deployment

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDeploymentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	wd := getWebDeploymentsProxy(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		deployments, resp, err := wd.getWebDeployments(ctx)

		if err != nil && resp.StatusCode == http.StatusNotFound {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No web deployment record found %s | error: %s", name, err), resp))
		}

		if err != nil && resp.StatusCode != http.StatusNotFound {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error retrieving web deployment %s | error: %s", name, err), resp))
		}

		for _, deployment := range *deployments.Entities {
			if name == *deployment.Name {
				d.SetId(*deployment.Id)
				return nil
			}
		}

		return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("No web deployment was found with the name %s", name), resp))
	})
}
