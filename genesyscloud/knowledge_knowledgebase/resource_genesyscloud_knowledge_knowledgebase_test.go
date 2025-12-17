package knowledge_knowledgebase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
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

	// Validate all knowledge bases are deleted
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		knowledgeBaseId := rs.Primary.ID

		// Retry for async deletion
		if err := util.WithRetries(context.Background(), 120*time.Second, func() *retry.RetryError {

			knowledgeBase, resp, err := knowledgeAPI.GetKnowledgeKnowledgebase(knowledgeBaseId)

			if knowledgeBase != nil {
				// Still exists
				return retry.RetryableError(
					fmt.Errorf("knowledge base (%s) still exists", knowledgeBaseId),
				)
			}

			if util.IsStatus404(resp) || util.IsStatus400(resp) {
				// Deleted successfully
				return nil
			}

			// Unexpected error
			return retry.NonRetryableError(fmt.Errorf("unexpected error: %v", err))

		}); err != nil {
			return fmt.Errorf("unexpected error: %v", err)
		}
	}

	// Success
	return nil
}

func cleanUpKnowledgeBase(knowledgeBaseName string) error {
	log.Printf("Cleaning up Knowledge Bases with name '%s'", knowledgeBaseName)
	knowledgeApi := platformclientv2.NewKnowledgeApi()

	var after string
	const pageSize = "100"

	rebuildError := rebuildDatabase()
	if rebuildError != nil {
		log.Printf("Failed to rebuild database: %v", rebuildError)
	}

	for {
		knowledgeBases, _, err := knowledgeApi.GetKnowledgeKnowledgebases("", after, "", pageSize, "", "", true, "", "")
		if err != nil {
			return fmt.Errorf("failed to get page %v of knowledge bases: %v", after, err)
		}

		if knowledgeBases.Entities == nil || len(*knowledgeBases.Entities) == 0 {
			break
		}

		for _, knowledgeBase := range *knowledgeBases.Entities {
			// Check if the knowledge base name starts with the Test name
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

			// Check if the knowledge base description starts with the Test description
			if knowledgeBase.Description != nil && strings.HasPrefix(*knowledgeBase.Description, "test-knowledgebase-description") {
				log.Printf("Deleting knowledge base %s, description: %s", *knowledgeBase.Name, *knowledgeBase.Description)
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
