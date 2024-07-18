package genesyscloud

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceKnowledgeDocumentBasic(t *testing.T) {
	var (
		knowledgeBaseResource1     = "test-knowledgebase1"
		categoryResource1          = "test-category1"
		categoryName               = "Terraform Knowledge Category " + uuid.NewString()
		categoryDescription        = "test-knowledge-category-description1"
		labelResource1             = "test-label1"
		labelName                  = "Terraform Knowledge Label " + uuid.NewString()
		labelColor                 = "#0F0F0F"
		knowledgeBaseName1         = "Terraform Knowledge Base " + uuid.NewString()
		knowledgeBaseDescription1  = "test-knowledgebase-description1"
		coreLanguage1              = "en-US"
		knowledgeDocumentResource1 = "test-knowledge-document1"
		title                      = "Terraform Knowledge Document"
		visible                    = true
		visible2                   = false
		published                  = false
		phrase                     = "Terraform Knowledge Document"
		autocomplete               = true
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
					coreLanguage1,
				) +
					generateKnowledgeCategoryResource(
						categoryResource1,
						knowledgeBaseResource1,
						categoryName,
						categoryDescription,
					) +
					generateKnowledgeLabelResource(
						labelResource1,
						knowledgeBaseResource1,
						labelName,
						labelColor,
					) +
					generateKnowledgeDocumentResource(
						knowledgeDocumentResource1,
						knowledgeBaseResource1,
						categoryResource1,
						labelResource1,
						categoryName,
						labelName,
						title,
						visible,
						published,
						phrase,
						autocomplete,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.title", title),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.visible", fmt.Sprintf("%v", visible)),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.alternatives.0.phrase", phrase),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.alternatives.0.autocomplete", fmt.Sprintf("%v", autocomplete)),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.label_names.0", labelName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.category_name", categoryName),
				),
			},
			{
				// Update
				Config: GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResource1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					coreLanguage1,
				) +
					generateKnowledgeCategoryResource(
						categoryResource1,
						knowledgeBaseResource1,
						categoryName,
						categoryDescription,
					) +
					generateKnowledgeLabelResource(
						labelResource1,
						knowledgeBaseResource1,
						labelName,
						labelColor,
					) +
					generateKnowledgeDocumentResource(
						knowledgeDocumentResource1,
						knowledgeBaseResource1,
						categoryResource1,
						labelResource1,
						categoryName,
						labelName,
						title,
						visible2,
						published,
						phrase,
						autocomplete,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.title", title),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.visible", fmt.Sprintf("%v", visible2)),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.alternatives.0.phrase", phrase),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.alternatives.0.autocomplete", fmt.Sprintf("%v", autocomplete)),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.category_name", categoryName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResource1, "knowledge_document.0.label_names.0", labelName),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_knowledge_document." + knowledgeDocumentResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgeDocumentDestroyed,
	})
}

func generateKnowledgeDocumentResource(resourceName string, knowledgeBaseResourceName string, knowledgeCategoryResourceName string, knowledgeLabelResourceName string, knowledgeCategoryName string, knowledgeLabelName string, title string, visible bool, published bool, phrase string, autocomplete bool) string {
	document := fmt.Sprintf(`
        resource "genesyscloud_knowledge_document" "%s" {
			depends_on=[genesyscloud_knowledge_category.%s, genesyscloud_knowledge_label.%s]
            knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
            published = %v
            %s
        }
        `, resourceName,
		knowledgeCategoryResourceName,
		knowledgeLabelResourceName,
		knowledgeBaseResourceName,
		published,
		generateKnowledgeDocumentRequestBody(knowledgeCategoryName, knowledgeLabelName, title, visible, phrase, autocomplete),
	)
	return document
}

func generateKnowledgeDocumentBasic(resourceName string, knowledgeBaseResourceName string, title string, visible bool, published bool, phrase string, autocomplete bool) string {
	document := fmt.Sprintf(`
        resource "genesyscloud_knowledge_document" "%s" {
            knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
            published = %v
            %s
        }
        `, resourceName,
		knowledgeBaseResourceName,
		published,
		generateKnowledgeDocumentRequestBodyBasic(title, visible, phrase, autocomplete),
	)
	return document
}

func generateKnowledgeDocumentRequestBodyBasic(title string, visible bool, phrase string, autocomplete bool) string {

	documentRequestBody := fmt.Sprintf(`
        knowledge_document {
			title = "%s"
			visible = %v
			%s
		}
        `, title,
		visible,
		generateKnowledgeDocumentAlternatives(phrase, autocomplete),
	)
	return documentRequestBody
}

func generateKnowledgeDocumentRequestBody(knowledgeCategoryName string, knowledgeLabelName string, title string, visible bool, phrase string, autocomplete bool) string {

	documentRequestBody := fmt.Sprintf(`
        knowledge_document {
			title = "%s"
			visible = %v
			%s
			category_name = "%s"
			label_names = ["%s"]
		}
        `, title,
		visible,
		generateKnowledgeDocumentAlternatives(phrase, autocomplete),
		knowledgeCategoryName,
		knowledgeLabelName,
	)
	return documentRequestBody
}

func generateKnowledgeDocumentAlternatives(phrase string, autocomplete bool) string {
	alternatives := fmt.Sprintf(`
        alternatives {
			phrase = "%s"
			autocomplete = %v
		}
        `, phrase,
		autocomplete,
	)
	return alternatives
}

func testVerifyKnowledgeDocumentDestroyed(state *terraform.State) error {
	knowledgeAPI := platformclientv2.NewKnowledgeApi()
	var knowledgeBaseId string
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_knowledge_knowledgebase" {
			knowledgeBaseId = rs.Primary.ID
			break
		}
	}
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_knowledge_document" {
			continue
		}
		id := strings.Split(rs.Primary.ID, " ")
		knowledgeDocumentId := id[0]
		knowledgeDocument, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseDocument(knowledgeBaseId, knowledgeDocumentId, nil, "")
		if knowledgeDocument != nil {
			return fmt.Errorf("Knowledge document (%s) still exists", knowledgeDocumentId)
		} else if util.IsStatus404(resp) || util.IsStatus400(resp) {
			// Knowledge base document not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All knowledge base documents destroyed
	return nil
}
