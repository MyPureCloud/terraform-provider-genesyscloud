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

// func TestAccResourceKnowledgeV1DocumentBasic(t *testing.T) {
// t.Skip("Skipping v1 knowledge tests since the test org is using v2")
// var (
// 	knowledgeBaseResource1       = "test-knowledgebase1"
// 	knowledgeCategoryResource1   = "test-category1"
// 	knowledgeCategoryName        = "Terraform Knowledge Category " + uuid.NewString()
// 	knowledgeCategoryDescription = "test-knowledge-category-description1"
// 	knowledgeBaseName1           = "Terraform Knowledge Base " + uuid.NewString()
// 	knowledgeBaseDescription1    = "test-knowledgebase-description1"
// 	knowledgeBaseCoreLanguage1   = "en-US"
// 	knowledgeDocumentResource1   = "test-knowledge-document1"
// 	varType1                     = "Faq"
// 	externalUrl                  = "http://example.com"
// 	question                     = "What should I ask?"
// 	answer                       = "I don't know"
// 	faqAlternatives              = []string{"\"faq-alt-1\"", "\"faq-alt-2\""}
// 	faqAlternatives2             = []string{"\"faq-alt-3\"", "\"faq-alt-4\""}
// 	title                        = "test-document-title1"
// 	title2                       = "test-document-title2"
// 	contentLocationUrl           = "http://example.com"
// 	articleAlternatives          = []string{"\"alt1\"", "\"alt2\""}
// )

// resource.Test(t, resource.TestCase{
// 	PreCheck:          func() { TestAccPreCheck(t) },
// 	ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
// 	Steps: []resource.TestStep{
// 		{
// 			// Create
// 			Config: GenerateKnowledgeKnowledgebaseResource(
// 				knowledgeBaseResource1,
// 				knowledgeBaseName1,
// 				knowledgeBaseDescription1,
// 				knowledgeBaseCoreLanguage1,
// 			) +
// 				generateKnowledgeV1Category(
// 					knowledgeCategoryResource1,
// 					knowledgeBaseResource1,
// 					knowledgeBaseCoreLanguage1,
// 					knowledgeCategoryName,
// 					knowledgeCategoryDescription,
// 				) +
// 				generateKnowledgeV1Document(
// 					knowledgeDocumentResource1,
// 					knowledgeCategoryResource1,
// 					knowledgeBaseResource1,
// 					knowledgeBaseCoreLanguage1,
// 					varType1,
// 					externalUrl,
// 					question,
// 					answer,
// 					faqAlternatives,
// 					title,
// 					contentLocationUrl,
// 					articleAlternatives,
// 				),
// 			Check: resource.ComposeTestCheckFunc(
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.type", varType1),
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.external_url", externalUrl),
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.faq.0.question", question),
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.faq.0.answer", answer),
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.faq.0.alternatives.0", faqAlternatives[0]),
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.categories.0", knowledgeCategoryName),
// 			),
// 		},
// 		{
// 			// Update
// 			Config: GenerateKnowledgeKnowledgebaseResource(
// 				knowledgeBaseResource1,
// 				knowledgeBaseName1,
// 				knowledgeBaseDescription1,
// 				knowledgeBaseCoreLanguage1,
// 			) +
// 				generateKnowledgeV1Category(
// 					knowledgeCategoryResource1,
// 					knowledgeBaseResource1,
// 					knowledgeBaseCoreLanguage1,
// 					knowledgeCategoryName,
// 					knowledgeCategoryDescription,
// 				) +
// 				generateKnowledgeV1Document(
// 					knowledgeDocumentResource1,
// 					knowledgeCategoryResource1,
// 					knowledgeBaseResource1,
// 					knowledgeBaseCoreLanguage1,
// 					varType1,
// 					externalUrl,
// 					question,
// 					answer,
// 					faqAlternatives2,
// 					title2,
// 					contentLocationUrl,
// 					articleAlternatives,
// 				),
// 			Check: resource.ComposeTestCheckFunc(
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.type", varType1),
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.external_url", externalUrl),
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.faq.0.question", question),
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.faq.0.answer", answer),
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.faq.0.alternatives.0", faqAlternatives2[0]),
// 				resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.categories.0", knowledgeCategoryName),
// 			),
// 		},
// 		{
// 			// Import/Read
// 			ResourceName:      "genesyscloud_knowledge_document." + knowledgeDocumentResource1,
// 			ImportState:       true,
// 			ImportStateVerify: true,
// 		},
// 	},
// 	CheckDestroy: testVerifyKnowledgeV1DocumentDestroyed,
// })
// }

// func generateKnowledgeV1Document(resourceName string, knowledgeCategoryResourceName string, knowledgeBaseResourceName string, languageCode string, varType string, externalUrl string, question string, answer string, faqAlternatives []string, title string, contentLocationUrl string, articleAlternatives []string) string {
// 	document := fmt.Sprintf(`
//         resource "genesyscloud_knowledge_v1_document" "%s" {
//             knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
//             language_code = "%s"
//             %s
//         }
//         `, resourceName,
// 		knowledgeBaseResourceName,
// 		languageCode,
// 		generateKnowledgeV1DocumentRequestBodyFaq(varType, externalUrl, question, answer, faqAlternatives, knowledgeCategoryResourceName),
// 	)
// 	return document
// }

// func generateKnowledgeV1DocumentRequestBodyFaq(varType string, externalUrl string, question string, answer string, faqAlternatives []string, knowledgeCategoryResourceName string) string {
// 	return fmt.Sprintf(`
//         knowledge_document {
//             type = "%s"
//             external_url = "%s"
//             %s
//             categories = [genesyscloud_knowledge_category.%s.knowledge_category.0.name]
//         }
//         `, varType,
// 		externalUrl,
// 		generateFaq(question, answer, faqAlternatives),
// 		knowledgeCategoryResourceName)
// }

// func generateKnowledgeV1DocumentRequestBodyArticle(varType string, externalUrl string, title string, contentLocationUrl string, alternatives []string, knowledgeCategoryResourceName []string) string {
// 	return fmt.Sprintf(`
//         knowledge_document {
//             type = "%s"
//             external_url = "%s"
//             %s
//             categories = [genesyscloud_knowledge_category.%s.knowledge_category.0.name]
//         }
//         `, varType,
// 		externalUrl,
// 		generateArticle(title, contentLocationUrl, alternatives),
// 		knowledgeCategoryResourceName)
// }

// func generateFaq(question string, answer string, alternatives []string) string {

// 	return fmt.Sprintf(`
//         faq {
//             question = "%s"
//             answer = "%s"
//             alternatives = [%s]
//         }
//         `, question,
// 		answer,
// 		alternatives)
// }

// func generateArticle(title string, contentLocationUrl string, alternatives []string) string {

// 	return fmt.Sprintf(`
//         article {
//             title = "%s"
//             content_location_url = "%s"
//             alternatives = [%s]
//         }
//         `, title,
// 		contentLocationUrl,
// 		alternatives)
// }

// func testVerifyKnowledgeV1DocumentDestroyed(state *terraform.State) error {
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
// 		if rs.Type != "genesyscloud_knowledge_v1_document" {
// 			continue
// 		}
// 		id := strings.Split(rs.Primary.ID, " ")
// 		knowledgeDocumentId := id[0]
// 		knowledgeDocument, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageDocument(knowledgeDocumentId, knowledgeBaseId, knowledgeBaseCoreLanguage1)
// 		if knowledgeDocument != nil {
// 			return fmt.Errorf("Knowledge document (%s) still exists", knowledgeDocumentId)
// 		} else if IsStatus404(resp) || IsStatus400(resp) {
// 			// Knowledge base document not found as expected
// 			continue
// 		} else {
// 			// Unexpected error
// 			return fmt.Errorf("Unexpected error: %s", err)
// 		}
// 	}
// 	// Success. All knowledge base documents destroyed
// 	return nil
// }
