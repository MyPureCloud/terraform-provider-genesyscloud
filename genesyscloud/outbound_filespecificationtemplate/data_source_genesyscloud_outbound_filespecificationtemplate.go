package outbound_filespecificationtemplate

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOutboundFileSpecificationTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundFilespecificationtemplateProxy(sdkConfig)
	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		fstId, retryable, err := proxy.getOutboundFilespecificationtemplateIdByName(ctx, name)

		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("Error requesting file specification template %s: %s", name, err))
		}
		if retryable {
			return retry.RetryableError(fmt.Errorf("No file specification template found with name %s", name))
		}
		d.SetId(fstId)
		return nil
	})
}
