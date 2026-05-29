package knowledge_document

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

func dataSourceKnowledgeDocumentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeDocumentProxy(sdkConfig)
	title := d.Get("title").(string)
	knowledgeBaseName := d.Get("knowledge_base_name").(string)
	categoryName, _ := d.Get("category_name").(string)

	return util.WithRetries(ctx, 10*time.Second, func() *retry.RetryError {
		documentId, retryable, resp, err := proxy.getKnowledgeDocumentByTitle(ctx, title, knowledgeBaseName, categoryName)

		if err != nil {
			errMsg := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get knowledge document by title %s | error: %s", title, err), resp)
			if !retryable {
				return retry.NonRetryableError(errMsg)
			}
			return retry.RetryableError(errMsg)
		}

		d.SetId(documentId)
		return nil
	})
}
