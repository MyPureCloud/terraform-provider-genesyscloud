package customer_intent

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
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
	intentsApi := platformclientv2.NewIntentsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_customer_intent" {
			continue
		}

		customerIntent, resp, err := intentsApi.GetIntentsCustomerintent(rs.Primary.ID)
		if customerIntent != nil {
			return fmt.Errorf("Customer intent (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
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

func TestAccResourceCustomerIntentWithSourceIntents(t *testing.T) {
	t.Parallel()
	var (
		resourcePath      = "genesyscloud_customer_intent.test_intent_with_sources"
		categoryResource  = "test_category_src"
		intentResource    = "test_intent_with_sources"
		sourceResource1   = "source_intent_1"
		sourceResource2   = "source_intent_2"
		sourceResource3   = "source_intent_3"
		categoryName      = "Test Category " + uuid.NewString()
		intentName        = "Test Customer Intent with Sources " + uuid.NewString()
		intentDesc        = "Test customer intent with source intents"
		expiryTime        = 24
		sourceIntentName1 = "Source Intent 1 " + uuid.NewString()
		sourceIntentName2 = "Source Intent 2 " + uuid.NewString()
		sourceIntentName3 = "Source Intent 3 " + uuid.NewString()
	)

	// Helper to build the base config: category + 3 source customer intents
	baseConfig := func() string {
		return generateIntentCategoryResource(categoryResource, categoryName, "Test category for source intents") +
			generateCustomerIntentResource(sourceResource1, sourceIntentName1, "Source intent 1", expiryTime,
				"genesyscloud_intent_category."+categoryResource+".id") +
			generateCustomerIntentResource(sourceResource2, sourceIntentName2, "Source intent 2", expiryTime,
				"genesyscloud_intent_category."+categoryResource+".id") +
			generateCustomerIntentResource(sourceResource3, sourceIntentName3, "Source intent 3", expiryTime,
				"genesyscloud_intent_category."+categoryResource+".id")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create customer intent with 2 source intents referencing real customer intent resources
				Config: baseConfig() + generateCustomerIntentWithSourceIntents(
					intentResource,
					intentName,
					intentDesc,
					expiryTime,
					"genesyscloud_intent_category."+categoryResource+".id",
					[]sourceIntentConfig{
						{
							sourceIntentId:   "genesyscloud_customer_intent." + sourceResource1 + ".id",
							sourceIntentName: sourceIntentName1,
							sourceType:       "Topic",
						},
						{
							sourceIntentId:   "genesyscloud_customer_intent." + sourceResource2 + ".id",
							sourceIntentName: sourceIntentName2,
							sourceType:       "Topic",
						},
					},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", intentName),
					resource.TestCheckResourceAttr(resourcePath, "description", intentDesc),
					resource.TestCheckResourceAttr(resourcePath, "expiry_time", fmt.Sprintf("%d", expiryTime)),
					resource.TestCheckResourceAttr(resourcePath, "source_intents.#", "2"),
				),
			},
			{
				// Update source intents - swap source1 for source3
				Config: baseConfig() + generateCustomerIntentWithSourceIntents(
					intentResource,
					intentName,
					intentDesc,
					expiryTime,
					"genesyscloud_intent_category."+categoryResource+".id",
					[]sourceIntentConfig{
						{
							sourceIntentId:   "genesyscloud_customer_intent." + sourceResource2 + ".id",
							sourceIntentName: sourceIntentName2,
							sourceType:       "Topic",
						},
						{
							sourceIntentId:   "genesyscloud_customer_intent." + sourceResource3 + ".id",
							sourceIntentName: sourceIntentName3,
							sourceType:       "Topic",
						},
					},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", intentName),
					resource.TestCheckResourceAttr(resourcePath, "source_intents.#", "2"),
				),
			},
			{
				// Remove all source intents
				Config: baseConfig() + generateCustomerIntentResource(
					intentResource,
					intentName,
					intentDesc,
					expiryTime,
					"genesyscloud_intent_category."+categoryResource+".id",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", intentName),
					resource.TestCheckResourceAttr(resourcePath, "source_intents.#", "0"),
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

type sourceIntentConfig struct {
	sourceIntentId   string
	sourceIntentName string
	sourceType       string
	sourceId         string
	sourceName       string
}

// generateCustomerIntentWithSourceIntents generates a Terraform config string for a customer intent with source intents
func generateCustomerIntentWithSourceIntents(
	resourceLabel string,
	name string,
	description string,
	expiryTime int,
	categoryId string,
	sourceIntents []sourceIntentConfig,
) string {
	config := fmt.Sprintf(`resource "genesyscloud_customer_intent" "%s" {
		name        = "%s"
		description = "%s"
		expiry_time = %d
		category_id = %s
`, resourceLabel, name, description, expiryTime, categoryId)

	for _, si := range sourceIntents {
		block := fmt.Sprintf(`
		source_intents {
			source_intent_id   = %s
			source_intent_name = "%s"
			source_type        = "%s"
`, si.sourceIntentId, si.sourceIntentName, si.sourceType)
		if si.sourceId != "" {
			block += fmt.Sprintf("\t\t\tsource_id = %s\n", si.sourceId)
		}
		if si.sourceName != "" {
			block += fmt.Sprintf("\t\t\tsource_name = \"%s\"\n", si.sourceName)
		}
		block += "\t\t}\n"
		config += block
	}

	config += "\t}\n"
	return config
}
