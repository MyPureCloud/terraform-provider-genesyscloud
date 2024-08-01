package outbound_campaignrule

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	outboundSequence "terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceOutboundCampaignRuleBasic(t *testing.T) {

	var (
		resourceId      = "campaign_rule"
		ruleName        = "Terraform test rule " + uuid.NewString()
		ruleNameUpdated = "Terraform test rule " + uuid.NewString()

		campaign1ResourceId  = "campaign1"
		campaign1Name        = "TF Test Campaign " + uuid.NewString()
		outboundFlowFilePath = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		campaign1FlowName    = "test flow " + uuid.NewString()
		campaign1Resource    = generateCampaignResourceForCampaignRuleTests(
			campaign1ResourceId,
			campaign1Name,
			"off",
			"contact-list",
			"test contact list"+uuid.NewString(),
			"location",
			"test location "+uuid.NewString(),
			fmt.Sprintf("+131783%v", 10000+rand.Intn(99999-10000)), // append random 5 digit number
			"site",
			"test site "+uuid.NewString(),
			"wrapupcode",
			"test wrapup code "+uuid.NewString(),
			"campaignrule-test-flow",
			outboundFlowFilePath,
			campaign1FlowName,
			"${data.genesyscloud_auth_division_home.home.name}",
			"car",
			"test car"+uuid.NewString(),
		)

		campaign2ResourceId = "campaign2"
		campaign2Name       = "TF Test Campaign " + uuid.NewString()
		campaign2FlowName   = "test flow " + uuid.NewString()
		campaign2Resource   = generateCampaignResourceForCampaignRuleTests(
			campaign2ResourceId,
			campaign2Name,
			"off",
			"contact-list-2",
			"test contact list"+uuid.NewString(),
			"location-2",
			"test location "+uuid.NewString(),
			fmt.Sprintf("+131782%v", 10000+rand.Intn(99999-10000)), // append random 5 digit number
			"site-2",
			"test site "+uuid.NewString(),
			"wrapupcode-2",
			"test wrapup code "+uuid.NewString(),
			"campaignrule-test-flow-2",
			outboundFlowFilePath,
			campaign2FlowName,
			"${data.genesyscloud_auth_division_home.home.name}",
			"car-2",
			"test car"+uuid.NewString(),
		)

		sequenceResourceId = "sequence"
		sequenceName       = "TF Test Sequence " + uuid.NewString()
		sequenceResource   = outboundSequence.GenerateOutboundSequence(
			sequenceResourceId,
			sequenceName,
			[]string{"genesyscloud_outbound_campaign." + campaign1ResourceId + ".id"},
			util.NullValue,
			util.NullValue,
		)

		campaignRuleEntityCampaignIds = []string{"genesyscloud_outbound_campaign." + campaign1ResourceId + ".id"}
		campaignRuleEntitySequenceIds = []string{"genesyscloud_outbound_sequence." + sequenceResourceId + ".id"}

		campaignRuleActionType                = "turnOnCampaign"
		campaignRuleActionCampaignIds         = []string{"genesyscloud_outbound_campaign." + campaign2ResourceId + ".id"}
		campaignRuleActionSequenceIds         = []string{"genesyscloud_outbound_sequence." + sequenceResourceId + ".id"}
		campaignRuleActionUseTriggeringEntity = util.FalseValue

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
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
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
						util.FalseValue, // enabled
						util.FalseValue, // matchAnyConditions
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "match_any_conditions", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "enabled", util.FalseValue),

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
						util.TrueValue, // enabled
						util.TrueValue, // matchAnyConditions
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "match_any_conditions", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "enabled", util.TrueValue),

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
						util.FalseValue, // enabled
						util.TrueValue,  // matchAnyConditions
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "enabled", util.FalseValue),
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

func TestAccResourceOutboundCampaignRuleEnabledAtCreation(t *testing.T) {
	var (
		resourceId      = "campaign_rule"
		ruleName        = "Terraform test rule " + uuid.NewString()
		ruleNameUpdated = "Terraform test rule " + uuid.NewString()

		campaign1ResourceId  = "campaign1"
		campaign1Name        = "TF Test Campaign " + uuid.NewString()
		outboundFlowFilePath = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		campaign1FlowName    = "test flow " + uuid.NewString()
		campaign1Resource    = generateCampaignResourceForCampaignRuleTests(
			campaign1ResourceId,
			campaign1Name,
			"off",
			"contact-list",
			"test contact list"+uuid.NewString(),
			"location",
			"test location "+uuid.NewString(),
			fmt.Sprintf("+131783%v", 10000+rand.Intn(99999-10000)), // append random 5 digit number
			"site",
			"test site "+uuid.NewString(),
			"wrapupcode",
			"test wrapup code "+uuid.NewString(),
			"campaignrule-test-flow",
			outboundFlowFilePath,
			campaign1FlowName,
			"${data.genesyscloud_auth_division_home.home.name}",
			"car",
			"test car"+uuid.NewString(),
		)

		campaign2ResourceId = "campaign2"
		campaign2Name       = "TF Test Campaign " + uuid.NewString()
		campaign2FlowName   = "test flow " + uuid.NewString()
		campaign2Resource   = generateCampaignResourceForCampaignRuleTests(
			campaign2ResourceId,
			campaign2Name,
			"off",
			"contact-list-2",
			"test contact list"+uuid.NewString(),
			"location-2",
			"test location "+uuid.NewString(),
			fmt.Sprintf("+131782%v", 10000+rand.Intn(99999-10000)), // append random 5 digit number
			"site-2",
			"test site "+uuid.NewString(),
			"wrapupcode-2",
			"test wrapup code "+uuid.NewString(),
			"campaignrule-test-flow-2",
			outboundFlowFilePath,
			campaign2FlowName,
			"${data.genesyscloud_auth_division_home.home.name}",
			"car-2",
			"test car"+uuid.NewString(),
		)

		sequenceResourceId = "sequence"
		sequenceName       = "TF Test Sequence " + uuid.NewString()
		sequenceResource   = outboundSequence.GenerateOutboundSequence(
			sequenceResourceId,
			sequenceName,
			[]string{"genesyscloud_outbound_campaign." + campaign1ResourceId + ".id"},
			util.NullValue,
			util.NullValue,
		)

		campaignRuleEntityCampaignIds = []string{"genesyscloud_outbound_campaign." + campaign1ResourceId + ".id"}
		campaignRuleEntitySequenceIds = []string{"genesyscloud_outbound_sequence." + sequenceResourceId + ".id"}

		campaignRuleActionType                = "turnOffCampaign"
		campaignRuleActionCampaignIds         = []string{"genesyscloud_outbound_campaign." + campaign2ResourceId + ".id"}
		campaignRuleActionSequenceIds         = []string{"genesyscloud_outbound_sequence." + sequenceResourceId + ".id"}
		campaignRuleActionUseTriggeringEntity = util.FalseValue

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
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
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
						util.TrueValue,  // enabled
						util.FalseValue, // matchAnyConditions
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "match_any_conditions", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "enabled", util.TrueValue),

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
						util.FalseValue, // enabled
						util.TrueValue,  // matchAnyConditions
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceId, "enabled", util.FalseValue),
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
		} else if util.IsStatus404(resp) {
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

func generateCampaignResourceForCampaignRuleTests(
	campaignResourceId,
	campaignName,
	campaignStatus,
	contactListResourceId,
	contactListName,
	locationResourceId,
	locationName,
	locationEmergencyNumber,
	siteResourceId,
	siteName,
	wrapupCodeResourceId,
	wrapupCodeName,
	flowResourceId,
	flowFilePath,
	flowName,
	flowDivisionName,
	carResourceId,
	carName string) string {

	fullyQualifiedPath, _ := filepath.Abs(flowFilePath)

	return fmt.Sprintf(`
resource "genesyscloud_outbound_campaign" "%s" {
	name                          = "%s"
	dialing_mode                  = "agentless"
	caller_name                   = "Test Name"
	caller_address                = "+353371111111"
	outbound_line_count           = 2
	campaign_status               = "%s"
	contact_list_id               = genesyscloud_outbound_contact_list.%s.id
	site_id                       = genesyscloud_telephony_providers_edges_site.%s.id
	call_analysis_response_set_id = genesyscloud_outbound_callanalysisresponseset.%s.id
	phone_columns {
		column_name = "Cell"
	}
}

resource "genesyscloud_outbound_contact_list" "%s" {
	name 						 = "%s"
	preview_mode_column_name     = "Cell"
	preview_mode_accepted_values = ["Cell"]
	column_names                 = ["Cell", "Home", "zipcode"]
	automatic_time_zone_mapping  = false
	phone_columns {
		column_name = "Cell"
		type        = "cell"
		callable_time_column = "Cell"
	}
	phone_columns {
		column_name = "Home"
		type        = "home"
		callable_time_column = "Home"
	}
}

resource "genesyscloud_location" "%s" {
    name  = "%s"
	notes = "HQ1"
	path  = []
	emergency_number {
		number = "%s"
		type   = null
	}
	address {
		street1  = "7601 Interactive Way"
		city     = "Indianapolis"
		state    = "IN"
		country  = "US"
		zip_code = "46278"
	}
}

resource "genesyscloud_telephony_providers_edges_site" "%s" {
	name                            = "%s"
	description                     = "TestAccResourceSite description 1"
	location_id                     = genesyscloud_location.%s.id
	media_model                     = "Cloud"
	media_regions_use_latency_based = false
}

resource "genesyscloud_routing_wrapupcode" "%s" {
	name = "%s"
}

resource "genesyscloud_flow" "%s" {
        filepath          = "%s"
        file_content_hash =  filesha256("%s")
        force_unlock      = false
        substitutions = {
			flow_name          = "%s"
			home_division_name = "%s"
			contact_list_name  = "${genesyscloud_outbound_contact_list.%s.name}"
			wrapup_code_name   = "${genesyscloud_routing_wrapupcode.%s.name}"
		}
}

resource "genesyscloud_outbound_callanalysisresponseset" "%s" {
	name                   = "%s"
	beep_detection_enabled = false
	responses {
		callable_person {
			reaction_type = "transfer_flow"
			name = "%s"
			data = "${genesyscloud_flow.%s.id}"
		}
	}
}
	`, campaignResourceId,
		campaignName,
		campaignStatus,
		contactListResourceId, // genesyscloud_outbound_campaign.contact_list_id
		siteResourceId,        // genesyscloud_outbound_campaign.site_id
		carResourceId,         // genesyscloud_outbound_campaign.call_analysis_response_set_id
		contactListResourceId,
		contactListName,
		locationResourceId,
		locationName,
		locationEmergencyNumber,
		siteResourceId,
		siteName,
		locationResourceId, // genesyscloud_telephony_providers_edges_site.location_id
		wrapupCodeResourceId,
		wrapupCodeName,
		flowResourceId,
		flowFilePath,
		fullyQualifiedPath,
		flowName,
		flowDivisionName,
		contactListResourceId, // genesyscloud_flow
		wrapupCodeResourceId,  // genesyscloud_flow
		carResourceId,
		carName,
		flowName,       // genesyscloud_outbound_callanalysisresponseset.responses.callable_person.name
		flowResourceId, // genesyscloud_outbound_callanalysisresponseset.responses.callable_person.data
	)
}
