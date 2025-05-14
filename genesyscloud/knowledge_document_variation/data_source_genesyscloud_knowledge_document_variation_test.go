package knowledge_document_variation

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	knowledgeKnowledgebase "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestAccDataSourceVariationRequest(t *testing.T) {
	var (
		// Knowledge Base
		knowledgeBaseResourceLabel1 = "test-knowledgebase1"
		knowledgeBaseName1          = uuid.NewString()
		knowledgeBaseDescription1   = "test-knowledgebase-description1"
		coreLanguage1               = "en-US"

		// Knowledge Document
		knowledgeDocumentResourceLabel1 = "test-knowledge-document1"
		title                           = "Terraform Knowledge Document"
		visible                         = true
		docPublished                    = false
		phrase                          = "Terraform Knowledge Document"
		autocomplete                    = true

		// Knowledge Document Variation
		variationResourceLabel  = "test-variation"
		published               = true
		bodyBlockType           = "Paragraph"
		contentBlockType1       = "Text"
		imageUrl                = "https://example.com/image"
		hyperlink               = "https://example.com/hyperlink"
		videoUrl                = "https://example.com/video"
		listType                = "ListItem"
		documentText            = "stuff"
		marks                   = []string{"Bold", "Italic", "Underline"}
		name                    = "Terraform Test Knowledge Document Variation"
		contextId               = uuid.NewString()
		valueId                 = uuid.NewString()
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
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
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
					) + generateKnowledgeDocumentVariation(
					variationResourceLabel,
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
				) + generateKnowledgeDocumentVariationDataSource(
					variationResourceLabel,
					name,
					knowledgeBaseResourceLabel1,
					knowledgeDocumentResourceLabel1,
				),
				Check: resource.ComposeTestCheckFunc(
					// As the ID is a concatenation of multiple IDs, this function will be used to test the ids
					func(state *terraform.State) error {

						// Get the Resource ID
						rs, ok := state.RootModule().Resources[ResourceType+"."+variationResourceLabel]
						if !ok {
							return fmt.Errorf("not found: %s", ResourceType+"."+variationResourceLabel)
						}
						id := rs.Primary.ID

						// Split the IDs
						resourceIDs, err := parseResourceIDs(id)
						if err != nil {
							return err
						}

						// Get the Data Source ID
						rs2, ok := state.RootModule().Resources["data."+ResourceType+"."+variationResourceLabel]
						if !ok {
							return fmt.Errorf("not found: %s", ResourceType+"."+variationResourceLabel)
						}

						variationID := rs2.Primary.ID
						knowledgeBaseID := rs2.Primary.Attributes["knowledge_base_id"]
						KnowledgeDocumentID := strings.Split(rs2.Primary.Attributes["knowledge_document_id"], ",")[0]

						// Ensure IDs are equal
						assert.Equal(t, resourceIDs.knowledgeDocumentVariationID, variationID, "Variation ID should be equal")
						assert.Equal(t, resourceIDs.knowledgeBaseID, knowledgeBaseID, "Knowledge Base ID should be equal")
						assert.Equal(t, resourceIDs.knowledgeDocumentID, KnowledgeDocumentID, "Knowledge Document ID should be equal")
						return nil
					},
					func(s *terraform.State) error {
						time.Sleep(45 * time.Second) // Wait for 45 seconds for resources to get deleted properly
						return nil
					},
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + variationResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateKnowledgeDocumentVariationDataSource(resourceLabel, variationName, knowledgeBaseID, knowledgeDocumentID string) string {
	return fmt.Sprintf(`data "genesyscloud_knowledge_document_variation" "%s" {
		knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
		knowledge_document_id = genesyscloud_knowledge_document.%s.id
		name = "%s"
	}
	`, resourceLabel, knowledgeBaseID, knowledgeDocumentID, variationName)
}
