package outbound_campaignrule

import (
	"fmt"
	"math/rand"
	"strconv"
	outboundCampaign "terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundCampaignRule(t *testing.T) {
	t.Parallel()

	var (
		campaignRuleResourceId = "campaign_rule"
		campaignRuleName       = "test-campaign-rule-" + uuid.NewString()

		campaign1ResourceId  = "campaign1"
		campaign1Name        = "TF Test Campaign " + uuid.NewString()
		outboundFlowFilePath = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		campaign1FlowName    = "test flow " + uuid.NewString()
		campaign1Resource    = outboundCampaign.GenerateOutboundCampaignBasic(
			campaign1ResourceId,
			campaign1Name,
			"contact-list",
			"site",
			fmt.Sprintf("+131784%v", 10000+rand.Intn(99999-10000)), // append random 5 digit number
			"car",
			strconv.Quote("off"),
			outboundFlowFilePath,
			"campaignrule-test-flow",
			campaign1FlowName,
			"${data.genesyscloud_auth_division_home.home.name}",
			"campaignrule-test-location",
			"campaignrule-test-wrapupcode",
		)

		campaign2ResourceId = "campaign2"
		campaign2Name       = "TF Test Campaign " + uuid.NewString()
		campaign2FlowName   = "test flow " + uuid.NewString()
		campaign2Resource   = outboundCampaign.GenerateOutboundCampaignBasic(
			campaign2ResourceId,
			campaign2Name,
			"contact-list-2",
			"site-2",
			fmt.Sprintf("+131785%v", 10000+rand.Intn(99999-10000)), // append random 5 digit number
			"car-1",
			strconv.Quote("off"),
			outboundFlowFilePath,
			"campaignrule-test-flow-2",
			campaign2FlowName,
			"${data.genesyscloud_auth_division_home.home.name}",
			"campaignrule-test-location-2",
			"campaignrule-test-wrapupcode-2",
		)

		dataSourceId = "campaign_rule_data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) +
					campaign1Resource +
					campaign2Resource +
					generateOutboundCampaignRule(
						campaignRuleResourceId,
						campaignRuleName,
						util.FalseValue,
						util.FalseValue,
						generateCampaignRuleEntity(
							[]string{"genesyscloud_outbound_campaign." + campaign1ResourceId + ".id"},
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
							[]string{"genesyscloud_outbound_campaign." + campaign2ResourceId + ".id"},
							[]string{},
							util.FalseValue,
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
