package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v94/platformclientv2"
)

func TestAccResourceOutboundCampaignRuleBasic(t *testing.T) {
	var (
		resourceId      = "campaign_rule"
		ruleName        = "Terraform test rule " + uuid.NewString()
		ruleNameUpdated = "Terraform test rule " + uuid.NewString()

		campaign1ResourceId  = "campaign1"
		campaign1Name        = "TF Test Campaign " + uuid.NewString()
		outboundFlowFilePath = "../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		campaign1FlowName    = "test flow " + uuid.NewString()
		campaign1Resource    = generateOutboundCampaignBasic(
			campaign1ResourceId,
			campaign1Name,
			"contact-list",
			"site",
			"+13178783419",
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
		campaign2Resource   = generateOutboundCampaignBasic(
			campaign2ResourceId,
			campaign2Name,
			"contact-list-2",
			"site-2",
			"+13178781119",
			"car-1",
			strconv.Quote("off"),
			outboundFlowFilePath,
			"campaignrule-test-flow-2",
			campaign2FlowName,
			"${data.genesyscloud_auth_division_home.home.name}",
			"campaignrule-test-location-2",
			"campaignrule-test-wrapupcode-2",
		)

		sequenceResourceId = "sequence"
		sequenceName       = "TF Test Sequence " + uuid.NewString()
		sequenceResource   = generateOutboundSequence(
			sequenceResourceId,
			sequenceName,
			[]string{"genesyscloud_outbound_campaign." + campaign1ResourceId + ".id"},
			nullValue,
			nullValue,
		)

		campaignRuleEntityCampaignIds = []string{"genesyscloud_outbound_campaign." + campaign1ResourceId + ".id"}
		campaignRuleEntitySequenceIds = []string{"genesyscloud_outbound_sequence." + sequenceResourceId + ".id"}

		campaignRuleActionType                = "turnOnCampaign"
		campaignRuleActionCampaignIds         = []string{"genesyscloud_outbound_campaign." + campaign2ResourceId + ".id"}
		campaignRuleActionSequenceIds         = []string{"genesyscloud_outbound_sequence." + sequenceResourceId + ".id"}
		campaignRuleActionUseTriggeringEntity = falseValue

		campaignRuleCondition1Type = "campaignProgress"
		paramRulesOperator         = "lessThan"
		paramRulesValue            = "0.4"
		paramRulesDialingMode      = "preview"
		paramRulesPriority         = "2"

		campaignRuleCondition1TypeUpdate = "campaignAgents"
		paramRulesOperatorUpdated        = "greaterThan"
		paramRulesValueUpdated           = "50"
		paramRulesDialingModeUpdated     = ""
		paramRulesPriorityUpdated        = ""
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) +
					sequenceResource +
					campaign1Resource +
					campaign2Resource +
					generateOutboundCampaignRule(
						resourceId,
						ruleName,
						falseValue,
						falseValue,
						generateCampaignRuleEntity(campaignRuleEntityCampaignIds, campaignRuleEntitySequenceIds),
						generateCampaignRuleConditions(
							"",
							campaignRuleCondition1Type,
							generateCampaignRuleParameters(
								paramRulesOperator,
								paramRulesValue,
								paramRulesDialingMode,
								paramRulesPriority,
							),
						),
						generateCampaignRuleActions(
							"",
							campaignRuleActionType,
							campaignRuleActionCampaignIds,
							campaignRuleActionSequenceIds,
							campaignRuleActionUseTriggeringEntity,
							generateCampaignRuleParameters(
								paramRulesOperator,
								paramRulesValue,
								paramRulesDialingMode,
								paramRulesPriority,
							),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "name", ruleName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "match_any_conditions", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "enabled", falseValue),

					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_entities.0.sequence_ids.0",
						"genesyscloud_outbound_sequence."+sequenceResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign1ResourceId, "id"),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_conditions.0.condition_type", campaignRuleCondition1Type),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_conditions.0.parameters.0.operator", paramRulesOperator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_conditions.0.parameters.0.value", paramRulesValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_conditions.0.parameters.0.dialing_mode", paramRulesDialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_conditions.0.parameters.0.priority", paramRulesPriority),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_actions.0.parameters.0.operator", paramRulesOperator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_actions.0.parameters.0.value", paramRulesValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_actions.0.parameters.0.dialing_mode", paramRulesDialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_actions.0.parameters.0.priority", paramRulesPriority),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_actions.0.campaign_rule_action_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign2ResourceId, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_actions.0.campaign_rule_action_entities.0.sequence_ids.0",
						"genesyscloud_outbound_sequence."+sequenceResourceId, "id"),
				),
			},
			// Update
			{
				Config: fmt.Sprintf(`
			data "genesyscloud_auth_division_home" "home" {}
			`) +
					sequenceResource +
					campaign1Resource +
					campaign2Resource +
					generateOutboundCampaignRule(
						resourceId,
						ruleNameUpdated,
						trueValue,
						trueValue,
						generateCampaignRuleEntity(
							campaignRuleEntityCampaignIds,
							campaignRuleEntitySequenceIds,
						),
						generateCampaignRuleConditions(
							"",
							campaignRuleCondition1TypeUpdate,
							generateCampaignRuleParameters(
								paramRulesOperatorUpdated,
								paramRulesValueUpdated,
								paramRulesDialingModeUpdated,
								paramRulesPriorityUpdated,
							),
						),
						generateCampaignRuleActions(
							"",
							campaignRuleActionType,
							campaignRuleActionCampaignIds,
							campaignRuleActionSequenceIds,
							campaignRuleActionUseTriggeringEntity,
							generateCampaignRuleParameters(
								paramRulesOperatorUpdated,
								paramRulesValueUpdated,
								paramRulesDialingModeUpdated,
								paramRulesPriorityUpdated,
							),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "name", ruleNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "match_any_conditions", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "enabled", trueValue),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_conditions.0.condition_type", campaignRuleCondition1TypeUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_conditions.0.parameters.0.operator", paramRulesOperatorUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_conditions.0.parameters.0.value", paramRulesValueUpdated),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_actions.0.parameters.0.operator", paramRulesOperatorUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_actions.0.parameters.0.value", paramRulesValueUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_actions.0.parameters.0.dialing_mode", paramRulesDialingModeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "campaign_rule_actions.0.parameters.0.priority", paramRulesPriorityUpdated),
				),
			},
			// Update (Setting 'enabled' back to false because we can't create or delete a rule with 'enabled' set to true)
			{
				Config: fmt.Sprintf(`
			data "genesyscloud_auth_division_home" "home" {}
			`) +
					sequenceResource +
					campaign1Resource +
					campaign2Resource +
					generateOutboundCampaignRule(
						resourceId,
						ruleNameUpdated,
						falseValue,
						trueValue,
						generateCampaignRuleEntity(
							campaignRuleEntityCampaignIds,
							campaignRuleEntitySequenceIds,
						),
						generateCampaignRuleConditions(
							"",
							campaignRuleCondition1TypeUpdate,
							generateCampaignRuleParameters(
								paramRulesOperatorUpdated,
								paramRulesValueUpdated,
								paramRulesDialingModeUpdated,
								paramRulesPriorityUpdated,
							),
						),
						generateCampaignRuleActions(
							"",
							campaignRuleActionType,
							campaignRuleActionCampaignIds,
							campaignRuleActionSequenceIds,
							campaignRuleActionUseTriggeringEntity,
							"",
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "enabled", falseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_campaignrule." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyCampaignRuleDestroyed,
	})
}

func generateOutboundCampaignRule(resourceId string, name string, enabled string, matchAnyConditions string, nestedBlocks ...string) string {
	return fmt.Sprintf(`
resource "genesyscloud_outbound_campaignrule" "%s" {
	name                 = "%s"
	enabled              = %s
	match_any_conditions = %s
	%s
}`, resourceId, name, enabled, matchAnyConditions, strings.Join(nestedBlocks, "\n"))
}

func generateCampaignRuleEntity(campaignIds []string, sequenceIds []string) string {
	return fmt.Sprintf(`
	campaign_rule_entities {
		campaign_ids = [%s]
		sequence_ids = [%s]
	}
`, strings.Join(campaignIds, ", "), strings.Join(sequenceIds, ", "))
}

func generateCampaignRuleActions(id string,
	actionType string,
	campaignIds []string,
	sequenceIds []string,
	useTriggeringEntity string,
	paramsBlock string,
) string {
	if useTriggeringEntity != "" {
		useTriggeringEntity = fmt.Sprintf("use_triggering_entity = %s", useTriggeringEntity)
	}
	return fmt.Sprintf(`
	campaign_rule_actions {
		id          = "%s" 
		action_type = "%s"
		campaign_rule_action_entities {
			campaign_ids          = [%s]
			sequence_ids          = [%s]
			%s
		}
		%s
	}
`, id, actionType, strings.Join(campaignIds, ", "), strings.Join(sequenceIds, ", "), useTriggeringEntity, paramsBlock)
}

func generateCampaignRuleConditions(id string, conditionType string, parametersBlock string) string {
	if id != "" {
		id = fmt.Sprintf(`id = "%s"`, id)
	}
	return fmt.Sprintf(`
	campaign_rule_conditions {
		%s
		condition_type = "%s"
		%s
	}
`, id, conditionType, parametersBlock)
}

func generateCampaignRuleParameters(operator string, value string, dialingMode string, priority string) string {
	if dialingMode != "" {
		dialingMode = fmt.Sprintf(`dialing_mode = "%s"`, dialingMode)
	}
	if priority != "" {
		priority = fmt.Sprintf(`priority = "%s"`, priority)
	}
	return fmt.Sprintf(`
		parameters {
			operator     = "%s"
			value        = "%s"
			%s	
			%s
		}
`, operator, value, dialingMode, priority)
}

func testVerifyCampaignRuleDestroyed(state *terraform.State) error {
	outboundApi := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_campaignrule" {
			continue
		}
		campaignRule, resp, err := outboundApi.GetOutboundCampaignrule(rs.Primary.ID)
		if campaignRule != nil {
			return fmt.Errorf("emergency group (%s) still exists", rs.Primary.ID)
		} else if isStatus404(resp) {
			// Campaign rule not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All campaign rules destroyed.
	return nil
}
