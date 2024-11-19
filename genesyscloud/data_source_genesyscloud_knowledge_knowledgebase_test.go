package genesyscloud

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceKnowledgeKnowledgebaseBasic(t *testing.T) {
	var (
		knowledgeBaseResource1     = "test-knowledgebase1"
		knowledgeBaseName1         = "Terraform Test Knowledge Base 1-" + uuid.NewString()
		knowledgeBaseDescription1  = "test-knowledgebase-description1"
		knowledgeBaseCoreLanguage1 = "en-US"

		knowledgeBaseDataSource = "test-knowledgebase-ds"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResource1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				) + generateKnowledgeKnowledgebaseDataSource(
					knowledgeBaseDataSource,
					"genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1+".name",
					"genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1+".core_language",
					"genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_knowledge_knowledgebase."+knowledgeBaseDataSource,
						"id", "genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "id",
					),
				),
			},
		},
	})
}

func generateKnowledgeKnowledgebaseDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	coreLanguage string,
	dependsOn string,
) string {
	return fmt.Sprintf(`data "genesyscloud_knowledge_knowledgebase" "%s" {
		name = %s
        core_language = %s
        depends_on=[%s]
	}
	`, resourceID, name, coreLanguage, dependsOn)
}
