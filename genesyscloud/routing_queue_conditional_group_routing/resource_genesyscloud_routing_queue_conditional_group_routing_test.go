package routing_queue_conditional_group_routing

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
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

var (
	mu sync.Mutex
)

func TestAccResourceRoutingQueueConditionalGroupRouting(t *testing.T) {
	var (
		conditionalGroupRoutingResourceLabel = "test-conditional-routing-group"

		queueResourceLabel = "test-queue"
		queueName1         = "Terraform Test Queue1-" + uuid.NewString()

		skillGroupResourceLabel                    = "skillgroup"
		skillGroupName                             = "test skillgroup " + uuid.NewString()
		conditionalGroupRoutingRule1Operator       = "LessThanOrEqualTo"
		conditionalGroupRoutingRule1Metric         = "EstimatedWaitTime"
		conditionalGroupRoutingRule1ConditionValue = "0"
		conditionalGroupRoutingRule1WaitSeconds    = "20"
		conditionalGroupRoutingRule1GroupType      = "SKILLGROUP"

		testUserResourceLabel = "user_resource1"
		testUserName          = "nameUser1" + uuid.NewString()
		testUserEmail         = uuid.NewString() + "@exampletest.com"

		groupResourceLabel = "group"
		groupName          = "terraform test group" + uuid.NewString()

		conditionalGroupRoutingRule2Operator       = "GreaterThanOrEqualTo"
		conditionalGroupRoutingRule2Metric         = "EstimatedWaitTime"
		conditionalGroupRoutingRule2ConditionValue = "5"
		conditionalGroupRoutingRule2WaitSeconds    = "15"
		conditionalGroupRoutingRule2GroupType      = "GROUP"
		userID                                     string
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
				// Create rule
				Config: routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceLabel,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName1,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id]",
				) + generateConditionalGroupRouting(
					conditionalGroupRoutingResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					generateConditionalGroupRoutingRuleBlock(
						conditionalGroupRoutingRule1Operator,
						conditionalGroupRoutingRule1Metric,
						conditionalGroupRoutingRule1ConditionValue,
						conditionalGroupRoutingRule1WaitSeconds,
						generateConditionalGroupRoutingRuleGroupBlock(
							"genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id",
							conditionalGroupRoutingRule1GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("genesyscloud_routing_queue."+queueResourceLabel, "id", checkQueueId(queueIdChan, false)),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+conditionalGroupRoutingResourceLabel, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.operator", conditionalGroupRoutingRule1Operator),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.metric", conditionalGroupRoutingRule1Metric),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.condition_value", conditionalGroupRoutingRule1ConditionValue),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.wait_seconds", conditionalGroupRoutingRule1WaitSeconds),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.groups.0.member_group_type", conditionalGroupRoutingRule1GroupType),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.groups.0.member_group_id", "genesyscloud_routing_skill_group."+skillGroupResourceLabel, "id",
					),
				),
			},
			{
				// Add rule
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
				) + generateConditionalGroupRouting(
					conditionalGroupRoutingResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					generateConditionalGroupRoutingRuleBlock(
						conditionalGroupRoutingRule1Operator,
						conditionalGroupRoutingRule1Metric,
						conditionalGroupRoutingRule1ConditionValue,
						conditionalGroupRoutingRule1WaitSeconds,
						generateConditionalGroupRoutingRuleGroupBlock(
							"genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id",
							conditionalGroupRoutingRule1GroupType,
						),
					),
					generateConditionalGroupRoutingRuleBlock(
						conditionalGroupRoutingRule2Operator,
						conditionalGroupRoutingRule2Metric,
						conditionalGroupRoutingRule2ConditionValue,
						conditionalGroupRoutingRule2WaitSeconds,
						"evaluated_queue_id = genesyscloud_routing_queue."+queueResourceLabel+".id",
						generateConditionalGroupRoutingRuleGroupBlock(
							"genesyscloud_group."+groupResourceLabel+".id",
							conditionalGroupRoutingRule2GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("genesyscloud_routing_queue."+queueResourceLabel, "id", checkQueueId(queueIdChan, false)),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+conditionalGroupRoutingResourceLabel, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),

					// Rule 1
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.operator", conditionalGroupRoutingRule1Operator),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.metric", conditionalGroupRoutingRule1Metric),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.condition_value", conditionalGroupRoutingRule1ConditionValue),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.wait_seconds", conditionalGroupRoutingRule1WaitSeconds),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.groups.0.member_group_type", conditionalGroupRoutingRule1GroupType),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResourceLabel, "rules.0.groups.0.member_group_id", "genesyscloud_routing_skill_group."+skillGroupResourceLabel, "id",
					),

					// Rule 2
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResourceLabel, "rules.1.evaluated_queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.1.operator", conditionalGroupRoutingRule2Operator),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.1.metric", conditionalGroupRoutingRule2Metric),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.1.condition_value", conditionalGroupRoutingRule2ConditionValue),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.1.wait_seconds", conditionalGroupRoutingRule2WaitSeconds),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.1.groups.0.member_group_type", conditionalGroupRoutingRule2GroupType),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.1.groups.0.member_group_id", "genesyscloud_group."+groupResourceLabel, "id",
					),
				),
			},
			{
				// Remove the skill group rule
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
				) + generateConditionalGroupRouting(
					conditionalGroupRoutingResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					generateConditionalGroupRoutingRuleBlock(
						conditionalGroupRoutingRule2Operator,
						conditionalGroupRoutingRule2Metric,
						conditionalGroupRoutingRule2ConditionValue,
						conditionalGroupRoutingRule2WaitSeconds,
						generateConditionalGroupRoutingRuleGroupBlock(
							"genesyscloud_group."+groupResourceLabel+".id",
							conditionalGroupRoutingRule2GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith("genesyscloud_routing_queue."+queueResourceLabel, "id", checkQueueId(queueIdChan, true)),
					resource.TestCheckResourceAttrPair(
						ResourceType+"."+conditionalGroupRoutingResourceLabel, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),

					// Rule 1
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.operator", conditionalGroupRoutingRule2Operator),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.metric", conditionalGroupRoutingRule2Metric),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.condition_value", conditionalGroupRoutingRule2ConditionValue),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.wait_seconds", conditionalGroupRoutingRule2WaitSeconds),
					resource.TestCheckResourceAttr(ResourceType+"."+conditionalGroupRoutingResourceLabel, "rules.0.groups.0.member_group_type", conditionalGroupRoutingRule2GroupType),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResourceLabel, "rules.0.groups.0.member_group_id", "genesyscloud_group."+groupResourceLabel, "id",
					),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["genesyscloud_user."+testUserResourceLabel]
						if !ok {
							return fmt.Errorf("not found: %s", "genesyscloud_user."+testUserResourceLabel)
						}
						userID = rs.Primary.ID
						log.Printf("User ID: %s\n", userID) // Print user ID
						return nil
					},
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + conditionalGroupRoutingResourceLabel,
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

func TestAccResourceRoutingQueueConditionalGroupRoutingExists(t *testing.T) {
	var (
		conditionalGroupRoutingResourceLabel = "test-conditional-routing-group"

		queueResourceLabel = "test-queue"
		queueName1         = "Terraform Test Queue-" + uuid.NewString()
		queueName2         = "Terraform Test Queue-" + uuid.NewString()

		skillGroupResourceLabel = "skillgroup"
		skillGroupName          = "test skillgroup " + uuid.NewString()

		conditionalGroupRoutingRule1Operator       = "LessThanOrEqualTo"
		conditionalGroupRoutingRule1Metric         = "EstimatedWaitTime"
		conditionalGroupRoutingRule1ConditionValue = "0"
		conditionalGroupRoutingRule1WaitSeconds    = "20"
		conditionalGroupRoutingRule1GroupType      = "SKILLGROUP"
	)

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
				// Create rule
				Config: routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceLabel,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName1,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id]",
				) + generateConditionalGroupRouting(
					conditionalGroupRoutingResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					generateConditionalGroupRoutingRuleBlock(
						conditionalGroupRoutingRule1Operator,
						conditionalGroupRoutingRule1Metric,
						conditionalGroupRoutingRule1ConditionValue,
						conditionalGroupRoutingRule1WaitSeconds,
						generateConditionalGroupRoutingRuleGroupBlock(
							"genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id",
							conditionalGroupRoutingRule1GroupType,
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResourceLabel, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResourceLabel, "rules.0.operator", conditionalGroupRoutingRule1Operator),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResourceLabel, "rules.0.metric", conditionalGroupRoutingRule1Metric),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResourceLabel, "rules.0.condition_value", conditionalGroupRoutingRule1ConditionValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResourceLabel, "rules.0.wait_seconds", conditionalGroupRoutingRule1WaitSeconds),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResourceLabel, "rules.0.groups.0.member_group_type", conditionalGroupRoutingRule1GroupType),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+conditionalGroupRoutingResourceLabel, "rules.0.groups.0.member_group_id", "genesyscloud_routing_skill_group."+skillGroupResourceLabel, "id",
					),
				),
			},
			{
				// Update queue
				Config: routingSkillGroup.GenerateRoutingSkillGroupResourceBasic(
					skillGroupResourceLabel,
					skillGroupName,
					"description",
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName2,
					"skill_groups = [genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id]",
				) + generateConditionalGroupRouting(
					conditionalGroupRoutingResourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".id",
					generateConditionalGroupRoutingRuleBlock(
						conditionalGroupRoutingRule1Operator,
						conditionalGroupRoutingRule1Metric,
						conditionalGroupRoutingRule1ConditionValue,
						conditionalGroupRoutingRule1WaitSeconds,
						generateConditionalGroupRoutingRuleGroupBlock(
							"genesyscloud_routing_skill_group."+skillGroupResourceLabel+".id",
							conditionalGroupRoutingRule1GroupType,
						),
					),
				),
				Check: verifyConditionalGroupRoutingExists("genesyscloud_routing_queue." + queueResourceLabel),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue_conditional_group_routing." + conditionalGroupRoutingResourceLabel,
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

func verifyConditionalGroupRoutingExists(queueResourcePath string) resource.TestCheckFunc {
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

		if queue.ConditionalGroupRouting == nil {
			return fmt.Errorf("no conditional group routing found for queue %s %s", queueID, *queue.Name)
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

func generateConditionalGroupRouting(resourceLabel string, queueId string, nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		queue_id = %s
		%s
	}`, ResourceType, resourceLabel, queueId, strings.Join(nestedBlocks, "\n"))
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

func testVerifyGroupsAndUsersDestroyed(state *terraform.State) error {
	groupsAPI := platformclientv2.NewGroupsApi()
	usersAPI := platformclientv2.NewUsersApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type == "genesyscloud_group" {
			group, resp, err := groupsAPI.GetGroup(rs.Primary.ID)
			if group != nil {
				return fmt.Errorf("Group (%s) still exists", rs.Primary.ID)
			} else if util.IsStatus404(resp) {
				// Group not found as expected
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
	return nil
}
