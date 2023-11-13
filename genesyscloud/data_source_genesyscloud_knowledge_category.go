package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func dataSourceKnowledgeCategory() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Knowledge Base Category. Select a category by name.",
		ReadContext: ReadWithPooledClient(dataSourceKnowledgeCategoryRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Knowledge base category name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"knowledge_base_name": {
				Description: "Knowledge base name",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceKnowledgeCategoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)
	knowledgeBaseName := d.Get("knowledge_base_name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageSize = 100
		publishedKnowledgeBases, _, getPublishedErr := knowledgeAPI.GetKnowledgeKnowledgebases("", "", "", fmt.Sprintf("%v", pageSize), knowledgeBaseName, "", true, "", "")
		unpublishedKnowledgeBases, _, getUnpublishedErr := knowledgeAPI.GetKnowledgeKnowledgebases("", "", "", fmt.Sprintf("%v", pageSize), knowledgeBaseName, "", false, "", "")

		if getPublishedErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to get knowledge base %s: %s", knowledgeBaseName, getPublishedErr))
		}
		if getUnpublishedErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to get knowledge base %s: %s", knowledgeBaseName, getUnpublishedErr))
		}

		noPublishedEntities := publishedKnowledgeBases.Entities == nil || len(*publishedKnowledgeBases.Entities) == 0
		noUnpublishedEntities := unpublishedKnowledgeBases.Entities == nil || len(*unpublishedKnowledgeBases.Entities) == 0
		if noPublishedEntities && noUnpublishedEntities {

			return retry.RetryableError(fmt.Errorf("no knowledge bases found with name %s", knowledgeBaseName))
		}

		// prefer published knowledge base
		for _, knowledgeBase := range *publishedKnowledgeBases.Entities {
			if knowledgeBase.Name != nil && *knowledgeBase.Name == knowledgeBaseName {
				knowledgeCategories, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseCategories(*knowledgeBase.Id, "", "", fmt.Sprintf("%v", pageSize), "", false, name, "", "", false)

				if getErr != nil {
					return retry.NonRetryableError(fmt.Errorf("Failed to get knowledge category %s: %s", name, getErr))
				}

				for _, knowledgeCategory := range *knowledgeCategories.Entities {
					if *knowledgeCategory.Name == name {
						id := fmt.Sprintf("%s,%s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id)
						d.SetId(id)
						return nil
					}
				}
			}
		}
		// use unpublished knowledge base if unpublished doesn't exist
		for _, knowledgeBase := range *unpublishedKnowledgeBases.Entities {
			if knowledgeBase.Name != nil && *knowledgeBase.Name == knowledgeBaseName {
				knowledgeCategories, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseCategories(*knowledgeBase.Id, "", "", fmt.Sprintf("%v", pageSize), "", false, name, "", "", false)

				if getErr != nil {
					return retry.NonRetryableError(fmt.Errorf("Failed to get knowledge category %s: %s", name, getErr))
				}

				for _, knowledgeCategory := range *knowledgeCategories.Entities {
					if *knowledgeCategory.Name == name {
						id := fmt.Sprintf("%s,%s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id)
						d.SetId(id)
						return nil
					}
				}
			}
		}
		return nil
	})
}
