package genesyscloud

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v154/platformclientv2"
)

func dataSourceWidgetDeployments() *schema.Resource {
	return &schema.Resource{
		Description:        "Data source for Genesys Cloud Widget Deployment. Select a widget deployment.",
		DeprecationMessage: "The CX as Code team will be removing the genesyscloud_widget_deployment resource and data source from the CX as Code Terraform provider in mid-April. If you are using these resources you must upgrade your CX as Code provider version after mid-April and before mid-June, you will experience errors in your CI/CD pipelines and CX as Code exports with the removal of /api/v2/widgets/deployments APIs.",
		ReadContext:        provider.ReadWithPooledClient(dataSourceWidgetDeploymentRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Widget Deployment Name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceWidgetDeploymentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	widgetAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query widget by name. Retry in case search has not yet indexed the widget.
	return util.WithRetries(ctx, 5*time.Second, func() *retry.RetryError {
		widgetDeployments, resp, getErr := widgetAPI.GetWidgetsDeployments()

		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_widget_deployment", fmt.Sprintf("Error requesting widget deployment %s | error: %s", name, getErr), resp))
		}

		if widgetDeployments.Entities == nil || len(*widgetDeployments.Entities) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_widget_deployment", fmt.Sprintf("No widget deployment found with name %s", name), resp))
		}

		for _, widgetDeployment := range *widgetDeployments.Entities {
			if *widgetDeployment.Name == name {
				d.SetId(*widgetDeployment.Id)
				return nil
			}
		}

		return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_widget_deployment", fmt.Sprintf("Unable to locate widget deployment name %s. It does not exist", name), resp))
	})
}
