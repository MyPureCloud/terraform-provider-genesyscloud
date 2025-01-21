package knowledgedocumentvariation

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"
)

func dataSourceKnowledgeDocumentVariationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newVariationRequestProxy(sdkConfig)

	name := d.Get("name").(string)
	knowledgeBaseID := d.Get("knowledge_base_id").(string)
	fullID := d.Get("knowledge_document_id").(string)
	knowledgeDocumentId := strings.Split(fullID, ",")[0]

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		documentVariationRequestId, resp, retryable, err := proxy.getVariationRequestIdByName(ctx, name, knowledgeBaseID, knowledgeDocumentId)

		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error searching variation request %s | error: %s", name, err), resp))
		}

		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No variation request found with name %s", name), resp))
		}

		d.SetId(documentVariationRequestId)
		return nil
	})
}
