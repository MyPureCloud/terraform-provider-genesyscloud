package routing_queue_conditional_group_activation

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/group"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingSkillGroup "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill_group"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

var (
	mu sync.Mutex
)

func TestAccResourceRoutingQueueConditionalGroupActivation(t *testing.T) {
	var (
		cgaResourceLabel = "test-cga"

		queueResourceLabel = "test-queue"
		queueName1         = "Terraform Test Queue CGA-" + uuid.NewString()

		skillGroupResourceLabel = "skillgroup"
		skillGroupName          = "test skillgroup cga " + uuid.NewString()

		rule1ConditionExpression = "C1"
		rule1Operator            = "GreaterThan"
		rule1Metric              = "EstimatedWaitTime"
		rule1Value               = "60"
		rule1GroupType           = "SKILLGROUP"

		testUserResourceLabel = "user_resource1"
		testUserName          = "nameUser1" + uuid.NewString()
		testUserEmail         = uuid.NewString() + "@examplecgatest.com"

		groupResourceLabel = "group"
		groupName          = "terraform test group cga " + uuid.NewString()

		rule2ConditionExpression = "C1 or C2"
		rule2Operator1           = "GreaterThan"
		rule2Metric1             = "EstimatedWaitTime"
		rule2Value1              = "120"
		rule2Operator2           = "LessThan"
		rule2Metric2             = "IdleAgentCount"
		rule2Value2              = "2"
		rule2GroupType           = "GROUP"
		userID                   string
	)

	queueIdChan := make(chan string, 1)
	err := os.Setenv(featureToggles.CGAToggleName(), "enabled")
	if err != nil {
		t.Errorf("%s is not set", featureToggles.CGAToggleName())
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			util.TestAccPreCheck(t)
		},
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceLabel,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName1,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id]",
				),
				Check: resource.ComposeTestCheckFunc(
					func(state *terraform.State) error {
						resourceState, ok := state.RootModule().Resources["genesyscloud_routing_queue."+queueResourceLabel]
						if !ok {
							return fmt.Errorf("failed to find resource %s in state", "genesyscloud_routing_queue."+queueResourceLabel)
						}
						queueIdChan <- resourceState.Primary.ID
						return nil
					},
				),
			},
			{
				Config: routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceLabel,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName1,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id]",
				) + generateConditionalGroupActivation(
					cgaResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					generateCgaRuleBlock(
						rule1ConditionExpression,
						generateCgaConditionBlock(rule1Metric, "", rule1Operator, rule1Value),
						generateCgaGroupBlock(
							"genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id",
							rule1GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("genesyscloud_routing_queue."+queueResourceLabel, "id", checkQueueId(queueIdChan, false)),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.condition_expression", rule1ConditionExpression),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.operator", rule1Operator),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.simple_metric.0.metric", rule1Metric),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.value", rule1Value),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.groups.0.member_group_type", rule1GroupType),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "rules.0.groups.0.member_group_id", "genesyscloud_routing_skill_group."+skillGroupResourceLabel, "id",
					),
				),
			},
			{
				Config: generateUserWithCustomAttrs(
					testUserResourceLabel,
					testUserEmail,
					testUserName,
				) + group.GenerateBasicGroupResource(
					groupResourceLabel,
					groupName,
					group.GenerateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				) + routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceLabel,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName1,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id]",
					"groups = [genesyscloud_group."+groupResourceLabel+".id]",
				) + generateConditionalGroupActivation(
					cgaResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					generateCgaRuleBlock(
						rule1ConditionExpression,
						generateCgaConditionBlock(rule1Metric, "", rule1Operator, rule1Value),
						generateCgaGroupBlock(
							"genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id",
							rule1GroupType,
						),
					),
					generateCgaRuleBlock(
						rule2ConditionExpression,
						generateCgaConditionBlock(rule2Metric1, "genesyscloud_routing_queue."+queueResourceLabel+".id", rule2Operator1, rule2Value1)+
							generateCgaConditionBlock(rule2Metric2, "genesyscloud_routing_queue."+queueResourceLabel+".id", rule2Operator2, rule2Value2),
						generateCgaGroupBlock(
							"genesyscloud_group."+groupResourceLabel+".id",
							rule2GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("genesyscloud_routing_queue."+queueResourceLabel, "id", checkQueueId(queueIdChan, false)),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),

					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.condition_expression", rule1ConditionExpression),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.operator", rule1Operator),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.simple_metric.0.metric", rule1Metric),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.value", rule1Value),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.groups.0.member_group_type", rule1GroupType),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "rules.0.groups.0.member_group_id", "genesyscloud_routing_skill_group."+skillGroupResourceLabel, "id",
					),

					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.1.condition_expression", rule2ConditionExpression),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.1.conditions.0.operator", rule2Operator1),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.1.conditions.0.simple_metric.0.metric", rule2Metric1),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.1.conditions.0.value", rule2Value1),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "rules.1.conditions.0.simple_metric.0.queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.1.conditions.1.operator", rule2Operator2),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.1.conditions.1.simple_metric.0.metric", rule2Metric2),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.1.conditions.1.value", rule2Value2),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.1.groups.0.member_group_type", rule2GroupType),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "rules.1.groups.0.member_group_id", "genesyscloud_group."+groupResourceLabel, "id",
					),
				),
			},
			{
				Config: generateUserWithCustomAttrs(
					testUserResourceLabel,
					testUserEmail,
					testUserName,
				) + group.GenerateBasicGroupResource(
					groupResourceLabel,
					groupName,
					group.GenerateGroupOwners("genesyscloud_user."+testUserResourceLabel+".id"),
				) + routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceLabel,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName1,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id]",
					"groups = [genesyscloud_group."+groupResourceLabel+".id]",
				) + generateConditionalGroupActivation(
					cgaResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					generateCgaRuleBlock(
						rule2ConditionExpression,
						generateCgaConditionBlock(rule2Metric1, "genesyscloud_routing_queue."+queueResourceLabel+".id", rule2Operator1, rule2Value1)+
							generateCgaConditionBlock(rule2Metric2, "genesyscloud_routing_queue."+queueResourceLabel+".id", rule2Operator2, rule2Value2),
						generateCgaGroupBlock(
							"genesyscloud_group."+groupResourceLabel+".id",
							rule2GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("genesyscloud_routing_queue."+queueResourceLabel, "id", checkQueueId(queueIdChan, true)),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),

					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.condition_expression", rule2ConditionExpression),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.operator", rule2Operator1),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.simple_metric.0.metric", rule2Metric1),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.value", rule2Value1),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.1.operator", rule2Operator2),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.1.simple_metric.0.metric", rule2Metric2),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.1.value", rule2Value2),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.groups.0.member_group_type", rule2GroupType),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "rules.0.groups.0.member_group_id", "genesyscloud_group."+groupResourceLabel, "id",
					),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["genesyscloud_user."+testUserResourceLabel]
						if !ok {
							return fmt.Errorf("not found: %s", "genesyscloud_user."+testUserResourceLabel)
						}
						userID = rs.Primary.ID
						log.Printf("User ID: %s\n", userID)
						return nil
					},
				),
			},
			{
				ResourceName:      ResourceType + "." + cgaResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(60 * time.Second)
			return testVerifyGroupsAndUsersDestroyed(state)
		},
	})
}

func TestAccResourceRoutingQueueConditionalGroupActivationExists(t *testing.T) {
	var (
		cgaResourceLabel = "test-cga"

		queueResourceLabel = "test-queue"
		queueName1         = "Terraform Test Queue CGA Exists-" + uuid.NewString()
		queueName2         = "Terraform Test Queue CGA Exists-" + uuid.NewString()

		skillGroupResourceLabel = "skillgroup"
		skillGroupName          = "test skillgroup cga exists " + uuid.NewString()

		rule1ConditionExpression = "C1"
		rule1Operator            = "GreaterThan"
		rule1Metric              = "EstimatedWaitTime"
		rule1Value               = "60"
		rule1GroupType           = "SKILLGROUP"
	)

	err := os.Setenv(featureToggles.CGAToggleName(), "enabled")
	if err != nil {
		t.Errorf("%s is not set", featureToggles.CGAToggleName())
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			util.TestAccPreCheck(t)
		},
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceLabel,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName1,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id]",
				) + generateConditionalGroupActivation(
					cgaResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					generateCgaRuleBlock(
						rule1ConditionExpression,
						generateCgaConditionBlock(rule1Metric, "", rule1Operator, rule1Value),
						generateCgaGroupBlock(
							"genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id",
							rule1GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.condition_expression", rule1ConditionExpression),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.operator", rule1Operator),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.simple_metric.0.metric", rule1Metric),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.value", rule1Value),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.groups.0.member_group_type", rule1GroupType),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "rules.0.groups.0.member_group_id", "genesyscloud_routing_skill_group."+skillGroupResourceLabel, "id",
					),
				),
			},
			{
				Config: routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceLabel,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName2,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id]",
				) + generateConditionalGroupActivation(
					cgaResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					generateCgaRuleBlock(
						rule1ConditionExpression,
						generateCgaConditionBlock(rule1Metric, "", rule1Operator, rule1Value),
						generateCgaGroupBlock(
							"genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id",
							rule1GroupType,
						),
					),
				),
				Check: verifyConditionalGroupActivationExists("genesyscloud_routing_queue." + queueResourceLabel),
			},
			{
				ResourceName:      ResourceType + "." + cgaResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(40 * time.Second)
			return testVerifyGroupsAndUsersDestroyed(state)
		},
	})
}

func TestAccResourceRoutingQueueConditionalGroupActivationWithPilotRule(t *testing.T) {
	var (
		cgaResourceLabel = "test-cga-pilot"

		queueResourceLabel = "test-queue"
		queueName          = "Terraform Test Queue CGA Pilot-" + uuid.NewString()

		skillGroupResourceLabel = "skillgroup"
		skillGroupName          = "test skillgroup cga pilot " + uuid.NewString()

		pilotConditionExpression = "C1"
		pilotOperator            = "GreaterThan"
		pilotMetric              = "EstimatedWaitTime"
		pilotValue               = "30"

		rule1ConditionExpression = "C1"
		rule1Operator            = "GreaterThan"
		rule1Metric              = "EstimatedWaitTime"
		rule1Value               = "60"
		rule1GroupType           = "SKILLGROUP"
	)

	err := os.Setenv(featureToggles.CGAToggleName(), "enabled")
	if err != nil {
		t.Errorf("%s is not set", featureToggles.CGAToggleName())
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			util.TestAccPreCheck(t)
		},
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceLabel,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id]",
				) + generateConditionalGroupActivationWithPilot(
					cgaResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					generateCgaPilotRuleBlock(
						pilotConditionExpression,
						generateCgaConditionBlock(pilotMetric, "", pilotOperator, pilotValue),
					),
					generateCgaRuleBlock(
						rule1ConditionExpression,
						generateCgaConditionBlock(rule1Metric, "", rule1Operator, rule1Value),
						generateCgaGroupBlock(
							"genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id",
							rule1GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "pilot_rule.0.condition_expression", pilotConditionExpression),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "pilot_rule.0.conditions.0.operator", pilotOperator),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "pilot_rule.0.conditions.0.simple_metric.0.metric", pilotMetric),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "pilot_rule.0.conditions.0.value", pilotValue),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.condition_expression", rule1ConditionExpression),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.operator", rule1Operator),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.simple_metric.0.metric", rule1Metric),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.conditions.0.value", rule1Value),
					resource.TestCheckResourceAttr(ResourceType+"."+cgaResourceLabel, "rules.0.groups.0.member_group_type", rule1GroupType),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+cgaResourceLabel, "rules.0.groups.0.member_group_id", "genesyscloud_routing_skill_group."+skillGroupResourceLabel, "id",
					),
				),
			},
			{
				ResourceName:      ResourceType + "." + cgaResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(40 * time.Second)
			return testVerifyGroupsAndUsersDestroyed(state)
		},
	})
}

func verifyConditionalGroupActivationExists(queueResourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		queueResource, ok := state.RootModule().Resources[queueResourcePath]
		if !ok {
			return fmt.Errorf("failed to find queue %s in state", queueResourcePath)
		}
		queueID := queueResource.Primary.ID

		routingApi := platformclientv2.NewRoutingApi()
		queue, _, err := routingApi.GetRoutingQueue(queueID, nil)
		if err != nil {
			return err
		}

		if queue.ConditionalGroupActivation == nil {
			return fmt.Errorf("no conditional group activation found for queue %s %s", queueID, *queue.Name)
		}

		return nil
	}
}

func checkQueueId(queueIdChan chan string, closeChannel bool) func(value string) error {
	return func(value string) error {
		queueId, ok := <-queueIdChan
		if !ok {
			return fmt.Errorf("queue id channel closed unexpectedly")
		}

		if value != queueId {
			return fmt.Errorf("queue id not equal to expected. Expected: %s, Actual: %s", queueId, value)
		}

		if closeChannel {
			close(queueIdChan)
		} else {
			queueIdChan <- queueId
		}

		return nil
	}
}

func generateConditionalGroupActivation(resourceLabel string, queueId string, ruleBlocks ...string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		queue_id = %s
		%s
	}`, ResourceType, resourceLabel, queueId, strings.Join(ruleBlocks, "\n"))
}

func generateConditionalGroupActivationWithPilot(resourceLabel string, queueId string, pilotBlock string, ruleBlocks ...string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		queue_id = %s
		%s
		%s
	}`, ResourceType, resourceLabel, queueId, pilotBlock, strings.Join(ruleBlocks, "\n"))
}

func generateCgaPilotRuleBlock(conditionExpression string, conditionBlocks ...string) string {
	return fmt.Sprintf(`
		pilot_rule {
			condition_expression = "%s"
			%s
		}
	`, conditionExpression, strings.Join(conditionBlocks, "\n"))
}

func generateCgaRuleBlock(conditionExpression string, conditionAndGroupBlocks ...string) string {
	return fmt.Sprintf(`
		rules {
			condition_expression = "%s"
			%s
		}
	`, conditionExpression, strings.Join(conditionAndGroupBlocks, "\n"))
}

func generateCgaConditionBlock(metric string, queueId string, operator string, value string) string {
	queueIdAttr := ""
	if queueId != "" {
		queueIdAttr = fmt.Sprintf("queue_id = %s", queueId)
	}
	return fmt.Sprintf(`
		conditions {
			simple_metric {
				metric = "%s"
				%s
			}
			operator = "%s"
			value    = %s
		}
	`, metric, queueIdAttr, operator, value)
}

func generateCgaGroupBlock(groupId, groupType string) string {
	return fmt.Sprintf(`groups {
		member_group_id   = %s
		member_group_type = "%s"
	}
	`, groupId, groupType)
}

func generateUserWithCustomAttrs(resourceLabel string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceLabel, email, name, strings.Join(attrs, "\n"))
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
	_, response, err := usersAPI.GetUser(id, nil, "", "")

	if response != nil && response.StatusCode == 404 {
		return true, nil
	}

	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return false, err
	}

	return false, nil
}

func testVerifyGroupsAndUsersDestroyed(state *terraform.State) error {
	groupsAPI := platformclientv2.NewGroupsApi()
	usersAPI := platformclientv2.NewUsersApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_group" {
			grp, resp, err := groupsAPI.GetGroup(rs.Primary.ID)
			if grp != nil {
				return fmt.Errorf("Group (%s) still exists", rs.Primary.ID)
			} else if util.IsStatus404(resp) {
				continue
			} else {
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
				continue
			} else {
				return fmt.Errorf("Unexpected error: %s", err)
			}
		}
	}
	return nil
}
