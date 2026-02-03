package intent_category

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_intent_category_test.go contains all of the test cases for running the resource
tests for intent_category.
*/

func TestAccResourceIntentCategory(t *testing.T) {
	t.Parallel()
	var ()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{},
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
