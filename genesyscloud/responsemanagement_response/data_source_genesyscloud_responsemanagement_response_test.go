package responsemanagement_response

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	respmanagementLibrary "terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceResponsemanagementResponse(t *testing.T) {
	var (
		responseResourceLabel = "response-resource"
		responseDataLabel     = "response-data"
		name                  = "Response-" + uuid.NewString()
		textsContent          = "Random text block content string"
		textsContentTypes     = []string{"text/plain", "text/html"}

		// Library resources variables
		libraryResourceLabel = "library-resource1"
		libraryName          = "Reference library1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Search by name
				Config: respmanagementLibrary.GenerateResponseManagementLibraryResource(
					libraryResourceLabel,
					libraryName,
				) + generateResponseManagementResponseResource(
					responseResourceLabel,
					name,
					[]string{"genesyscloud_responsemanagement_library." + libraryResourceLabel + ".id"},
					util.NullValue,
					util.NullValue,
					util.NullValue,
					[]string{},
					generateTextsBlock(
						textsContent,
						textsContentTypes[0],
						util.NullValue,
					),
				) + generateResponsemanagementResponseDataSource(
					responseDataLabel,
					name,
					"genesyscloud_responsemanagement_library."+libraryResourceLabel+".id",
					"genesyscloud_responsemanagement_response."+responseResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_responsemanagement_response."+responseDataLabel, "id",
						"genesyscloud_responsemanagement_response."+responseResourceLabel, "id",
					),
				),
			},
		},
		CheckDestroy: testVerifyResponseManagementResponseDestroyed,
	})
}

func generateResponsemanagementResponseDataSource(
	resourceLabel string,
	name string,
	libraryID string,
	dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_responsemanagement_response" "%s" {
			name = "%s"
			library_id = %s
			depends_on=[%s]
		}
	`, resourceLabel, name, libraryID, dependsOn)
}
