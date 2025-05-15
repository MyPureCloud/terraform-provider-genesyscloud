package knowledge_document_variation

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"time"
)

func dataSourceKnowledgeDocumentVariationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newVariationRequestProxy(sdkConfig)

	name, _ := d.Get("name").(string)
	ids := getKnowledgeIdsFromResourceData(d)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		documentVariationRequestId, resp, retryable, err := proxy.getVariationRequestIdByName(ctx, name, ids.knowledgeBaseID, ids.knowledgeDocumentID)

		if err != nil {
			if retryable {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No variation request found with name %s", name), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching variation request %s | error: %s", name, err), resp))
		}

		d.SetId(documentVariationRequestId)
		return nil
	})
}
