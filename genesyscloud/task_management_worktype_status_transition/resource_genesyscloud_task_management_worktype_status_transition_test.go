package task_management_worktype_status_transition

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	workType "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The resource_genesyscloud_task_management_worktype_status_test.go contains all of the test cases for running the resource
tests for task_management_worktype_status.
*/

/*
	setup worktype status and workbin outside the test step , since the genesyscloud_worktype_status_transition does not

does not create a new resource but will update the existing  genesyscloud_worktype_status and we get plan non-empty errors for
genesyscloud_worktype_status. referring  it as data source in a different test step too will not work since the resources are
independent in each test step.
*/
func TestAccResourceTaskManagementWorktypeStatusTransition(t *testing.T) {
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

		// Status 2
		statusResourceLabel2 = "status2"
		status2Name          = "status2-" + uuid.NewString()
		status2Category      = "Closed"

		statusResourceType = "genesyscloud_task_management_worktype_status"
	)

	initialConfig := workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
		workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceLabel, wsName, wsDescription) +
		workType.GenerateWorktypeResourceBasic(
			wtResourceLabel,
			wtName,
			wtDescription,
			fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
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
		) +

		// This status is used as a reference in the first status
		GenerateWorktypeStatusResourceWithDependsOn(
			statusResourceLabel2,
			fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
			status2Name,
			status2Category,
			"",
			util.NullValue,
			"",
			"genesyscloud_task_management_worktype_status"+"."+statusResourceLabel1,
			"default = false",
		)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Update worktype status and add another status so we can test destination_status_ids and default_destination_status_id
				Config: initialConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(statusResourceType+"."+statusResourceLabel1, "worktype_id", fmt.Sprintf("genesyscloud_task_management_worktype.%s", wtResourceLabel), "id"),
					resource.TestCheckResourceAttr(statusResourceType+"."+statusResourceLabel1, "name", status1Name1),
					resource.TestCheckResourceAttr(statusResourceType+"."+statusResourceLabel1, "category", status1Category),
					resource.TestCheckResourceAttr(statusResourceType+"."+statusResourceLabel1, "description", ""),
				),
			},
			{
				// Update worktype status and add another status so we can test destination_status_ids and default_destination_status_id
				Config: initialConfig +
					generateWorktypeStatusDataSourceForTransition(
						statusResourceLabel1,
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						status1Name1,
					) +
					GenerateWorkTypeStatusResourceTransition(
						statusResourceLabel1+"transition",
						fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s.id", statusResourceLabel1),
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s.id", statusResourceLabel2),
						fmt.Sprintf("genesyscloud_task_management_worktype_status.%s.id", statusResourceLabel2),
						"60",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(statusResourceType+"."+statusResourceLabel1, "worktype_id", fmt.Sprintf("genesyscloud_task_management_worktype.%s", wtResourceLabel), "id"),
					resource.TestCheckResourceAttr(statusResourceType+"."+statusResourceLabel1, "name", status1Name1),
					resource.TestCheckResourceAttr(statusResourceType+"."+statusResourceLabel1, "category", status1Category),
					resource.TestCheckResourceAttr(statusResourceType+"."+statusResourceLabel1, "description", ""),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_task_management_worktype_status_transition." + statusResourceLabel1 + "transition",
				ImportState:       true,
				ImportStateVerify: false,
				//ImportStateVerifyIgnore: true,
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

		worktypeId, statusId := splitWorktypeStatusTerraformTransitionId(res.Primary.ID)
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

	//All worktype statuses deleted
	return nil
}
