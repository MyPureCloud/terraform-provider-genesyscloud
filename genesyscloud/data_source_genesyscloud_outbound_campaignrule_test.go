package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"testing"
)

func TestAccDataSourceCampaignRule(t *testing.T) {
	var (
		campaignRuleResourceId = "campaign_rule"
		campaignRuleName       = "test-campaign-rule-" + uuid.NewString()

		// TODO: Replace campaign ID with reference to campaign resource.
		campaignRuleEntityCampaignIds = []string{strconv.Quote("e6b0a237-2620-4644-b1bb-f0e58e923a93")}

		// TODO: Replace campaign ID with reference to campaign resource.
		campaignRuleActionCampaignIds = []string{strconv.Quote("7fc6b00a-f2f5-44d2-9fc5-169f339f6c4b")}

		dataSourceId = "campaign_rule_data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOutboundCampaignRule(
					campaignRuleResourceId,
					campaignRuleName,
					falseValue,
					falseValue,
					generateCampaignRuleEntity(
						campaignRuleEntityCampaignIds,
						[]string{},
					),
					generateCampaignRuleConditions(
						"",
						"campaignProgress",
						generateCampaignRuleParameters(
							"lessThan",
							"0.5",
							"preview",
							"2",
						),
					),
					generateCampaignRuleActions(
						"",
						"turnOnCampaign",
						campaignRuleActionCampaignIds,
						[]string{},
						falseValue,
						generateCampaignRuleParameters(
							"lessThan",
							"0.5",
							"preview",
							"2",
						),
					),
				) + generateCampaignRuleDataSource(
					dataSourceId,
					campaignRuleName,
					"genesyscloud_outbound_campaignrule."+campaignRuleResourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_campaignrule."+dataSourceId, "id",
						"genesyscloud_outbound_campaignrule."+campaignRuleResourceId, "id"),
				),
			},
		},
	})
}

func generateCampaignRuleDataSource(dataSourceId string, campaignRuleName string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_campaignrule" "%s" {
		name = "%s"
		depends_on = [%s]
}`, dataSourceId, campaignRuleName, dependsOn)
}
