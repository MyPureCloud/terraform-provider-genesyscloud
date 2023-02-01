package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceResponsemanagementResponses(t *testing.T) {
	var (
		responseResource  = "response-resource"
		responseData      = "response-data"
		name              = "Response-" + uuid.NewString()
		textsContent      = "Random text block content string"
		textsContentTypes = []string{"text/plain", "text/html"}

		// Library resources variables
		libraryResource = "library-resource1"
		libraryName     = "Reference library1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: generateResponseManagementLibraryResource(
					libraryResource,
					libraryName,
				) + generateResponseManagementResponsesResource(
					responseResource,
					name,
					[]string{"genesyscloud_responsemanagement_library." + libraryResource + ".id"},
					"",
					"",
					"",
					[]string{},
					generateTextsBlock(
						textsContent,
						textsContentTypes[0],
					),
				) + generateResponsemanagementResponsesDataSource(
					responseData,
					name,
					"genesyscloud_responsemanagement_library."+libraryResource+".id",
					"genesyscloud_responsemanagement_responses."+responseResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_responsemanagement_responses."+responseData, "id",
						"genesyscloud_responsemanagement_responses."+responseResource, "id",
					),
				),
			},
		},
	})
}

func generateResponsemanagementResponsesDataSource(
	resourceID string,
	name string,
	libraryID string,
	dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_responsemanagement_responses" "%s" {
			name = "%s"
			library_id = %s
			depends_on=[%s]
		}
	`, resourceID, name, libraryID, dependsOn)
}
