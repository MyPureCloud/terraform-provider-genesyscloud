package task_management_workbin

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_task_management_workbin_test.go contains all of the test cases for running the resource
tests for task_management_workbin.
*/

func TestAccResourceTaskManagementWorkbin(t *testing.T) {
	t.Parallel()
	var (
		workbinResId    = "workbin_1"
		workbinName     = "tf_workbin_" + uuid.NewString()
		workDescription = "created for CX as Code test case"

		divisionResId1 = "div_1"
		divisionName1  = "tf_div_1_" + uuid.NewString()
		divisionResId2 = "div_2"
		divisionName2  = "tf_div_2_" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Default division
			{
				Config: GenerateWorkbinResource(workbinResId, workbinName, workDescription, nullValue) +
					"\n data \"genesyscloud_auth_division_home\" \"home\" {}",
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+workbinResId, "name", workbinName),
					resource.TestCheckResourceAttr(resourceName+"."+workbinResId, "description", workDescription),
					resource.TestCheckResourceAttrPair(resourceName+"."+workbinResId, "division_id", "data.genesyscloud_auth_division_home.home", "id"),
				),
			},
			// Change division
			{
				Config: gcloud.GenerateAuthDivisionBasic(divisionResId1, divisionName1) +
					gcloud.GenerateAuthDivisionBasic(divisionResId2, divisionName2) +
					GenerateWorkbinResource(workbinResId, workbinName, workDescription, "genesyscloud_auth_division."+divisionResId1+".id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+workbinResId, "name", workbinName),
					resource.TestCheckResourceAttr(resourceName+"."+workbinResId, "description", workDescription),
					resource.TestCheckResourceAttrPair(resourceName+"."+workbinResId, "division_id", "genesyscloud_auth_division."+divisionResId1, "id"),
				),
			},
			{
				Config: gcloud.GenerateAuthDivisionBasic(divisionResId1, divisionName1) +
					gcloud.GenerateAuthDivisionBasic(divisionResId2, divisionName2) +
					GenerateWorkbinResource(workbinResId, workbinName, workDescription, "genesyscloud_auth_division."+divisionResId2+".id"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName+"."+workbinResId, "division_id", "genesyscloud_auth_division."+divisionResId2, "id"),
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
