package routing_queue_conditional_group_routing

import (
	"fmt"
	"os"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/group"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceRoutingQueueConditionalGroupRouting(t *testing.T) {
	var (
		conditionalGroupRoutingResource = "test-conditional-routing-group"

		queueResource = "test-queue"
		queueName1    = "Terraform Test Queue1-" + uuid.NewString()

		skillGroupResourceId = "skillgroup"
		skillGroupName       = "test skillgroup " + uuid.NewString()

		conditionalGroupRoutingRule1Operator       = "LessThanOrEqualTo"
		conditionalGroupRoutingRule1Metric         = "EstimatedWaitTime"
		conditionalGroupRoutingRule1ConditionValue = "0"
		conditionalGroupRoutingRule1WaitSeconds    = "20"
		conditionalGroupRoutingRule1GroupType      = "SKILLGROUP"

		testUserResource = "user_resource1"
		testUserName     = "nameUser1" + uuid.NewString()
		testUserEmail    = uuid.NewString() + "@example.com"

		groupResourceId = "group"
		groupName       = "terraform test group" + uuid.NewString()

		conditionalGroupRoutingRule2Operator       = "GreaterThanOrEqualTo"
		conditionalGroupRoutingRule2Metric         = "EstimatedWaitTime"
		conditionalGroupRoutingRule2ConditionValue = "5"
		conditionalGroupRoutingRule2WaitSeconds    = "15"
		conditionalGroupRoutingRule2GroupType      = "GROUP"
	)

	// Use this to save the id of the parent queue
	queueIdChan := make(chan string, 1)
	err := os.Setenv(featureToggles.CSGToggleName(), "enabled")
	if err != nil {
		t.Errorf("%s is not set", featureToggles.CSGToggleName())
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			util.TestAccPreCheck(t)
		},
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create the queue first so we can save the id to a channel and use it in the later test steps
				// The reason we are doing this is that we need to verify the parent queue is never dropped and recreated because of CGR
				Config: gcloud.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceId,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName1,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceId+".id]",
				),
				Check: resource.ComposeTestCheckFunc(
					func(state *terraform.State) error {
						resourceState, ok := state.RootModule().Resources["genesyscloud_routing_queue."+queueResource]
						if !ok {
							return fmt.Errorf("failed to find resource %s in state", "genesyscloud_routing_queue."+queueResource)
						}
						queueIdChan <- resourceState.Primary.ID

						return nil
					},
				),
			},
			{
				// Create rule
				Config: gcloud.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceId,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName1,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceId+".id]",
				) + generateConditionalGroupRouting(
					conditionalGroupRoutingResource,
					"genesyscloud_routing_queue."+queueResource+".id",
					generateConditionalGroupRoutingRuleBlock(
						conditionalGroupRoutingRule1Operator,
						conditionalGroupRoutingRule1Metric,
						conditionalGroupRoutingRule1ConditionValue,
						conditionalGroupRoutingRule1WaitSeconds,
						generateConditionalGroupRoutingRuleGroupBlock(
							"genesyscloud_routing_skill_group."+skillGroupResourceId+".id",
							conditionalGroupRoutingRule1GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("genesyscloud_routing_queue."+queueResource, "id", checkQueueId(queueIdChan, false)),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "queue_id", "genesyscloud_routing_queue."+queueResource, "id",
					),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.operator", conditionalGroupRoutingRule1Operator),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.metric", conditionalGroupRoutingRule1Metric),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.condition_value", conditionalGroupRoutingRule1ConditionValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.wait_seconds", conditionalGroupRoutingRule1WaitSeconds),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.groups.0.member_group_type", conditionalGroupRoutingRule1GroupType),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.groups.0.member_group_id", "genesyscloud_routing_skill_group."+skillGroupResourceId, "id",
					),
				),
			},
			{
				// Add rule
				Config: generateUserWithCustomAttrs(
					testUserResource,
					testUserEmail,
					testUserName,
				) + group.GenerateBasicGroupResource(
					groupResourceId,
					groupName,
					group.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + gcloud.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceId,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName1,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceId+".id]",
					"groups = [genesyscloud_group."+groupResourceId+".id]",
				) + generateConditionalGroupRouting(
					conditionalGroupRoutingResource,
					"genesyscloud_routing_queue."+queueResource+".id",
					generateConditionalGroupRoutingRuleBlock(
						conditionalGroupRoutingRule1Operator,
						conditionalGroupRoutingRule1Metric,
						conditionalGroupRoutingRule1ConditionValue,
						conditionalGroupRoutingRule1WaitSeconds,
						generateConditionalGroupRoutingRuleGroupBlock(
							"genesyscloud_routing_skill_group."+skillGroupResourceId+".id",
							conditionalGroupRoutingRule1GroupType,
						),
					),
					generateConditionalGroupRoutingRuleBlock(
						conditionalGroupRoutingRule2Operator,
						conditionalGroupRoutingRule2Metric,
						conditionalGroupRoutingRule2ConditionValue,
						conditionalGroupRoutingRule2WaitSeconds,
						"evaluated_queue_id = genesyscloud_routing_queue."+queueResource+".id",
						generateConditionalGroupRoutingRuleGroupBlock(
							"genesyscloud_group."+groupResourceId+".id",
							conditionalGroupRoutingRule2GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("genesyscloud_routing_queue."+queueResource, "id", checkQueueId(queueIdChan, false)),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "queue_id", "genesyscloud_routing_queue."+queueResource, "id",
					),

					// Rule 1
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.operator", conditionalGroupRoutingRule1Operator),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.metric", conditionalGroupRoutingRule1Metric),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.condition_value", conditionalGroupRoutingRule1ConditionValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.wait_seconds", conditionalGroupRoutingRule1WaitSeconds),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.groups.0.member_group_type", conditionalGroupRoutingRule1GroupType),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.groups.0.member_group_id", "genesyscloud_routing_skill_group."+skillGroupResourceId, "id",
					),

					// Rule 2
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.1.evaluated_queue_id", "genesyscloud_routing_queue."+queueResource, "id",
					),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.1.operator", conditionalGroupRoutingRule2Operator),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.1.metric", conditionalGroupRoutingRule2Metric),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.1.condition_value", conditionalGroupRoutingRule2ConditionValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.1.wait_seconds", conditionalGroupRoutingRule2WaitSeconds),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.1.groups.0.member_group_type", conditionalGroupRoutingRule2GroupType),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.1.groups.0.member_group_id", "genesyscloud_group."+groupResourceId, "id",
					),
				),
			},
			{
				// Remove a rule
				Config: generateUserWithCustomAttrs(
					testUserResource,
					testUserEmail,
					testUserName,
				) + group.GenerateBasicGroupResource(
					groupResourceId,
					groupName,
					group.GenerateGroupOwners("genesyscloud_user."+testUserResource+".id"),
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName1,
					"groups = [genesyscloud_group."+groupResourceId+".id]",
				) + generateConditionalGroupRouting(
					conditionalGroupRoutingResource,
					"genesyscloud_routing_queue."+queueResource+".id",
					generateConditionalGroupRoutingRuleBlock(
						conditionalGroupRoutingRule2Operator,
						conditionalGroupRoutingRule2Metric,
						conditionalGroupRoutingRule2ConditionValue,
						conditionalGroupRoutingRule2WaitSeconds,
						generateConditionalGroupRoutingRuleGroupBlock(
							"genesyscloud_group."+groupResourceId+".id",
							conditionalGroupRoutingRule2GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("genesyscloud_routing_queue."+queueResource, "id", checkQueueId(queueIdChan, true)),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "queue_id", "genesyscloud_routing_queue."+queueResource, "id",
					),

					// Rule 1
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.operator", conditionalGroupRoutingRule2Operator),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.metric", conditionalGroupRoutingRule2Metric),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.condition_value", conditionalGroupRoutingRule2ConditionValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.wait_seconds", conditionalGroupRoutingRule2WaitSeconds),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.groups.0.member_group_type", conditionalGroupRoutingRule2GroupType),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResource, "rules.0.groups.0.member_group_id", "genesyscloud_group."+groupResourceId, "id",
					),
					func(s *terraform.State) error {
						time.Sleep(60 * time.Second) // Wait for 60 seconds for resources to get deleted properly
						return nil
					},
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue_conditional_group_routing." + conditionalGroupRoutingResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
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

func generateConditionalGroupRouting(resourceId string, queueId string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		queue_id = %s
		%s
	}`, resourceName, resourceId, queueId, strings.Join(nestedBlocks, "\n"))
}

func generateConditionalGroupRoutingRuleBlock(operator, metric, conditionValue, waitSeconds string, nestedBlocks ...string) string {
	return fmt.Sprintf(`
		rules {
			operator = "%s"
			metric = "%s"
			condition_value = %s
			wait_seconds = %s
			%s
		}
	`, operator, metric, conditionValue, waitSeconds, strings.Join(nestedBlocks, "\n"))
}

func generateConditionalGroupRoutingRuleGroupBlock(groupId, groupType string) string {
	return fmt.Sprintf(`groups {
		member_group_id   = %s
		member_group_type = "%s"
	}
	`, groupId, groupType)
}

func generateUserWithCustomAttrs(resourceID string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceID, email, name, strings.Join(attrs, "\n"))
}
