package task_management_workbin

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

/*
The resource_genesyscloud_task_management_workbin_test.go contains all of the test cases for running the resource
tests for task_management_workbin.
*/

func TestAccResourceTaskManagementWorkbin(t *testing.T) {
	t.Parallel()
	var (
		workbinResourceLabel = "workbin_1"
		workbinName          = "tf_workbin_" + uuid.NewString()
		workDescription      = "created for CX as Code test case"

		divisionResourceLabel1 = "div_1"
		divisionName1          = "tf_div_1_" + uuid.NewString()
		divisionResourceLabel2 = "div_2"
		divisionName2          = "tf_div_2_" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Default division
			{
				Config: GenerateWorkbinResource(workbinResourceLabel, workbinName, workDescription, nullValue) +
					"\n data \"genesyscloud_auth_division_home\" \"home\" {}",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+workbinResourceLabel, "name", workbinName),
					resource.TestCheckResourceAttr(ResourceType+"."+workbinResourceLabel, "description", workDescription),
					resource.TestCheckResourceAttrPair(ResourceType+"."+workbinResourceLabel, "division_id", "data.genesyscloud_auth_division_home.home", "id"),
				),
			},
			// Change division
			{
				Config: authDivision.GenerateAuthDivisionBasic(divisionResourceLabel1, divisionName1) +
					authDivision.GenerateAuthDivisionBasic(divisionResourceLabel2, divisionName2) +
					GenerateWorkbinResource(workbinResourceLabel, workbinName, workDescription, "genesyscloud_auth_division."+divisionResourceLabel1+".id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+workbinResourceLabel, "name", workbinName),
					resource.TestCheckResourceAttr(ResourceType+"."+workbinResourceLabel, "description", workDescription),
					resource.TestCheckResourceAttrPair(ResourceType+"."+workbinResourceLabel, "division_id", "genesyscloud_auth_division."+divisionResourceLabel1, "id"),
				),
			},
			{
				Config: authDivision.GenerateAuthDivisionBasic(divisionResourceLabel1, divisionName1) +
					authDivision.GenerateAuthDivisionBasic(divisionResourceLabel2, divisionName2) +
					GenerateWorkbinResource(workbinResourceLabel, workbinName, workDescription, "genesyscloud_auth_division."+divisionResourceLabel2+".id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ResourceType+"."+workbinResourceLabel, "division_id", "genesyscloud_auth_division."+divisionResourceLabel2, "id"),
				),
			},
		},
		CheckDestroy: testVerifyTaskManagementWorkbinDestroyed,
	})
}

func testVerifyTaskManagementWorkbinDestroyed(state *terraform.State) error {
	taskMgmtApi := platformclientv2.NewTaskManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_task_management_workbin" {
			continue
		}

		workbin, resp, err := taskMgmtApi.GetTaskmanagementWorkbin(rs.Primary.ID)
		if workbin != nil {
			return fmt.Errorf("Task management workbin (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Workbin not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All workbins destroyed
	return nil
}
