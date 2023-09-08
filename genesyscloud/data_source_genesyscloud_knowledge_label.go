package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func dataSourceKnowledgeLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Knowledge Base Label. Select a label by name.",
		ReadContext: ReadWithPooledClient(dataSourceKnowledgeLabelRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Knowledge base label name",
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

func dataSourceKnowledgeLabelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
				knowledgeLabels, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLabels(*knowledgeBase.Id, "", "", fmt.Sprintf("%v", pageSize), name, false)

				if getErr != nil {
					return retry.NonRetryableError(fmt.Errorf("Failed to get knowledge label %s: %s", name, getErr))
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
				knowledgeLabels, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLabels(*knowledgeBase.Id, "", "", fmt.Sprintf("%v", pageSize), name, false)

				if getErr != nil {
					return retry.NonRetryableError(fmt.Errorf("Failed to get knowledge label %s: %s", name, getErr))
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
