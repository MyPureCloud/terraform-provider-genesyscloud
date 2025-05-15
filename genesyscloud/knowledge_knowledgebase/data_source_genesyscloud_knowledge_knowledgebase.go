package knowledge_knowledgebase

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKnowledgeKnowledgebaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	knowledgebaseProxy := GetKnowledgebaseProxy(sdkConfig)

	name := d.Get("name").(string)
	coreLanguage := d.Get("core_language").(string)

	// Find first non-deleted knowledge base by name. Retry in case new knowledge base is not yet indexed by search
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		publishedKnowledgeBases, publishedResp, getPublishedErr := knowledgebaseProxy.getAllKnowledgebaseEntities(ctx, true)
		unpublishedKnowledgeBases, unpublishedResp, getUnpublishedErr := knowledgebaseProxy.getAllKnowledgebaseEntities(ctx, false)

		if getPublishedErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("error requesting knowledge base %s | error: %s", name, getPublishedErr), publishedResp))
		}
		if getUnpublishedErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("error requesting knowledge base %s | error: %s", name, getUnpublishedErr), unpublishedResp))
		}

		noPublishedEntities := publishedKnowledgeBases == nil || len(*publishedKnowledgeBases) == 0
		noUnpublishedEntities := unpublishedKnowledgeBases == nil || len(*unpublishedKnowledgeBases) == 0
		if noPublishedEntities && noUnpublishedEntities {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("no knowledge bases found with name %s", name), publishedResp))
		}

		// prefer published knowledge base
		for _, knowledgeBase := range *publishedKnowledgeBases {
			if knowledgeBase.Name != nil && *knowledgeBase.Name == name &&
				*knowledgeBase.CoreLanguage == coreLanguage {
				d.SetId(*knowledgeBase.Id)
				return nil
			}
		}
		// use unpublished knowledge base if unpublished doesn't exist
		for _, knowledgeBase := range *unpublishedKnowledgeBases {
			if knowledgeBase.Name != nil && *knowledgeBase.Name == name &&
				*knowledgeBase.CoreLanguage == coreLanguage {
				d.SetId(*knowledgeBase.Id)
				return nil
			}
		}

		return nil
	})
}
