package routing_queue

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"terraform-provider-genesyscloud/genesyscloud/architect_user_prompt"
	"terraform-provider-genesyscloud/genesyscloud/group"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingSkillGroup "terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	"terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var (
	mu sync.Mutex
)

func TestAccResourceRoutingQueueBasic(t *testing.T) {
	var (
		queueResource1           = "test-queue"
		queueName1               = "Terraform Test Queue1-" + uuid.NewString()
		queueName2               = "Terraform Test Queue2-" + uuid.NewString()
		queueDesc1               = "This is a test"
		queueDesc2               = "This is still a test"
		alertTimeout1            = "7"
		alertTimeout2            = "100"
		slPercent1               = "0.5"
		slPercent2               = "0.9"
		slDuration1              = "1000"
		slDuration2              = "10000"
		wrapupPromptOptional     = "OPTIONAL"
		wrapupPromptMandTimeout  = "MANDATORY_TIMEOUT"
		routingRuleOpAny         = "ANY"
		routingRuleOpMeetsThresh = "MEETS_THRESHOLD"
		skillEvalAll             = "ALL"
		skillEvalBest            = "BEST"
		callingPartyName         = "Acme"
		callingPartyNumber       = "3173416548"
		scoringMethod            = "TimestampAndPriority"
		queueSkillResource       = "test-queue-skill"
		queueSkillName           = "Terraform Skill " + uuid.NewString()

		bullseyeMemberGroupName = "test_membergroup_series6"
		bullseyeMemberGroupType = "GROUP"
		testUserResource        = "user_resource1"
		testUserName            = "nameUser1" + uuid.NewString()
		testUserEmail           = uuid.NewString() + "@examplestest.com"
		callbackHours           = "7"
		callbackHours2          = "7"
		userID                  string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + routingSkill.GenerateRoutingSkillResource(queueSkillResource, queueSkillName) +
					group.GenerateGroupResource(
						bullseyeMemberGroupName,
						"MySeries6Groupv20",
						strconv.Quote("TestGroupForSeries6"),
						util.NullValue, // Default type
						util.NullValue, // Default visibility
						util.NullValue, // Default rules_visible
						group.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
					) + GenerateRoutingQueueResource(
					queueResource1,
					queueName1,
					queueDesc1,
					util.NullValue,               // MANDATORY_TIMEOUT
					"200000",                     // acw_timeout
					util.NullValue,               // ALL
					util.NullValue,               // auto_answer_only true
					util.NullValue,               // No calling party name
					util.NullValue,               // No calling party number
					util.NullValue,               // enable_manual_assignment false
					util.FalseValue,              // suppress_in_queue_call_recording false
					util.NullValue,               // enable_transcription false
					strconv.Quote(scoringMethod), // scoring Method
					GenerateAgentOwnedRouting("agent_owned_routing", util.TrueValue, callbackHours, callbackHours),
					GenerateMediaSettings("media_settings_call", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettingsCallBack("media_settings_callback", alertTimeout1, util.FalseValue, slPercent1, slDuration1, util.TrueValue, slDuration1, slDuration1),
					GenerateMediaSettings("media_settings_chat", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_email", alertTimeout1, util.TrueValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_message", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateBullseyeSettingsWithMemberGroup(alertTimeout1, "genesyscloud_group."+bullseyeMemberGroupName+".id", bullseyeMemberGroupType, "genesyscloud_routing_skill."+queueSkillResource+".id"),
					GenerateRoutingRules(routingRuleOpAny, "50", util.NullValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "name", queueName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "description", queueDesc1),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_wrapup_prompt", wrapupPromptMandTimeout),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_timeout_ms", "200000"),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "skill_evaluation_method", skillEvalAll),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "auto_answer_only", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "suppress_in_queue_call_recording", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_manual_assignment", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_transcription", util.FalseValue),
					provider.TestDefaultHomeDivision("genesyscloud_routing_queue."+queueResource1),
					validateMediaSettings(queueResource1, "media_settings_call", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_callback", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_chat", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_email", alertTimeout1, util.TrueValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_message", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateBullseyeSettings(queueResource1, 1, alertTimeout1, "genesyscloud_routing_skill."+queueSkillResource),
					validateRoutingRules(queueResource1, 0, routingRuleOpAny, "50", "5"),
					validateAgentOwnedRouting(queueResource1, "agent_owned_routing", util.TrueValue, callbackHours, callbackHours),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["genesyscloud_user."+testUserResource]
						if !ok {
							return fmt.Errorf("not found: %s", "genesyscloud_user."+testUserResource)
						}
						userID = rs.Primary.ID
						log.Printf("User ID: %s\n", userID) // Print user ID
						return nil
					},
				),
			},
			{
				// Update
				Config: GenerateRoutingQueueResource(
					queueResource1,
					queueName2,
					queueDesc2,
					strconv.Quote(wrapupPromptOptional),
					util.NullValue, // acw_timeout
					strconv.Quote(skillEvalBest),
					util.FalseValue, // auto_answer_only false
					strconv.Quote(callingPartyName),
					strconv.Quote(callingPartyNumber),
					util.TrueValue, // suppress_in_queue_call_recording true
					util.TrueValue, // enable_manual_assignment true
					util.TrueValue, // enable_transcription true
					strconv.Quote(scoringMethod),
					GenerateAgentOwnedRouting("agent_owned_routing", util.TrueValue, callbackHours2, callbackHours2),
					GenerateMediaSettings("media_settings_call", alertTimeout2, util.FalseValue, slPercent2, slDuration2),
					GenerateMediaSettings("media_settings_callback", alertTimeout2, util.TrueValue, slPercent2, slDuration2),
					GenerateMediaSettings("media_settings_chat", alertTimeout2, util.FalseValue, slPercent2, slDuration2),
					GenerateMediaSettings("media_settings_email", alertTimeout2, util.FalseValue, slPercent2, slDuration2),
					GenerateMediaSettings("media_settings_message", alertTimeout2, util.FalseValue, slPercent2, slDuration2),
					GenerateBullseyeSettings(alertTimeout2),
					GenerateBullseyeSettings(alertTimeout2),
					GenerateBullseyeSettings(alertTimeout2),
					GenerateRoutingRules(routingRuleOpMeetsThresh, "90", "30"),
					GenerateRoutingRules(routingRuleOpAny, "45", "15"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "name", queueName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "description", queueDesc2),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_wrapup_prompt", wrapupPromptOptional),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "skill_evaluation_method", skillEvalBest),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "auto_answer_only", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "calling_party_name", callingPartyName),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "calling_party_number", callingPartyNumber),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "scoring_method", scoringMethod),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "suppress_in_queue_call_recording", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_manual_assignment", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_transcription", util.TrueValue),
					provider.TestDefaultHomeDivision("genesyscloud_routing_queue."+queueResource1),
					validateMediaSettings(queueResource1, "media_settings_call", alertTimeout2, util.FalseValue, slPercent2, slDuration2),
					validateMediaSettings(queueResource1, "media_settings_callback", alertTimeout2, util.TrueValue, slPercent2, slDuration2),
					validateMediaSettings(queueResource1, "media_settings_chat", alertTimeout2, util.FalseValue, slPercent2, slDuration2),
					validateMediaSettings(queueResource1, "media_settings_email", alertTimeout2, util.FalseValue, slPercent2, slDuration2),
					validateMediaSettings(queueResource1, "media_settings_message", alertTimeout2, util.FalseValue, slPercent2, slDuration2),
					validateBullseyeSettings(queueResource1, 3, alertTimeout2, ""),
					validateRoutingRules(queueResource1, 0, routingRuleOpMeetsThresh, "90", "30"),
					validateRoutingRules(queueResource1, 1, routingRuleOpAny, "45", "15"),
					validateAgentOwnedRouting(queueResource1, "agent_owned_routing", util.TrueValue, callbackHours2, callbackHours2),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for resources to get deleted properly
						return nil
					},
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResource1,
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					checkUserDeleted(userID),
				),
			},
		},
		CheckDestroy: testVerifyQueuesAndUsersDestroyed,
	})
}

func TestAccResourceRoutingQueueConditionalRouting(t *testing.T) {
	if exists := featureToggles.CSGToggleExists(); exists {
		t.Skip("conditional group routing is deprecated in this resource, skipping test")
	}

	var (
		queueResource1          = "test-queue"
		queueName1              = "Terraform Test Queue1-" + uuid.NewString()
		queueDesc1              = "This is a test"
		alertTimeout1           = "7"
		slPercent1              = "0.5"
		slDuration1             = "1000"
		wrapupPromptMandTimeout = "MANDATORY_TIMEOUT"
		skillEvalAll            = "ALL"

		skillGroupResourceId = "skillgroup"
		skillGroupName       = "test skillgroup " + uuid.NewString()

		groupResourceId = "group"
		groupName       = "terraform test group" + uuid.NewString()
		queueResource2  = "test-queue-2"
		queueName2      = "Terraform Test Queue2-" + uuid.NewString()

		conditionalGroupRouting1Operator       = "LessThanOrEqualTo"
		conditionalGroupRouting1Metric         = "EstimatedWaitTime"
		conditionalGroupRouting1ConditionValue = "0"
		conditionalGroupRouting1WaitSeconds    = "20"
		conditionalGroupRouting1GroupType      = "SKILLGROUP"

		conditionalGroupRouting2Operator       = "GreaterThanOrEqualTo"
		conditionalGroupRouting2Metric         = "EstimatedWaitTime"
		conditionalGroupRouting2ConditionValue = "5"
		conditionalGroupRouting2WaitSeconds    = "15"
		conditionalGroupRouting2GroupType      = "GROUP"
		testUserResource                       = "user_resource1"
		testUserName                           = "nameUser1" + uuid.NewString()
		testUserEmail                          = uuid.NewString() + "@example.com"
		userID                                 string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
				// Create
				Config: routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceId,
					skillGroupName,
					"description",
				) + GenerateRoutingQueueResource(
					queueResource1,
					queueName1,
					queueDesc1,
					util.NullValue,  // MANDATORY_TIMEOUT
					"200000",        // acw_timeout
					util.NullValue,  // ALL
					util.NullValue,  // auto_answer_only true
					util.NullValue,  // No calling party name
					util.NullValue,  // No calling party number
					util.NullValue,  // enable_transcription false
					util.FalseValue, // suppress_in_queue_call_recording false
					util.NullValue,  // enable_manual_assignment false
					strconv.Quote("TimestampAndPriority"),
					GenerateMediaSettings(
						"media_settings_call",
						alertTimeout1,
						util.TrueValue,
						slPercent1,
						slDuration1),
					GenerateMediaSettings(
						"media_settings_callback",
						alertTimeout1,
						util.TrueValue,
						slPercent1,
						slDuration1),
					GenerateMediaSettings(
						"media_settings_chat",
						alertTimeout1,
						util.FalseValue,
						slPercent1,
						slDuration1),
					GenerateMediaSettings(
						"media_settings_email",
						alertTimeout1,
						util.TrueValue,
						slPercent1,
						slDuration1),
					GenerateMediaSettings(
						"media_settings_message",
						alertTimeout1,
						util.TrueValue, slPercent1,
						slDuration1),
					GenerateConditionalGroupRoutingRules(
						util.NullValue,                         // queue_id (queue_id in the first rule should be omitted)
						conditionalGroupRouting1Operator,       // operator
						conditionalGroupRouting1Metric,         // metric
						conditionalGroupRouting1ConditionValue, // condition_value
						conditionalGroupRouting1WaitSeconds,    // wait_seconds
						GenerateConditionalGroupRoutingRuleGroup(
							"genesyscloud_routing_skill_group."+skillGroupResourceId+".id", // group_id
							conditionalGroupRouting1GroupType,                              // group_type
						),
					),
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceId+".id]",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "name", queueName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "description", queueDesc1),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_wrapup_prompt", wrapupPromptMandTimeout),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_timeout_ms", "200000"),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "skill_evaluation_method", skillEvalAll),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "auto_answer_only", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "suppress_in_queue_call_recording", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_manual_assignment", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_transcription", util.FalseValue),

					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.operator", conditionalGroupRouting1Operator),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.metric", conditionalGroupRouting1Metric),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.condition_value", conditionalGroupRouting1ConditionValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.wait_seconds", conditionalGroupRouting1WaitSeconds),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.groups.0.member_group_type", conditionalGroupRouting1GroupType),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.groups.0.member_group_id", "genesyscloud_routing_skill_group."+skillGroupResourceId, "id"),

					provider.TestDefaultHomeDivision("genesyscloud_routing_queue."+queueResource1),
					validateMediaSettings(queueResource1, "media_settings_call", alertTimeout1, util.TrueValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_callback", alertTimeout1, util.TrueValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_chat", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_email", alertTimeout1, util.TrueValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_message", alertTimeout1, util.TrueValue, slPercent1, slDuration1),
				),
			},
			{
				// Update
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + group.GenerateBasicGroupResource(
					groupResourceId,
					groupName,
					group.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) +
					generateRoutingQueueResourceBasic(
						queueResource2,
						queueName2,
					) +
					routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
						skillGroupResourceId,
						skillGroupName,
						"description",
					) + GenerateRoutingQueueResource(
					queueResource1,
					queueName1,
					queueDesc1,
					util.NullValue,  // MANDATORY_TIMEOUT
					"200000",        // acw_timeout
					util.NullValue,  // ALL
					util.NullValue,  // auto_answer_only true
					util.NullValue,  // No calling party name
					util.NullValue,  // No calling party number
					util.NullValue,  // enable_transcription false
					util.FalseValue, // suppress_in_queue_call_recording false
					util.NullValue,  // enable_manual_assignment false
					strconv.Quote("TimestampAndPriority"),
					GenerateMediaSettings("media_settings_call", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_callback", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_chat", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_email", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_message", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateConditionalGroupRoutingRules(
						util.NullValue,                         // queue_id (queue_id in the first rule should be omitted)
						conditionalGroupRouting1Operator,       // operator
						conditionalGroupRouting1Metric,         // metric
						conditionalGroupRouting1ConditionValue, // condition_value
						conditionalGroupRouting1WaitSeconds,    // wait_seconds
						GenerateConditionalGroupRoutingRuleGroup(
							"genesyscloud_routing_skill_group."+skillGroupResourceId+".id", // group_id
							conditionalGroupRouting1GroupType,                              // group_type
						),
					),
					GenerateConditionalGroupRoutingRules(
						"genesyscloud_routing_queue."+queueResource2+".id", // queue_id
						conditionalGroupRouting2Operator,                   // operator
						conditionalGroupRouting2Metric,                     // metric
						conditionalGroupRouting2ConditionValue,             // condition_value
						conditionalGroupRouting2WaitSeconds,                // wait_seconds
						GenerateConditionalGroupRoutingRuleGroup(
							"genesyscloud_group."+groupResourceId+".id", // group_id
							conditionalGroupRouting2GroupType,           // group_type
						),
					),
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceId+".id]",
					"groups = [genesyscloud_group."+groupResourceId+".id]",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "name", queueName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "description", queueDesc1),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_wrapup_prompt", wrapupPromptMandTimeout),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_timeout_ms", "200000"),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "skill_evaluation_method", skillEvalAll),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "auto_answer_only", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "suppress_in_queue_call_recording", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_manual_assignment", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_transcription", util.FalseValue),

					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.operator", conditionalGroupRouting1Operator),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.metric", conditionalGroupRouting1Metric),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.condition_value", conditionalGroupRouting1ConditionValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.wait_seconds", conditionalGroupRouting1WaitSeconds),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.groups.0.member_group_type", conditionalGroupRouting1GroupType),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.0.groups.0.member_group_id", "genesyscloud_routing_skill_group."+skillGroupResourceId, "id"),

					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.1.operator", conditionalGroupRouting2Operator),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.1.metric", conditionalGroupRouting2Metric),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.1.condition_value", conditionalGroupRouting2ConditionValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.1.wait_seconds", conditionalGroupRouting2WaitSeconds),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.1.groups.0.member_group_type", conditionalGroupRouting2GroupType),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "conditional_group_routing_rules.1.groups.0.member_group_id", "genesyscloud_group."+groupResourceId, "id"),

					provider.TestDefaultHomeDivision("genesyscloud_routing_queue."+queueResource1),
					validateMediaSettings(queueResource1, "media_settings_call", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_callback", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_chat", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_email", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_message", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["genesyscloud_user."+testUserResource]
						if !ok {
							return fmt.Errorf("not found: %s", "genesyscloud_user."+testUserResource)
						}
						userID = rs.Primary.ID
						log.Printf("User ID: %s\n", userID) // Print user ID
						return nil
					},
				),

				PreventPostDestroyRefresh: true,
			},
			{
				Config: GenerateRoutingQueueResource(
					queueResource1,
					queueName1,
					queueDesc1,
					util.NullValue,  // MANDATORY_TIMEOUT
					"200000",        // acw_timeout
					util.NullValue,  // ALL
					util.NullValue,  // auto_answer_only true
					util.NullValue,  // No calling party name
					util.NullValue,  // No calling party number
					util.NullValue,  // enable_transcription false
					util.FalseValue, // suppress_in_queue_call_recording false
					util.NullValue,  // enable_manual_assignment false
					strconv.Quote("TimestampAndPriority"),
					GenerateMediaSettings("media_settings_call", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_callback", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_chat", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_email", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_message", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateConditionalGroupRoutingRules(
						util.NullValue,                         // queue_id (queue_id in the first rule should be omitted)
						conditionalGroupRouting1Operator,       // operator
						conditionalGroupRouting1Metric,         // metric
						conditionalGroupRouting1ConditionValue, // condition_value
						conditionalGroupRouting1WaitSeconds,    // wait_seconds
						GenerateConditionalGroupRoutingRuleGroup(
							"genesyscloud_routing_skill_group."+skillGroupResourceId+".id", // group_id
							conditionalGroupRouting1GroupType,                              // group_type
						),
					),
					GenerateConditionalGroupRoutingRules(
						"genesyscloud_routing_queue."+queueResource2+".id", // queue_id
						conditionalGroupRouting2Operator,                   // operator
						conditionalGroupRouting2Metric,                     // metric
						conditionalGroupRouting2ConditionValue,             // condition_value
						conditionalGroupRouting2WaitSeconds,                // wait_seconds
						GenerateConditionalGroupRoutingRuleGroup(
							"genesyscloud_group."+groupResourceId+".id", // group_id
							conditionalGroupRouting2GroupType,           // group_type
						),
					),
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceId+".id]",
					"groups = [genesyscloud_group."+groupResourceId+".id]",
				),
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResource1,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(45 * time.Second)
			return testVerifyQueuesAndUsersDestroyed(state)
		},
	})
}

func TestAccResourceRoutingQueueParToCGR(t *testing.T) {
	var (
		queueResource1          = "test-queue"
		queueName1              = "Terraform Test Queue1-" + uuid.NewString()
		queueDesc1              = "This is a test"
		alertTimeout1           = "7"
		slPercent1              = "0.5"
		slDuration1             = "1000"
		wrapupPromptMandTimeout = "MANDATORY_TIMEOUT"
		routingRuleOpAny        = "ANY"
		skillEvalAll            = "ALL"
		callbackHours           = "7"
		scoringMethod           = "TimestampAndPriority"
		skillGroupResourceId    = "skillgroup"
		skillGroupName          = "test skillgroup " + uuid.NewString()
	)

	// Create CGR queue with routing rules
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceId,
					skillGroupName,
					"description",
				) + GenerateRoutingQueueResource(
					queueResource1,
					queueName1,
					queueDesc1,
					util.NullValue,  // MANDATORY_TIMEOUT
					"200000",        // acw_timeout
					util.NullValue,  // ALL
					util.NullValue,  // auto_answer_only true
					util.NullValue,  // No calling party name
					util.NullValue,  // No calling party number
					util.NullValue,  // enable_transcription false
					util.FalseValue, // suppress_in_queue_call_recording false
					util.NullValue,  // enable_manual_assignment false
					strconv.Quote(scoringMethod),
					GenerateAgentOwnedRouting("agent_owned_routing", util.TrueValue, callbackHours, callbackHours),
					GenerateMediaSettings("media_settings_call", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_callback", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_chat", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_email", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateMediaSettings("media_settings_message", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					GenerateRoutingRules(routingRuleOpAny, "50", "6"),
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceId+".id]",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "name", queueName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "description", queueDesc1),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_wrapup_prompt", wrapupPromptMandTimeout),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_timeout_ms", "200000"),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "skill_evaluation_method", skillEvalAll),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "auto_answer_only", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_manual_assignment", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "suppress_in_queue_call_recording", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_transcription", util.FalseValue),

					provider.TestDefaultHomeDivision("genesyscloud_routing_queue."+queueResource1),
					validateMediaSettings(queueResource1, "media_settings_call", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_callback", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_chat", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_email", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_message", alertTimeout1, util.FalseValue, slPercent1, slDuration1),
					validateAgentOwnedRouting(queueResource1, "agent_owned_routing", util.TrueValue, callbackHours, callbackHours),
					validateRoutingRules(queueResource1, 0, routingRuleOpAny, "50", "6"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyQueuesDestroyed,
	})
}

func TestAccResourceRoutingQueueFlows(t *testing.T) {
	var (
		queueResource1 = "test-queue"
		queueName1     = "Terraform Test Queue1-" + uuid.NewString()

		queueFlowResource1          = "test_flow1"
		queueFlowResource2          = "test_flow2"
		emailInQueueFlowResource1   = "email_test_flow1"
		emailInQueueFlowResource2   = "email_test_flow2"
		messageInQueueFlowResource1 = "message_test_flow1"
		messageInQueueFlowResource2 = "message_test_flow2"
		queueFlowName1              = "Terraform Flow Test-" + uuid.NewString()
		queueFlowName2              = "Terraform Flow Test-" + uuid.NewString()
		queueFlowName3              = "Terraform Flow Test-" + uuid.NewString()
		queueFlowFilePath1          = "../../examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml"
		queueFlowFilePath2          = "../../examples/resources/genesyscloud_flow/inboundcall_flow_example2.yaml"
		queueFlowFilePath3          = "../../examples/resources/genesyscloud_flow/inboundcall_flow_example3.yaml"

		queueFlowInboundcallConfig1          = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", queueFlowName1)
		messageInQueueFlowInboundcallConfig3 = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", queueFlowName3)

		//variables for testing 'on_hold_prompt_id'
		userPromptResource1         = "test-user_prompt_1"
		userPromptName1             = "TestUserPrompt_1" + strings.Replace(uuid.NewString(), "-", "", -1)
		userPromptDescription1      = "Test description"
		userPromptResourceLang1     = "en-us"
		userPromptResourceText1     = "This is a test greeting!"
		userPromptResourceFileName2 = "../" + testrunner.GetTestDataPath("test-prompt-02.wav")
		userPromptResourceTTS1      = "This is a test greeting!"
		userPromptAsset1            = architect_user_prompt.UserPromptResourceStruct{
			Language:        userPromptResourceLang1,
			Tts_string:      strconv.Quote(userPromptResourceTTS1),
			Text:            util.NullValue,
			Filename:        util.NullValue,
			FileContentHash: util.NullValue,
		}
		userPromptAsset2 = architect_user_prompt.UserPromptResourceStruct{
			Language:        userPromptResourceLang1,
			Tts_string:      util.NullValue,
			Text:            strconv.Quote(userPromptResourceText1),
			Filename:        strconv.Quote(userPromptResourceFileName2),
			FileContentHash: userPromptResourceFileName2,
		}

		userPromptResources1 = []*architect_user_prompt.UserPromptResourceStruct{&userPromptAsset1}
		userPromptResources2 = []*architect_user_prompt.UserPromptResourceStruct{&userPromptAsset2}
	)

	var homeDivisionName string
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: "data \"genesyscloud_auth_division_home\" \"home\" {}",
				Check: resource.ComposeTestCheckFunc(
					util.GetHomeDivisionName("data.genesyscloud_auth_division_home.home", &homeDivisionName),
				),
			},
		},
	})

	emailInQueueFlowInboundcallConfig2 := fmt.Sprintf(`inboundEmail:
    name: %s
    division: %s
    startUpRef: "/inboundEmail/states/state[Initial State_10]"
    defaultLanguage: en-us
    supportedLanguages:
        en-us:
            defaultLanguageSkill:
                noValue: true
    settingsInboundEmailHandling:
        emailHandling:
            disconnect:
                none: true
    settingsErrorHandling:
        errorHandling:
            disconnect:
                none: true
    states:
        - state:
            name: Initial State
            refId: Initial State_10
            actions:
                - disconnect:
                    name: Disconnect
`, queueFlowName2, homeDivisionName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: architect_flow.GenerateFlowResource(
					queueFlowResource1,
					queueFlowFilePath1,
					queueFlowInboundcallConfig1,
					false,
				) + architect_flow.GenerateFlowResource(
					emailInQueueFlowResource1,
					queueFlowFilePath2,
					emailInQueueFlowInboundcallConfig2,
					false,
				) + architect_flow.GenerateFlowResource(
					messageInQueueFlowResource1,
					queueFlowFilePath3,
					messageInQueueFlowInboundcallConfig3,
					false,
				) + architect_user_prompt.GenerateUserPromptResource(&architect_user_prompt.UserPromptStruct{
					ResourceID:  userPromptResource1,
					Name:        userPromptName1,
					Description: strconv.Quote(userPromptDescription1),
					Resources:   userPromptResources1,
				}) + GenerateRoutingQueueResourceBasic(
					queueResource1,
					queueName1,
					"queue_flow_id = genesyscloud_flow."+queueFlowResource1+".id",
					"email_in_queue_flow_id = genesyscloud_flow."+emailInQueueFlowResource1+".id",
					"message_in_queue_flow_id = genesyscloud_flow."+messageInQueueFlowResource1+".id",
					"on_hold_prompt_id = genesyscloud_architect_user_prompt."+userPromptResource1+".id",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "queue_flow_id", "genesyscloud_flow."+queueFlowResource1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "email_in_queue_flow_id", "genesyscloud_flow."+emailInQueueFlowResource1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "message_in_queue_flow_id", "genesyscloud_flow."+messageInQueueFlowResource1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "on_hold_prompt_id", "genesyscloud_architect_user_prompt."+userPromptResource1, "id"),
				),
			},
			{
				// Update the flows
				Config: architect_flow.GenerateFlowResource(
					queueFlowResource2,
					queueFlowFilePath1,
					queueFlowInboundcallConfig1,
					false,
				) + architect_flow.GenerateFlowResource(
					emailInQueueFlowResource2,
					queueFlowFilePath2,
					emailInQueueFlowInboundcallConfig2,
					false,
				) + architect_flow.GenerateFlowResource(
					messageInQueueFlowResource2,
					queueFlowFilePath3,
					messageInQueueFlowInboundcallConfig3,
					false,
				) + architect_user_prompt.GenerateUserPromptResource(&architect_user_prompt.UserPromptStruct{
					ResourceID:  userPromptResource1,
					Name:        userPromptName1,
					Description: strconv.Quote(userPromptDescription1),
					Resources:   userPromptResources2,
				}) + GenerateRoutingQueueResourceBasic(
					queueResource1,
					queueName1,
					"queue_flow_id = genesyscloud_flow."+queueFlowResource2+".id",
					"email_in_queue_flow_id = genesyscloud_flow."+emailInQueueFlowResource2+".id",
					"message_in_queue_flow_id = genesyscloud_flow."+messageInQueueFlowResource2+".id",
					"on_hold_prompt_id = genesyscloud_architect_user_prompt."+userPromptResource1+".id",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "queue_flow_id", "genesyscloud_flow."+queueFlowResource2, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "email_in_queue_flow_id", "genesyscloud_flow."+emailInQueueFlowResource2, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "message_in_queue_flow_id", "genesyscloud_flow."+messageInQueueFlowResource2, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "on_hold_prompt_id", "genesyscloud_architect_user_prompt."+userPromptResource1, "id"),
					func(s *terraform.State) error {
						time.Sleep(45 * time.Second) // Wait for 45 seconds for proper deletion of user
						return nil
					},
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyQueuesDestroyed,
	})
}

func TestAccResourceRoutingQueueSkillgroupMembers(t *testing.T) {
	var (
		queueResourceId = "test-queue"
		queueName       = "tf test queue" + uuid.NewString()

		user1ResourceId = "user1"
		user1Name       = "user " + uuid.NewString()
		user1Email      = "user" + strings.Replace(uuid.NewString(), "-", "", -1) + "@example.com"

		user2ResourceId = "user2"
		user2Name       = "user " + uuid.NewString()
		user2Email      = "user" + strings.Replace(uuid.NewString(), "-", "", -1) + "@example.com"

		skillResourceId = "test-skill"
		skillName       = "Skill " + uuid.NewString()

		skillGroupResourceId = "test-skill-group"
		skillGroupName       = "tf test skillgroup " + uuid.NewString()
	)

	skillGroupConfig := fmt.Sprintf(`
	resource "genesyscloud_routing_skill_group" "%s" {
		name = "%s"
		skill_conditions = jsonencode(
			[
			  {
				"routingSkillConditions" : [
				  {
					"routingSkill" : "%s",
					"comparator" : "GreaterThan",
					"proficiency" : 2,
					"childConditions" : [{
					  "routingSkillConditions" : [],
					  "languageSkillConditions" : [],
					  "operation" : "And"
					}]
				  }
				],
				"languageSkillConditions" : [],
				"operation" : "And"
			}]
		)
	
		depends_on = [ genesyscloud_routing_skill.%s ]
	}	
	`, skillGroupResourceId, skillGroupName, skillName, skillResourceId)

	user2Config := fmt.Sprintf(`
	resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		routing_skills {
			skill_id    = genesyscloud_routing_skill.%s.id
			proficiency = 4.5
		}
	}
	`, user2ResourceId, user2Email, user2Name, skillResourceId)

	/*
		Assign 1 user to the queue via the members set
		Assign 1 members based on a skill group
		Confirm that the length of `skill_groups` and `members` both equal 1
	*/
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: routingSkill.GenerateRoutingSkillResource(
					skillResourceId,
					skillName,
				) + skillGroupConfig + user2Config +
					genesyscloud.GenerateBasicUserResource(
						user1ResourceId,
						user1Email,
						user1Name,
					) + GenerateRoutingQueueResourceBasic(
					queueResourceId,
					queueName,
					GenerateMemberBlock("genesyscloud_user."+user1ResourceId+".id", util.NullValue),
					fmt.Sprintf("skill_groups = [genesyscloud_routing_skill_group.%s.id]", skillGroupResourceId),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResourceId, "skill_groups.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResourceId, "members.#", "1"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyQueuesDestroyed,
	})
}

func TestAccResourceRoutingQueueMembers(t *testing.T) {
	var (
		queueResource        = "test-queue-members"
		queueName            = "Terraform Test Queue3-" + uuid.NewString()
		queueMemberResource1 = "test-queue-user1"
		queueMemberResource2 = "test-queue-user2"
		queueMemberEmail1    = "terraform1-" + uuid.NewString() + "@queue1.com"
		queueMemberEmail2    = "terraform2-" + uuid.NewString() + "@queue2.com"
		queueMemberName1     = "Henry Terraform Test"
		queueMemberName2     = "Amanda Terraform Test"
		defaultQueueRingNum  = "1"
		queueRingNum         = "3"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: genesyscloud.GenerateBasicUserResource(
					queueMemberResource1,
					queueMemberEmail1,
					queueMemberName1,
				) + genesyscloud.GenerateBasicUserResource(
					queueMemberResource2,
					queueMemberEmail2,
					queueMemberName2,
				) + GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
					GenerateMemberBlock("genesyscloud_user."+queueMemberResource1+".id", util.NullValue),
				),
				Check: resource.ComposeTestCheckFunc(
					validateMember("genesyscloud_routing_queue."+queueResource, "genesyscloud_user."+queueMemberResource1, defaultQueueRingNum),
				),
			},
			{
				PreConfig: func() {
					// Wait for a specified duration to avoid runtime error
					time.Sleep(30 * time.Second)
				},
				// Update with another queue member and modify rings
				Config: genesyscloud.GenerateBasicUserResource(
					queueMemberResource1,
					queueMemberEmail1,
					queueMemberName1,
				) + genesyscloud.GenerateBasicUserResource(
					queueMemberResource2,
					queueMemberEmail2,
					queueMemberName2,
				) + GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
					GenerateMemberBlock("genesyscloud_user."+queueMemberResource1+".id", queueRingNum),
					GenerateMemberBlock("genesyscloud_user."+queueMemberResource2+".id", queueRingNum),
					GenerateBullseyeSettings("10"),
					GenerateBullseyeSettings("10"),
					GenerateBullseyeSettings("10"),
				),
				Check: resource.ComposeTestCheckFunc(
					validateMember("genesyscloud_routing_queue."+queueResource, "genesyscloud_user."+queueMemberResource1, queueRingNum),
					validateMember("genesyscloud_routing_queue."+queueResource, "genesyscloud_user."+queueMemberResource2, queueRingNum),
				),
			},
			{
				// Remove a queue member
				Config: genesyscloud.GenerateBasicUserResource(
					queueMemberResource2,
					queueMemberEmail2,
					queueMemberName2,
				) + GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
					GenerateMemberBlock("genesyscloud_user."+queueMemberResource2+".id", queueRingNum),
					GenerateBullseyeSettings("10"),
					GenerateBullseyeSettings("10"),
					GenerateBullseyeSettings("10"),
				),
				Check: resource.ComposeTestCheckFunc(
					validateMember("genesyscloud_routing_queue."+queueResource, "genesyscloud_user."+queueMemberResource2, queueRingNum),
				),
				Destroy: true,
			},
			{
				// Remove all queue members
				Config: GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
					"members = []",
					GenerateBullseyeSettings("10"),
					GenerateBullseyeSettings("10"),
					GenerateBullseyeSettings("10"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_routing_queue."+queueResource, "members.%"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResource,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
			},
		},
		CheckDestroy: testVerifyQueuesAndUsersDestroyed,
	})
}
func TestAccResourceRoutingQueueWrapupCodes(t *testing.T) {
	var (
		queueResource       = "test-queue-wrapup"
		queueName           = "Terraform Test Queue-" + uuid.NewString()
		wrapupCodeResource1 = "test-wrapup-1"
		wrapupCodeResource2 = "test-wrapup-2"
		wrapupCodeResource3 = "test-wrapup-3"
		wrapupCodeName1     = "Terraform Test Code1-" + uuid.NewString()
		wrapupCodeName2     = "Terraform Test Code2-" + uuid.NewString()
		wrapupCodeName3     = "Terraform Test Code3-" + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with two wrapup codes
				Config: GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
					GenerateQueueWrapupCodes("genesyscloud_routing_wrapupcode."+wrapupCodeResource1+".id",
						"genesyscloud_routing_wrapupcode."+wrapupCodeResource2+".id"),
				) + genesyscloud.GenerateRoutingWrapupcodeResource(
					wrapupCodeResource1,
					wrapupCodeName1,
				) + genesyscloud.GenerateRoutingWrapupcodeResource(
					wrapupCodeResource2,
					wrapupCodeName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateQueueWrapupCode("genesyscloud_routing_queue."+queueResource, "genesyscloud_routing_wrapupcode."+wrapupCodeResource1),
					validateQueueWrapupCode("genesyscloud_routing_queue."+queueResource, "genesyscloud_routing_wrapupcode."+wrapupCodeResource2),
				),
			},
			{
				// Update with another wrapup code
				Config: GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
					GenerateQueueWrapupCodes(
						"genesyscloud_routing_wrapupcode."+wrapupCodeResource1+".id",
						"genesyscloud_routing_wrapupcode."+wrapupCodeResource2+".id",
						"genesyscloud_routing_wrapupcode."+wrapupCodeResource3+".id"),
				) + genesyscloud.GenerateRoutingWrapupcodeResource(
					wrapupCodeResource1,
					wrapupCodeName1,
				) + genesyscloud.GenerateRoutingWrapupcodeResource(
					wrapupCodeResource2,
					wrapupCodeName2,
				) + genesyscloud.GenerateRoutingWrapupcodeResource(
					wrapupCodeResource3,
					wrapupCodeName3,
				),
				Check: resource.ComposeTestCheckFunc(
					validateQueueWrapupCode("genesyscloud_routing_queue."+queueResource, "genesyscloud_routing_wrapupcode."+wrapupCodeResource1),
					validateQueueWrapupCode("genesyscloud_routing_queue."+queueResource, "genesyscloud_routing_wrapupcode."+wrapupCodeResource2),
				),
			},
			{
				// Remove two wrapup codes
				Config: GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
					GenerateQueueWrapupCodes("genesyscloud_routing_wrapupcode."+wrapupCodeResource2+".id"),
				) + genesyscloud.GenerateRoutingWrapupcodeResource(
					wrapupCodeResource2,
					wrapupCodeName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateQueueWrapupCode("genesyscloud_routing_queue."+queueResource, "genesyscloud_routing_wrapupcode."+wrapupCodeResource2),
				),
			},
			{
				// Remove all wrapup codes
				Config: GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
					GenerateQueueWrapupCodes(),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_routing_queue."+queueResource, "wrapup_codes.%"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyQueuesDestroyed,
	})
}

func TestAccResourceRoutingQueueDirectRouting(t *testing.T) {
	var (
		queueResource1    = "test-queue-direct"
		queueResource2    = "test-queue"
		queueName1        = "Terraform Test Queue1-" + uuid.NewString()
		queueName2        = "Terraform Test Queue2-" + uuid.NewString()
		queueName3        = "Terraform Test Queue3-" + uuid.NewString()
		agentWaitSeconds1 = "200"
		waitForAgent1     = "true"
		agentWaitSeconds2 = "300"
		waitForAgent2     = "false"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingQueueResourceBasic(queueResource2, queueName2) +
					generateRoutingQueueResourceBasicWithDepends(
						queueResource1,
						"genesyscloud_routing_queue."+queueResource2,
						queueName1,
						generateDirectRouting(
							agentWaitSeconds1, // agentWaitSeconds
							waitForAgent1,     // waitForAgent
							"true",            // callUseAgentAddressOutbound
							"true",            // emailUseAgentAddressOutbound
							"true",            // messageUseAgentAddressOutbound
							"backup_queue_id = genesyscloud_routing_queue."+queueResource2+".id",
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "name", queueName1),
					validateDirectRouting(queueResource1, agentWaitSeconds1, waitForAgent1, "true", "true", "true"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "direct_routing.0.backup_queue_id", "genesyscloud_routing_queue."+queueResource2, "id"),
				),
			},
			{
				// Update
				Config: generateRoutingQueueResourceBasic(queueResource2, queueName3) +
					generateRoutingQueueResourceBasicWithDepends(
						queueResource1,
						"genesyscloud_routing_queue."+queueResource2,
						queueName1,
						generateDirectRouting(
							agentWaitSeconds2, // agentWaitSeconds
							waitForAgent2,     // waitForAgent
							"true",            // callUseAgentAddressOutbound
							"true",            // emailUseAgentAddressOutbound
							"true",            // messageEnabled
							"backup_queue_id = genesyscloud_routing_queue."+queueResource2+".id",
						),
					),
				Check: resource.ComposeTestCheckFunc(
					validateDirectRouting(queueResource1, agentWaitSeconds2, waitForAgent2, "true", "true", "true"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "direct_routing.0.backup_queue_id", "genesyscloud_routing_queue."+queueResource2, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyQueuesDestroyed,
	})
}

func TestAccResourceRoutingQueueDirectRoutingNoBackup(t *testing.T) {
	var (
		queueResource1    = "test-queue-direct"
		queueName1        = "Terraform Test Queue1-" + uuid.NewString()
		queueName2        = "Terraform Test Queue2-" + uuid.NewString()
		agentWaitSeconds1 = "200"
		waitForAgent1     = "true"
		agentWaitSeconds2 = "300"
		waitForAgent2     = "false"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingQueueResourceBasic(
					queueResource1,
					queueName1,
					generateDirectRouting(
						agentWaitSeconds1, // agentWaitSeconds
						waitForAgent1,     // waitForAgent
						"true",            // callUseAgentAddressOutbound
						"true",            // emailUseAgentAddressOutbound
						"true",            // messageUseAgentAddressOutbound
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "name", queueName1),
					validateDirectRouting(queueResource1, agentWaitSeconds1, waitForAgent1, "true", "true", "true"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "direct_routing.0.backup_queue_id", "genesyscloud_routing_queue."+queueResource1, "id"), // set to itself by Backend logic
				),
			},
			{
				// Update
				Config: generateRoutingQueueResourceBasic(
					queueResource1,
					queueName2,
					generateDirectRouting(
						agentWaitSeconds2, // agentWaitSeconds
						waitForAgent2,     // waitForAgent
						"true",            // callUseAgentAddressOutbound
						"true",            // emailUseAgentAddressOutbound
						"true",            // messageEnabled
					),
				),
				Check: resource.ComposeTestCheckFunc(
					validateDirectRouting(queueResource1, agentWaitSeconds2, waitForAgent2, "true", "true", "true"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queueResource1, "direct_routing.0.backup_queue_id", "genesyscloud_routing_queue."+queueResource1, "id"), // set to itself by Backend logic
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyQueuesDestroyed,
	})
}

// TestAccResourceRoutingQueueMembersOutsideOfConfig
// Creates a queue and a user, and then adds the user to that queue outside Terraform.
// On the next apply, we expect an empty plan and therefore no errors (achieved through 'members' being a computed field)
// Although members should not be a computed field, it was always computed in the past. As a result, some CX as Code users got used
// to the behaviour described above, so we don't want to break that behaviour.
func TestAccResourceRoutingQueueMembersOutsideOfConfig(t *testing.T) {
	var (
		userResourceId  = "user"
		userEmail       = fmt.Sprintf("user%s@test.com", strings.Replace(uuid.NewString(), "-", "", -1))
		queueResourceId = "queue"
		queueName       = "tf test queue " + uuid.NewString()
		userID          string
	)

	queueResource := fmt.Sprintf(`
resource "genesyscloud_routing_queue" "%s" {
	name = "%s"
}
`, queueResourceId, queueName)

	userResource := fmt.Sprintf(`
resource "genesyscloud_user" "%s" {
	name  = "tf test user"
	email = "%s"
}
`, userResourceId, userEmail)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: queueResource + userResource,
				Check: resource.ComposeTestCheckFunc(
					addMemberToQueue("genesyscloud_routing_queue."+queueResourceId, "genesyscloud_user."+userResourceId),
				),
			},
			{
				Config:             queueResource + userResource,
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["genesyscloud_user."+userResourceId]
						if !ok {
							return fmt.Errorf("not found: %s", "genesyscloud_user."+userResourceId)
						}
						userID = rs.Primary.ID
						log.Printf("User ID: %s\n", userID) // Print user ID
						return nil
					},
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResourceId,
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					checkUserDeleted(userID),
				),
			},
		},
		CheckDestroy: testVerifyQueuesDestroyed,
	})
}

func addMemberToQueue(queueResourceName, userResourceName string) resource.TestCheckFunc {
	getResourceGuidFromState := func(state *terraform.State, resourceName string) (string, error) {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("failed to find resourceState %s in state", resourceName)
		}
		return resourceState.Primary.ID, nil
	}

	return func(state *terraform.State) error {
		sdkConfig, err := provider.AuthorizeSdk()
		if err != nil {
			log.Fatal(err)
		}

		apiInstance := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

		queueID, err := getResourceGuidFromState(state, queueResourceName)
		if err != nil {
			return err
		}

		userID, err := getResourceGuidFromState(state, userResourceName)
		if err != nil {
			return err
		}

		log.Printf("adding member %s to queue %s", userID, queueID)

		const deleteMembers = false
		body := []platformclientv2.Writableentity{{Id: &userID}}
		if _, err := apiInstance.PostRoutingQueueMembers(queueID, body, deleteMembers); err != nil {
			return fmt.Errorf("failed to add member to queue %s: %v", queueID, err)
		}

		log.Printf("added member %s to queue %s", userID, queueID)

		time.Sleep(3 * time.Second)
		return nil
	}
}

func testVerifyQueuesDestroyed(state *terraform.State) error {
	routingAPI := platformclientv2.NewRoutingApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_queue" {
			continue
		}
		queue, resp, err := routingAPI.GetRoutingQueue(rs.Primary.ID)
		if queue != nil {
			return fmt.Errorf("Queue (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Queue not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All queues destroyed
	return nil
}

func testVerifyQueuesAndUsersDestroyed(state *terraform.State) error {
	routingAPI := platformclientv2.NewRoutingApi()
	usersAPI := platformclientv2.NewUsersApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_routing_queue" {
			queue, resp, err := routingAPI.GetRoutingQueue(rs.Primary.ID)
			if queue != nil {
				return fmt.Errorf("Queue (%s) still exists", rs.Primary.ID)
			} else if util.IsStatus404(resp) {
				// Queue not found as expected
				continue
			} else {
				// Unexpected error
				return fmt.Errorf("Unexpected error: %s", err)
			}
		}
		if rs.Type == "genesyscloud_user" {
			err := checkUserDeleted(rs.Primary.ID)(state)
			if err != nil {
				continue
			}
			user, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")
			if user != nil {
				return fmt.Errorf("User Resource (%s) still exists", rs.Primary.ID)
			} else if util.IsStatus404(resp) {
				// User not found as expected
				continue
			} else {
				// Unexpected error
				return fmt.Errorf("Unexpected error: %s", err)
			}
		}
	}
	// Success. All queues destroyed
	return nil
}

func validateMediaSettings(resourceName, settingsAttr, alertingTimeout, enableAutoAnswer, slPercent, slDurationMs string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, settingsAttr+".0.alerting_timeout_sec", alertingTimeout),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, settingsAttr+".0.service_level_percentage", slPercent),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, settingsAttr+".0.service_level_duration_ms", slDurationMs),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, settingsAttr+".0.enable_auto_answer", enableAutoAnswer),
	)
}

func validateAgentOwnedRouting(resourceName string, agentattr, enableAgentOwnedCallBacks string, maxOwnedCallBackHours string, maxOwnedCallBackDelayHours string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, agentattr+".0.enable_agent_owned_callbacks", enableAgentOwnedCallBacks),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, agentattr+".0.max_owned_callback_hours", maxOwnedCallBackHours),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, agentattr+".0.max_owned_callback_delay_hours", maxOwnedCallBackDelayHours),
	)
}

func generateRoutingQueueResourceBasic(resourceID string, name string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		name = "%s"
		%s
	}
	`, resourceID, name, strings.Join(nestedBlocks, "\n"))
}

// Used when testing skills group dependencies.
func generateRoutingQueueResourceBasicWithDepends(resourceID string, dependsOn string, name string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		depends_on = [%s]
		name = "%s"
		%s
	}
	`, resourceID, dependsOn, name, strings.Join(nestedBlocks, "\n"))
}

func generateDirectRouting(
	agentWaitSeconds string,
	waitForAgent string,
	callUseAgentAddressOutbound string,
	emailUseAgentAddressOutbound string,
	messageUseAgentAddressOutbound string,
	extraArgs ...string) string {
	return fmt.Sprintf(` direct_routing {
		agent_wait_seconds = %s
		wait_for_agent = %s
		call_use_agent_address_outbound = %s
		email_use_agent_address_outbound = %s
		message_use_agent_address_outbound = %s
		%s
	}
	`,
		agentWaitSeconds,
		waitForAgent,
		callUseAgentAddressOutbound,
		emailUseAgentAddressOutbound,
		messageUseAgentAddressOutbound,
		strings.Join(extraArgs, "\n"))
}

func validateRoutingRules(resourceName string, ringNum int, operator string, threshold string, waitSec string) resource.TestCheckFunc {
	ringNumStr := strconv.Itoa(ringNum)
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, "routing_rules."+ringNumStr+".operator", operator),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, "routing_rules."+ringNumStr+".threshold", threshold),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, "routing_rules."+ringNumStr+".wait_seconds", waitSec),
	)
}

func validateBullseyeSettings(resourceName string, numRings int, timeout string, skillToRemove string) resource.TestCheckFunc {
	var checks []resource.TestCheckFunc
	for i := 0; i < numRings; i++ {
		ringNum := strconv.Itoa(i)
		checks = append(checks,
			resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, "bullseye_rings."+ringNum+".expansion_timeout_seconds", timeout))

		if skillToRemove != "" {
			checks = append(checks,
				resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+resourceName, "bullseye_rings."+ringNum+".skills_to_remove.0", skillToRemove, "id"))
		} else {
			checks = append(checks,
				resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, "bullseye_rings."+ringNum+".skills_to_remove.#", "0"))
		}
	}
	return resource.ComposeAggregateTestCheckFunc(checks...)
}

func validateMember(queueResourceName string, userResourceName string, ringNum string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		queueResource, ok := state.RootModule().Resources[queueResourceName]
		if !ok {
			return fmt.Errorf("Failed to find queue %s in state", queueResourceName)
		}
		queueID := queueResource.Primary.ID

		userResource, ok := state.RootModule().Resources[userResourceName]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userResourceName)
		}
		userID := userResource.Primary.ID

		numMembersAttr, ok := queueResource.Primary.Attributes["members.#"]
		if !ok {
			return fmt.Errorf("No members found for queue %s in state", queueID)
		}

		numMembers, _ := strconv.Atoi(numMembersAttr)
		for i := 0; i < numMembers; i++ {
			if queueResource.Primary.Attributes["members."+strconv.Itoa(i)+".user_id"] == userID {
				if queueResource.Primary.Attributes["members."+strconv.Itoa(i)+".ring_num"] == ringNum {
					// Found user with correct ring
					return nil
				}
				return fmt.Errorf("Member %s found for queue %s with incorrect ring_num", userID, queueID)
			}
		}

		return fmt.Errorf("Member %s not found for queue %s in state", userID, queueID)
	}
}

// Validate groups and skill group fields.
func validateGroups(queueResourceName string, skillGroupResourceName string, groupResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		skillGroupResource, ok := state.RootModule().Resources[skillGroupResourceName]
		if !ok {
			return fmt.Errorf("Failed to find skillGroup %s in state", skillGroupResourceName)
		}

		groupResource, ok := state.RootModule().Resources[groupResourceName]
		if !ok {
			return fmt.Errorf("Failed to find group %s in state", groupResourceName)
		}

		queueResource, ok := state.RootModule().Resources[queueResourceName]
		if !ok {
			return fmt.Errorf("Failed to find queue %s in state", queueResourceName)
		}

		queueID := queueResource.Primary.ID
		skillGroupID := skillGroupResource.Primary.ID
		groupID := groupResource.Primary.ID

		numSkillGroupAttr, ok := queueResource.Primary.Attributes["skill_groups.#"]
		if !ok {
			return fmt.Errorf("No skill_groups found for queue %s in state", queueID)
		}

		numGroupAttr, ok := queueResource.Primary.Attributes["groups.#"]
		if !ok {
			return fmt.Errorf("No groups found for queue %s in state", queueID)
		}

		foundSkillGroup := false
		numSkillGroups, _ := strconv.Atoi(numSkillGroupAttr)
		for i := 0; i < numSkillGroups; i++ {
			if queueResource.Primary.Attributes["skill_groups."+strconv.Itoa(i)] == skillGroupID {
				foundSkillGroup = true
				break
			}
		}
		if !foundSkillGroup {
			return fmt.Errorf("Skill group id %s not found for queue %s in state", skillGroupID, queueID)
		}

		numGroups, _ := strconv.Atoi(numGroupAttr)
		for i := 0; i < numGroups; i++ {
			if queueResource.Primary.Attributes["groups."+strconv.Itoa(i)] == groupID {
				// Found  group
				return nil
			}
		}
		return fmt.Errorf("Group id %s not found for queue %s in state", groupID, queueID)
	}
}

func validateQueueWrapupCode(queueResourceName string, codeResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		queueResource, ok := state.RootModule().Resources[queueResourceName]
		if !ok {
			return fmt.Errorf("Failed to find queue %s in state", queueResourceName)
		}
		queueID := queueResource.Primary.ID

		codeResource, ok := state.RootModule().Resources[codeResourceName]
		if !ok {
			return fmt.Errorf("Failed to find code %s in state", codeResourceName)
		}
		codeID := codeResource.Primary.ID

		numCodesAttr, ok := queueResource.Primary.Attributes["wrapup_codes.#"]
		if !ok {
			return fmt.Errorf("No wrapup codes found for queue %s in state", queueID)
		}

		numCodes, _ := strconv.Atoi(numCodesAttr)
		for i := 0; i < numCodes; i++ {
			if queueResource.Primary.Attributes["wrapup_codes."+strconv.Itoa(i)] == codeID {
				// Found wrapup code
				return nil
			}
		}
		return fmt.Errorf("Wrapup code %s not found for queue %s in state", codeID, queueID)
	}
}

func validateDirectRouting(resourceName string,
	agentWaitSeconds string,
	waitForAgent string,
	callUseAgentAddressOutbound string,
	emailUseAgentAddressOutbound string,
	messageUseAgentAddressOutbound string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, "direct_routing.0.agent_wait_seconds", agentWaitSeconds),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, "direct_routing.0.wait_for_agent", waitForAgent),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, "direct_routing.0.call_use_agent_address_outbound", callUseAgentAddressOutbound),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, "direct_routing.0.email_use_agent_address_outbound", emailUseAgentAddressOutbound),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, "direct_routing.0.message_use_agent_address_outbound", messageUseAgentAddressOutbound),
	)
}

func TestAccResourceRoutingQueueSkillGroups(t *testing.T) {
	var (
		queueResource         = "test-queue-members-seg"
		queueName             = "Terraform-Test-QueueSkillGroup-" + uuid.NewString()
		groupResource         = "routing-group"
		groupName             = "group" + uuid.NewString()
		skillGroupResource    = "routing-skill-group"
		skillGroupName        = "Skillgroup" + uuid.NewString()
		skillGroupDescription = "description-" + uuid.NewString()
		testUserResource      = "user_resource1"
		testUserName          = "nameUser1" + uuid.NewString()
		testUserEmail         = uuid.NewString() + "@example.com"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(skillGroupResource, skillGroupName, skillGroupDescription) +
					group.GenerateBasicGroupResource(groupResource, groupName,
						group.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
					) +
					GenerateRoutingQueueResourceBasicWithDepends(
						queueResource,
						"genesyscloud_routing_skill_group."+skillGroupResource,
						queueName,
						"members = []",
						"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResource+".id]",
						"groups = [genesyscloud_group."+groupResource+".id]",
						GenerateBullseyeSettings("10"),
						GenerateBullseyeSettings("10"),
						GenerateBullseyeSettings("10")),
				Check: resource.ComposeTestCheckFunc(
					validateGroups("genesyscloud_routing_queue."+queueResource, "genesyscloud_routing_skill_group."+skillGroupResource, "genesyscloud_group."+groupResource),
				),

				PreventPostDestroyRefresh: true,
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResource,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"suppress_in_queue_call_recording",
				},
				Destroy: true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(45 * time.Second)
			return testVerifyQueuesAndUsersDestroyed(state)
		},
	})
}

func generateUserWithCustomAttrs(resourceID string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceID, email, name, strings.Join(attrs, "\n"))
}

func checkUserDeleted(id string) resource.TestCheckFunc {
	log.Printf("Fetching user with ID: %s\n", id)
	return func(s *terraform.State) error {
		maxAttempts := 30
		for i := 0; i < maxAttempts; i++ {
			deleted, err := isUserDeleted(id)
			if err != nil {
				return err
			}
			if deleted {
				return nil
			}
			time.Sleep(10 * time.Second)
		}
		return fmt.Errorf("user %s was not deleted properly", id)
	}
}

func isUserDeleted(id string) (bool, error) {
	mu.Lock()
	defer mu.Unlock()

	usersAPI := platformclientv2.NewUsersApi()
	// Attempt to get the user
	_, response, err := usersAPI.GetUser(id, nil, "", "")

	// Check if the user is not found (deleted)
	if response != nil && response.StatusCode == 404 {
		return true, nil // User is deleted
	}

	// Handle other errors
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return false, err
	}

	// If user is found, it means the user is not deleted
	return false, nil
}
