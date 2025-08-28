package knowledge_knowledgebase

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceKnowledgeKnowledgebaseBasic(t *testing.T) {
	var (
		knowledgeBaseResourceLabel1 = "test-knowledgebase1"
		knowledgeBaseName1          = "Test-Terraform-Knowledge-Base" + uuid.NewString()
		knowledgeBaseDescription1   = "test-knowledgebase-description1"
		knowledgeBaseCoreLanguage1  = "en-US"

		knowledgeBaseDataSourceLabel = "test-knowledgebase-ds"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResourceLabel1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				) + generateKnowledgeKnowledgebaseDataSource(
					knowledgeBaseDataSourceLabel,
					ResourceType+"."+knowledgeBaseResourceLabel1+".name",
					ResourceType+"."+knowledgeBaseResourceLabel1+".core_language",
					ResourceType+"."+knowledgeBaseResourceLabel1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+knowledgeBaseDataSourceLabel,
						"id", ResourceType+"."+knowledgeBaseResourceLabel1, "id",
					),
				),
			},
		},
	})
}

func generateKnowledgeKnowledgebaseDataSource(
	resourceLabel string,
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
	`, resourceLabel, name, coreLanguage, dependsOn)
}
