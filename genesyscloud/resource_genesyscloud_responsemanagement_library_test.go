package genesyscloud

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceResponseManagementLibrary(t *testing.T) {

	var (
		libraryResource = "response_management_library"
		name1           = "Library " + uuid.NewString()
		name2           = "Library " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
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
		CheckDestroy: testVerifyResponseManagementLibraryDestroyed,
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

func testVerifyResponseManagementLibraryDestroyed(state *terraform.State) error {
	responseAPI := platformclientv2.NewResponseManagementApi()

	diagErr := WithRetries(context.Background(), 180*time.Second, func() *retry.RetryError {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "genesyscloud_responsemanagement_library" {
				continue
			}
			_, resp, err := responseAPI.GetResponsemanagementLibrary(rs.Primary.ID)
			if err != nil {
				if IsStatus404(resp) {
					continue
				}
				return retry.NonRetryableError(fmt.Errorf("Unexpected error: %s", err))
			}

			return retry.RetryableError(fmt.Errorf("Library %s still exists", rs.Primary.ID))
		}
		return nil
	})

	if diagErr != nil {
		return fmt.Errorf(fmt.Sprintf("%v", diagErr))
	}

	// Success. All Libraries destroyed
	return nil
}
