package responsemanagement_response

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	respmanagementLibrary "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// TestAccResourceResponseManagementResponseWhatsappCarousel needs an org with a WhatsApp integration and the
// PURE-6496 carousel toggle enabled. It uses WHATSAPP_INTEGRATION_ID when set, otherwise the org's first
// WhatsApp integration, and is skipped rather than failed when the org has none.
func TestAccResourceResponseManagementResponseWhatsappCarousel(t *testing.T) {
	var (
		responseResourceLabel = "response-carousel"
		responseResourceName  = ResourceType + "." + responseResourceLabel
		libraryResourceLabel  = "library-carousel"
		libraryName           = "ReferencelibraryCarousel-" + uuid.NewString()
		responseName          = "Response-" + uuid.NewString()
		// Carousel template names may only contain lowercase letters, numbers and underscores.
		templateName1 = "tf_carousel_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
		templateName2 = "tf_carousel_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
		text          = "Check out this week's offers"
	)

	// Step configs are rendered before PreCheck runs, so the integration has to be resolved up front.
	util.TestAccPreCheck(t)
	integrationId := getTestWhatsappIntegrationId(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create a two-card image carousel
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(libraryResourceLabel, libraryName) +
					generateCarouselResponse(responseResourceLabel, responseName, libraryResourceLabel, integrationId, templateName1, text, "Buy now"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(responseResourceName, "response_type", "MessagingTemplate"),
					resource.TestCheckResourceAttr(responseResourceName, "messaging_template.0.whats_app.0.name", templateName1),
					resource.TestCheckResourceAttr(responseResourceName, "messaging_template.0.whats_app.0.integration_id", integrationId),
					resource.TestCheckResourceAttr(responseResourceName, "messaging_template.0.whats_app.0.category", "marketing"),
					resource.TestCheckResourceAttrSet(responseResourceName, "messaging_template.0.whats_app.0.template_status"),
					resource.TestCheckResourceAttr(responseResourceName, "messaging_template.0.whats_app.0.carousel.0.cards.#", "2"),
					resource.TestCheckResourceAttr(responseResourceName, "messaging_template.0.whats_app.0.carousel.0.cards.0.header.0.type", "Image"),
					resource.TestCheckResourceAttr(responseResourceName, "messaging_template.0.whats_app.0.carousel.0.cards.0.buttons.0.payload", "card-1"),
					resource.TestCheckResourceAttr(responseResourceName, "messaging_template.0.whats_app.0.carousel.0.cards.1.buttons.0.payload", "card-2"),
					resource.TestCheckResourceAttr(responseResourceName, "messaging_template.0.whats_app.0.carousel.0.cards.1.card_body_text.0.content", "Second card"),
				),
			},
			{
				// Changing a locked field must plan a replacement instead of failing with a 400 on apply.
				// If WhatsApp review already moved the template to ERROR the update happens in place, which is
				// also valid, so only the new values are asserted here; the replace decision is unit tested.
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(libraryResourceLabel, libraryName) +
					generateCarouselResponse(responseResourceLabel, responseName, libraryResourceLabel, integrationId, templateName2, text, "Shop now"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(responseResourceName, "messaging_template.0.whats_app.0.name", templateName2),
					resource.TestCheckResourceAttr(responseResourceName, "messaging_template.0.whats_app.0.carousel.0.cards.0.buttons.0.content_text", "Shop now"),
					resource.TestCheckResourceAttrSet(responseResourceName, "messaging_template.0.whats_app.0.template_status"),
				),
			},
			{
				ResourceName:      responseResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// template_status and status_info are updated asynchronously by WhatsApp review.
				ImportStateVerifyIgnore: []string{
					"messaging_template.0.whats_app.0.template_status",
					"messaging_template.0.whats_app.0.status_info",
				},
			},
		},
		CheckDestroy: testVerifyResponseManagementResponseDestroyed,
	})
}

func generateCarouselResponse(resourceLabel, name, libraryLabel, integrationId, templateName, text, buttonText string) string {
	return fmt.Sprintf(`
resource "genesyscloud_responsemanagement_response" "%s" {
  name          = "%s"
  library_ids   = [genesyscloud_responsemanagement_library.%s.id]
  response_type = "MessagingTemplate"
  texts {
    content      = "%s"
    content_type = "text/plain"
  }
  messaging_template {
    whats_app {
      name           = "%s"
      language       = "en_US"
      integration_id = "%s"
      category       = "marketing"
      carousel {
        cards {
          header {
            type    = "Image"
            content = "https://example.com/card-1.jpg"
          }
          buttons {
            type         = "QuickReply"
            content_text = "%s"
            payload      = "card-1"
          }
          card_body_text {
            content = "First card"
          }
        }
        cards {
          header {
            type    = "Image"
            content = "https://example.com/card-2.jpg"
          }
          buttons {
            type         = "QuickReply"
            content_text = "%s"
            payload      = "card-2"
          }
          card_body_text {
            content = "Second card"
          }
        }
      }
    }
  }
}
`, resourceLabel, name, libraryLabel, text, templateName, integrationId, buttonText, buttonText)
}

func getTestWhatsappIntegrationId(t *testing.T) string {
	if id := os.Getenv("WHATSAPP_INTEGRATION_ID"); id != "" {
		return id
	}
	config, err := provider.AuthorizeSdk()
	if err != nil {
		t.Fatalf("failed to authorize SDK: %v", err)
	}
	api := platformclientv2.NewConversationsApiWithConfig(config)
	integrations, resp, err := api.GetConversationsMessagingIntegrationsWhatsapp(25, 1, "", "", "")
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		t.Fatalf("failed to list WhatsApp integrations (status %d): %v", statusCode, err)
	}
	if integrations.Entities != nil {
		for _, integration := range *integrations.Entities {
			if integration.Id != nil {
				return *integration.Id
			}
		}
	}
	t.Skip("skipping: no WhatsApp integration in the test org; carousel templates require one")
	return ""
}
