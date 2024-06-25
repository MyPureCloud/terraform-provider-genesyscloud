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

func TestAccResourceKnowledgeLabelBasic(t *testing.T) {
	var (
		knowledgeBaseResource1     = "test-knowledgebase1"
		knowledgeBaseName1         = "Terraform Knowledge Base" + uuid.NewString()
		knowledgeBaseDescription1  = "test-knowledgebase-description1"
		knowledgeBaseCoreLanguage1 = "en-US"
		knowledgeLabelResource1    = "test-knowledge-label1"
		labelName                  = "Terraform Knowledge Label" + uuid.NewString()
		labelColor                 = "#0F0F0F"
		labelColor2                = "#FFFFFF"
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
				) +
					generateKnowledgeLabelResource(
						knowledgeLabelResource1,
						knowledgeBaseResource1,
						labelName,
						labelColor,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_knowledge_label."+knowledgeLabelResource1, "knowledge_base_id", "genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_label."+knowledgeLabelResource1, "knowledge_label.0.name", labelName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_label."+knowledgeLabelResource1, "knowledge_label.0.color", labelColor),
				),
			},
			{
				// Update
				Config: GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResource1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				) +
					generateKnowledgeLabelResource(
						knowledgeLabelResource1,
						knowledgeBaseResource1,
						labelName,
						labelColor2,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_knowledge_label."+knowledgeLabelResource1, "knowledge_base_id", "genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_label."+knowledgeLabelResource1, "knowledge_label.0.name", labelName),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_label."+knowledgeLabelResource1, "knowledge_label.0.color", labelColor2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_knowledge_label." + knowledgeLabelResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgeLabelDestroyed,
	})
}

func generateKnowledgeLabelResource(resourceName string, knowledgeBaseResource string, labelName string, labelColor string) string {
	label := fmt.Sprintf(`
        resource "genesyscloud_knowledge_label" "%s" {
            knowledge_base_id = genesyscloud_knowledge_knowledgebase.%s.id
            %s
        }
        `, resourceName,
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
		if rs.Type == "genesyscloud_knowledge_knowledgebase" {
			knowledgeBaseId = rs.Primary.ID
			break
		}
	}
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_knowledge_label" {
			continue
		}
		id := strings.Split(rs.Primary.ID, " ")
		knowledgeLabelId := id[0]
		knowledgeLabel, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseLabel(knowledgeBaseId, knowledgeLabelId)
		if knowledgeLabel != nil {
			return fmt.Errorf("Knowledge label (%s) still exists", knowledgeLabelId)
		} else if util.IsStatus404(resp) || util.IsStatus400(resp) {
			// Knowledge base label not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All knowledge labels destroyed
	return nil
}
