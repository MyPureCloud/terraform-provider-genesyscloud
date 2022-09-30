package genesyscloud

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
)

func dataSourceKnowledgeKnowledgebase() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Knowledge Base. Select a knowledge base by name.",
		ReadContext: readWithPooledClient(dataSourceKnowledgeKnowledgebaseRead),
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
	// if running in test env, sleep to allow resource records to persist
	if strings.EqualFold(os.Getenv("TF_ACC"), "1") {
		time.Sleep(8 * time.Second)
	}
	sdkConfig := m.(*providerMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	name := d.Get("name").(string)
	coreLanguage := d.Get("core_language").(string)

	// Find first non-deleted knowledge base by name. Retry in case new knowledge base is not yet indexed by search
	return withRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			knowledgeBases, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebases("", "", "", fmt.Sprintf("%v", pageSize), name, coreLanguage, false, "", "")
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("error requesting knowledge base %s: %s", name, getErr))
			}

			if knowledgeBases.Entities == nil || len(*knowledgeBases.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("no knowledge bases found with name %s", name))
			}

			for _, knowledgeBase := range *knowledgeBases.Entities {
				if knowledgeBase.Name != nil && *knowledgeBase.Name == name &&
					*knowledgeBase.CoreLanguage == coreLanguage {
					d.SetId(*knowledgeBase.Id)
					return nil
				}
			}
		}
	})
}
