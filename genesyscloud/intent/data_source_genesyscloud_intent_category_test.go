package intent_category

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
Test Class for the intent category Data Source
*/

func TestAccDataSourceIntentCategoryBasic(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel    = "test-intent-category"
		dataSourceLabel  = "test-intent-category-ds"
		categoryName     = "Terraform Intent Category DS " + uuid.NewString()
		categoryDesc     = "Test intent category for data source"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create resource and data source
				Config: generateIntentCategoryResource(
					resourceLabel,
					categoryName,
					categoryDesc,
				) + generateIntentCategoryDataSource(
					dataSourceLabel,
					categoryName,
					"genesyscloud_intent_category."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_intent_category."+dataSourceLabel, "id",
						"genesyscloud_intent_category."+resourceLabel, "id",
					),
				),
			},
		},
		CheckDestroy: testVerifyIntentCategoryDestroyed,
	})
}

func TestAccDataSourceIntentCategoryNotFound(t *testing.T) {
	t.Parallel()
	var (
		dataSourceLabel = "test-intent-category-ds-not-found"
		categoryName    = "Non-existent Intent Category " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search for non-existent category
				Config: generateIntentCategoryDataSource(
					dataSourceLabel,
					categoryName,
					"", // No dependency
				),
				ExpectError: regexp.MustCompile(fmt.Sprintf("No intent category found with name %s", regexp.QuoteMeta(categoryName))),
			},
		},
	})
}

// generateIntentCategoryDataSource generates a Terraform config string for an intent category data source
func generateIntentCategoryDataSource(
	dataSourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOn string,
) string {
	if dependsOn != "" {
		return fmt.Sprintf(`data "genesyscloud_intent_category" "%s" {
		name       = "%s"
		depends_on = [%s]
	}
	`, dataSourceLabel, name, dependsOn)
	}
	return fmt.Sprintf(`data "genesyscloud_intent_category" "%s" {
		name = "%s"
	}
	`, dataSourceLabel, name)
}
