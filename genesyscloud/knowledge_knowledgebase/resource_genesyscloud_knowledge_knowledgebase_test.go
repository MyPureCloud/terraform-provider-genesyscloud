package knowledge_knowledgebase

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceKnowledgeKnowledgebaseBasic(t *testing.T) {
	var (
		knowledgeBaseResourceLabel1 = "test-knowledgebase1"
		knowledgeBaseName1          = "Test-Terraform-Knowledge-Base" + uuid.NewString()
		knowledgeBaseDescription1   = "test-knowledgebase-description1"
		knowledgeBaseDescription2   = "test-knowledgebase-description2"
		knowledgeBaseCoreLanguage1  = "en-US"
	)

	err := cleanUpKnowledgeBase("Test-Terraform-Knowledge-Base")
	if err != nil {
		log.Printf("Failed to clean up knowledge base: %v", err)
	}

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

func cleanUpKnowledgeBase(knowledgeBaseName string) error {
	log.Printf("Cleaning up Knowledge Bases with name '%s'", knowledgeBaseName)
	knowledgeApi := platformclientv2.NewKnowledgeApi()

	var after string
	const pageSize = "100"

	for {
		knowledgeBases, _, err := knowledgeApi.GetKnowledgeKnowledgebases("", after, "", pageSize, "", "", true, "", "")
		if err != nil {
			return fmt.Errorf("failed to get page %v of knowledge bases: %v", after, err)
		}

		if knowledgeBases.Entities == nil || len(*knowledgeBases.Entities) == 0 {
			break
		}

		for _, knowledgeBase := range *knowledgeBases.Entities {
			if knowledgeBase.Name != nil && strings.HasPrefix(*knowledgeBase.Name, knowledgeBaseName) {
				log.Printf("Deleting knowledge base %s", *knowledgeBase.Name)
				_, _, err := knowledgeApi.DeleteKnowledgeKnowledgebase(*knowledgeBase.Id)
				if err != nil {
					// Logging the error rather than returning it to ensure the deletion of other knowledge bases
					log.Printf("Failed to delete knowledge base %s: %v", *knowledgeBase.Name, err)
					continue
				}
				log.Printf("Deleted knowledge base %s", *knowledgeBase.Name)
				time.Sleep(5 * time.Second)
			}
		}

		previousAfter := after
		if knowledgeBases.NextUri == nil {
			break
		}
		after, err = util.GetQueryParamValueFromUri(*knowledgeBases.NextUri, "after")
		if err != nil {
			return err
		}
		if after == "" || after == previousAfter {
			break
		}
	}
	log.Printf("Cleaned up Knowledge Bases with name '%s'", knowledgeBaseName)
	return nil
}
