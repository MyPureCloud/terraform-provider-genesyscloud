package task_management_worktype

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
)

/*
The resource_genesyscloud_task_management_worktype_test.go contains all of the test cases for running the resource
tests for task_management_worktype.
*/

// Basic test with create and update excluding workitem statusses
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

			// No statuses
			statuses:          []worktypeStatusConfig{},
			defaultStatusName: "",

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
			// All optional properties update (except statuses)
			{
				Config: workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
					routingQueue.GenerateRoutingQueueResourceBasic(queueResId, queueName) +
					gcloud.GenerateRoutingLanguageResource(langResId, langName) +
					gcloud.GenerateRoutingSkillResource(skillResId1, skillResName1) +
					gcloud.GenerateRoutingSkillResource(skillResId2, skillResName2) +
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
		wtRes = worktypeConfig{
			resID:            "worktype_1",
			name:             "tf_worktype_" + uuid.NewString(),
			description:      "worktype created for CX as Code test case",
			defaultWorkbinId: fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceId),
			schemaId:         fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceId),

			statuses: []worktypeStatusConfig{
				{
					name:        "Open Status",
					description: "Description of open status",
					category:    "Open",
				},
				{
					name:        "Close Status",
					description: "Description of close status",
					category:    "Closed",
				},
			},
			defaultStatusName: "Open Status",
		}

		// Updated statuses
		statusUpdates = worktypeConfig{
			statuses: []worktypeStatusConfig{
				{
					name:                         "Open Status",
					description:                  "Description of open status. Updated",
					defaultDestinationStatusName: "WIP",
					destinationStatusNames:       []string{"WIP", "Waiting Status"},
					statusTransitionTime:         "10:09:08",
					transitionDelay:              86500,
					category:                     "Open",
				},
				{
					name:        "WIP",
					description: "Description of in progress status. Updated",
					category:    "InProgress",
				},
				{
					name:        "Waiting Status",
					description: "Description of waiting status. Updated",
					category:    "Waiting",
				},
				{
					name:        "Close Status",
					description: "Description of close status. Updated",
					category:    "Closed",
				},
			},
		}

		// Updated statuses 2
		statusUpdates2 = worktypeConfig{
			statuses: []worktypeStatusConfig{
				{
					name:                         "Open Status",
					description:                  "Description of open status. Updated 2",
					defaultDestinationStatusName: "Close Status",
					transitionDelay:              300,

					category: "Open",
				},
				{
					name:        "Close Status",
					description: "Description of close status. Updated 2",
					category:    "Closed",
				},
			},
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Initial basic statuses
			{
				Config: workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
					GenerateWorktypeResourceBasic(wtRes.resID, wtRes.name, wtRes.description, wtRes.defaultWorkbinId, wtRes.schemaId, generateWorktypeAllStatuses(wtRes)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "name", wtRes.name),
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "description", wtRes.description),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "default_workbin_id", fmt.Sprintf("genesyscloud_task_management_workbin.%s", wbResourceId), "id"),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "schema_id", fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s", wsResourceId), "id"),

					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "statuses.#", fmt.Sprintf("%v", len(wtRes.statuses))),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName+"."+wtRes.resID, "statuses.*", map[string]*regexp.Regexp{
						"name":        regexp.MustCompile(wtRes.statuses[0].name),
						"description": regexp.MustCompile(wtRes.statuses[0].description),
						"category":    regexp.MustCompile(wtRes.statuses[0].category),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName+"."+wtRes.resID, "statuses.*", map[string]*regexp.Regexp{
						"name":        regexp.MustCompile(wtRes.statuses[1].name),
						"description": regexp.MustCompile(wtRes.statuses[1].description),
						"category":    regexp.MustCompile(wtRes.statuses[1].category),
					}),
				),
			},
			// Add statuses and destination references
			{
				Config: workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
					GenerateWorktypeResourceBasic(wtRes.resID, wtRes.name, wtRes.description, wtRes.defaultWorkbinId, wtRes.schemaId, generateWorktypeAllStatuses(statusUpdates)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "name", wtRes.name),
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "description", wtRes.description),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "default_workbin_id", fmt.Sprintf("genesyscloud_task_management_workbin.%s", wbResourceId), "id"),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "schema_id", fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s", wsResourceId), "id"),

					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "statuses.#", fmt.Sprintf("%v", len(statusUpdates.statuses))),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName+"."+wtRes.resID, "statuses.*", map[string]*regexp.Regexp{
						"name":                            regexp.MustCompile(statusUpdates.statuses[0].name),
						"description":                     regexp.MustCompile(statusUpdates.statuses[0].description),
						"category":                        regexp.MustCompile(statusUpdates.statuses[0].category),
						"default_destination_status_name": regexp.MustCompile(statusUpdates.statuses[0].defaultDestinationStatusName),
						"status_transition_delay_seconds": regexp.MustCompile(fmt.Sprintf("%v", statusUpdates.statuses[0].transitionDelay)),
						"status_transition_time":          regexp.MustCompile(fmt.Sprintf("%v", statusUpdates.statuses[0].statusTransitionTime)),
						"destination_status_names.0":      regexp.MustCompile(statusUpdates.statuses[0].destinationStatusNames[0]),
						"destination_status_names.1":      regexp.MustCompile(statusUpdates.statuses[0].destinationStatusNames[1]),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName+"."+wtRes.resID, "statuses.*", map[string]*regexp.Regexp{
						"name":        regexp.MustCompile(statusUpdates.statuses[1].name),
						"description": regexp.MustCompile(statusUpdates.statuses[1].description),
						"category":    regexp.MustCompile(statusUpdates.statuses[1].category),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName+"."+wtRes.resID, "statuses.*", map[string]*regexp.Regexp{
						"name":        regexp.MustCompile(statusUpdates.statuses[2].name),
						"description": regexp.MustCompile(statusUpdates.statuses[2].description),
						"category":    regexp.MustCompile(statusUpdates.statuses[2].category),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName+"."+wtRes.resID, "statuses.*", map[string]*regexp.Regexp{
						"name":        regexp.MustCompile(statusUpdates.statuses[3].name),
						"description": regexp.MustCompile(statusUpdates.statuses[3].description),
						"category":    regexp.MustCompile(statusUpdates.statuses[3].category),
					}),
				),
			},
			// Removing statuses and update
			{
				Config: workbin.GenerateWorkbinResource(wbResourceId, wbName, wbDescription, util.NullValue) +
					workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceId, wsName, wsDescription) +
					GenerateWorktypeResourceBasic(wtRes.resID, wtRes.name, wtRes.description, wtRes.defaultWorkbinId, wtRes.schemaId, generateWorktypeAllStatuses(statusUpdates2)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "name", wtRes.name),
					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "description", wtRes.description),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "default_workbin_id", fmt.Sprintf("genesyscloud_task_management_workbin.%s", wbResourceId), "id"),
					resource.TestCheckResourceAttrPair(resourceName+"."+wtRes.resID, "schema_id", fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s", wsResourceId), "id"),

					resource.TestCheckResourceAttr(resourceName+"."+wtRes.resID, "statuses.#", fmt.Sprintf("%v", len(statusUpdates2.statuses))),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName+"."+wtRes.resID, "statuses.*", map[string]*regexp.Regexp{
						"name":        regexp.MustCompile(statusUpdates2.statuses[0].name),
						"description": regexp.MustCompile(statusUpdates2.statuses[0].description),
						"category":    regexp.MustCompile(statusUpdates2.statuses[0].category),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName+"."+wtRes.resID, "statuses.*", map[string]*regexp.Regexp{
						"name":        regexp.MustCompile(statusUpdates2.statuses[1].name),
						"description": regexp.MustCompile(statusUpdates2.statuses[1].description),
						"category":    regexp.MustCompile(statusUpdates2.statuses[1].category),
					}),
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
	statuses := generateWorktypeAllStatuses(wt)

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
		%s
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
		statuses,
	)
	return tfConfig
}

func generateWorktypeAllStatuses(wt worktypeConfig) string {
	statuses := []string{}

	for _, s := range wt.statuses {
		statuses = append(statuses, generateWorktypeStatus(s))
	}

	return strings.Join(statuses, "\n")
}

func generateWorktypeStatus(wtStatus worktypeStatusConfig) string {
	additional := []string{}
	if len(wtStatus.destinationStatusNames) > 0 {
		additional = append(additional, util.GenerateMapProperty("destination_status_names", util.GenerateStringArrayEnquote(wtStatus.destinationStatusNames...)))
	}
	if wtStatus.defaultDestinationStatusName != "" {
		additional = append(additional, util.GenerateMapProperty("default_destination_status_name", strconv.Quote(wtStatus.defaultDestinationStatusName)))
	}
	if wtStatus.transitionDelay != 0 {
		additional = append(additional, util.GenerateMapProperty("status_transition_delay_seconds", strconv.Itoa(wtStatus.transitionDelay)))
	}
	if wtStatus.statusTransitionTime != "" {
		additional = append(additional, util.GenerateMapProperty("status_transition_time", strconv.Quote(wtStatus.statusTransitionTime)))
	}

	return fmt.Sprintf(`statuses {
		name = "%s"
		description = "%s"
		category = "%s"
		%s
	}
	`, wtStatus.name, wtStatus.description, wtStatus.category, strings.Join(additional, "\n"))
}
