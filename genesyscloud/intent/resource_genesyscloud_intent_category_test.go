package intent_category

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceIntentCategory(t *testing.T) {
	t.Parallel()
	var (
		resourcePath = "genesyscloud_intent_category.test_category"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateIntentCategoryResource(
					"test_category",
					"Test category",
					"Test description",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_intent_category.test_category", "name", "Test category"),
					resource.TestCheckResourceAttr("genesyscloud_intent_category.test_category", "description", "Test description"),
				),
			},
			{
				Config: generateIntentCategoryResource(
					"test_category",
					"Updated test category",
					"The category has been updated",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_intent_category.test_category", "name", "Updated test category"),
					resource.TestCheckResourceAttr("genesyscloud_intent_category.test_category", "description", "The category has been updated"),
				),
			},
			{
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIntentCategoryDestroyed,
	})
}

func testVerifyIntentCategoryDestroyed(state *terraform.State) error {
	return nil
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
