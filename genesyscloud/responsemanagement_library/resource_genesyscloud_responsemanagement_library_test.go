package responsemanagement_library

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

func TestAccResourceResponseManagementLibrary(t *testing.T) {

	var (
		libraryResourceLabel  = "response_management_library"
		libraryFullPath       = ResourceType + "." + libraryResourceLabel
		name1                 = "Library " + uuid.NewString()
		name2                 = "Library " + uuid.NewString()
		divisionResourceLabel = "test_div"
		divisionName          = "test tf division " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with default (home) division
				Config: GenerateResponseManagementLibraryResource(libraryResourceLabel, name1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(libraryFullPath, "name", name1),
					resource.TestCheckResourceAttrSet(libraryFullPath, "division_id"),
				),
			},
			{
				// Update name and move to an explicit division
				Config: GenerateResponseManagementLibraryResourceWithDivision(
					libraryResourceLabel, name2, "genesyscloud_auth_division."+divisionResourceLabel+".id") +
					authDivision.GenerateAuthDivisionBasic(divisionResourceLabel, divisionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(libraryFullPath, "name", name2),
					resource.TestCheckResourceAttrPair(libraryFullPath, "division_id",
						authDivision.ResourceType+"."+divisionResourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      libraryFullPath,
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
		return errors.New(fmt.Sprintf("%v", diagErr))
	}

	// Success. All Libraries destroyed
	return nil
}
