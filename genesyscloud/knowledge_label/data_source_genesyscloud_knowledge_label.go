package knowledge_label

import (
	"context"
	"fmt"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func dataSourceKnowledgeLabelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)
	knowledgeBaseName := d.Get("knowledge_base_name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageSize = 100
		publishedKnowledgeBases, publishedResp, getPublishedErr := knowledgeAPI.GetKnowledgeKnowledgebases("", "", "", fmt.Sprintf("%v", pageSize), knowledgeBaseName, "", true, "", "")
		unpublishedKnowledgeBases, unpublishedResp, getUnpublishedErr := knowledgeAPI.GetKnowledgeKnowledgebases("", "", "", fmt.Sprintf("%v", pageSize), knowledgeBaseName, "", false, "", "")

		if getPublishedErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to get knowledge base %s | error: %s", knowledgeBaseName, getPublishedErr), publishedResp))
		}
		if getUnpublishedErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to get knowledge base %s | error: %s", knowledgeBaseName, getUnpublishedErr), unpublishedResp))
		}

		noPublishedEntities := publishedKnowledgeBases.Entities == nil || len(*publishedKnowledgeBases.Entities) == 0
		noUnpublishedEntities := unpublishedKnowledgeBases.Entities == nil || len(*unpublishedKnowledgeBases.Entities) == 0
		if noPublishedEntities && noUnpublishedEntities {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("no knowledge bases found with name %s", knowledgeBaseName), publishedResp))
		}

		// prefer published knowledge base
		for _, knowledgeBase := range *publishedKnowledgeBases.Entities {
			if knowledgeBase.Name != nil && *knowledgeBase.Name == knowledgeBaseName {
				knowledgeLabels, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLabels(*knowledgeBase.Id, "", "", fmt.Sprintf("%v", pageSize), name, false)

				if getErr != nil {
					return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to get knowledge label %s | error: %s", name, getErr), resp))
				}

				for _, knowledgeLabel := range *knowledgeLabels.Entities {
					if *knowledgeLabel.Name == name {
						id := fmt.Sprintf("%s,%s", *knowledgeLabel.Id, *knowledgeBase.Id)
						d.SetId(id)
						return nil
					}
				}
			}
		}
		// use unpublished knowledge base if unpublished doesn't exist
		for _, knowledgeBase := range *unpublishedKnowledgeBases.Entities {
			if knowledgeBase.Name != nil && *knowledgeBase.Name == knowledgeBaseName {
				knowledgeLabels, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLabels(*knowledgeBase.Id, "", "", fmt.Sprintf("%v", pageSize), name, false)

				if getErr != nil {
					return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to get knowledge label %s | error: %s", name, getErr), resp))
				}

				for _, knowledgeLabel := range *knowledgeLabels.Entities {
					if *knowledgeLabel.Name == name {
						id := fmt.Sprintf("%s,%s", *knowledgeLabel.Id, *knowledgeBase.Id)
						d.SetId(id)
						return nil
					}
				}
			}
		}
		return nil
	})
}
