package guide_version

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func dataSourceGuideVersionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideVersionProxy(sdkConfig)

	guideId := d.Get("guide_id").(string)
	versionId := d.Get("version_id").(string)

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		version, resp, getErr := proxy.getGuideVersionById(ctx, versionId, guideId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read guide version %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read guide version %s | error: %s", d.Id(), getErr), resp))
		}

		if version.Version == "" {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, "version ID is empty in response", resp))
		}

		d.SetId(version.Version)
		return nil
	})
}
