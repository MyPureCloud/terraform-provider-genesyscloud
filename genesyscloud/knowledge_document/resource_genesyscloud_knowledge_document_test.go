package knowledge_document

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func TestAccResourceKnowledgeDocumentBasic(t *testing.T) {
	var (
		knowledgeBaseResourceLabel1     = "test-knowledgebase1"
		categoryResourceLabel1          = "test-category1"
		categoryName                    = "Terraform Knowledge Category " + uuid.NewString()
		categoryDescription             = "test-knowledge-category-description1"
		labelResourceLabel1             = "test-label1"
		labelName                       = "Terraform Knowledge Label " + uuid.NewString()
		labelColor                      = "#0F0F0F"
		knowledgeBaseName1              = "Terraform Knowledge Base " + uuid.NewString()
		knowledgeBaseDescription1       = "test-knowledgebase-description1"
		coreLanguage1                   = "en-US"
		knowledgeDocumentResourceLabel1 = "test-knowledge-document1"
		title                           = "Terraform Knowledge Document"
		visible                         = true
		visible2                        = false
		published                       = false
		phrase                          = "Terraform Knowledge Document"
		autocomplete                    = true
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateKnowledgeKnowledgebaseResource(
					knowledgeBaseResourceLabel1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					coreLanguage1,
				) +
					generateKnowledgeCategoryResource(
						categoryResourceLabel1,
						knowledgeBaseResourceLabel1,
						categoryName,
						categoryDescription,
					) +
					generateKnowledgeLabelResource(
						labelResourceLabel1,
						knowledgeBaseResourceLabel1,
						labelName,
						labelColor,
					) +
					generateKnowledgeDocumentResource(
						knowledgeDocumentResourceLabel1,
						knowledgeBaseResourceLabel1,
						categoryResourceLabel1,
						labelResourceLabel1,
						categoryName,
						labelName,
						title,
						visible,
						published,
						phrase,
						autocomplete,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.title", title),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.visible", fmt.Sprintf("%v", visible)),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.alternatives.0.phrase", phrase),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.alternatives.0.autocomplete", fmt.Sprintf("%v", autocomplete)),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.label_names.0", labelName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.category_name", categoryName),
				),
			},
			{
				// Update
				Config: generateKnowledgeKnowledgebaseResource(
					knowledgeBaseResourceLabel1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					coreLanguage1,
				) +
					generateKnowledgeCategoryResource(
						categoryResourceLabel1,
						knowledgeBaseResourceLabel1,
						categoryName,
						categoryDescription,
					) +
					generateKnowledgeLabelResource(
						labelResourceLabel1,
						knowledgeBaseResourceLabel1,
						labelName,
						labelColor,
					) +
					generateKnowledgeDocumentResource(
						knowledgeDocumentResourceLabel1,
						knowledgeBaseResourceLabel1,
						categoryResourceLabel1,
						labelResourceLabel1,
						categoryName,
						labelName,
						title,
						visible2,
						published,
						phrase,
						autocomplete,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.title", title),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.visible", fmt.Sprintf("%v", visible2)),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.alternatives.0.phrase", phrase),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.alternatives.0.autocomplete", fmt.Sprintf("%v", autocomplete)),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.category_name", categoryName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document."+knowledgeDocumentResourceLabel1, "knowledge_document.0.label_names.0", labelName),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_knowledge_document." + knowledgeDocumentResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgeDocumentDestroyed,
	})
}

func generateKnowledgeDocumentResource(resourceLabel string, knowledgeBaseResourceLabel string, knowledgeCategoryResourceLabel string, knowledgeLabelResourceLabel string, knowledgeCategoryName string, knowledgeLabelName string, title string, visible bool, published bool, phrase string, autocomplete bool) string {
	document := fmt.Sprintf(`
        resource "genesyscloud_knowledge_document" "%s" {
			depends_on=[genesyscloud_knowledge_category.%s, genesyscloud_knowledge_label.%s]
            knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
            published = %v
            %s
        }
        `, resourceLabel,
		knowledgeCategoryResourceLabel,
		knowledgeLabelResourceLabel,
		knowledgeBaseResourceLabel,
		published,
		generateKnowledgeDocumentRequestBody(knowledgeCategoryName, knowledgeLabelName, title, visible, phrase, autocomplete),
	)
	return document
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

func generateKnowledgeKnowledgebaseResource(
	resourceLabel string,
	name string,
	description string,
	coreLanguage string) string {
	return fmt.Sprintf(`resource "genesyscloud_knowledge_knowledgebase" "%s" {
		name = "%s"
        description = "%s"
        core_language = "%s"
	}
	`, resourceLabel, name, description, coreLanguage)
}
func generateKnowledgeCategoryResource(resourceLabel string, knowledgeBaseResource string, categoryName string, categoryDescription string) string {
	category := fmt.Sprintf(`
        resource "genesyscloud_knowledge_category" "%s" {
            knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
            %s
        }
        `, resourceLabel,
		knowledgeBaseResource,
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

func generateKnowledgeLabelResource(resourceLabel string, knowledgeBaseResource string, labelName string, labelColor string) string {
	label := fmt.Sprintf(`
        resource "genesyscloud_knowledge_label" "%s" {
            knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
            %s
        }
        `, resourceLabel,
		knowledgeBaseResource,
		generateKnowledgeLabelRequestBody(labelName, labelColor),
	)
	return label
}

func generateKnowledgeLabelRequestBody(labelName string, labelColor string) string {

	return fmt.Sprintf(`
        knowledge_label {
            name = "%s"
            color = "%s"
        }
        `, labelName,
		labelColor,
	)
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
