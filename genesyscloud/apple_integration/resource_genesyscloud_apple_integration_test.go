package apple_integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceAppleIntegrationBasic(t *testing.T) {
	var (
		resourceLabel   = "test-apple-integration"
		randomString    = uuid.NewString()
		integrationName = "Test Apple Integration " + randomString
		businessId      = "test-business-" + randomString
		updatedName     = "Updated Apple Integration " + randomString
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
			{
				// Update
				Config: generateBasicAppleIntegrationResource(
					resourceLabel,
					updatedName,
					businessId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+resourceLabel, "name", updatedName),
					resource.TestCheckResourceAttr(resourceName+"."+resourceLabel, "messages_for_business_id", businessId),
				),
			},
			{
				// Import/Read
				ResourceName:      resourceName + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}