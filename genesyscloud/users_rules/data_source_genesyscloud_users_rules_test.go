package users_rules

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUsersRules(t *testing.T) {
	var (
		usersRuleResourceLabel     = "users-rule"
		usersRuleDataResourceLabel = "users-rule-data"

		usersRuleName        = "terraform-users-rule-" + uuid.NewString()
		usersRuleDescription = "terraform-users-rule-description-" + uuid.NewString()
		usersRuleType        = "Learning"
	)

	criteria := []UsersRulesCriteriaStruct{
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
				Config: GenerateUsersRulesResource(
					usersRuleResourceLabel,
					usersRuleName,
					usersRuleDescription,
					usersRuleType,
					criteria,
				) + generateUsersRulesDataSource(
					usersRuleDataResourceLabel,
					usersRuleName,
					ResourceType+"."+usersRuleResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+usersRuleDataResourceLabel, "id",
						"genesyscloud_users_rules."+usersRuleResourceLabel, "id",
					),
				),
			},
		},
	})
}

func generateUsersRulesDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, ResourceType, resourceLabel, name, dependsOnResource)
}
