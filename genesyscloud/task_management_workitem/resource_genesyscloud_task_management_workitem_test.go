package task_management_workitem

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	externalContact "terraform-provider-genesyscloud/genesyscloud/external_contacts"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v116/platformclientv2"
)

/*
The resource_genesyscloud_task_management_workitem_test.go contains all of the test cases for running the resource
tests for task_management_workitem.
*/

type scoredAgentConfig struct {
	agent_id string
	score    int
}

type workitemConfig struct {
	name                   string
	worktype_id            string
	description            string
	language_id            string
	priority               int
	date_due               string
	date_expires           string
	duration_seconds       int
	ttl                    int
	status_id              string
	workbin_id             string
	assignee_id            string
	external_contact_id    string
	external_tag           string
	queue_id               string
	skills_ids             []string
	preferred_agents_ids   []string
	auto_status_transition bool
	scored_agents          []scoredAgentConfig
	custom_fields          string
}

func TestAccResourceTaskManagementWorkitem(t *testing.T) {
	t.Parallel()
	var (
		// Workbin
		wbResourceId  = "workbin_1"
		wbName        = "wb_" + uuid.NewString()
		wbDescription = "workbin created for CX as Code test case"

		wb2ResourceId  = "workbin_2"
		wb2Name        = "wb_" + uuid.NewString()
		wb2Description = "workbin created for CX as Code test case"

		// Schema
		wsResourceId  = "schema_1"
		wsName        = "ws_" + uuid.NewString()
		wsDescription = "workitem schema created for CX as Code test case"

		// worktype
		wtResName         = "tf_worktype_1"
		wtName            = "tf-worktype" + uuid.NewString()
		wtDescription     = "tf-worktype-description"
		wtOStatusName     = "Open Status"
		wtOStatusDesc     = "Description of open status"
		wtOStatusCategory = "Open"
		wtCStatusName     = "Closed Status"
		wtCStatusDesc     = "Description of closed status"
		wtCStatusCategory = "Closed"

		// language
		resLang = "language_1"
		lang    = "en-us"

		// queue
		resQueue  = "queue_1"
		queueName = "tf_queue_" + uuid.NewString()

		// skill
		skillResId1   = "skill_1"
		skillResName1 = "tf_skill_1" + uuid.NewString()

		// user
		userResId1 = "user_1"
		userName1  = "tf_user_1" + uuid.NewString()
		userEmail1 = "tf_user_1" + uuid.NewString() + "@example.com"

		// external contact
		externalContactResId1 = "external_contact_1"
		externalContactTitle1 = "tf_external_contact_1" + uuid.NewString()

		// basic workitem
		workitemRes = "workitem_1"
		workitem1   = workitemConfig{
			name:        "tf-workitem" + uuid.NewString(),
			worktype_id: fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResName),
		}
		workitem1Update = workitemConfig{
			name:                   "tf-workitem" + uuid.NewString(),
			worktype_id:            fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResName),
			description:            "test workitem created by CX as Code",
			language_id:            fmt.Sprintf("genesyscloud_routing_language.%s.id", resLang),
			priority:               42,
			date_due:               time.Now().Add(time.Hour * 24).Format("2006-01-02T15:04Z"),
			date_expires:           time.Now().Add(time.Hour * 42).Format("2006-01-02T15:04Z"),
			duration_seconds:       99999,
			ttl:                    888888,
			status_id:              gcloud.NullValue,
			workbin_id:             fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceId),
			assignee_id:            fmt.Sprintf("genesyscloud_user.%s.id", userResId1),
			external_contact_id:    fmt.Sprintf("genesyscloud_externalcontacts_contact.%s.id", externalContactResId1),
			external_tag:           "external tag",
			queue_id:               fmt.Sprintf("genesyscloud_routing_queue.%s.id", resQueue),
			skills_ids:             []string{fmt.Sprintf("genesyscloud_routing_skill.%s.id", skillResId1)},
			preferred_agents_ids:   []string{fmt.Sprintf("genesyscloud_user.%s.id", userResId1)},
			auto_status_transition: false,
			custom_fields:          "",
			scored_agents: []scoredAgentConfig{{
				agent_id: fmt.Sprintf("genesyscloud_user.%s.id", userResId1),
				score:    42,
			}},
		}

		// String configuration of task management objects needed for the workitem: schema, workbin, workitem.
		// They don't really change so they are defined here instead of in each step.
		taskMgmtConfig = workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, gcloud.NullValue) +
			workbin.GenerateWorkbinResource(wb2ResourceId, wb2Name, wb2Description, gcloud.NullValue) +
			workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
			worktype.GenerateWorktypeResourceBasic(
				wtResName,
				wtName,
				wtDescription,
				fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceId),
				fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceId),
				// Needs both an open and closed status or workitems cannot be created for worktype.
				fmt.Sprintf(`
				statuses {
					name = "%s"
					description = "%s"
					category = "%s"
				}
				statuses {
					name = "%s"
					description = "%s"
					category = "%s"
				}
				default_status_name = "%s"
				`, wtOStatusName, wtOStatusDesc, wtOStatusCategory,
					wtCStatusName, wtCStatusDesc, wtCStatusCategory,
					wtOStatusName),
			)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Create basic workitem
			{
				Config: taskMgmtConfig +
					generateWorkitemResourceBasic(workitemRes, workitem1.name, workitem1.worktype_id, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "name", workitem1.name),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemRes, "worktype_id", "genesyscloud_task_management_worktype."+wtResName, "id"),
				),
			},
			// Update workitem with more fields
			{
				Config: taskMgmtConfig +
					gcloud.GenerateRoutingLanguageResource(resLang, lang) +
					gcloud.GenerateRoutingQueueResourceBasic(resQueue, queueName) +
					gcloud.GenerateRoutingSkillResource(skillResId1, skillResName1) +
					gcloud.GenerateBasicUserResource(userResId1, userEmail1, userName1) +
					externalContact.GenerateBasicExternalContactResource(externalContactResId1, externalContactTitle1) +
					generateWorkitemResource(workitemRes, workitem1Update),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "name", workitem1Update.name),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "description", workitem1Update.description),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemRes, "language_id", "genesyscloud_routing_language."+resLang, "id"),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "priority", fmt.Sprintf("%d", workitem1Update.priority)),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "date_due", workitem1Update.date_due),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "date_expires", workitem1Update.date_expires),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "duration_seconds", fmt.Sprintf("%d", workitem1Update.duration_seconds)),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "ttl", fmt.Sprintf("%d", workitem1Update.ttl)),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemRes, "status_id", "genesyscloud_task_management_worktype."+wtResName, "default_status_id"),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemRes, "workbin_id", "genesyscloud_task_management_workbin."+wbResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemRes, "assignee_id", "genesyscloud_user."+userResId1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemRes, "external_contact_id", "genesyscloud_externalcontacts_contact."+externalContactResId1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "external_tag", workitem1Update.external_tag),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemRes, "queue_id", "genesyscloud_routing_queue."+resQueue, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemRes, "skills_ids.0", "genesyscloud_routing_skill."+skillResId1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemRes, "preferred_agents_ids.0", "genesyscloud_user."+userResId1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "auto_status_transition", fmt.Sprintf("%t", workitem1Update.auto_status_transition)),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "custom_fields", workitem1Update.custom_fields),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemRes, "scored_agents.0.agent_id", "genesyscloud_user."+userResId1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemRes, "scored_agents.0.score", fmt.Sprintf("%d", workitem1Update.scored_agents[0].score)),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_task_management_workitem." + workitemRes,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTaskManagementWorkitemDestroyed,
	})
}

func generateWorkitemResourceBasic(resName string, wName string, wWorktypeId string, attrs string) string {
	return fmt.Sprintf(`resource "genesyscloud_task_management_workitem" "%s" {
		name = "%s"
		worktype_id = %s
		%s
	}`, resName, wName, wWorktypeId, attrs)
}

func generateWorkitemResource(resName string, wt workitemConfig) string {
	return fmt.Sprintf(`resource "genesyscloud_task_management_workitem" "%s" {
		name = "%s"
		worktype_id = %s
		description = "%s"
		language_id = %s
		priority = %d
		date_due = "%s"
		date_expires = "%s"
		duration_seconds = %d
		ttl = %d
		status_id = %s
		workbin_id = %s
		assignee_id = "%s"
		external_contact_id = %s
		external_tag = "%s"
		queue_id = %s
		skills_ids = %s
		preferred_agents_ids = %s
		auto_status_transition = %v
		custom_fields = "%s"
		%s
	}
	`, resName,
		wt.name,
		wt.worktype_id,
		wt.description,
		wt.language_id,
		wt.priority,
		wt.date_due,
		wt.date_expires,
		wt.duration_seconds,
		wt.ttl,
		wt.status_id,
		wt.workbin_id,
		wt.assignee_id,
		wt.external_contact_id,
		wt.external_tag,
		wt.queue_id,
		gcloud.GenerateStringArrayEnquote(wt.skills_ids...),
		gcloud.GenerateStringArrayEnquote(wt.preferred_agents_ids...),
		wt.auto_status_transition,
		wt.custom_fields,
		generateScoredAgents(&wt.scored_agents),
	)
}

func generateScoredAgents(scoredAgents *[]scoredAgentConfig) string {
	result := ""
	if scoredAgents == nil {
		return result
	}

	for _, scoredAgent := range *scoredAgents {
		result += fmt.Sprintf(`
		scored_agents {
			agent_id = "%s"
			score = %d
		}
		`, scoredAgent.agent_id, scoredAgent.score)
	}
	return result
}

func testVerifyTaskManagementWorkitemDestroyed(state *terraform.State) error {
	taskMgmtApi := platformclientv2.NewTaskManagementApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_task_management_workitem" {
			continue
		}

		worktype, resp, err := taskMgmtApi.GetTaskmanagementWorkitem(rs.Primary.ID, "")
		if worktype != nil {
			return fmt.Errorf("task management workitem (%s) still exists", rs.Primary.ID)
		} else if gcloud.IsStatus404(resp) {
			// Workitem not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All worktypes destroyed
	return nil
}
