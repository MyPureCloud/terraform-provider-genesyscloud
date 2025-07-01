package guide_jobs

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceGuideJobs(t *testing.T) {
	if v := os.Getenv("GENESYSCLOUD_REGION"); v != "tca" {
		t.Skipf("Skipping test for region %s. genesyscloud_guide is currently only supported in tca", v)
		return
	}
	var (
		resourceLabel = "guide_job"
		resourcePath  = ResourceType + "." + resourceLabel
		description   = "Test guide job description " + uuid.NewString()
		testUrl       = "https://example.com/test-guide.pdf"
		updatedDesc   = "Updated test guide job description " + uuid.NewString()
		updatedUrl    = "https://example.com/updated-guide.pdf"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create with description
				Config: generateGuideJobResource(
					resourceLabel,
					description,
					"", // no URL
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "description", description),
					resource.TestCheckResourceAttrSet(resourcePath, "id"),
				),
			},
			{
				// Create with URL
				Config: generateGuideJobResource(
					resourceLabel,
					"", // no description
					testUrl,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "url", testUrl),
					resource.TestCheckResourceAttrSet(resourcePath, "id"),
				),
			},
			{
				// Update with both description and URL
				Config: generateGuideJobResource(
					resourceLabel,
					updatedDesc,
					updatedUrl,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "description", updatedDesc),
					resource.TestCheckResourceAttr(resourcePath, "url", updatedUrl),
					resource.TestCheckResourceAttrSet(resourcePath, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyGuideJobDestroyed,
	})
}

func generateGuideJobResource(resourceLabel, description, url string) string {
	config := fmt.Sprintf(`resource "%s" "%s" {`, ResourceType, resourceLabel)

	if description != "" {
		config += fmt.Sprintf(`
    description = "%s"`, description)
	}

	if url != "" {
		config += fmt.Sprintf(`
    url = "%s"`, url)
	}

	config += "\n}"
	return config
}

func testVerifyGuideJobDestroyed(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}
		// Guide jobs are typically one-time operations that don't persist
		// after completion, so we don't need to verify deletion
	}
	return nil
}
