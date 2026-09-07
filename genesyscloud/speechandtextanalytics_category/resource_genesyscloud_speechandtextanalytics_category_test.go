package speechandtextanalytics_category

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceSpeechAndTextAnalyticsCategory(t *testing.T) {
	t.Parallel()

	var (
		resourceLabel = "category_" + uuid.NewString()

		name1        = "tfacc-category-" + uuid.NewString()
		description1 = "Terraform acceptance test category"

		name2        = "tfacc-category-" + uuid.NewString()
		description2 = "Terraform acceptance test category updated"

		interactionType = "Voice"

		// The top two operand levels must be "OperandGroup"; leaves are "Term" or "Topic".
		// The API requires "inverted" and "occurrence" on every operand.
		criteria1 = `jsonencode({
    type       = "OperandGroup"
    inverted   = false
    occurrence = 1
    operands = [
      {
        type       = "OperandGroup"
        inverted   = false
        occurrence = 1
        operands = [
          {
            type       = "Term"
            inverted   = false
            occurrence = 1
            term = {
              word            = "refund"
              participantType = "External"
            }
          }
        ]
      }
    ]
  })`

		criteria2 = `jsonencode({
    type       = "OperandGroup"
    inverted   = false
    occurrence = 1
    operands = [
      {
        type       = "OperandGroup"
        inverted   = false
        occurrence = 1
        operands = [
          {
            type       = "Term"
            inverted   = false
            occurrence = 1
            term = {
              word            = "cancel"
              participantType = "External"
            }
          }
        ]
      }
    ]
  })`
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateCategoryResource(
					resourceLabel,
					name1,
					description1,
					interactionType,
					criteria1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "interaction_type", interactionType),
					resource.TestCheckResourceAttrSet(ResourceType+"."+resourceLabel, "criteria"),
				),
			},
			{
				// Update
				Config: generateCategoryResource(
					resourceLabel,
					name2,
					description2,
					interactionType,
					criteria2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "description", description2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "interaction_type", interactionType),
					resource.TestCheckResourceAttrSet(ResourceType+"."+resourceLabel, "criteria"),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyCategoryDestroyed,
	})
}

func testVerifyCategoryDestroyed(state *terraform.State) error {
	sttAPI := platformclientv2.NewSpeechTextAnalyticsApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		category, resp, err := sttAPI.GetSpeechandtextanalyticsCategory(rs.Primary.ID)
		if category != nil {
			return fmt.Errorf("category (%s) still exists", rs.Primary.ID)
		}
		if util.IsStatus404(resp) {
			// category not found as expected
			continue
		}

		return fmt.Errorf("unexpected error checking category destruction: %v", err)
	}

	return nil
}

func generateCategoryResource(
	resourceLabel, name, description, interactionType, criteria string,
) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
  name             = %q
  description      = %q
  interaction_type = %q
  criteria         = %s
}
`, ResourceType, resourceLabel, name, description, interactionType, criteria)
}
