package task_management_worktype_status

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	workType "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

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
		wbResourceLabel = "workbin_1"
		wbName          = "wb_" + uuid.NewString()
		wbDescription   = "workbin created for CX as Code test case"

		// Schema
		wsResourceLabel = "schema_1"
		wsName          = "ws_" + uuid.NewString()
		wsDescription   = "workitem schema created for CX as Code test case"

		// Worktype
		wtResourceLabel = "worktype_id"
		wtName          = "wt_" + uuid.NewString()
		wtDescription   = "test worktype description"

		// Status 1
		statusResourceLabel1 = "status1"
		status1Name1         = "status1-" + uuid.NewString()
		status1Category      = "Open"
		status1Name2         = "status1-" + uuid.NewString()
		status1Description   = "test description"

		// Status 2
		statusResourceLabel2 = "status2"
		status2Name          = "status2-" + uuid.NewString()
		status2Category      = "Closed"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create worktype status
				Config: workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceLabel, wsName, wsDescription) +
					workType.GenerateWorktypeResourceBasic(
						wtResourceLabel,
						wtName,
						wtDescription,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
						fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceLabel),
						"",
					) +
					GenerateWorktypeStatusResource(
						statusResourceLabel1,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						status1Name1,
						status1Category,
						"",
						util.NullValue,
						"",
						"default = true",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ResourceType+"."+statusResourceLabel1, "worktype_id", fmt.Sprintf("genesyscloud_task_management_worktype.%s", wtResourceLabel), "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+statusResourceLabel1, "name", status1Name1),
					resource.TestCheckResourceAttr(ResourceType+"."+statusResourceLabel1, "category", status1Category),
					resource.TestCheckResourceAttr(ResourceType+"."+statusResourceLabel1, "status_transition_delay_seconds", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+statusResourceLabel1, "default", util.TrueValue),
				),
			},
			{
				// Update worktype status and add another status so we can test destination_status_ids and default_destination_status_id
				Config: workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceLabel, wsName, wsDescription) +
					workType.GenerateWorktypeResourceBasic(
						wtResourceLabel,
						wtName,
						wtDescription,
						fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
						fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceLabel),
						"",
					) +
					GenerateWorktypeStatusResource(
						statusResourceLabel1,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						status1Name2,
						status1Category,
						status1Description,
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s.id", statusResourceLabel2),
						"12:04:21",
						generateDestinationStatusIdsArray([]string{fmt.Sprintf("genesyscloud_task_management_worktype_status.%s.id", statusResourceLabel2)}),
						fmt.Sprintf("status_transition_delay_seconds = %d", 90000),
						"default = false",
					) +
					// This status is used as a reference in the first status
					GenerateWorktypeStatusResource(
						statusResourceLabel2,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						status2Name,
						status2Category,
						"",
						util.NullValue,
						"",
						"default = true",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ResourceType+"."+statusResourceLabel1, "worktype_id", fmt.Sprintf("genesyscloud_task_management_worktype.%s", wtResourceLabel), "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+statusResourceLabel1, "name", status1Name2),
					resource.TestCheckResourceAttr(ResourceType+"."+statusResourceLabel1, "category", status1Category),
					resource.TestCheckResourceAttr(ResourceType+"."+statusResourceLabel1, "description", status1Description),
					ValidateStatusIds(ResourceType+"."+statusResourceLabel1, "destination_status_ids.0", fmt.Sprintf("%s.%s", ResourceType, statusResourceLabel2), "id"),
					ValidateStatusIds(ResourceType+"."+statusResourceLabel1, "default_destination_status_id", fmt.Sprintf("%s.%s", ResourceType, statusResourceLabel2), "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+statusResourceLabel1, "status_transition_delay_seconds", "90000"),
					resource.TestCheckResourceAttr(ResourceType+"."+statusResourceLabel1, "status_transition_time", "12:04:21"),
					resource.TestCheckResourceAttr(ResourceType+"."+statusResourceLabel1, "default", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+statusResourceLabel2, "default", util.TrueValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_task_management_worktype_status." + statusResourceLabel1,
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
		if res.Type != ResourceType {
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
