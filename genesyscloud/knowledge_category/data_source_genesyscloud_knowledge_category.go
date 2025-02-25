package knowledge_category

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKnowledgeCategory() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Knowledge Base Category. Select a category by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceKnowledgeCategoryRead),
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
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeCategoryProxy(sdkConfig)
	name := d.Get("name").(string)
	knowledgeBaseName := d.Get("knowledge_base_name").(string)

	return util.WithRetries(ctx, 10*time.Second, func() *retry.RetryError {
		knowledgeCategoryId, retryable, resp, getErr := proxy.getKnowledgeCategoryByName(ctx, name, knowledgeBaseName)

		if getErr != nil {
			errMsg := util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Failed to get knowledge category by name %s | error: %s", name, getErr), resp)
			if !retryable {
				return retry.NonRetryableError(errMsg)
			}
			return retry.RetryableError(errMsg)

		}

		d.SetId(knowledgeCategoryId)

		return nil
	})
}
