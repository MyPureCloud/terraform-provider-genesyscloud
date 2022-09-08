package genesyscloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
)

func TestAccResourceKnowledgeCategoryBasic(t *testing.T) {
	var (
		knowledgeBaseResource1     = "test-knowledgebase1"
		knowledgeBaseName1         = "Terraform Knowledge Base" + uuid.NewString()
		knowledgeBaseDescription1  = "test-knowledgebase-description1"
		knowledgeBaseCoreLanguage1 = "en-US"
		knowledgeCategoryResource1 = "test-knowledge-category1"
		categoryName               = "Terraform Knowledge Category" + uuid.NewString()
		categoryDescription        = "test-description1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateKnowledgeKnowledgebaseResource(
					knowledgeBaseResource1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				) +
					generateKnowledgeCategory(
						knowledgeCategoryResource1,
						knowledgeBaseResource1,
						knowledgeBaseCoreLanguage1,
						categoryName,
						categoryDescription,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResource1, "knowledge_category.0.name", categoryName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_category."+knowledgeCategoryResource1, "knowledge_category.0.description", categoryDescription),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_knowledge_category." + knowledgeCategoryResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgeCategoryDestroyed,
	})
}

func generateKnowledgeCategory(resourceName string, knowledgeBaseResource string, languageCode string, categoryName string, categoryDescription string) string {
	category := fmt.Sprintf(`
        resource "genesyscloud_knowledge_category" "%s" {
            knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
            language_code = "%s"
            %s
        }
        `, resourceName,
		knowledgeBaseResource,
		languageCode,
		generateKnowledgeCategoryRequestBody(categoryName, categoryDescription),
	)
	return category
}

func generateKnowledgeCategoryRequestBody(categoryName string, categoryDescription string) string {

	return fmt.Sprintf(`
        knowledge_category {
            name = "%s"
            description = "%s"
        }
        `, categoryName,
		categoryDescription,
	)
}

func testVerifyKnowledgeCategoryDestroyed(state *terraform.State) error {
	knowledgeAPI := platformclientv2.NewKnowledgeApi()
	knowledgeBaseCoreLanguage1 := "en-US"
	var knowledgeBaseId string
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_knowledge_knowledgebase" {
			knowledgeBaseId = rs.Primary.ID
			break
		}
	}
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_knowledge_category" {
			continue
		}
		id := strings.Split(rs.Primary.ID, " ")
		knowledgeCategoryId := id[0]
		knowledgeCategory, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, knowledgeBaseCoreLanguage1)
		if knowledgeCategory != nil {
			return fmt.Errorf("Knowledge category (%s) still exists", knowledgeCategoryId)
		} else if isStatus404(resp) || isStatus400(resp) {
			// Knowledge base not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All knowledge categories destroyed
	return nil
}
