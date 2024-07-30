package task_management_worktype

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingLanguage "terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"

	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_task_management_worktype_test.go contains all of the test cases for running the resource
tests for task_management_worktype.
*/

// Basic test with create and update
func TestAccResourceTaskManagementWorktype(t *testing.T) {
	t.Parallel()
	var (
		// Home division
		divData = "home"

		// Workbin
		wbResourceId  = "workbin_1"
		wbName        = "wb_" + uuid.NewString()
		wbDescription = "workbin created for CX as Code test case"

		// Schema
		wsResourceId  = "schema_1"
		wsName        = "ws_" + uuid.NewString()
		wsDescription = "workitem schema created for CX as Code test case"

		// Queue
		queueResId = "queue_1"
		queueName  = "tf_queue_" + uuid.NewString()

		// Language
		langResId = "lang_1"
		langName  = "tf_lang_" + uuid.NewString()

		// SKills
		skillResId1   = "skill_1"
		skillResName1 = "tf_skill_1" + uuid.NewString()
		skillResId2   = "skill_2"
		skillResName2 = "tf_skill_2" + uuid.NewString()

		// Worktype
		wtRes = worktypeConfig{
			resID:            "worktype_1",
			name:             "tf_worktype_" + uuid.NewString(),
			description:      "worktype created for CX as Code test case",
			divisionId:       fmt.Sprintf("data.genesyscloud_auth_division_home.%s.id", divData),
			defaultWorkbinId: fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceId),

			defaultDurationS:    86400,
			defaultExpirationS:  86400,
			defaultDueDurationS: 86400,
			defaultPriority:     100,
			defaultTtlS:         86400,

			defaultLanguageId: fmt.Sprintf("genesyscloud_routing_language.%s.id", langResId),
			defaultQueueId:    fmt.Sprintf("genesyscloud_routing_queue.%s.id", queueResId),
			defaultSkillIds: []string{
				fmt.Sprintf("genesyscloud_routing_skill.%s.id", skillResId1),
				fmt.Sprintf("genesyscloud_routing_skill.%s.id", skillResId2),
			},
			assignmentEnabled: false,

			schemaId:      fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceId),
			schemaVersion: 1,
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Most basic config, barebones to create a worktype
			{
				Config: workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
					GenerateWorktypeResourceBasic(wtRes.resID, wtRes.name, wtRes.description, wtRes.defaultWorkbinId, wtRes.schemaId, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "name", wtRes.name),
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "description", wtRes.description),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "default_workbin_id", fmt.Sprintf("genesyscloud_task_management_workbin.%s", wbResourceId), "id"),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "schema_id", fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s", wsResourceId), "id"),
				),
			},
			// All optional properties update
			{
				Config: workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
					routingQueue.GenerateRoutingQueueResourceBasic(queueResId, queueName) +
					routingLanguage.GenerateRoutingLanguageResource(langResId, langName) +
					routingSkill.GenerateRoutingSkillResource(skillResId1, skillResName1) +
					routingSkill.GenerateRoutingSkillResource(skillResId2, skillResName2) +
					generateWorktypeResource(wtRes) +
					fmt.Sprintf("\n data \"genesyscloud_auth_division_home\" \"%s\" {}\n", divData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "name", wtRes.name),
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "description", wtRes.description),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "default_workbin_id", fmt.Sprintf("genesyscloud_task_management_workbin.%s", wbResourceId), "id"),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "schema_id", fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s", wsResourceId), "id"),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "division_id", fmt.Sprintf("data.genesyscloud_auth_division_home.%s", divData), "id"),

					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "default_duration_seconds", fmt.Sprintf("%v", wtRes.defaultDurationS)),
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "default_expiration_seconds", fmt.Sprintf("%v", wtRes.defaultExpirationS)),
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "default_due_duration_seconds", fmt.Sprintf("%v", wtRes.defaultDueDurationS)),
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "default_priority", fmt.Sprintf("%v", wtRes.defaultPriority)),
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "default_ttl_seconds", fmt.Sprintf("%v", wtRes.defaultTtlS)),

					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "default_language_id", fmt.Sprintf("genesyscloud_routing_language.%s", langResId), "id"),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "default_queue_id", fmt.Sprintf("genesyscloud_routing_queue.%s", queueResId), "id"),

					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "default_skills_ids.0", fmt.Sprintf("genesyscloud_routing_skill.%s", skillResId1), "id"),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "default_skills_ids.1", fmt.Sprintf("genesyscloud_routing_skill.%s", skillResId2), "id"),

					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "assignment_enabled", fmt.Sprintf("%v", wtRes.assignmentEnabled)),
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "schema_version", fmt.Sprintf("%v", wtRes.schemaVersion)),
				),
			},
		},
		CheckDestroy: testVerifyTaskManagementWorktypeDestroyed,
	})
}

func testVerifyTaskManagementWorktypeDestroyed(state *terraform.State) error {
	taskMgmtApi := platformclientv2.NewTaskManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_task_management_worktype" {
			continue
		}

		worktype, resp, err := taskMgmtApi.GetTaskmanagementWorktype(rs.Primary.ID, nil)
		if worktype != nil {
			return fmt.Errorf("Task management worktype (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Worktype not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All worktypes destroyed
	return nil
}

func generateWorktypeResource(wt worktypeConfig) string {
	tfConfig := fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
		default_workbin_id = %s
		schema_id = %s
		division_id = %s

		default_duration_seconds = %v
		default_expiration_seconds = %v
		default_due_duration_seconds = %v
		default_priority = %v
		default_ttl_seconds = %v

		default_language_id = %s
		default_queue_id = %s
		default_skills_ids = %s

		assignment_enabled = %v
		schema_version = %v
	}
		`, resourceName,
		wt.resID,
		wt.name,
		wt.description,
		wt.defaultWorkbinId,
		wt.schemaId,
		wt.divisionId,
		wt.defaultDurationS,
		wt.defaultExpirationS,
		wt.defaultDueDurationS,
		wt.defaultPriority,
		wt.defaultTtlS,
		wt.defaultLanguageId,
		wt.defaultQueueId,
		util.GenerateStringArray(wt.defaultSkillIds...),
		wt.assignmentEnabled,
		wt.schemaVersion,
	)
	return tfConfig
}
