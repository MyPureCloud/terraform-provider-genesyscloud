package responsemanagement_library

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func TestAccResourceResponseManagementLibrary(t *testing.T) {

	var (
		libraryResourceLabel = "response_management_library"
		name1                = "Library " + uuid.NewString()
		name2                = "Library " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateResponseManagementLibraryResource(libraryResourceLabel, name1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_library."+libraryResourceLabel, "name", name1),
				),
			},
			{
				// Update
				Config: GenerateResponseManagementLibraryResource(libraryResourceLabel, name2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_responsemanagement_library."+libraryResourceLabel, "name", name2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_responsemanagement_library." + libraryResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyResponseManagementLibraryDestroyed,
	})
}

func testVerifyResponseManagementLibraryDestroyed(state *terraform.State) error {
	responseAPI := platformclientv2.NewResponseManagementApi()

	diagErr := util.WithRetries(context.Background(), 180*time.Second, func() *retry.RetryError {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "genesyscloud_responsemanagement_library" {
				continue
			}
			_, resp, err := responseAPI.GetResponsemanagementLibrary(rs.Primary.ID)
			if err != nil {
				if util.IsStatus404(resp) {
					continue
				}
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Unexpected error: %s", err), resp))
			}

			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Library %s still exists", rs.Primary.ID), resp))
		}
		return nil
	})

	if diagErr != nil {
		return fmt.Errorf(fmt.Sprintf("%v", diagErr))
	}

	// Success. All Libraries destroyed
	return nil
}
