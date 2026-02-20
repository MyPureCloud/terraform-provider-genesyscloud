package knowledge_label

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	knowledgeKnowledgebase "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_knowledgebase"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
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

	// Find the knowledge base ID
	for _, rs := range state.RootModule().Resources {
		if rs.Type == knowledgeKnowledgebase.ResourceType {
			knowledgeBaseId = rs.Primary.ID
			break
		}
	}

	// Validate all labels are deleted
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		id := strings.Split(rs.Primary.ID, " ")
		knowledgeLabelId := id[0]

		// Retry up to 120 seconds
		if err := util.WithRetries(context.Background(), 120*time.Second, func() *retry.RetryError {

			knowledgeLabel, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseLabel(
				knowledgeBaseId,
				knowledgeLabelId,
			)

			if knowledgeLabel != nil {
				// Still exists
				return retry.RetryableError(fmt.Errorf("knowledge label (%s) still exists", knowledgeLabelId))
			}

			if util.IsStatus404(resp) || util.IsStatus400(resp) {
				// Deleted successfully
				return nil
			}

			// Fail if Any other error
			return retry.NonRetryableError(fmt.Errorf("unexpected error: %v", err))

		}); err != nil {
			return fmt.Errorf("unexpected error: %v", err)
		}
	}

	return nil
}
