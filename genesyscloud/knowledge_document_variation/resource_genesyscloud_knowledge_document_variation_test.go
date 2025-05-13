package knowledge_document_variation

import (
	"fmt"
	knowledgeKnowledgebase "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
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
		name                            = "Terraform Test Knowledge Document Variation"
		contextId                       = uuid.NewString()
		valueId                         = uuid.NewString()
		paragraphTestProperties         = map[string]string{
			"fSize":   "Large",
			"fType":   "Heading1",
			"tColor":  "#FFFFFF",
			"bgColor": "#000000",
			"align":   "Right",
			"indent":  "3.14",
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create
				Config: knowledgeKnowledgebase.GenerateKnowledgeKnowledgebaseResource(
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
						name,
						contextId,
						valueId,
						paragraphTestProperties,
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.type", bodyBlockType),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.type", contentBlockType1),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.text.0.text", documentText),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.text.0.marks.#", fmt.Sprintf("%v", len(marks))),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.text.0.hyperlink", hyperlink),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.name", name),
				),
			},
			{
				// Update
				Config: knowledgeKnowledgebase.GenerateKnowledgeKnowledgebaseResource(
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
						contentBlockType2,
						imageUrl,
						hyperlink,
						videoUrl,
						listType,
						documentText,
						marks,
						name,
						contextId,
						valueId,
						paragraphTestProperties,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.type", bodyBlockType),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.type", contentBlockType2),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.image.0.url", imageUrl),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.blocks.0.image.0.hyperlink", hyperlink),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.name", name),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + variationResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgeDocumentVariationDestroyed,
	})
}

func TestAccResourceKnowledgeDocumentVariationDifferentTypes(t *testing.T) {
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
		bodyBlockTypeList               = "UnorderedList"
		bodyBlockTypeVideo              = "Video"
		bodyBlockTypeImage              = "Image"
		bodyBlockTypeParagraph          = "Paragraph"
		contentBlockType1               = "Text"
		imageUrl                        = "https://example.com/image"
		hyperlink                       = "https://example.com/hyperlink"
		videoUrl                        = "https://example.com/video"
		listType                        = "ListItem"
		marks                           = []string{"Bold", "Italic", "Underline"}
		name                            = "Terraform Test Knowledge Document Variation"
		documentText                    = "stuff"
		contextId                       = uuid.NewString()
		valueId                         = uuid.NewString()

		listTestProperties = map[string]string{
			"unordered_type": "Square",
			"ordered_type":   "Number",
		}
		videoImageTestProperties = map[string]string{
			"bgColor": "#000000",
			"align":   "Right",
			"indent":  "3.14",
		}
		paragraphTestProperties = map[string]string{
			"fSize":   "Large",
			"fType":   "Heading1",
			"tColor":  "#FFFFFF",
			"bgColor": "#000000",
			"align":   "Right",
			"indent":  "3.14",
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create Type List
				Config: knowledgeKnowledgebase.GenerateKnowledgeKnowledgebaseResource(
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
						bodyBlockTypeList,
						contentBlockType1,
						imageUrl,
						hyperlink,
						videoUrl,
						listType,
						documentText,
						marks,
						name,
						contextId,
						valueId,
						listTestProperties,
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.type", bodyBlockTypeList),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.list.0.properties.0.unordered_type", listTestProperties["unordered_type"]),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.list.0.properties.0.ordered_type", listTestProperties["ordered_type"]),
				),
			},
			{
				// Create Type Image
				Config: knowledgeKnowledgebase.GenerateKnowledgeKnowledgebaseResource(
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
						bodyBlockTypeImage,
						contentBlockType1,
						imageUrl,
						hyperlink,
						videoUrl,
						listType,
						documentText,
						marks,
						name,
						contextId,
						valueId,
						videoImageTestProperties,
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.type", bodyBlockTypeImage),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.image.0.url", imageUrl),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.image.0.properties.0.align", videoImageTestProperties["align"]),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.image.0.properties.0.background_color", videoImageTestProperties["bgColor"]),
				),
			},
			{
				// Create Type Video
				Config: knowledgeKnowledgebase.GenerateKnowledgeKnowledgebaseResource(
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
						bodyBlockTypeVideo,
						contentBlockType1,
						imageUrl,
						hyperlink,
						videoUrl,
						listType,
						documentText,
						marks,
						name,
						contextId,
						valueId,
						videoImageTestProperties,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.type", bodyBlockTypeVideo),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.video.0.url", videoUrl),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.video.0.properties.0.align", videoImageTestProperties["align"]),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.video.0.properties.0.background_color", videoImageTestProperties["bgColor"]),
				),
			},
			{
				// Create Type Paragraph
				Config: knowledgeKnowledgebase.GenerateKnowledgeKnowledgebaseResource(
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
						bodyBlockTypeParagraph,
						contentBlockType1,
						imageUrl,
						hyperlink,
						videoUrl,
						listType,
						documentText,
						marks,
						name,
						contextId,
						valueId,
						paragraphTestProperties,
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.type", bodyBlockTypeParagraph),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.properties.0.align", paragraphTestProperties["align"]),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.properties.0.background_color", paragraphTestProperties["bgColor"]),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.properties.0.font_size", paragraphTestProperties["fSize"]),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.properties.0.font_type", paragraphTestProperties["fType"]),
					resource.TestCheckResourceAttr(ResourceType+"."+variationResourceLabel1, "knowledge_document_variation.0.body.0.blocks.0.paragraph.0.properties.0.text_color", paragraphTestProperties["tColor"]),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + variationResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgeDocumentVariationDestroyed,
	})
}

func generateKnowledgeDocumentVariation(resourceLabel string, knowledgeBaseResourceLabel string, knowledgeDocumentResourceLabel string, published bool, bodyBlockType string, contentBlockType string, imageUrl string, hyperlink string, videoUrl string, listType string, documentText string, marks []string, name string, contextId, valueId string, properties map[string]string) string {
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
		generateKnowledgeDocumentVariationBody(bodyBlockType, contentBlockType, imageUrl, hyperlink, videoUrl, listType, documentText, marks, name, contextId, valueId, properties),
	)
	return variation
}

func generateKnowledgeContexts(contextId, valueId string) string {
	context := fmt.Sprintf(`
        contexts {
			context {
				context_id = "%s"
			}
			values {
				value_id = "%s"
			}
		}
        `, contextId, valueId,
	)
	return context
}

func generateKnowledgeDocumentVariationBody(bodyBlockType string, contentBlockType string, imageUrl string, hyperlink string, videoUrl string, listType string, documentText string, marks []string, name string, contextId, valueId string, properties map[string]string) string {
	variationBody := fmt.Sprintf(`
        knowledge_document_variation {
		name = "%s"
			%v
			%v
		}
        `, name, generateKnowledgeContexts(contextId, valueId), generateDocumentBody(bodyBlockType, contentBlockType, imageUrl, hyperlink, videoUrl, listType, documentText, marks, properties),
	)
	return variationBody
}

func generateDocumentBody(bodyBlockType string, contentBlockType string, imageUrl string, hyperlink string, videoUrl string, listType string, documentText string, marks []string, properties map[string]string) string {
	documentBody := fmt.Sprintf(`
        body {
			%v
		}
        `, generateDocumentBodyBlocks(bodyBlockType, contentBlockType, imageUrl, hyperlink, videoUrl, listType, documentText, marks, properties),
	)
	return documentBody
}

func generateDocumentBodyBlocks(bodyBlockType string, contentBlockType string, imageUrl string, hyperlink string, videoUrl string, listType string, documentText string, marks []string, properties map[string]string) string {
	bodyBlocks := ""
	if bodyBlockType == "Paragraph" {
		bodyBlocks = fmt.Sprintf(`
			blocks {
				type = "%s"
				%v
			}
			`, bodyBlockType,
			generateDocumentBodyParagraph(documentText, imageUrl, hyperlink, marks, contentBlockType, properties),
		)
	}
	if bodyBlockType == "Image" {
		bodyBlocks = fmt.Sprintf(`
			blocks {
				type = "%s"
				%v
			}
			`, bodyBlockType,
			generateDocumentBodyImage(imageUrl, hyperlink, properties),
		)
	}
	if bodyBlockType == "Video" {
		bodyBlocks = fmt.Sprintf(`
			blocks {
				type = "%s"
				%v
			}
			`, bodyBlockType,
			generateDocumentBodyVideo(videoUrl, properties),
		)
	}
	if bodyBlockType == "OrderedList" || bodyBlockType == "UnorderedList" {
		bodyBlocks = fmt.Sprintf(`
			blocks {
				type = "%s"
				%v
			}
			`, bodyBlockType,
			generateDocumentBodyList(listType, documentText, imageUrl, hyperlink, marks, contentBlockType, properties),
		)
	}

	return bodyBlocks
}

func generateDocumentBodyParagraph(documentText string, imageUrl string, hyperlink string, marks []string, contentBlockType string, properties map[string]string) string {
	paragraph := fmt.Sprintf(`
        paragraph {
			%v
			properties {
				align = "%s"
				background_color = "%s"
				indentation = %v
				font_size = "%s"
				font_type = "%s"
				text_color = "%s"
			}
		}
        `, generateDocumentContentBlocks(documentText, imageUrl, hyperlink, marks, contentBlockType, properties), properties["align"], properties["bgColor"], properties["indent"], properties["fSize"], properties["fType"], properties["tColor"],
	)
	return paragraph
}

func generateDocumentContentBlocks(documentText string, imageUrl string, hyperlink string, marks []string, contentBlockType string, properties map[string]string) string {
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
			generateDocumentBodyImage(imageUrl, hyperlink, properties),
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

func generateDocumentBodyImage(imageUrl string, hyperlink string, properties map[string]string) string {
	image := fmt.Sprintf(`
        image {
			url = "%s"
			hyperlink = "%s"
			properties {
				align = "%s"
				background_color = "%s"
				indentation = %v
			}
		}
        `, imageUrl, hyperlink, properties["align"], properties["bgColor"], properties["indent"],
	)
	return image
}

func generateDocumentBodyVideo(videoUrl string, properties map[string]string) string {
	video := fmt.Sprintf(`
        video {
			url = "%s"
			properties {
				align = "%s"
				background_color = "%s"
				indentation = %v
			}
		}
        `, videoUrl, properties["align"], properties["bgColor"], properties["indent"],
	)
	return video
}

func generateDocumentBodyList(listType string, documentText string, imageUrl string, hyperlink string, marks []string, contentBlockType1 string, properties map[string]string) string {
	list := fmt.Sprintf(`
        list {
			%v
			%v
		}
        `, generateDocumentBodyListProperties(properties["unordered_type"], properties["ordered_type"]), generateDocumentBodyListBlocks(listType, documentText, imageUrl, hyperlink, marks, contentBlockType1, properties),
	)
	return list
}

func generateDocumentBodyListProperties(unorderedType, orderedType string) string {
	properties := fmt.Sprintf(`
        properties {
			unordered_type = "%s"
			ordered_type = "%s"
		}
        `, unorderedType, orderedType)
	return properties
}

func generateDocumentBodyListBlocks(listType string, documentText string, imageUrl string, hyperlink string, marks []string, contentBlockType1 string, properties map[string]string) string {
	listBlocks := fmt.Sprintf(`
        blocks {
			type = "%s"
			%v
		}
        `, listType,
		generateDocumentContentBlocks(documentText, imageUrl, hyperlink, marks, contentBlockType1, properties),
	)
	return listBlocks
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
		if rs.Type != ResourceType {
			continue
		}

		id := strings.Split(rs.Primary.ID, " ")
		knowledgeDocumentVariationId := id[0]
		knowledgeBaseId := id[1]
		knowledgeDocumentId := id[2]

		publishedKnowledgeDocumentVariation, publishedResp, publishedErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(knowledgeDocumentVariationId, knowledgeDocumentId, knowledgeBaseId, "Published", nil)
		// check both published and draft variations
		if publishedKnowledgeDocumentVariation != nil {
			return fmt.Errorf("knowledge document variation (%s) still exists", knowledgeDocumentVariationId)
		} else if util.IsStatus404(publishedResp) || util.IsStatus400(publishedResp) {
			draftKnowledgeDocumentVariation, draftResp, draftErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocumentVariation(knowledgeDocumentVariationId, knowledgeDocumentId, knowledgeBaseId, "Draft", nil)

			if draftKnowledgeDocumentVariation != nil {
				return fmt.Errorf("knowledge document variation (%s) still exists", knowledgeDocumentVariationId)
			} else if util.IsStatus404(draftResp) || util.IsStatus400(draftResp) {
				// Knowledge base document not found as expected
				continue
			} else {
				return fmt.Errorf("unexpected error: %s", draftErr)
			}
		} else {
			return fmt.Errorf("unexpected error: %s", publishedErr)
		}
	}
	return nil
}
