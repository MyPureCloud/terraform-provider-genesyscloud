package conversations_messaging_integrations_apple

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// TestAccDataSourceAppleIntegration tests the apple integration data source.
// Uses APPLE_MESSAGES_BUSINESS_ID environment variable for real business ID, otherwise uses fake ID.
// - Fake ID: Creates integration with incomplete status (expected for testing)
// - Real ID: Creates fully functional integration for comprehensive testing
func TestAccDataSourceAppleIntegration(t *testing.T) {
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
					ResourceType+"."+resourceLabel+".name",
					ResourceType+"."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+dataSourceLabel, "id", ResourceType+"."+resourceLabel, "id"),
				),
			},
		},
	})
}

func generateBasicAppleIntegrationResource(resourceLabel, name, businessId string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		messages_for_business_id = "%s"
	}
	`, ResourceType, resourceLabel, name, businessId)
}

func generateAppleIntegrationDataSource(resourceLabel, name, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = %s
		depends_on = [%s]
	}
	`, ResourceType, resourceLabel, name, dependsOnResource)
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
