package genesyscloud

// import (
// 	"fmt"
// 	"strings"
// 	"testing"

// 	"github.com/google/uuid"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
// 	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
// )

// func TestAccResourceKnowledgeV1CategoryBasic(t *testing.T) {
// 	t.Skip("Skipping v1 knowledge tests since the test org is using v2")
// 	var (
// 		knowledgeBaseResource1     = "test-knowledgebase1"
// 		knowledgeBaseName1         = "Terraform Knowledge Base" + uuid.NewString()
// 		knowledgeBaseDescription1  = "test-knowledgebase-description1"
// 		knowledgeBaseCoreLanguage1 = "en-US"
// 		knowledgeCategoryResource1 = "test-knowledge-category1"
// 		categoryName               = "Terraform Knowledge Category" + uuid.NewString()
// 		categoryDescription        = "test-description1"
// 		categoryDescription2       = "test-description2"
// 	)

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:          func() { TestAccPreCheck(t) },
// 		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
// 		Steps: []resource.TestStep{
// 			{
// 				// Create
// 				Config: GenerateKnowledgeKnowledgebaseResource(
// 					knowledgeBaseResource1,
// 					knowledgeBaseName1,
// 					knowledgeBaseDescription1,
// 					knowledgeBaseCoreLanguage1,
// 				) +
// 					generateKnowledgeV1Category(
// 						knowledgeCategoryResource1,
// 						knowledgeBaseResource1,
// 						knowledgeBaseCoreLanguage1,
// 						categoryName,
// 						categoryDescription,
// 					),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttrPair("genesyscloud_knowledge_v1_category."+knowledgeCategoryResource1, "knowledge_base_id", "genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "id"),
// 					resource.TestCheckResourceAttr("genesyscloud_knowledge_v1_category."+knowledgeCategoryResource1, "language_code", knowledgeBaseCoreLanguage1),
// 					resource.TestCheckResourceAttr("genesyscloud_knowledge_v1_category."+knowledgeCategoryResource1, "knowledge_category.0.name", categoryName),
// 					resource.TestCheckResourceAttr("genesyscloud_knowledge_v1_category."+knowledgeCategoryResource1, "knowledge_category.0.description", categoryDescription),
// 				),
// 			},
// 			{
// 				// Update
// 				Config: GenerateKnowledgeKnowledgebaseResource(
// 					knowledgeBaseResource1,
// 					knowledgeBaseName1,
// 					knowledgeBaseDescription1,
// 					knowledgeBaseCoreLanguage1,
// 				) +
// 					generateKnowledgeV1Category(
// 						knowledgeCategoryResource1,
// 						knowledgeBaseResource1,
// 						knowledgeBaseCoreLanguage1,
// 						categoryName,
// 						categoryDescription2,
// 					),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttrPair("genesyscloud_knowledge_v1_category."+knowledgeCategoryResource1, "knowledge_base_id", "genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "id"),
// 					resource.TestCheckResourceAttr("genesyscloud_knowledge_v1_category."+knowledgeCategoryResource1, "language_code", knowledgeBaseCoreLanguage1),
// 					resource.TestCheckResourceAttr("genesyscloud_knowledge_v1_category."+knowledgeCategoryResource1, "knowledge_category.0.name", categoryName),
// 					resource.TestCheckResourceAttr("genesyscloud_knowledge_v1_category."+knowledgeCategoryResource1, "knowledge_category.0.description", categoryDescription2),
// 				),
// 			},
// 			{
// 				// Import/Read
// 				ResourceName:      "genesyscloud_knowledge_v1_category." + knowledgeCategoryResource1,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 			},
// 		},
// 		CheckDestroy: testVerifyKnowledgeV1CategoryDestroyed,
// 	})
// }

// func generateKnowledgeV1Category(resourceName string, knowledgeBaseResource string, languageCode string, categoryName string, categoryDescription string) string {
// 	category := fmt.Sprintf(`
//         resource "genesyscloud_knowledge_v1_category" "%s" {
//             knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
//             language_code = "%s"
//             %s
//         }
//         `, resourceName,
// 		knowledgeBaseResource,
// 		languageCode,
// 		generateKnowledgeV1CategoryRequestBody(categoryName, categoryDescription),
// 	)
// 	return category
// }

// func generateKnowledgeV1CategoryRequestBody(categoryName string, categoryDescription string) string {

// 	return fmt.Sprintf(`
//         knowledge_category {
//             name = "%s"
//             description = "%s"
//         }
//         `, categoryName,
// 		categoryDescription,
// 	)
// }

// func testVerifyKnowledgeV1CategoryDestroyed(state *terraform.State) error {
// 	knowledgeAPI := platformclientv2.NewKnowledgeApi()
// 	knowledgeBaseCoreLanguage1 := "en-US"
// 	var knowledgeBaseId string
// 	for _, rs := range state.RootModule().Resources {
// 		if rs.Type == "genesyscloud_knowledge_knowledgebase" {
// 			knowledgeBaseId = rs.Primary.ID
// 			break
// 		}
// 	}
// 	for _, rs := range state.RootModule().Resources {
// 		if rs.Type != "genesyscloud_knowledge_v1_category" {
// 			continue
// 		}
// 		id := strings.Split(rs.Primary.ID, " ")
// 		knowledgeCategoryId := id[0]
// 		knowledgeCategory, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, knowledgeBaseCoreLanguage1)
// 		if knowledgeCategory != nil {
// 			return fmt.Errorf("Knowledge category (%s) still exists", knowledgeCategoryId)
// 		} else if IsStatus404(resp) || IsStatus400(resp) {
// 			// Knowledge base category not found as expected
// 			continue
// 		} else {
// 			// Unexpected error
// 			return fmt.Errorf("Unexpected error: %s", err)
// 		}
// 	}
// 	// Success. All knowledge categories destroyed
// 	return nil
// }
