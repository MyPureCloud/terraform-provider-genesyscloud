package task_management_workitem

import (
	"fmt"
	"strconv"
	"strings"
	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingLanguage "terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"

	"terraform-provider-genesyscloud/genesyscloud/user"

	"terraform-provider-genesyscloud/genesyscloud/user_roles"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	externalContact "terraform-provider-genesyscloud/genesyscloud/external_contacts"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	worktypeStatus "terraform-provider-genesyscloud/genesyscloud/task_management_worktype_status"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
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
	var (
		// home division
		homeDivResourceLabel = "home"

		// Workbin
		wbResourceLabel = "workbin_1"
		wbName          = "wb_" + uuid.NewString()
		wbDescription   = "workbin created for CX as Code test case"

		wb2ResourceLabel = "workbin_2"
		wb2Name          = "wb_" + uuid.NewString()
		wb2Description   = "workbin created for CX as Code test case"

		// Schema
		wsResourceLabel = "schema_1"
		wsName          = "ws_" + uuid.NewString()
		wsDescription   = "workitem schema created for CX as Code test case"

		// worktype
		wtResourceLabel = "tf_worktype_1"
		wtName          = "tf-worktype" + uuid.NewString()
		wtDescription   = "tf-worktype-description"

		// Worktype statuses
		statusResourceLabelOpen   = "open-status"
		wtOStatusName             = "Open Status"
		wtOStatusDesc             = "Description of open status"
		wtOStatusCategory         = "Open"
		statusResourceLabelClosed = "closed-status"
		wtCStatusName             = "Closed Status"
		wtCStatusDesc             = "Description of closed status"
		wtCStatusCategory         = "Closed"

		// language
		resourceLabelLang = "language_1"
		lang              = "tf_lang_" + uuid.NewString()

		// queue
		resLabelQueue = "queue_1"
		queueName     = "tf_queue_" + uuid.NewString()

		// skill
		skillResourceLabel = "skill_1"
		skillResName1      = "tf_skill_1" + uuid.NewString()

		// role (for user to be assigned a workitem)
		roleResourceLabel1 = "role_1"
		roleName1          = "tf_role_1" + uuid.NewString()

		// user
		userResourceLabel1 = "user_1"
		userName1          = "tf_user_1" + uuid.NewString()
		userEmail1         = "tf_user_1" + uuid.NewString() + "@example.com"

		// external contact
		externalContactResourceLabel1 = "external_contact_1"
		externalContactTitle1         = "tf_external_contact_1" + uuid.NewString()

		// basic workitem
		workitemResourceLabel = "workitem_1"
		workitem1             = workitemConfig{
			name:        "tf-workitem" + uuid.NewString(),
			worktype_id: fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
		}
		workitem1Update = workitemConfig{
			name:                   "tf-workitem" + uuid.NewString(),
			worktype_id:            fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
			description:            "test workitem created by CX as Code",
			language_id:            fmt.Sprintf("genesyscloud_routing_language.%s.id", resourceLabelLang),
			priority:               42,
			date_due:               time.Now().Add(time.Hour * 10).Format(resourcedata.TimeParseFormat), // 1 day from now
			date_expires:           time.Now().Add(time.Hour * 20).Format(resourcedata.TimeParseFormat), // 2 days from now
			duration_seconds:       99999,
			ttl:                    int(time.Now().Add(time.Hour * 24 * 30 * 6).Unix()), // ~6 months from now
			status_id:              util.NullValue,
			workbin_id:             fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
			assignee_id:            fmt.Sprintf("genesyscloud_user.%s.id", userResourceLabel1),
			external_contact_id:    fmt.Sprintf("genesyscloud_externalcontacts_contact.%s.id", externalContactResourceLabel1),
			external_tag:           "external tag",
			queue_id:               fmt.Sprintf("genesyscloud_routing_queue.%s.id", resLabelQueue),
			skills_ids:             []string{fmt.Sprintf("genesyscloud_routing_skill.%s.id", skillResourceLabel)},
			preferred_agents_ids:   []string{fmt.Sprintf("genesyscloud_user.%s.id", userResourceLabel1)},
			auto_status_transition: false,
			custom_fields:          "", // tested on a separate test case
			scored_agents: []scoredAgentConfig{{
				agent_id: fmt.Sprintf("genesyscloud_user.%s.id", userResourceLabel1),
				score:    42,
			}},
		}

		// String configuration of task management objects needed for the workitem: schema, workbin, worktype, worktype_status.
		// They don't really change so they are defined here instead of in each step.
		taskMgmtConfig = workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
			workbin.GenerateWorkbinResource(wb2ResourceLabel, wb2Name, wb2Description, util.NullValue) +
			workitemSchema.GenerateWorkitemSchemaResourceBasic(wsResourceLabel, wsName, wsDescription) +
			worktype.GenerateWorktypeResourceBasic(
				wtResourceLabel,
				wtName,
				wtDescription,
				fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
				fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceLabel),
				"",
			) +
			worktypeStatus.GenerateWorktypeStatusResource(
				statusResourceLabelOpen,
				fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
				wtOStatusName,
				wtOStatusCategory,
				wtOStatusDesc,
				util.NullValue,
				"",
			) +
			worktypeStatus.GenerateWorktypeStatusResource(
				statusResourceLabelClosed,
				fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
				wtCStatusName,
				wtCStatusCategory,
				wtCStatusDesc,
				util.NullValue,
				"",
			)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Create basic workitem
			{
				Config: taskMgmtConfig +
					generateWorkitemResourceBasic(
						workitemResourceLabel,
						workitem1.name,
						workitem1.worktype_id,
						fmt.Sprintf("status_id = genesyscloud_task_management_worktype_status.%s.id", statusResourceLabelOpen),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "name", workitem1.name),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemResourceLabel, "worktype_id", "genesyscloud_task_management_worktype."+wtResourceLabel, "id"),
					worktypeStatus.ValidateStatusIds("genesyscloud_task_management_workitem."+workitemResourceLabel, "status_id", "genesyscloud_task_management_worktype_status."+statusResourceLabelOpen, "id"),
				),
			},
			{
				//Add user roles first
				Config: gcloud.GenerateAuthDivisionHomeDataSource(homeDivResourceLabel) +
					user.GenerateBasicUserResource(userResourceLabel1, userEmail1, userName1) +
					authRole.GenerateAuthRoleResource(roleResourceLabel1, roleName1, "test role description",
						authRole.GenerateRolePermPolicy("workitems", "*", strconv.Quote("*")),
					) +
					user_roles.GenerateUserRoles("user_role_1", userResourceLabel1,
						generateResourceRoles(
							"genesyscloud_auth_role."+roleResourceLabel1+".id",
							"data.genesyscloud_auth_division_home."+homeDivResourceLabel+".id",
						),
					) + taskMgmtConfig + generateWorkitemResourceBasic(
					workitemResourceLabel,
					workitem1.name,
					workitem1.worktype_id,
					fmt.Sprintf("status_id = genesyscloud_task_management_worktype_status.%s.id", statusResourceLabelOpen),
				),
			},
			// Update workitem with more fields
			{
				Config: taskMgmtConfig +
					gcloud.GenerateAuthDivisionHomeDataSource(homeDivResourceLabel) +
					routingLanguage.GenerateRoutingLanguageResource(resourceLabelLang, lang) +
					routingQueue.GenerateRoutingQueueResourceBasic(resLabelQueue, queueName) +
					routingSkill.GenerateRoutingSkillResource(skillResourceLabel, skillResName1) +
					user.GenerateBasicUserResource(userResourceLabel1, userEmail1, userName1) +
					externalContact.GenerateBasicExternalContactResource(externalContactResourceLabel1, externalContactTitle1) +
					authRole.GenerateAuthRoleResource(roleResourceLabel1, roleName1, "test role description",
						authRole.GenerateRolePermPolicy("workitems", "*", strconv.Quote("*")),
					) +
					user_roles.GenerateUserRoles("user_role_1", userResourceLabel1,
						generateResourceRoles(
							"genesyscloud_auth_role."+roleResourceLabel1+".id",
							"data.genesyscloud_auth_division_home."+homeDivResourceLabel+".id",
						),
					) +
					generateWorkitemResource(workitemResourceLabel, workitem1Update, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "name", workitem1Update.name),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "description", workitem1Update.description),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemResourceLabel, "language_id", "genesyscloud_routing_language."+resourceLabelLang, "id"),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "priority", fmt.Sprintf("%d", workitem1Update.priority)),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "date_due", workitem1Update.date_due),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "date_expires", workitem1Update.date_expires),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "duration_seconds", fmt.Sprintf("%d", workitem1Update.duration_seconds)),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "ttl", fmt.Sprintf("%d", workitem1Update.ttl)),
					worktypeStatus.ValidateStatusIds("genesyscloud_task_management_workitem."+workitemResourceLabel, "status_id", "genesyscloud_task_management_worktype_status."+statusResourceLabelOpen, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemResourceLabel, "workbin_id", "genesyscloud_task_management_workbin."+wbResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemResourceLabel, "assignee_id", "genesyscloud_user."+userResourceLabel1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemResourceLabel, "external_contact_id", "genesyscloud_externalcontacts_contact."+externalContactResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "external_tag", workitem1Update.external_tag),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemResourceLabel, "queue_id", "genesyscloud_routing_queue."+resLabelQueue, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemResourceLabel, "skills_ids.0", "genesyscloud_routing_skill."+skillResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemResourceLabel, "preferred_agents_ids.0", "genesyscloud_user."+userResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "auto_status_transition", fmt.Sprintf("%t", workitem1Update.auto_status_transition)),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "custom_fields", workitem1Update.custom_fields),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemResourceLabel, "scored_agents.0.agent_id", "genesyscloud_user."+userResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "scored_agents.0.score", fmt.Sprintf("%d", workitem1Update.scored_agents[0].score)),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_task_management_workitem." + workitemResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTaskManagementWorkitemDestroyed,
	})
}

// TestAccResourceTaskManagementWorkitemCustomFields tests having custom field values
// for the workitem. Creation and updating.
func TestAccResourceTaskManagementWorkitemCustomFields(t *testing.T) {
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
		wsProperties    = `jsonencode({
			"custom_attribute_1_text" : {
			  "allOf" : [
				{
				  "$ref" : "#/definitions/text"
				}
			  ],
			  "title" : "custom_attribute_1",
			  "description" : "Custom attribute for text",
			  "minLength" : 0,
			  "maxLength" : 100
			},
			"custom_attribute_2_longtext" : {
			  "allOf" : [
				{
				  "$ref" : "#/definitions/longtext"
				}
			  ],
			  "title" : "custom_attribute_2",
			  "description" : "Custom attribute for long text",
			  "minLength" : 0,
			  "maxLength" : 1000
			},
			"custom_attribute_3_url" : {
			  "allOf" : [
				{
				  "$ref" : "#/definitions/url"
				}
			  ],
			  "title" : "custom_attribute_3",
			  "description" : "Custom attribute for url",
			  "minLength" : 0,
			  "maxLength" : 200
			},
			"custom_attribute_4_identifier" : {
			  "allOf" : [
				{
				  "$ref" : "#/definitions/identifier"
				}
			  ],
			  "title" : "custom_attribute_4",
			  "description" : "Custom attribute for identifier",
			  "minLength" : 0,
			  "maxLength" : 100
			},
			"custom_attribute_5_enum" : {
			  "allOf" : [
				{
				  "$ref" : "#/definitions/enum"
				}
			  ],
			  "title" : "custom_attribute_5",
			  "description" : "Custom attribute for enum",
			  "enum" : ["option_1", "option_2", "option_3"],
			  "_enumProperties" : {
				"option_1" : {
				  "title" : "Option 1",
				  "_disabled" : false
				},
				"option_2" : {
				  "title" : "Option 2",
				  "_disabled" : false
				},
				"option_3" : {
				  "title" : "Option 3",
				  "_disabled" : false
				},
			  },
			},
			"custom_attribute_6_date" : {
			  "allOf" : [
				{
				  "$ref" : "#/definitions/date"
				}
			  ],
			  "title" : "custom_attribute_6",
			  "description" : "Custom attribute for date",
			},
			"custom_attribute_7_datetime" : {
			  "allOf" : [
				{
				  "$ref" : "#/definitions/datetime"
				}
			  ],
			  "title" : "custom_attribute_7",
			  "description" : "Custom attribute for datetime",
			},
			"custom_attribute_8_integer" : {
			  "allOf" : [
				{
				  "$ref" : "#/definitions/integer"
				}
			  ],
			  "title" : "custom_attribute_8",
			  "description" : "Custom attribute for integer",
			  "minimum" : 1,
			  "maximum" : 1000
			},
			"custom_attribute_9_number" : {
			  "allOf" : [
				{
				  "$ref" : "#/definitions/number"
				}
			  ],
			  "title" : "custom_attribute_9",
			  "description" : "Custom attribute for number",
			  "minimum" : 1,
			  "maximum" : 1000
			},
			"custom_attribute_10_checkbox" : {
			  "allOf" : [
				{
				  "$ref" : "#/definitions/checkbox"
				}
			  ],
			  "title" : "custom_attribute_10",
			  "description" : "Custom attribute for checkbox"
			},
			"custom_attribute_11_tag" : {
			  "allOf" : [
				{
				  "$ref" : "#/definitions/tag"
				}
			  ],
			  "title" : "custom_attribute_11",
			  "description" : "Custom attribute for tag",
			  "items" : {
				"minLength" : 1,
				"maxLength" : 100
			  },
			  "minItems" : 0,
			  "maxItems" : 10,
			  "uniqueItems" : true
			},
		  })
		`

		// worktype
		wtResourceLabel = "tf_worktype_1"
		wtName          = "tf-worktype" + uuid.NewString()
		wtDescription   = "tf-worktype-description"

		// Worktype statuses
		statusResourceLabelOpen   = "open-status"
		wtOStatusName             = "Open Status"
		wtOStatusDesc             = "Description of open status"
		wtOStatusCategory         = "Open"
		statusResourceLabelClosed = "closed-status"
		wtCStatusName             = "Closed Status"
		wtCStatusDesc             = "Description of closed status"
		wtCStatusCategory         = "Closed"

		// basic workitem
		workitemResourceLabel = "workitem_1"
		workitem1             = workitemConfig{
			name:        "tf-workitem" + uuid.NewString(),
			worktype_id: fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
			custom_fields: `
			  {
				"custom_attribute_1_text" : "value_1 text",
				"custom_attribute_2_longtext" : "value_2 longtext",
				"custom_attribute_3_url" : "https://www.test.com",
				"custom_attribute_4_identifier" : "value_4 identifier",
				"custom_attribute_5_enum" : "option_1",
				"custom_attribute_6_date" : "2021-01-01",
				"custom_attribute_7_datetime" : "2021-01-01T00:00:00.000Z",
				"custom_attribute_8_integer" : 8,
				"custom_attribute_9_number" : 9,
				"custom_attribute_10_checkbox" : true,
				"custom_attribute_11_tag" : ["tag_1", "tag_2"]
			  }
			`,
		}
		workitem1update = workitemConfig{
			name:        "tf-workitem" + uuid.NewString(),
			worktype_id: fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
			custom_fields: `
			  {
				"custom_attribute_1_text" : "value_1 text update",
				"custom_attribute_2_longtext" : "value_2 longtext update",
				"custom_attribute_3_url" : "https://www.test-update.com",
				"custom_attribute_4_identifier" : "value_4 identifier update",
				"custom_attribute_5_enum" : "option_2",
				"custom_attribute_6_date" : "2022-02-02",
				"custom_attribute_7_datetime" : "2022-02-02T00:00:00.000Z",
				"custom_attribute_8_integer" : 82,
				"custom_attribute_9_number" : 92,
				"custom_attribute_10_checkbox" : false,
				"custom_attribute_11_tag" : ["tag_1_update", "tag_2_update"]
			  }
			`,
		}

		// String configuration of task management objects needed for the workitem: schema, workbin, workitem.
		// They don't really change so they are defined here instead of in each step.
		taskMgmtConfig = workbin.GenerateWorkbinResource(wbResourceLabel, wbName, wbDescription, util.NullValue) +
			workitemSchema.GenerateWorkitemSchemaResource(wsResourceLabel, wsName, wsDescription, wsProperties, util.TrueValue) +
			worktype.GenerateWorktypeResourceBasic(
				wtResourceLabel,
				wtName,
				wtDescription,
				fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
				fmt.Sprintf("genesyscloud_task_management_workitem_schema.%s.id", wsResourceLabel),
				"",
			) +
			worktypeStatus.GenerateWorktypeStatusResource(
				statusResourceLabelOpen,
				fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
				wtOStatusName,
				wtOStatusCategory,
				wtOStatusDesc,
				util.NullValue,
				"",
			) +
			worktypeStatus.GenerateWorktypeStatusResource(
				statusResourceLabelClosed,
				fmt.Sprintf("genesyscloud_task_management_worktype.%s.id", wtResourceLabel),
				wtCStatusName,
				wtCStatusCategory,
				wtCStatusDesc,
				util.NullValue,
				"",
			)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: taskMgmtConfig +
					generateWorkitemResourceBasic(
						workitemResourceLabel,
						workitem1.name,
						workitem1.worktype_id,
						fmt.Sprintf(`custom_fields = jsonencode(%s)`, workitem1.custom_fields),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "name", workitem1.name),
					resource.TestCheckResourceAttrPair("genesyscloud_task_management_workitem."+workitemResourceLabel, "worktype_id", "genesyscloud_task_management_worktype."+wtResourceLabel, "id"),
					validateWorkitemCustomFields("genesyscloud_task_management_workitem."+workitemResourceLabel, workitem1.custom_fields),
				),
			},
			{
				Config: taskMgmtConfig +
					generateWorkitemResourceBasic(
						workitemResourceLabel,
						workitem1update.name,
						workitem1update.worktype_id,
						fmt.Sprintf(`custom_fields = jsonencode(%s)`, workitem1update.custom_fields),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_task_management_workitem."+workitemResourceLabel, "name", workitem1update.name),
					validateWorkitemCustomFields("genesyscloud_task_management_workitem."+workitemResourceLabel, workitem1update.custom_fields),
				),
			},
		},
		CheckDestroy: testVerifyTaskManagementWorkitemDestroyed,
	})
}

// validateWorkitemCustomFields validates the custom fields of the workitem
func validateWorkitemCustomFields(resourcePath string, jsonFields string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("Failed to find resource %s in state", resourcePath)
		}
		resourceLabel := resourceState.Primary.ID

		stateCustomFields, ok := resourceState.Primary.Attributes["custom_fields"]
		if !ok {
			return fmt.Errorf("No custom_fields found for %s in state", resourceLabel)
		}

		if !util.EquivalentJsons(stateCustomFields, jsonFields) {
			return fmt.Errorf("%s custom_fields does not match %s", stateCustomFields, jsonFields)
		}

		return nil
	}
}

func generateWorkitemResourceBasic(resourceLabel string, wName string, wWorktypeId string, attrs string) string {
	return fmt.Sprintf(`resource "genesyscloud_task_management_workitem" "%s" {
		name = "%s"
		worktype_id = %s
		%s
	}`, resourceLabel, wName, wWorktypeId, attrs)
}

func generateWorkitemResource(resourceLabel string, wt workitemConfig, attrs string) string {
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
		assignee_id = %s
		external_contact_id = %s
		external_tag = "%s"
		queue_id = %s
		skills_ids = %s
		preferred_agents_ids = %s
		auto_status_transition = %v
		custom_fields = "%s"
		%s
		%s
	}
	`, resourceLabel,
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
		util.GenerateStringArray(wt.skills_ids...),
		util.GenerateStringArray(wt.preferred_agents_ids...),
		wt.auto_status_transition,
		wt.custom_fields,
		generateScoredAgents(&wt.scored_agents),
		attrs,
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
			agent_id = %s
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
		} else if util.IsStatus404(resp) {
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

func generateResourceRoles(skillID string, divisionIds ...string) string {
	var divAttr string
	if len(divisionIds) > 0 {
		divAttr = "division_ids = [" + strings.Join(divisionIds, ",") + "]"
	}
	return fmt.Sprintf(`roles {
		role_id = %s
		%s
	}
	`, skillID, divAttr)
}
