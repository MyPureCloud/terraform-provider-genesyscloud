package task_management_worktype_status

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	workType "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_task_management_worktype_status_test.go contains all of the test cases for running the resource
tests for task_management_worktype_status.
*/

func TestAccResourceTaskManagementWorktypeStatus(t *testing.T) {
	t.Parallel()
	var (
		// Workbin
		wbResourceId  = "workbin_1"
		wbName        = "wb_" + uuid.NewString()
		wbDescription = "workbin created for CX as Code test case"

		// Schema
		wsResourceId  = "schema_1"
		wsName        = "ws_" + uuid.NewString()
		wsDescription = "workitem schema created for CX as Code test case"

		// Worktype
		wtResourceId  = "worktype_id"
		wtName        = "wt_" + uuid.NewString()
		wtDescription = "test worktype description"

		// Status 1
		statusResource1    = "status1"
		status1Name1       = "status1-" + uuid.NewString()
		status1Category    = "Open"
		status1Name2       = "status1-" + uuid.NewString()
		status1Description = "test description"

		// Status 2
		statusResource2 = "status2"
		status2Name     = "status2-" + uuid.NewString()
		status2Category = "Closed"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create worktype status
				Config: workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
					workType.GenerateWorktypeResourceBasic(
						wtResourceId,
						wtName,
						wtDescription,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceId),
						fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceId),
						"",
					) +
					GenerateWorktypeStatusResource(
						statusResource1,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceId),
						status1Name1,
						status1Category,
						"",
						util.NullValue,
						"",
						"default = true",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName+"."+statusResource1, "worktype_id", fmt.Sprintf("genesyscloud_task_management_worktype.%s", wtResourceId), "id"),
					resource.TestCheckResourceAttr(resourceName+"."+statusResource1, "name", status1Name1),
					resource.TestCheckResourceAttr(resourceName+"."+statusResource1, "category", status1Category),
					resource.TestCheckResourceAttr(resourceName+"."+statusResource1, "status_transition_delay_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName+"."+statusResource1, "default", util.TrueValue),
				),
			},
			{
				// Update worktype status and add another status so we can test destination_status_ids and default_destination_status_id
				Config: workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
					workType.GenerateWorktypeResourceBasic(
						wtResourceId,
						wtName,
						wtDescription,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceId),
						fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceId),
						"",
					) +
					GenerateWorktypeStatusResource(
						statusResource1,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceId),
						status1Name2,
						status1Category,
						status1Description,
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s.id", statusResource2),
						"12:04:21",
						generateDestinationStatusIdsArray([]string{fmt.Sprintf("genesyscloud_task_management_worktype_status.%s.id", statusResource2)}),
						fmt.Sprintf("status_transition_delay_seconds = %d", 90000),
						"default = false",
					) +
					// This status is used as a reference in the first status
					GenerateWorktypeStatusResource(
						statusResource2,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceId),
						status2Name,
						status2Category,
						"",
						util.NullValue,
						"",
						"default = true",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName+"."+statusResource1, "worktype_id", fmt.Sprintf("genesyscloud_task_management_worktype.%s", wtResourceId), "id"),
					resource.TestCheckResourceAttr(resourceName+"."+statusResource1, "name", status1Name2),
					resource.TestCheckResourceAttr(resourceName+"."+statusResource1, "category", status1Category),
					resource.TestCheckResourceAttr(resourceName+"."+statusResource1, "description", status1Description),
					ValidateStatusIds(resourceName+"."+statusResource1, "destination_status_ids.0", fmt.Sprintf("%s.%s", resourceName, statusResource2), "id"),
					ValidateStatusIds(resourceName+"."+statusResource1, "default_destination_status_id", fmt.Sprintf("%s.%s", resourceName, statusResource2), "id"),
					resource.TestCheckResourceAttr(resourceName+"."+statusResource1, "status_transition_delay_seconds", "90000"),
					resource.TestCheckResourceAttr(resourceName+"."+statusResource1, "status_transition_time", "12:04:21"),
					resource.TestCheckResourceAttr(resourceName+"."+statusResource1, "default", util.FalseValue),
					resource.TestCheckResourceAttr(resourceName+"."+statusResource2, "default", util.TrueValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_task_management_worktype_status." + statusResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTaskManagementWorktypeStatusDestroyed,
	})
}

func testVerifyTaskManagementWorktypeStatusDestroyed(state *terraform.State) error {
	taskManagementApi := platformclientv2.NewTaskManagementApi()
	for _, res := range state.RootModule().Resources {
		if res.Type != resourceName {
			continue
		}

		worktypeId, statusId := SplitWorktypeStatusTerraformId(res.Primary.ID)
		worktypeStatus, resp, err := taskManagementApi.GetTaskmanagementWorktypeStatus(worktypeId, statusId)
		if worktypeStatus != nil {
			return fmt.Errorf("task management worktype status (%s) still exists", res.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Worktype no found, as expected
			continue
		} else {
			return fmt.Errorf("unexpected error: %s", err)
		}
	}

	// All worktype statuses deleted
	return nil
}

func generateDestinationStatusIdsArray(destinationIds []string) string {
	return fmt.Sprintf(`destination_status_ids = [%s]`, strings.Join(destinationIds, ", "))
}
