package apple_integration

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// TestAccResourceAppleIntegrationBasic tests basic CRUD operations for apple integration.
// Uses environment variable APPLE_MESSAGES_BUSINESS_ID for real business ID, otherwise uses fake ID.
// - With fake ID: Tests error handling (update fails due to incomplete async creation)
// - With real ID: Tests full CRUD operations (update succeeds)
func TestAccResourceAppleIntegrationBasic(t *testing.T) {
	var (
		resourceLabel   = "test-apple-integration"
		randomString    = uuid.NewString()
		integrationName = "Test Apple Integration " + randomString
		// Business ID source: APPLE_MESSAGES_BUSINESS_ID env var or fake "test-business-{uuid}"
		// Real ID example: "97181fc7-0454-46d6-931a-a0784641e794" (works in DEV only!)
		businessId  = getTestBusinessId()
		updatedName = "Updated Apple Integration " + randomString
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateBasicAppleIntegrationResource(
					resourceLabel,
					integrationName,
					businessId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+resourceLabel, "name", integrationName),
					resource.TestCheckResourceAttr(resourceName+"."+resourceLabel, "messages_for_business_id", businessId),
					resource.TestCheckResourceAttrSet(resourceName+"."+resourceLabel, "id"),
				),
			},
			func() resource.TestStep {
				step := resource.TestStep{
					// Update - Behavior depends on business ID type:
					// - Fake ID: Expects error (async creation incomplete)
					// - Real ID: Expects success (full CRUD validation)
					Config: generateBasicAppleIntegrationResource(
						resourceLabel,
						updatedName,
						businessId,
					),
				}
				// Conditional test behavior based on business ID source
				if !isUsingRealBusinessId() {
					// Fake ID: Test error handling for incomplete integrations
					step.ExpectError = regexp.MustCompile(`Create integration has not completed|INVALID_MESSAGES_FOR_BUSINESS_ID`)
				} else {
					// Real ID: Test successful update operation
					step.Check = resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName+"."+resourceLabel, "name", updatedName),
						resource.TestCheckResourceAttr(resourceName+"."+resourceLabel, "messages_for_business_id", businessId),
					)
				}
				return step
			}(),
			{
				// Import/Read
				ResourceName:            resourceName + "." + resourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_status", "create_error", "recipient_id", "name"},
			},
		},
	})
}
