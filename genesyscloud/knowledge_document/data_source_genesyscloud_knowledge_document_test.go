package knowledge_document

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	knowledgeBases "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccDataSourceKnowledgeDocumentBasic(t *testing.T) {
	var (
		knowledgeBaseResourceLabel = "test-knowledgebase1"
		knowledgeBaseName          = "Test-Terraform-Knowledge-Base" + uuid.NewString()
		knowledgeBaseDescription   = "test-knowledgebase-description1"
		coreLanguage               = "en-US"

		categoryResourceLabel = "test-category1"
		categoryName          = "Terraform Knowledge Category " + uuid.NewString()
		categoryDescription   = "test-knowledge-category-description1"

		labelResourceLabel = "test-label1"
		labelName          = "Terraform Knowledge Label " + uuid.NewString()
		labelColor         = "#0F0F0F"

		documentResourceLabel = "test-knowledge-document1"
		title                 = "Terraform Knowledge Document " + uuid.NewString()
		visible               = true
		published             = false
		phrase                = "Terraform Knowledge Document"
		autocomplete          = true

		dataSourceLabel        = "test-knowledge-document-ds"
		dataSourceWithCatLabel = "test-knowledge-document-ds-cat"
	)

	baseConfig := knowledgeBases.GenerateKnowledgeKnowledgebaseResource(
		knowledgeBaseResourceLabel,
		knowledgeBaseName,
		knowledgeBaseDescription,
		coreLanguage,
	) + generateKnowledgeCategoryResource(
		categoryResourceLabel,
		knowledgeBaseResourceLabel,
		categoryName,
		categoryDescription,
	) + generateKnowledgeLabelResource(
		labelResourceLabel,
		knowledgeBaseResourceLabel,
		labelName,
		labelColor,
	) + generateKnowledgeDocumentResource(
		documentResourceLabel,
		knowledgeBaseResourceLabel,
		categoryResourceLabel,
		labelResourceLabel,
		categoryName,
		labelName,
		title,
		visible,
		published,
		phrase,
		autocomplete,
	)

	dependsOn := ResourceType + "." + documentResourceLabel + ", genesyscloud_knowledge_knowledgebase." + knowledgeBaseResourceLabel

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		CheckDestroy:      testVerifyKnowledgeDocumentDestroyed,
		Steps: []resource.TestStep{
			{
				// Lookup by title only
				Config: baseConfig + generateKnowledgeDocumentDataSource(
					dataSourceLabel,
					title,
					knowledgeBaseName,
					"",
					dependsOn,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data."+DataSourceType+"."+dataSourceLabel, "id",
						ResourceType+"."+documentResourceLabel, "id",
					),
				),
			},
			{
				// Lookup by title and category name
				Config: baseConfig + generateKnowledgeDocumentDataSource(
					dataSourceWithCatLabel,
					title,
					knowledgeBaseName,
					categoryName,
					dependsOn,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data."+DataSourceType+"."+dataSourceWithCatLabel, "id",
						ResourceType+"."+documentResourceLabel, "id",
					),
					func(s *terraform.State) error {
						time.Sleep(20 * time.Second) // Wait for proper deletion of knowledgebase
						return nil
					},
				),
			},
		},
	})
}

func generateKnowledgeDocumentDataSource(resourceLabel string, title string, knowledgeBaseName string, categoryName string, dependsOn string) string {
	categoryAttr := ""
	if categoryName != "" {
		categoryAttr = fmt.Sprintf(`category_name = "%s"`, categoryName)
	}
	return fmt.Sprintf(`data "genesyscloud_knowledge_document" "%s" {
		title               = "%s"
		knowledge_base_name = "%s"
		%s
		depends_on          = [%s]
	}
	`, resourceLabel, title, knowledgeBaseName, categoryAttr, dependsOn)
}
