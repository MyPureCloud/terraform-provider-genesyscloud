package guide_jobs

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		description   = "create a guide that handles customer service inquiry, greeting the customer initalliy"
		testUrl       = "https://example.com/test-guide.pdf"
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
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
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
