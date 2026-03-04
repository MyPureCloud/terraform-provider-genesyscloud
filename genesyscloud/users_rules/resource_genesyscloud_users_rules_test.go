package users_rules

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

func TestAccResourceUsersRules(t *testing.T) {
	usersRuleResourceLabel1 := "test-users-rule-1"
	usersRuleName := "terraform-users-rule-" + uuid.NewString()
	usersRuleNameUpdated := "terraform-users-rule-updated-" + uuid.NewString()
	usersRuleDescription := "terraform-users-rule-description-" + uuid.NewString()
	usersRuleDescriptionUpdated := "terraform-users-rule-description-updated-" + uuid.NewString()
	usersRuleType := "Learning"
	usersRuleContextId := uuid.NewString()
	usersRuleMuId := uuid.NewString()

	criteria := []UsersRulesCriteriaStruct{
		{
			Operator: "Or",
			Group: []UsersRulesGroupItemStruct{
				{
					Operator:  "And",
					Container: "ManagementUnit",
					Values: []UsersRulesValueStruct{
						{
							ContextId: usersRuleContextId,
							Ids:       []string{usersRuleMuId},
						},
					},
				},
			},
		},
	}

	updatedCriteria := []UsersRulesCriteriaStruct{
		{
			Operator: "Or",
			Group: []UsersRulesGroupItemStruct{
				{
					Operator:  "And",
					Container: "ManagementUnit",
					Values: []UsersRulesValueStruct{
						{
							ContextId: uuid.NewString(),
							Ids:       []string{uuid.NewString(), uuid.NewString()},
						},
						{
							ContextId: uuid.NewString(),
							Ids:       []string{uuid.NewString()},
						},
					},
				},
				{
					Operator:  "Not",
					Container: "User",
					Values: []UsersRulesValueStruct{
						{
							Ids: []string{uuid.NewString(), uuid.NewString()},
						},
					},
				},
			},
		},
		{
			Operator: "Or",
			Group: []UsersRulesGroupItemStruct{
				{
					Operator:  "And",
					Container: "Language",
					Values: []UsersRulesValueStruct{
						{
							Ids: []string{uuid.NewString()},
						},
					},
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateUsersRulesResource(
					usersRuleResourceLabel1,
					usersRuleName,
					usersRuleDescription,
					usersRuleType,
					criteria,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "name", usersRuleName),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "description", usersRuleDescription),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "type", usersRuleType),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.operator", "Or"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.0.operator", "And"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.0.container", "ManagementUnit"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.0.values.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.0.values.0.context_id", usersRuleContextId),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.0.values.0.ids.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.0.values.0.ids.0", usersRuleMuId),
				),
			},
			{
				// Update
				Config: GenerateUsersRulesResource(
					usersRuleResourceLabel1,
					usersRuleNameUpdated,
					usersRuleDescriptionUpdated,
					usersRuleType,
					updatedCriteria,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "name", usersRuleNameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "description", usersRuleDescriptionUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "type", usersRuleType),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.operator", "Or"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.0.operator", "And"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.0.container", "ManagementUnit"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.0.values.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.0.values.0.ids.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.0.values.1.ids.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.1.operator", "Not"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.1.container", "User"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.1.values.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.1.values.0.context_id", ""),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.0.group.1.values.0.ids.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.1.operator", "Or"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.1.group.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.1.group.0.operator", "And"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.1.group.0.container", "Language"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.1.group.0.values.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.1.group.0.values.0.ids.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+usersRuleResourceLabel1, "criteria.1.group.0.values.0.context_id", ""),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + usersRuleResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUsersRulesDestroyed,
	})
}

func testVerifyUsersRulesDestroyed(state *terraform.State) error {
	usersRulesApi := platformclientv2.NewUsersRulesApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		usersRule, resp, err := usersRulesApi.GetUsersRule(rs.Primary.ID)
		if usersRule != nil {
			continue
		}

		if usersRule != nil {
			return fmt.Errorf("Users rule (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			// Users rule not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All Users Rules destroyed
	return nil
}
