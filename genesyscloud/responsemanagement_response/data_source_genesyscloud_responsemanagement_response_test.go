package responsemanagement_response

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	respmanagementLibrary "terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	"testing"
)

func TestAccDataSourceResponsemanagementResponse(t *testing.T) {
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
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResource,
					libraryName,
				) + generateResponseManagementResponseResource(
					responseResource,
					name,
					[]string{"genesyscloud_responsemanagement_library." + libraryResource + ".id"},
					gcloud.NullValue,
					gcloud.NullValue,
					gcloud.NullValue,
					[]string{},
					generateTextsBlock(
						textsContent,
						textsContentTypes[0],
					),
				) + generateResponsemanagementResponseDataSource(
					responseData,
					name,
					"genesyscloud_responsemanagement_library."+libraryResource+".id",
					"genesyscloud_responsemanagement_response."+responseResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_responsemanagement_response."+responseData, "id",
						"genesyscloud_responsemanagement_response."+responseResource, "id",
					),
				),
			},
		},
	})
}

func generateResponsemanagementResponseDataSource(
	resourceID string,
	name string,
	libraryID string,
	dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_responsemanagement_response" "%s" {
			name = "%s"
			library_id = %s
			depends_on=[%s]
		}
	`, resourceID, name, libraryID, dependsOn)
}
