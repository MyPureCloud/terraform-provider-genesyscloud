package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v95/platformclientv2"
)

func dataSourceWidgetDeployments() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Widget Deployment. Select a widget deployment.",
		ReadContext: readWithPooledClient(dataSourceWidgetDeploymentRead),
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
	sdkConfig := m.(*ProviderMeta).ClientConfig
	widgetAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	// Query widget by name. Retry in case search has not yet indexed the widget.
	return withRetries(ctx, 5*time.Second, func() *resource.RetryError {
		widgetDeployments, _, getErr := widgetAPI.GetWidgetsDeployments()

		if getErr != nil {
			return resource.NonRetryableError(fmt.Errorf("Error requesting widget deployment %s: %s", name, getErr))
		}

		if widgetDeployments.Entities == nil || len(*widgetDeployments.Entities) == 0 {
			return resource.RetryableError(fmt.Errorf("No widget deployment found with name %s", name))
		}

		for _, widgetDeployment := range *widgetDeployments.Entities {
			if *widgetDeployment.Name == name {
				d.SetId(*widgetDeployment.Id)
				return nil
			}
		}

		return resource.NonRetryableError(fmt.Errorf("Unable to locate widget deployment name %s. It does not exist", name))
	})
}
