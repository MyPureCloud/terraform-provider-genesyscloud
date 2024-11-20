package outbound_campaignrule

import (
	"fmt"
	"math/rand"
	"strconv"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
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
		campaignRuleResourceLabel = "campaign_rule"
		campaignRuleName          = "test-campaign-rule-" + uuid.NewString()
		divResourceLabel          = "test-outbound-campaignrule-division"
		divName                   = "terraform-" + uuid.NewString()

		campaign1ResourceLabel = "campaign1"
		campaign1Name          = "TF Test Campaign " + uuid.NewString()
		outboundFlowFilePath   = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		campaign1FlowName      = "test flow " + uuid.NewString()
		campaign1Resource      = outboundCampaign.GenerateOutboundCampaignBasic(
			campaign1ResourceLabel,
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
			divResourceLabel,
		)

		campaign2ResourceLabel = "campaign2"
		campaign2Name          = "TF Test Campaign " + uuid.NewString()
		campaign2FlowName      = "test flow " + uuid.NewString()
		campaign2Resource      = outboundCampaign.GenerateOutboundCampaignBasic(
			campaign2ResourceLabel,
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
			divResourceLabel,
		)
		dataSourceLabel = "campaign_rule_data"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: `data "genesyscloud_auth_division_home" "home" {}` + "\n" +
					authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					campaign1Resource +
					campaign2Resource +
					generateOutboundCampaignRule(
						campaignRuleResourceLabel,
						campaignRuleName,
						util.FalseValue,
						util.FalseValue,
						generateCampaignRuleEntity(
							[]string{"genesyscloud_outbound_campaign." + campaign1ResourceLabel + ".id"},
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
							[]string{"genesyscloud_outbound_campaign." + campaign2ResourceLabel + ".id"},
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
					dataSourceLabel,
					campaignRuleName,
					"genesyscloud_outbound_campaignrule."+campaignRuleResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_campaignrule."+dataSourceLabel, "id",
						"genesyscloud_outbound_campaignrule."+campaignRuleResourceLabel, "id"),
				),
			},
		},
	})
}

func generateCampaignRuleDataSource(dataSourceLabel string, campaignRuleName string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_campaignrule" "%s" {
	name = "%s"
	depends_on = [%s]
}`, dataSourceLabel, campaignRuleName, dependsOn)
}
