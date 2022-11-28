package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
	"strings"
	"testing"
)

func TestAccResourceResponseManagementResponses(t *testing.T) {
	t.Parallel()
	var (
		responseResource = "response-resource"
		name1            = "Test response" + uuid.NewString()
		interactionTypes = []string{"chat", "email", "twitter"}
		responseTypes    = []string{`MessagingTemplate`, `CampaignSmsTemplate`, `CampaignEmailTemplate`}
		contentTypes     = []string{"text/plain", "text/html"}

		libraryResource = "library-resource"
		libraryName     = "Reference library" + uuid.NewString()
		//name2 = "Test response"+uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateResponseManagementLibraryResource(
					libraryResource,
					libraryName,
				) + generateResponseManagementResponsesResource(
					responseResource,
					name1,
					[]string{"genesyscloud_responsemanagement_library." + libraryResource + ".id"},
					interactionTypes[0],
					responseTypes[0],
					generateTextsBlock(
						uuid.NewString(),
						contentTypes[0],
					),
					generateSubstitutionsBlock(
						uuid.NewString(),
						uuid.NewString(),
					),
					generateMessagingTemplateBlock(
						generateWhatsappBlock(
							uuid.NewString(),
							uuid.NewString(),
							"en_US",
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_responsemanagement_responses." + responseResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyResponseManagementResponsesDestroyed,
	})
}

func generateResponseManagementResponsesResource(
	resourceId string,
	name string,
	libraryIds []string,
	interactionType string,
	responseType string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
		resource "genesyscloud_responsemanagement_responses" "%s" {
			name = "%s"
			library_ids = [%s]
			interaction_type = "%s"
			response_type = "%s"
			%s
		}
	`, resourceId, name, strings.Join(libraryIds, ", "), interactionType, responseType, strings.Join(nestedBlocks, "\n"))
}

func generateTextsBlock(
	content string,
	contentType string,
) string {
	return fmt.Sprintf(`
		texts {
			content = "%s"
			content_type = "%s"
		}
	`, content, contentType)
}

func generateSubstitutionsBlock(
	description string,
	defaultValue string,
) string {
	return fmt.Sprintf(`
		substitutions {
			description = "%s"
			default_value = "%s"
		}
	`, description, defaultValue)
}

func generateMessagingTemplateBlock(
	attrs ...string,
) string {
	return fmt.Sprintf(`
		messaging_template {
			%s
		}
	`, strings.Join(attrs, "\n"))
}

func generateWhatsappBlock(
	name string,
	nameSpace string,
	language string,
) string {
	return fmt.Sprintf(`
		whats_app{
			name = "%s"
			namespace = "%s"
			language = "%s"
		}
	`, name, nameSpace, language)
}

func testVerifyResponseManagementResponsesDestroyed(state *terraform.State) error {
	managementAPI := platformclientv2.NewResponseManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_responsemanagement_responses" {
			continue
		}
		responses, resp, err := managementAPI.GetResponsemanagementResponse(rs.Primary.ID, "")
		if responses != nil {
			return fmt.Errorf("response (%s) still exists", rs.Primary.ID)
		} else if isStatus404(resp) {
			// response not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All responses destroyed
	return nil
}
