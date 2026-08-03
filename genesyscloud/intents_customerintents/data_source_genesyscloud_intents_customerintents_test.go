package intents_customerintents

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
					"genesyscloud_intents_categories."+categoryResource+".id",
				) + generateCustomerIntentDataSource(
					dataSourceLabel,
					intentName,
					"genesyscloud_intents_customerintents."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_intents_customerintents."+dataSourceLabel, "id",
						"genesyscloud_intents_customerintents."+resourceLabel, "id",
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
					"",
				),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No customer intent found with name %s", regexp.QuoteMeta(intentName))),
			},
		},
	})
}

func generateCustomerIntentDataSource(dataSourceLabel string, name string, dependsOn string) string {
	if dependsOn != "" {
		return fmt.Sprintf(`data "genesyscloud_intents_customerintents" "%s" {
		name       = "%s"
		depends_on = [%s]
	}
	`, dataSourceLabel, name, dependsOn)
	}
	return fmt.Sprintf(`data "genesyscloud_intents_customerintents" "%s" {
		name = "%s"
	}
	`, dataSourceLabel, name)
}
