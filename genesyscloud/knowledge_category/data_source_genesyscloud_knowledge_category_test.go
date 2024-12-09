package knowledge_category

import (
	"fmt"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceKnowledgeCategoryBasic(t *testing.T) {
	var (
		knowledgeBaseResourceLabel1 = "test-knowledgebase1"
		categoryResourceLabel1      = "test-category1"
		categoryName                = "Terraform Test Category 1-" + uuid.NewString()
		categoryDescription         = "category description"
		knowledgeBaseName1          = "Terraform Test Knowledge Base 1-" + uuid.NewString()
		knowledgeBaseDescription1   = "test-knowledgebase-description1"
		knowledgeBaseCoreLanguage1  = "en-US"

		categoryDataSourceLabel = "test-category-ds"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: gcloud.GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResourceLabel1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				) + generateKnowledgeCategoryResource(
					categoryResourceLabel1,
					knowledgeBaseResourceLabel1,
					categoryName,
					categoryDescription,
				) + generateKnowledgeCategoryDataSource(
					categoryDataSourceLabel,
					categoryName,
					knowledgeBaseName1,
					"genesyscloud_knowledge_category."+categoryResourceLabel1+", genesyscloud_knowledge_knowledgebase."+knowledgeBaseResourceLabel1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_knowledge_category."+categoryDataSourceLabel,
						"id", "genesyscloud_knowledge_category."+categoryResourceLabel1, "id",
					),
				),
			},
		},
	})
}

func generateKnowledgeCategoryDataSource(
	resourceLabel string,
	name string,
	knowledgeBaseName string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOn string,
) string {
	return fmt.Sprintf(`data "genesyscloud_knowledge_category" "%s" {
		name = "%s"
        knowledge_base_name = "%s"
        depends_on=[%s]
	}
	`, resourceLabel, name, knowledgeBaseName, dependsOn)
}
