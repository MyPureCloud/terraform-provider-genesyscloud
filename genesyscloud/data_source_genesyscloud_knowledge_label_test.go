package genesyscloud

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceKnowledgeLabelBasic(t *testing.T) {
	var (
		knowledgeBaseResource1     = "test-knowledgebase1"
		labelResource1             = "test-label1"
		labelName                  = "Terraform Test Label 1-" + uuid.NewString()
		labelColor                 = "#ffffff"
		knowledgeBaseName1         = "Terraform Test Knowledge Base 1-" + uuid.NewString()
		knowledgeBaseDescription1  = "test-knowledgebase-description1"
		knowledgeBaseCoreLanguage1 = "en-US"

		labelDataSource = "test-label-ds"
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
				) + generateKnowledgeLabelResource(
					labelResource1,
					knowledgeBaseResource1,
					labelName,
					labelColor,
				) + generateKnowledgeLabelDataSource(
					labelDataSource,
					labelName,
					knowledgeBaseName1,
					"genesyscloud_knowledge_label."+labelResource1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_knowledge_label."+labelDataSource,
						"id", "genesyscloud_knowledge_label."+labelResource1, "id",
					),
				),
			},
		},
	})
}

func generateKnowledgeLabelDataSource(
	resourceID string,
	name string,
	knowledgeBaseName string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOn string,
) string {
	return fmt.Sprintf(`data "genesyscloud_knowledge_label" "%s" {
		name = "%s"
		knowledge_base_name = "%s"
        depends_on=[%s]
	}
	`, resourceID, name, knowledgeBaseName, dependsOn)
}
