package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccResourceResponseManagementLibrary(t *testing.T) {

	var (
		libraryResource = "response_management_library"
		name1           = "Library " + uuid.NewString()
		name2           = "Library " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateResponseManagementLibraryResource(libraryResource, name1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_library."+libraryResource, "name", name1),
				),
			},
			{
				// Update
				Config: generateResponseManagementLibraryResource(libraryResource, name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_library."+libraryResource, "name", name2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_responsemanagement_library." + libraryResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateResponseManagementLibraryResource(
	resourceId string,
	name string) string {
	return fmt.Sprintf(`
		resource "genesyscloud_responsemanagement_library" "%s" {
			name = "%s"
		}
	`, resourceId, name)
}
