package knowledge_label

import (
	"fmt"
	"strings"
	"testing"

	knowledgeKnowledgebase "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceKnowledgeLabelBasic(t *testing.T) {
	var (
		knowledgeBaseResourceLabel1   = "test-knowledgebase1"
		knowledgeBaseName1            = "Test-Terraform-Knowledge-Base" + uuid.NewString()
		knowledgeBaseDescription1     = "test-knowledgebase-description1"
		knowledgeBaseCoreLanguage1    = "en-US"
		knowledgeBaseFullResourcePath = knowledgeKnowledgebase.ResourceType + "." + knowledgeBaseResourceLabel1

		knowledgeLabelResourceLabel1   = "test-knowledge-label1"
		knowledgeLabelFullResourcePath = ResourceType + "." + knowledgeLabelResourceLabel1
		labelName                      = "Terraform Knowledge Label" + uuid.NewString()
		labelColor                     = "#0F0F0F"
		labelColor2                    = "#FFFFFF"
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
				) +
					generateKnowledgeLabelResource(
						knowledgeLabelResourceLabel1,
						knowledgeBaseResourceLabel1,
						labelName,
						labelColor,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						knowledgeLabelFullResourcePath, "knowledge_base_id",
						knowledgeBaseFullResourcePath, "id",
					),
					resource.TestCheckResourceAttr(knowledgeLabelFullResourcePath, "knowledge_label.0.name", labelName),
					resource.TestCheckResourceAttr(knowledgeLabelFullResourcePath, "knowledge_label.0.color", labelColor),
				),
			},
			{
				// Update
				Config: knowledgeKnowledgebase.GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResourceLabel1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				) +
					generateKnowledgeLabelResource(
						knowledgeLabelResourceLabel1,
						knowledgeBaseResourceLabel1,
						labelName,
						labelColor2,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						knowledgeLabelFullResourcePath, "knowledge_base_id",
						knowledgeBaseFullResourcePath, "id",
					),
					resource.TestCheckResourceAttr(knowledgeLabelFullResourcePath, "knowledge_label.0.name", labelName),
					resource.TestCheckResourceAttr(knowledgeLabelFullResourcePath, "knowledge_label.0.color", labelColor2),
				),
			},
			{
				// Import/Read
				ResourceName:      knowledgeLabelFullResourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgeLabelDestroyed,
	})
}

func generateKnowledgeLabelResource(resourceLabel string, knowledgeBaseResource string, labelName string, labelColor string) string {
	label := fmt.Sprintf(`
        resource "%s" "%s" {
            knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
            %s
        }
        `, ResourceType,
		resourceLabel,
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

func testVerifyKnowledgeLabelDestroyed(state *terraform.State) error {
	knowledgeAPI := platformclientv2.NewKnowledgeApi()
	var knowledgeBaseId string
	for _, rs := range state.RootModule().Resources {
		if rs.Type == knowledgeKnowledgebase.ResourceType {
			knowledgeBaseId = rs.Primary.ID
			break
		}
	}
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}
		id := strings.Split(rs.Primary.ID, " ")
		knowledgeLabelId := id[0]
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseLabel(knowledgeBaseId, knowledgeLabelId)
		if err == nil {
			return fmt.Errorf("knowledge label (%s) still exists", knowledgeLabelId)
		} else if util.IsStatus404(resp) || util.IsStatus400(resp) {
			// Knowledge base label not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err.Error())
		}
	}
	// Success. All knowledge labels destroyed
	return nil
}
