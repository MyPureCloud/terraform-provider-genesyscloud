package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/MyPureCloud/platform-client-sdk-go/platformclientv2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	queueResource1           = "test-queue"
	queueResource2           = "test-queue-members"
	queueName1               = "Terraform Test Queue1-" + uuid.NewString()
	queueName2               = "Terraform Test Queue2-" + uuid.NewString()
	queueName3               = "Terraform Test Queue3-" + uuid.NewString()
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
	queueMemberResource1     = "test-queue-user1"
	queueMemberResource2     = "test-queue-user2"
	queueMemberEmail1        = "terraform1-" + uuid.NewString() + "@example.com"
	queueMemberEmail2        = "terraform2-" + uuid.NewString() + "@example.com"
	queueMemberName1         = "Henry Terraform"
	queueMemberName2         = "Amanda Terraform"
	defaultQueueRingNum      = "1"
	queueRingNum             = "3"
)

func TestAccResourceRoutingQueueBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingQueueResource(
					queueResource1,
					queueName1,
					queueDesc1,
					nullValue, // MANDATORY_TIMEOUT
					"200000",
					nullValue, // ALL
					nullValue, // auto_answer_only true
					nullValue, // No calling party name
					nullValue, // No calling party number
					nullValue, // enable_manual_assignment false
					nullValue, // enable_transcription false
					generateMediaSettings("media_settings_call", alertTimeout1, slPercent1, slDuration1),
					generateMediaSettings("media_settings_callback", alertTimeout1, slPercent1, slDuration1),
					generateMediaSettings("media_settings_chat", alertTimeout1, slPercent1, slDuration1),
					generateMediaSettings("media_settings_email", alertTimeout1, slPercent1, slDuration1),
					generateMediaSettings("media_settings_message", alertTimeout1, slPercent1, slDuration1),
					generateBullseyeSettings(alertTimeout1),
					generateBullseyeSettings(alertTimeout1),
					generateRoutingRules(routingRuleOpAny, "50", nullValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "name", queueName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "description", queueDesc1),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_wrapup_prompt", wrapupPromptMandTimeout),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_timeout_ms", "200000"),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "skill_evaluation_method", skillEvalAll),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "auto_answer_only", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_manual_assignment", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_transcription", falseValue),
					testDefaultHomeDivision("genesyscloud_routing_queue."+queueResource1),
					validateMediaSettings(queueResource1, "media_settings_call", alertTimeout1, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_callback", alertTimeout1, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_chat", alertTimeout1, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_email", alertTimeout1, slPercent1, slDuration1),
					validateMediaSettings(queueResource1, "media_settings_message", alertTimeout1, slPercent1, slDuration1),
					validateBullseyeSettings(queueResource1, 2, alertTimeout1, ""),
					validateRoutingRules(queueResource1, 0, routingRuleOpAny, "50", "5"),
				),
			},
			{
				// Update
				Config: generateRoutingQueueResource(
					queueResource1,
					queueName2,
					queueDesc2,
					strconv.Quote(wrapupPromptOptional),
					nullValue,
					strconv.Quote(skillEvalBest),
					falseValue, // auto_answer_only false
					strconv.Quote(callingPartyName),
					strconv.Quote(callingPartyNumber),
					trueValue, // enable_manual_assignment true
					trueValue, // enable_transcription true
					generateMediaSettings("media_settings_call", alertTimeout2, slPercent2, slDuration2),
					generateMediaSettings("media_settings_callback", alertTimeout2, slPercent2, slDuration2),
					generateMediaSettings("media_settings_chat", alertTimeout2, slPercent2, slDuration2),
					generateMediaSettings("media_settings_email", alertTimeout2, slPercent2, slDuration2),
					generateMediaSettings("media_settings_message", alertTimeout2, slPercent2, slDuration2),
					generateMediaSettings("media_settings_social", alertTimeout2, slPercent2, slDuration2),
					generateMediaSettings("media_settings_video", alertTimeout2, slPercent2, slDuration2),
					generateBullseyeSettings(alertTimeout2),
					generateBullseyeSettings(alertTimeout2),
					generateBullseyeSettings(alertTimeout2),
					generateRoutingRules(routingRuleOpMeetsThresh, "90", "30"),
					generateRoutingRules(routingRuleOpAny, "45", "15"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "name", queueName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "description", queueDesc2),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "acw_wrapup_prompt", wrapupPromptOptional),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "skill_evaluation_method", skillEvalBest),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "auto_answer_only", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "calling_party_name", callingPartyName),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "calling_party_number", callingPartyNumber),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_manual_assignment", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueResource1, "enable_transcription", trueValue),
					testDefaultHomeDivision("genesyscloud_routing_queue."+queueResource1),
					validateMediaSettings(queueResource1, "media_settings_call", alertTimeout2, slPercent2, slDuration2),
					validateMediaSettings(queueResource1, "media_settings_callback", alertTimeout2, slPercent2, slDuration2),
					validateMediaSettings(queueResource1, "media_settings_chat", alertTimeout2, slPercent2, slDuration2),
					validateMediaSettings(queueResource1, "media_settings_email", alertTimeout2, slPercent2, slDuration2),
					validateMediaSettings(queueResource1, "media_settings_message", alertTimeout2, slPercent2, slDuration2),
					validateMediaSettings(queueResource1, "media_settings_social", alertTimeout2, slPercent2, slDuration2),
					validateMediaSettings(queueResource1, "media_settings_video", alertTimeout2, slPercent2, slDuration2),
					validateBullseyeSettings(queueResource1, 3, alertTimeout2, ""),
					validateRoutingRules(queueResource1, 0, routingRuleOpMeetsThresh, "90", "30"),
					validateRoutingRules(queueResource1, 1, routingRuleOpAny, "45", "15"),
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

func TestAccResourceRoutingQueueMembers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingQueueResourceBasic(
					queueResource2,
					queueName3,
					generateMemberBlock("genesyscloud_user."+queueMemberResource1+".id", nullValue),
				) + generateBasicUserResource(
					queueMemberResource1,
					queueMemberEmail1,
					queueMemberName1,
				) + generateBasicUserResource(
					queueMemberResource2,
					queueMemberEmail2,
					queueMemberName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateMember("genesyscloud_routing_queue."+queueResource2, "genesyscloud_user."+queueMemberResource1, defaultQueueRingNum),
				),
			},
			{
				// Update with another queue member and modify rings
				Config: generateRoutingQueueResourceBasic(
					queueResource2,
					queueName3,
					generateMemberBlock("genesyscloud_user."+queueMemberResource1+".id", queueRingNum),
					generateMemberBlock("genesyscloud_user."+queueMemberResource2+".id", queueRingNum),
				) + generateBasicUserResource(
					queueMemberResource1,
					queueMemberEmail1,
					queueMemberName1,
				) + generateBasicUserResource(
					queueMemberResource2,
					queueMemberEmail2,
					queueMemberName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateMember("genesyscloud_routing_queue."+queueResource2, "genesyscloud_user."+queueMemberResource1, queueRingNum),
					validateMember("genesyscloud_routing_queue."+queueResource2, "genesyscloud_user."+queueMemberResource2, queueRingNum),
				),
			},
			{
				// Remove a queue member
				Config: generateRoutingQueueResourceBasic(
					queueResource2,
					queueName3,
					generateMemberBlock("genesyscloud_user."+queueMemberResource2+".id", queueRingNum),
				) + generateBasicUserResource(
					queueMemberResource1,
					queueMemberEmail1,
					queueMemberName1,
				) + generateBasicUserResource(
					queueMemberResource2,
					queueMemberEmail2,
					queueMemberName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateMember("genesyscloud_routing_queue."+queueResource2, "genesyscloud_user."+queueMemberResource2, queueRingNum),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue." + queueResource2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyQueuesDestroyed,
	})
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
		} else if resp != nil && resp.StatusCode == 404 {
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

func validateMediaSettings(resourceName string, settingsAttr string, alertingTimeout string, slPercent string, slDurationMs string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, settingsAttr+".0.alerting_timeout_sec", alertingTimeout),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, settingsAttr+".0.service_level_percentage", slPercent),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, settingsAttr+".0.service_level_duration_ms", slDurationMs),
	)
}

func generateRoutingQueueResourceBasic(resourceID string, name string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		name = "%s"
		%s
	}
	`, resourceID, name, strings.Join(nestedBlocks, "\n"))
}

func generateRoutingQueueResource(
	resourceID string,
	name string,
	desc string,
	acwWrapupPrompt string,
	acwTimeout string,
	skillEvalMethod string,
	autoAnswerOnly string,
	callingPartyName string,
	callingPartyNumber string,
	enableTranscription string,
	enableManualAssignment string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		name = "%s"
		description = "%s"
		acw_wrapup_prompt = %s
		acw_timeout_ms = %s
		skill_evaluation_method = %s
		auto_answer_only = %s
		calling_party_name = %s
		calling_party_number = %s
		enable_transcription = %s
  		enable_manual_assignment = %s
		%s
	}
	`, resourceID,
		name,
		desc,
		acwWrapupPrompt,
		acwTimeout,
		skillEvalMethod,
		autoAnswerOnly,
		callingPartyName,
		callingPartyNumber,
		enableTranscription,
		enableManualAssignment,
		strings.Join(nestedBlocks, "\n"))
}

func generateMediaSettings(attrName string, alertingTimeout string, slPercent string, slDurationMs string) string {
	return fmt.Sprintf(`%s {
		alerting_timeout_sec = %s
		service_level_percentage = %s
		service_level_duration_ms = %s
	}
	`, attrName, alertingTimeout, slPercent, slDurationMs)
}

func generateRoutingRules(operator string, threshold string, waitSeconds string) string {
	return fmt.Sprintf(`routing_rules {
		operator = "%s"
		threshold = %s
		wait_seconds = %s
	}
	`, operator, threshold, waitSeconds)
}

func generateBullseyeSettings(expTimeout string, skillsToRemove ...string) string {
	return fmt.Sprintf(`bullseye_rings {
		expansion_timeout_seconds = %s
		skills_to_remove = [%s]
	}
	`, expTimeout, strings.Join(skillsToRemove, ", "))
}

func generateMemberBlock(userID string, ringNum string) string {
	return fmt.Sprintf(`members {
		user_id = %s
		ring_num = %s
	}
	`, userID, ringNum)
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
				resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceName, "bullseye_rings."+ringNum+".skills_to_remove.0", skillToRemove))
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
