package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func dataSourceKnowledgeKnowledgebase() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Knowledge Base. Select a knowledge base by name.",
		ReadContext: ReadWithPooledClient(dataSourceKnowledgeKnowledgebaseRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Knowledge base name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"core_language": {
				Description:  "Core language for knowledge base in which initial content must be created, language codes [en-US, en-UK, en-AU, de-DE] are supported currently, however the new DX knowledge will support all these language codes",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"en-US", "en-UK", "en-AU", "de-DE", "es-US", "es-ES", "fr-FR", "pt-BR", "nl-NL", "it-IT", "fr-CA"}, false),
			},
		},
	}
}

func dataSourceKnowledgeKnowledgebaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)
	coreLanguage := d.Get("core_language").(string)

	// Find first non-deleted knowledge base by name. Retry in case new knowledge base is not yet indexed by search
	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageSize = 100
		publishedKnowledgeBases, _, getPublishedErr := knowledgeAPI.GetKnowledgeKnowledgebases("", "", "", fmt.Sprintf("%v", pageSize), name, coreLanguage, true, "", "")
		unpublishedKnowledgeBases, _, getUnpublishedErr := knowledgeAPI.GetKnowledgeKnowledgebases("", "", "", fmt.Sprintf("%v", pageSize), name, coreLanguage, false, "", "")

		if getPublishedErr != nil {
			return retry.NonRetryableError(fmt.Errorf("error requesting knowledge base %s: %s", name, getPublishedErr))
		}
		if getUnpublishedErr != nil {
			return retry.NonRetryableError(fmt.Errorf("error requesting knowledge base %s: %s", name, getUnpublishedErr))
		}

		noPublishedEntities := publishedKnowledgeBases.Entities == nil || len(*publishedKnowledgeBases.Entities) == 0
		noUnpublishedEntities := unpublishedKnowledgeBases.Entities == nil || len(*unpublishedKnowledgeBases.Entities) == 0
		if noPublishedEntities && noUnpublishedEntities {
			return retry.RetryableError(fmt.Errorf("no knowledge bases found with name %s", name))
		}

		// prefer published knowledge base
		for _, knowledgeBase := range *publishedKnowledgeBases.Entities {
			if knowledgeBase.Name != nil && *knowledgeBase.Name == name &&
				*knowledgeBase.CoreLanguage == coreLanguage {
				d.SetId(*knowledgeBase.Id)
				return nil
			}
		}
		// use unpublished knowledge base if unpublished doesn't exist
		for _, knowledgeBase := range *unpublishedKnowledgeBases.Entities {
			if knowledgeBase.Name != nil && *knowledgeBase.Name == name &&
				*knowledgeBase.CoreLanguage == coreLanguage {
				d.SetId(*knowledgeBase.Id)
				return nil
			}
		}

		return nil
	})
}
