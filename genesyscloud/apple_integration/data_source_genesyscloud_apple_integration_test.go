package apple_integration

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccDataSourceAppleIntegration(t *testing.T) {
	var (
		resourceLabel    = "test-apple-integration"
		dataSourceLabel  = "test-apple-integration-data"
		randomString     = uuid.NewString()
		integrationName  = "Test Apple Integration " + randomString
		businessId       = "test-business-" + randomString
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