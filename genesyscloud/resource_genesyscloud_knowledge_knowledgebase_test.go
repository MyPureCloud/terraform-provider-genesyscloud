package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceKnowledgeKnowledgebaseBasic(t *testing.T) {
	var (
		knowledgeBaseResource1     = "test-knowledgebase1"
		knowledgeBaseName1         = "Terraform Knowledge Base" + uuid.NewString()
		knowledgeBaseDescription1  = "test-knowledgebase-description1"
		knowledgeBaseDescription2  = "test-knowledgebase-description2"
		knowledgeBaseCoreLanguage1 = "en-US"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateKnowledgeKnowledgebaseResource(
					knowledgeBaseResource1,
					knowledgeBaseName1,
					knowledgeBaseDescription1,
					knowledgeBaseCoreLanguage1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "name", knowledgeBaseName1),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "description", knowledgeBaseDescription1),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "core_language", knowledgeBaseCoreLanguage1),
				),
			},
			{
				// Update
				Config: generateKnowledgeKnowledgebaseResource(
					knowledgeBaseResource1,
					knowledgeBaseName1,
					knowledgeBaseDescription2,
					knowledgeBaseCoreLanguage1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "name", knowledgeBaseName1),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "description", knowledgeBaseDescription2),
					resource.TestCheckResourceAttr("genesyscloud_knowledge_knowledgebase."+knowledgeBaseResource1, "core_language", knowledgeBaseCoreLanguage1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_knowledge_knowledgebase." + knowledgeBaseResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyKnowledgebasesDestroyed,
	})
}

func generateKnowledgeKnowledgebaseResource(
	resourceID string,
	name string,
	description string,
	coreLanguage string) string {
	return fmt.Sprintf(`resource "genesyscloud_knowledge_knowledgebase" "%s" {
		name = "%s"
        description = "%s"
        core_language = "%s"
	}
	`, resourceID, name, description, coreLanguage)
}

func testVerifyKnowledgebasesDestroyed(state *terraform.State) error {
	knowledgeAPI := platformclientv2.NewKnowledgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_knowledge_knowledgebase" {
			continue
		}

		knowledgeBase, resp, err := knowledgeAPI.GetKnowledgeKnowledgebase(rs.Primary.ID)
		if knowledgeBase != nil {
			return fmt.Errorf("Knowledge base (%s) still exists", rs.Primary.ID)
		} else if IsStatus404(resp) {
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
