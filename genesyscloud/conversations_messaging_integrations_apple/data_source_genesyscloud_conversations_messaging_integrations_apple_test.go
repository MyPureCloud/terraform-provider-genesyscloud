package conversations_messaging_integrations_apple

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v178/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// TestAccDataSourceAppleIntegration tests the apple integration data source.
// Uses APPLE_MESSAGES_BUSINESS_ID environment variable for real business ID, otherwise uses fake ID.
// - Fake ID: Creates integration with incomplete status (expected for testing)
// - Real ID: Creates fully functional integration for comprehensive testing
func TestAccDataSourceAppleIntegration(t *testing.T) {
	if !checkAppleIntegrationEndpointsEnabled() {
		t.Skip("Skipping test as apple integration endpoints are not enabled")
	}
	var (
		resourceLabel   = "test-apple-integration"
		dataSourceLabel = "test-apple-integration-data"
		randomString    = uuid.NewString()
		integrationName = "Test Apple Integration " + randomString
		businessId      = getTestBusinessId()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateBasicAppleIntegrationResource(
					resourceLabel,
					integrationName,
					businessId,
				) + generateAppleIntegrationDataSource(
					dataSourceLabel,
					resourceName+"."+resourceLabel+".name",
					resourceName+"."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+resourceName+"."+dataSourceLabel, "id", resourceName+"."+resourceLabel, "id"),
				),
			},
		},
		CheckDestroy: testVerifyAppleIntegrationDeleted,
	})
}

func generateBasicAppleIntegrationResource(resourceLabel, name, businessId string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		messages_for_business_id = "%s"
	}
	`, resourceName, resourceLabel, name, businessId)
}

func generateAppleIntegrationDataSource(resourceLabel, name, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = %s
		depends_on = [%s]
	}
	`, resourceName, resourceLabel, name, dependsOnResource)
}

// getTestBusinessId returns business ID for testing:
// - If APPLE_MESSAGES_BUSINESS_ID env var is set: returns real business ID
// - Otherwise: returns fake "test-business-{uuid}" for error scenario testing
func getTestBusinessId() string {
	if businessId := os.Getenv("APPLE_MESSAGES_BUSINESS_ID"); businessId != "" {
		return businessId
	}
	// Return fake business ID - creates integration with incomplete status for error testing
	return "test-business-" + uuid.NewString()
}

// isUsingRealBusinessId returns true when APPLE_MESSAGES_BUSINESS_ID environment variable is set
// Used to conditionally expect errors (fake ID) vs success (real ID) in tests
func isUsingRealBusinessId() bool {
	return os.Getenv("APPLE_MESSAGES_BUSINESS_ID") != ""
}

func testVerifyAppleIntegrationDeleted(state *terraform.State) error {
	conversationApi := platformclientv2.NewConversationsApi()
	var integrationId string

	// Find the Apple integration ID from state
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_conversations_messaging_integrations_apple" {
			integrationId = rs.Primary.ID
			break
		}
	}

	if integrationId == "" {
		return fmt.Errorf("Apple integration ID not found in state")
	}

	// Retry for up to 120 seconds, checking if the integration is deleted
	if err := util.WithRetries(context.Background(), 120*time.Second, func() *retry.RetryError {

		integration, resp, err := conversationApi.GetConversationsMessagingIntegrationsAppleIntegrationId(integrationId, "")

		if integration != nil {
			// Still exists → retry
			return retry.RetryableError(fmt.Errorf("Apple integration (%s) still exists", integrationId))
		}

		if util.IsStatus404(resp) || util.IsStatus400(resp) {
			// Deleted successfully
			return nil
		}

		// Any unexpected error → non-retryable failure
		return retry.NonRetryableError(fmt.Errorf("unexpected error: %v", err))

	}); err != nil {
		return fmt.Errorf("error verifying Apple integration deletion: %v", err)
	}

	return nil
}
