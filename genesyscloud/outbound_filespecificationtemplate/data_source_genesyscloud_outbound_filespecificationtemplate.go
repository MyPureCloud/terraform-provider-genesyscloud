package outbound_filespecificationtemplate

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

func dataSourceOutboundFileSpecificationTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		fileSpecificationTemplates, _, getErr := outboundAPI.GetOutboundFilespecificationtemplates(pageSize, pageNum, true, "", name, "", "")
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("error requesting file specification template %s: %s", name, getErr))
		}
		if fileSpecificationTemplates.Entities == nil || len(*fileSpecificationTemplates.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("no file specification templates found with name %s", name))
		}
		fileSpecificationTemplate := (*fileSpecificationTemplates.Entities)[0]
		d.SetId(*fileSpecificationTemplate.Id)
		return nil
	})
}
