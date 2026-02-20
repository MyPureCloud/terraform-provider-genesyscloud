package customer_intent

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The resource_genesyscloud_customer_intent_test.go contains all of the test cases for running the resource
tests for customer_intent.
*/

func TestAccResourceCustomerIntent(t *testing.T) {
	t.Parallel()
	var (
		resourcePath     = "genesyscloud_customer_intent.test_intent"
		categoryResource = "test_category"
		intentResource   = "test_intent"
		categoryName     = "Test Category " + uuid.NewString()
		intentName1      = "Test Customer Intent " + uuid.NewString()
		intentDesc1      = "Test customer intent description"
		expiryTime1      = 24
		intentName2      = "Updated Customer Intent " + uuid.NewString()
		intentDesc2      = "Updated customer intent description"
		expiryTime2      = 48
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create customer intent
				Config: generateIntentCategoryResource(
					categoryResource,
					categoryName,
					"Test category for customer intent",
				) + generateCustomerIntentResource(
					intentResource,
					intentName1,
					intentDesc1,
					expiryTime1,
					"genesyscloud_intent_category."+categoryResource+".id",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", intentName1),
					resource.TestCheckResourceAttr(resourcePath, "description", intentDesc1),
					resource.TestCheckResourceAttr(resourcePath, "expiry_time", fmt.Sprintf("%d", expiryTime1)),
					resource.TestCheckResourceAttrPair(resourcePath, "category_id", "genesyscloud_intent_category."+categoryResource, "id"),
				),
			},
			{
				// Update customer intent
				Config: generateIntentCategoryResource(
					categoryResource,
					categoryName,
					"Test category for customer intent",
				) + generateCustomerIntentResource(
					intentResource,
					intentName2,
					intentDesc2,
					expiryTime2,
					"genesyscloud_intent_category."+categoryResource+".id",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", intentName2),
					resource.TestCheckResourceAttr(resourcePath, "description", intentDesc2),
					resource.TestCheckResourceAttr(resourcePath, "expiry_time", fmt.Sprintf("%d", expiryTime2)),
					resource.TestCheckResourceAttrPair(resourcePath, "category_id", "genesyscloud_intent_category."+categoryResource, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyCustomerIntentDestroyed,
	})
}

func testVerifyCustomerIntentDestroyed(state *terraform.State) error {
	return nil
}

// generateCustomerIntentResource generates a Terraform config string for a customer intent resource
func generateCustomerIntentResource(
	resourceLabel string,
	name string,
	description string,
	expiryTime int,
	categoryId string,
) string {
	return fmt.Sprintf(`resource "genesyscloud_customer_intent" "%s" {
		name        = "%s"
		description = "%s"
		expiry_time = %d
		category_id = %s
	}
	`, resourceLabel, name, description, expiryTime, categoryId)
}

// generateIntentCategoryResource generates a Terraform config string for an intent category resource
func generateIntentCategoryResource(
	resourceLabel string,
	name string,
	description string,
) string {
	return fmt.Sprintf(`resource "genesyscloud_intent_category" "%s" {
		name        = "%s"
		description = "%s"
	}
	`, resourceLabel, name, description)
}
