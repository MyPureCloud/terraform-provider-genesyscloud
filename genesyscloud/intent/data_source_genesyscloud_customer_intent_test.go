package customer_intent

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
Test Class for the customer intent Data Source
*/

func TestAccDataSourceCustomerIntentBasic(t *testing.T) {
	t.Parallel()
	var (
		categoryResource = "test_category"
		resourceLabel    = "test-customer-intent"
		dataSourceLabel  = "test-customer-intent-ds"
		categoryName     = "Test Category " + uuid.NewString()
		intentName       = "Terraform Customer Intent DS " + uuid.NewString()
		intentDesc       = "Test customer intent for data source"
		expiryTime       = 24
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create resource and data source
				Config: generateIntentCategoryResource(
					categoryResource,
					categoryName,
					"Test category for customer intent",
				) + generateCustomerIntentResource(
					resourceLabel,
					intentName,
					intentDesc,
					expiryTime,
					"genesyscloud_intent_category."+categoryResource+".id",
				) + generateCustomerIntentDataSource(
					dataSourceLabel,
					intentName,
					"genesyscloud_customer_intent."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_customer_intent."+dataSourceLabel, "id",
						"genesyscloud_customer_intent."+resourceLabel, "id",
					),
				),
			},
		},
		CheckDestroy: testVerifyCustomerIntentDestroyed,
	})
}

func TestAccDataSourceCustomerIntentNotFound(t *testing.T) {
	t.Parallel()
	var (
		dataSourceLabel = "test-customer-intent-ds-not-found"
		intentName      = "Non-existent Customer Intent " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search for non-existent customer intent
				Config: generateCustomerIntentDataSource(
					dataSourceLabel,
					intentName,
					"", // No dependency
				),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No customer intent found with name %s", regexp.QuoteMeta(intentName))),
			},
		},
	})
}

// generateCustomerIntentDataSource generates a Terraform config string for a customer intent data source
func generateCustomerIntentDataSource(
	dataSourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOn string,
) string {
	if dependsOn != "" {
		return fmt.Sprintf(`data "genesyscloud_customer_intent" "%s" {
		name       = "%s"
		depends_on = [%s]
	}
	`, dataSourceLabel, name, dependsOn)
	}
	return fmt.Sprintf(`data "genesyscloud_customer_intent" "%s" {
		name = "%s"
	}
	`, dataSourceLabel, name)
}
