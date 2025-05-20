package knowledge_knowledgebase

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func TestAccResourceKnowledgeKnowledgebaseBasic(t *testing.T) {
	var (
		knowledgeBaseResourceLabel1 = "test-knowledgebase1"
		knowledgeBaseName1          = "Terraform Knowledge Base" + uuid.NewString()
		knowledgeBaseDescription1   = "test-knowledgebase-description1"
		knowledgeBaseDescription2   = "test-knowledgebase-description2"
		knowledgeBaseCoreLanguage1  = "en-US"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResourceLabel1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+knowledgeBaseResourceLabel1, "name", knowledgeBaseName1),
					resource.TestCheckResourceAttr(ResourceType+"."+knowledgeBaseResourceLabel1, "description", knowledgeBaseDescription1),
					resource.TestCheckResourceAttr(ResourceType+"."+knowledgeBaseResourceLabel1, "core_language", knowledgeBaseCoreLanguage1),
				),
			},
			{
				// Update
				Config: GenerateKnowledgeKnowledgebaseResource(
					knowledgeBaseResourceLabel1,
					knowledgeBaseName1,
					knowledgeBaseDescription2,
					knowledgeBaseCoreLanguage1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+knowledgeBaseResourceLabel1, "name", knowledgeBaseName1),
					resource.TestCheckResourceAttr(ResourceType+"."+knowledgeBaseResourceLabel1, "description", knowledgeBaseDescription2),
					resource.TestCheckResourceAttr(ResourceType+"."+knowledgeBaseResourceLabel1, "core_language", knowledgeBaseCoreLanguage1),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + knowledgeBaseResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgebasesDestroyed,
	})
}

func testVerifyKnowledgebasesDestroyed(state *terraform.State) error {
	knowledgeAPI := platformclientv2.NewKnowledgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		knowledgeBase, resp, err := knowledgeAPI.GetKnowledgeKnowledgebase(rs.Primary.ID)
		if knowledgeBase != nil {
			return fmt.Errorf("Knowledge base (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Knowledge base not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All knowledge bases destroyed
	return nil
}
