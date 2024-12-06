package knowledge

import (
	"fmt"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func TestAccResourceKnowledgeDocumentVariationBasic(t *testing.T) {
	var (
		variationResourceLabel1         = "test-variation1"
		knowledgeBaseResourceLabel1     = "test-knowledgebase1"
		knowledgeBaseName1              = "Terraform Knowledge Base " + uuid.NewString()
		knowledgeBaseDescription1       = "test-knowledgebase-description1"
		coreLanguage1                   = "en-US"
		knowledgeDocumentResourceLabel1 = "test-knowledge-document1"
		title                           = "Terraform Knowledge Document"
		visible                         = true
		docPublished                    = false
		published                       = true
		phrase                          = "Terraform Knowledge Document"
		autocomplete                    = true
		bodyBlockType                   = "Paragraph"
		contentBlockType1               = "Text"
		contentBlockType2               = "Image"
		imageUrl                        = "https://example.com/image"
		hyperlink                       = "https://example.com/hyperlink"
		videoUrl                        = "https://example.com/video"
		listType                        = "ListItem"
		documentText                    = "stuff"
		marks                           = []string{"Bold", "Italic", "Underline"}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create
				Config: gcloud.GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResourceLabel1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					coreLanguage1,
				) +
					generateKnowledgeDocumentBasic(
						knowledgeDocumentResourceLabel1,
						knowledgeBaseResourceLabel1,
						title,
						visible,
						docPublished,
						phrase,
						autocomplete,
					) +
					generateKnowledgeDocumentVariation(
						variationResourceLabel1,
						knowledgeBaseResourceLabel1,
						knowledgeDocumentResourceLabel1,
						published,
						bodyBlockType,
						contentBlockType1,
						imageUrl,
						hyperlink,
						videoUrl,
						listType,
						documentText,
						marks,
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document_variation."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.type", bodyBlockType),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document_variation."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.type", contentBlockType1),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document_variation."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.text.0.text", documentText),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document_variation."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.text.0.marks.#", fmt.Sprintf("%v", len(marks))),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document_variation."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.text.0.hyperlink", hyperlink),
				),
			},
			{
				// Update
				Config: gcloud.GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResourceLabel1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					coreLanguage1,
				) +
					generateKnowledgeDocumentBasic(
						knowledgeDocumentResourceLabel1,
						knowledgeBaseResourceLabel1,
						title,
						visible,
						published,
						phrase,
						autocomplete,
					) +
					generateKnowledgeDocumentVariation(
						variationResourceLabel1,
						knowledgeBaseResourceLabel1,
						knowledgeDocumentResourceLabel1,
						published,
						bodyBlockType,
						contentBlockType2,
						imageUrl,
						hyperlink,
						videoUrl,
						listType,
						documentText,
						marks,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document_variation."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.type", bodyBlockType),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document_variation."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.type", contentBlockType2),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document_variation."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.image.0.url", imageUrl),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_document_variation."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.image.0.hyperlink", hyperlink),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_knowledge_document_variation." + variationResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgeDocumentVariationDestroyed,
	})
}

func generateKnowledgeDocumentVariation(resourceLabel string, knowledgeBaseResourceLabel string, knowledgeDocumentResourceLabel string, published bool, bodyBlockType string, contentBlockType string, imageUrl string, hyperlink string, videoUrl string, listType string, documentText string, marks []string) string {
	variation := fmt.Sprintf(`
        resource "genesyscloud_knowledge_document_variation" "%s" {
			depends_on=[genesyscloud_knowledge_document.%s]
			knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
			knowledge_document_id = genesyscloud_knowledge_document.%s.id
			published = %v
			%v
        }
        `, resourceLabel,
		knowledgeDocumentResourceLabel,
		knowledgeBaseResourceLabel,
		knowledgeDocumentResourceLabel,
		published,
		generateKnowledgeDocumentVariationBody(bodyBlockType, contentBlockType, imageUrl, hyperlink, videoUrl, listType, documentText, marks),
	)
	return variation
}

func generateKnowledgeDocumentVariationBody(bodyBlockType string, contentBlockType string, imageUrl string, hyperlink string, videoUrl string, listType string, documentText string, marks []string) string {
	variationBody := fmt.Sprintf(`
        knowledge_document_variation {
			%v
		}
        `, generateDocumentBody(bodyBlockType, contentBlockType, imageUrl, hyperlink, videoUrl, listType, documentText, marks),
	)
	return variationBody
}

func generateDocumentBody(bodyBlockType string, contentBlockType string, imageUrl string, hyperlink string, videoUrl string, listType string, documentText string, marks []string) string {
	documentBody := fmt.Sprintf(`
        body {
			%v
		}
        `, generateDocumentBodyBlocks(bodyBlockType, contentBlockType, imageUrl, hyperlink, videoUrl, listType, documentText, marks),
	)
	return documentBody
}

func generateDocumentBodyBlocks(bodyBlockType string, contentBlockType string, imageUrl string, hyperlink string, videoUrl string, listType string, documentText string, marks []string) string {
	bodyBlocks := ""
	if bodyBlockType == "Paragraph" {
		bodyBlocks = fmt.Sprintf(`
			blocks {
				type = "%s"
				%v
			}
			`, bodyBlockType,
			generateDocumentBodyParagraph(documentText, imageUrl, hyperlink, marks, contentBlockType),
		)
	}
	if bodyBlockType == "Image" {
		bodyBlocks = fmt.Sprintf(`
			blocks {
				type = "%s"
				%v
			}
			`, bodyBlockType,
			generateDocumentBodyImage(imageUrl, hyperlink),
		)
	}
	if bodyBlockType == "Video" {
		bodyBlocks = fmt.Sprintf(`
			blocks {
				type = "%s"
				%v
			}
			`, bodyBlockType,
			generateDocumentBodyVideo(videoUrl),
		)
	}
	if bodyBlockType == "OrderedList" || bodyBlockType == "UnorderedList" {
		bodyBlocks = fmt.Sprintf(`
			blocks {
				type = "%s"
				%v
			}
			`, bodyBlockType,
			generateDocumentBodyList(listType, documentText, imageUrl, hyperlink, marks, contentBlockType),
		)
	}

	return bodyBlocks
}

func generateDocumentBodyParagraph(documentText string, imageUrl string, hyperlink string, marks []string, contentBlockType string) string {
	paragraph := fmt.Sprintf(`
        paragraph {
			%v
		}
        `, generateDocumentContentBlocks(documentText, imageUrl, hyperlink, marks, contentBlockType),
	)
	return paragraph
}

func generateDocumentContentBlocks(documentText string, imageUrl string, hyperlink string, marks []string, contentBlockType string) string {
	contentBlocks := ""
	if contentBlockType == "Text" {
		contentBlocks = fmt.Sprintf(`
			blocks {
				type = "%s"
				%v
			}
			`,
			contentBlockType,
			generateDocumentText(documentText, marks, hyperlink),
		)
	} else {
		contentBlocks = fmt.Sprintf(`
			blocks {
				type = "%s"
				%v
			}
			`,
			contentBlockType,
			generateDocumentBodyImage(imageUrl, hyperlink),
		)
	}
	return contentBlocks
}

func generateDocumentText(documentText string, marks []string, hyperlink string) string {
	markString := ""
	for i, mark := range marks {
		markString += fmt.Sprintf("\"%s\"", mark)

		if i < len(marks)-1 {
			markString += ","
		}
	}

	contentBlocks := fmt.Sprintf(`
        text {
			text = "%s"
			marks = [%s]
			hyperlink = "%s"
		}
        `, documentText,
		markString,
		hyperlink,
	)
	return contentBlocks
}

func generateDocumentBodyImage(imageUrl string, hyperlink string) string {
	image := fmt.Sprintf(`
        image {
			url = "%s"
			hyperlink = "%s"
		}
        `, imageUrl,
		hyperlink,
	)
	return image
}

func generateDocumentBodyVideo(videoUrl string) string {
	video := fmt.Sprintf(`
        video {
			url = "%s"
		}
        `, videoUrl,
	)
	return video
}

func generateDocumentBodyList(listType string, documentText string, imageUrl string, hyperlink string, marks []string, contentBlockType1 string) string {
	list := fmt.Sprintf(`
        list {
			%v
		}
        `, generateDocumentBodyListBlocks(listType, documentText, imageUrl, hyperlink, marks, contentBlockType1),
	)
	return list
}

func generateDocumentBodyListBlocks(listType string, documentText string, imageUrl string, hyperlink string, marks []string, contentBlockType1 string) string {
	listBlocks := fmt.Sprintf(`
        blocks {
			type = "%s"
			%v
		}
        `, listType,
		generateDocumentContentBlocks(documentText, imageUrl, hyperlink, marks, contentBlockType1),
	)
	return listBlocks
}

func generateAddressableEntityRef(versionId string) string {
	variationBody := fmt.Sprintf(`
        document_version {
			id = "%s"
		}
        `, versionId,
	)
	return variationBody
}

func generateKnowledgeDocumentBasic(resourceLabel string, knowledgeBaseResourceLabel string, title string, visible bool, published bool, phrase string, autocomplete bool) string {
	document := fmt.Sprintf(`
        resource "genesyscloud_knowledge_document" "%s" {
            knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
            published = %v
            %s
        }
        `, resourceLabel,
		knowledgeBaseResourceLabel,
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

func testVerifyKnowledgeDocumentVariationDestroyed(state *terraform.State) error {
	knowledgeAPI := platformclientv2.NewKnowledgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_knowledge_document_variation" {
			continue
		}
		id := strings.Split(rs.Primary.ID, " ")
		knowledgeDocumentVariationId := id[0]
		knowledgeBaseId := id[1]
		knowledgeDocumentId := id[2]
		publishedKnowledgeDocumentVariation, publishedResp, publishedErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(knowledgeDocumentVariationId, knowledgeDocumentId, knowledgeBaseId, "Published")
		// check both published and draft variations
		if publishedKnowledgeDocumentVariation != nil {
			return fmt.Errorf("Knowledge document variation (%s) still exists", knowledgeDocumentVariationId)
		} else if util.IsStatus404(publishedResp) || util.IsStatus400(publishedResp) {
			draftKnowledgeDocumentVariation, draftResp, draftErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(knowledgeDocumentVariationId, knowledgeDocumentId, knowledgeBaseId, "Draft")

			if draftKnowledgeDocumentVariation != nil {
				return fmt.Errorf("Knowledge document variation (%s) still exists", knowledgeDocumentVariationId)
			} else if util.IsStatus404(draftResp) || util.IsStatus400(draftResp) {
				// Knowledge base document not found as expected
				continue
			} else {
				// Unexpected error
				return fmt.Errorf("Unexpected error: %s", draftErr)
			}
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", publishedErr)
		}
	}
	// Success. All knowledge base documents destroyed
	return nil
}
