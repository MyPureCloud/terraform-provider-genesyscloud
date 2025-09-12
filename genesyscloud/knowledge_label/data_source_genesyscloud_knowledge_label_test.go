package knowledge_label

import (
	"fmt"
	"testing"

	knowledgeKnowledgebase "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceKnowledgeLabelBasic(t *testing.T) {
	var (
		knowledgeBaseResourceLabel1 = "test-knowledgebase1"
		labelResourceLabel1         = "test-label1"
		labelResourceFullPath       = ResourceType + "." + labelResourceLabel1

		labelName                  = "Terraform Test Label 1-" + uuid.NewString()
		labelColor                 = "#ffffff"
		knowledgeBaseName1         = "Test-Terraform-Knowledge-Base" + uuid.NewString()
		knowledgeBaseDescription1  = "test-knowledgebase-description1"
		knowledgeBaseCoreLanguage1 = "en-US"

		labelDataSourceLabel    = "test-label-ds"
		labelDataSourceFullPath = "data." + ResourceType + "." + labelDataSourceLabel
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: knowledgeKnowledgebase.GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResourceLabel1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				) + generateKnowledgeLabelResource(
					labelResourceLabel1,
					knowledgeBaseResourceLabel1,
					labelName,
					labelColor,
				) + generateKnowledgeLabelDataSource(
					labelDataSourceLabel,
					labelName,
					knowledgeBaseName1,
					labelResourceFullPath,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						labelDataSourceFullPath, "id",
						labelResourceFullPath, "id",
					),
				),
			},
		},
	})
}

func generateKnowledgeLabelDataSource(
	resourceLabel,
	name,
	knowledgeBaseName,
	dependsOn string,
) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		knowledge_base_name = "%s"
        depends_on=[%s]
	}
	`, ResourceType, resourceLabel, name, knowledgeBaseName, dependsOn)
}
