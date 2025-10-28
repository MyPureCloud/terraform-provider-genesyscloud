package outbound_campaignrule

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	outboundSequence "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceOutboundCampaignRuleBasic(t *testing.T) {

	var (
		resourceLabel   = "campaign_rule"
		ruleName        = "Terraform test rule " + uuid.NewString()
		ruleNameUpdated = "Terraform test rule " + uuid.NewString()

		campaign1ResourceLabel = "campaign1"
		campaign1Name          = "TF Test Campaign " + uuid.NewString()
		outboundFlowFilePath   = filepath.Join(testrunner.RootDir, "examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml")
		campaign1FlowName      = "test flow " + uuid.NewString()
		campaign1Resource      = generateCampaignResourceForCampaignRuleTests(
			campaign1ResourceLabel,
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

		campaign2ResourceLabel = "campaign2"
		campaign2Name          = "TF Test Campaign " + uuid.NewString()
		campaign2FlowName      = "test flow " + uuid.NewString()
		campaign2Resource      = generateCampaignResourceForCampaignRuleTests(
			campaign2ResourceLabel,
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

		sequenceResourceLabel = "sequence"
		sequenceName          = "TF Test Sequence " + uuid.NewString()
		sequenceResource      = outboundSequence.GenerateOutboundSequence(
			sequenceResourceLabel,
			sequenceName,
			[]string{"genesyscloud_outbound_campaign." + campaign1ResourceLabel + ".id"},
			util.NullValue,
			util.NullValue,
		)

		campaignRuleEntityCampaignIds = []string{"genesyscloud_outbound_campaign." + campaign1ResourceLabel + ".id"}
		campaignRuleEntitySequenceIds = []string{"genesyscloud_outbound_sequence." + sequenceResourceLabel + ".id"}

		campaignRuleActionType                = "turnOnCampaign"
		campaignRuleActionCampaignIds         = []string{"genesyscloud_outbound_campaign." + campaign2ResourceLabel + ".id"}
		campaignRuleActionSequenceIds         = []string{"genesyscloud_outbound_sequence." + sequenceResourceLabel + ".id"}
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
						resourceLabel,
						ruleName,
						util.FalseValue, // enabled
						util.FalseValue, // matchAnyConditions
						generateCampaignRuleEntity(
							campaignRuleEntityCampaignIds,
							campaignRuleEntitySequenceIds,
							[]string{},
							[]string{},
						),
						generateCampaignRuleConditions(
							"",
							campaignRuleCondition1Type,
							generateCampaignRuleParameters(
								paramRulesOperator,
								paramRulesValue,
								paramRulesDialingMode,
								paramRulesPriority,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleActions(
							"",
							campaignRuleActionType,
							campaignRuleActionCampaignIds,
							campaignRuleActionSequenceIds,
							[]string{},
							[]string{},
							campaignRuleActionUseTriggeringEntity,
							generateCampaignRuleParameters(
								paramRulesOperator,
								paramRulesValue,
								paramRulesDialingMode,
								paramRulesPriority,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "name", ruleName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "match_any_conditions", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "enabled", util.FalseValue),

					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_entities.0.sequence_ids.0",
						"genesyscloud_outbound_sequence."+sequenceResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign1ResourceLabel, "id"),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.condition_type", campaignRuleCondition1Type),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.operator", paramRulesOperator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.value", paramRulesValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.dialing_mode", paramRulesDialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.priority", paramRulesPriority),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.operator", paramRulesOperator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.value", paramRulesValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.dialing_mode", paramRulesDialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.priority", paramRulesPriority),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.campaign_rule_action_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign2ResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.campaign_rule_action_entities.0.sequence_ids.0",
						"genesyscloud_outbound_sequence."+sequenceResourceLabel, "id"),
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
						resourceLabel,
						ruleNameUpdated,
						util.TrueValue, // enabled
						util.TrueValue, // matchAnyConditions
						generateCampaignRuleEntity(
							campaignRuleEntityCampaignIds,
							campaignRuleEntitySequenceIds,
							[]string{},
							[]string{},
						),
						generateCampaignRuleConditions(
							"",
							campaignRuleCondition1TypeUpdate,
							generateCampaignRuleParameters(
								paramRulesOperatorUpdated,
								paramRulesValueUpdated,
								paramRulesDialingModeUpdated,
								paramRulesPriorityUpdated,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleActions(
							"",
							campaignRuleActionType,
							campaignRuleActionCampaignIds,
							campaignRuleActionSequenceIds,
							[]string{},
							[]string{},
							campaignRuleActionUseTriggeringEntity,
							generateCampaignRuleParameters(
								paramRulesOperatorUpdated,
								paramRulesValueUpdated,
								paramRulesDialingModeUpdated,
								paramRulesPriorityUpdated,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "name", ruleNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "match_any_conditions", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "enabled", util.TrueValue),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.condition_type", campaignRuleCondition1TypeUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.operator", paramRulesOperatorUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.value", paramRulesValueUpdated),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.operator", paramRulesOperatorUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.value", paramRulesValueUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.dialing_mode", paramRulesDialingModeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.priority", paramRulesPriorityUpdated),
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
						resourceLabel,
						ruleNameUpdated,
						util.FalseValue, // enabled
						util.TrueValue,  // matchAnyConditions
						generateCampaignRuleEntity(
							campaignRuleEntityCampaignIds,
							campaignRuleEntitySequenceIds,
							[]string{},
							[]string{},
						),
						generateCampaignRuleConditions(
							"",
							campaignRuleCondition1TypeUpdate,
							generateCampaignRuleParameters(
								paramRulesOperatorUpdated,
								paramRulesValueUpdated,
								paramRulesDialingModeUpdated,
								paramRulesPriorityUpdated,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleActions(
							"",
							campaignRuleActionType,
							campaignRuleActionCampaignIds,
							campaignRuleActionSequenceIds,
							[]string{},
							[]string{},
							campaignRuleActionUseTriggeringEntity,
							"",
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "enabled", util.FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_campaignrule." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyCampaignRuleDestroyed,
	})
}

func TestAccResourceOutboundCampaignRuleEnabledAtCreation(t *testing.T) {
	var (
		resourceLabel   = "campaign_rule"
		ruleName        = "Terraform test rule " + uuid.NewString()
		ruleNameUpdated = "Terraform test rule " + uuid.NewString()

		campaign1ResourceLabel = "campaign1"
		campaign1Name          = "TF Test Campaign " + uuid.NewString()
		outboundFlowFilePath   = filepath.Join(testrunner.RootDir, "examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml")
		campaign1FlowName      = "test flow " + uuid.NewString()
		campaign1Resource      = generateCampaignResourceForCampaignRuleTests(
			campaign1ResourceLabel,
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

		campaign2ResourceLabel = "campaign2"
		campaign2Name          = "TF Test Campaign " + uuid.NewString()
		campaign2FlowName      = "test flow " + uuid.NewString()
		campaign2Resource      = generateCampaignResourceForCampaignRuleTests(
			campaign2ResourceLabel,
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

		sequenceResourceLabel = "sequence"
		sequenceName          = "TF Test Sequence " + uuid.NewString()
		sequenceResource      = outboundSequence.GenerateOutboundSequence(
			sequenceResourceLabel,
			sequenceName,
			[]string{"genesyscloud_outbound_campaign." + campaign1ResourceLabel + ".id"},
			util.NullValue,
			util.NullValue,
		)

		campaignRuleEntityCampaignIds = []string{"genesyscloud_outbound_campaign." + campaign1ResourceLabel + ".id"}
		campaignRuleEntitySequenceIds = []string{"genesyscloud_outbound_sequence." + sequenceResourceLabel + ".id"}

		campaignRuleActionType                = "turnOffCampaign"
		campaignRuleActionCampaignIds         = []string{"genesyscloud_outbound_campaign." + campaign2ResourceLabel + ".id"}
		campaignRuleActionSequenceIds         = []string{"genesyscloud_outbound_sequence." + sequenceResourceLabel + ".id"}
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
						resourceLabel,
						ruleName,
						util.TrueValue,  // enabled
						util.FalseValue, // matchAnyConditions
						generateCampaignRuleEntity(
							campaignRuleEntityCampaignIds,
							campaignRuleEntitySequenceIds,
							[]string{},
							[]string{},
						),
						generateCampaignRuleConditions(
							"",
							campaignRuleCondition1Type,
							generateCampaignRuleParameters(
								paramRulesOperator,
								paramRulesValue,
								paramRulesDialingMode,
								paramRulesPriority,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleActions(
							"",
							campaignRuleActionType,
							campaignRuleActionCampaignIds,
							campaignRuleActionSequenceIds,
							[]string{},
							[]string{},
							campaignRuleActionUseTriggeringEntity,
							generateCampaignRuleParameters(
								paramRulesOperator,
								paramRulesValue,
								paramRulesDialingMode,
								paramRulesPriority,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "name", ruleName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "match_any_conditions", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "enabled", util.TrueValue),

					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_entities.0.sequence_ids.0",
						"genesyscloud_outbound_sequence."+sequenceResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign1ResourceLabel, "id"),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.condition_type", campaignRuleCondition1Type),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.operator", paramRulesOperator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.value", paramRulesValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.dialing_mode", paramRulesDialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.priority", paramRulesPriority),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.operator", paramRulesOperator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.value", paramRulesValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.dialing_mode", paramRulesDialingMode),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.priority", paramRulesPriority),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.campaign_rule_action_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign2ResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.campaign_rule_action_entities.0.sequence_ids.0",
						"genesyscloud_outbound_sequence."+sequenceResourceLabel, "id"),
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
						resourceLabel,
						ruleNameUpdated,
						util.FalseValue, // enabled
						util.TrueValue,  // matchAnyConditions
						generateCampaignRuleEntity(
							campaignRuleEntityCampaignIds,
							campaignRuleEntitySequenceIds,
							[]string{},
							[]string{},
						),
						generateCampaignRuleConditions(
							"",
							campaignRuleCondition1TypeUpdate,
							generateCampaignRuleParameters(
								paramRulesOperatorUpdated,
								paramRulesValueUpdated,
								paramRulesDialingModeUpdated,
								paramRulesPriorityUpdated,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleActions(
							"",
							campaignRuleActionType,
							campaignRuleActionCampaignIds,
							campaignRuleActionSequenceIds,
							[]string{},
							[]string{},
							campaignRuleActionUseTriggeringEntity,
							"",
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "enabled", util.FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_campaignrule." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyCampaignRuleDestroyed,
	})
}

func TestAccResourceOutboundCampaignRuleActions(t *testing.T) {

	var (
		resourceLabel = "campaign_rule"
		ruleName      = "Terraform test rule " + uuid.NewString()

		queueLabel    = "queue1"
		queueNameAttr = "Terraform test queue " + uuid.NewString()

		campaign1ResourceLabel = "campaign1"
		campaign1Name          = "TF Test Campaign " + uuid.NewString()
		outboundFlowFilePath   = filepath.Join(testrunner.RootDir, "examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml")
		campaign1FlowName      = "test flow " + uuid.NewString()
		campaign1Resource      = generateCampaignResourceForCampaignRuleTests(
			campaign1ResourceLabel,
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

		campaign2ResourceLabel = "campaign2"
		campaign2Name          = "TF Test Campaign " + uuid.NewString()
		campaign2FlowName      = "test flow " + uuid.NewString()
		campaign2Resource      = generateCampaignResourceForCampaignRuleTests(
			campaign2ResourceLabel,
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

		campaignRuleEntityCampaignIds = []string{"genesyscloud_outbound_campaign." + campaign1ResourceLabel + ".id"}

		// Condition 1
		condition1Type     = "campaignRecordsAttempted"
		condition1Operator = "lessThan"
		condition1Value    = "10"

		// Condition 2
		condition2Type     = "campaignValidAttempts"
		condition2Operator = "lessThanEqualTo"
		condition2Value    = "0.2"

		// Condition 3
		condition3Type     = "campaignRightPartyContacts"
		condition3Operator = "greaterThan"
		condition3Value    = "30"

		// Condition 1 update
		condition1TypeUpdate      = "campaignBusinessSuccess"
		condition1OperatorUpdated = "greaterThan"
		condition1ValueUpdated    = "5"

		// Condition 2 update
		condition2TypeUpdate      = "campaignBusinessNeutral"
		condition2OperatorUpdated = "greaterThanEqualTo"
		condition2ValueUpdated    = "15"

		// Condition 3 update
		condition3TypeUpdate      = "campaignBusinessFailure"
		condition3OperatorUpdated = "equals"
		condition3ValueUpdated    = "25"

		// Action 1
		action1Type       = "setCampaignAbandonRate"
		actionCampaignIds = []string{"genesyscloud_outbound_campaign." + campaign2ResourceLabel + ".id"}
		abandonRate       = "6.7"

		// Action 2
		action2Type   = "setCampaignNumberOfLines"
		outboundLines = "0"

		// Action 1 update
		action1TypeUpdated = "setCampaignWeight"
		relativeWeight     = "50"

		// Action 2 update
		action2TypeUpdated = "setCampaignMaxCallsPerAgent"
		maxCpa             = "7.8"

		// Action 3 update
		action3TypeUpdated = "changeCampaignQueue"
		queueId            = fmt.Sprintf("%s.%s.id", routingQueue.ResourceType, queueLabel)
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
					campaign1Resource +
					campaign2Resource +
					generateOutboundCampaignRule(
						resourceLabel,
						ruleName,
						util.FalseValue, // enabled
						util.FalseValue, // matchAnyConditions
						generateCampaignRuleEntity(
							campaignRuleEntityCampaignIds,
							[]string{},
							[]string{},
							[]string{},
						),
						generateCampaignRuleConditions(
							"",
							condition1Type,
							generateCampaignRuleParameters(
								condition1Operator,
								condition1Value,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleConditions(
							"",
							condition2Type,
							generateCampaignRuleParameters(
								condition2Operator,
								condition2Value,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleConditions(
							"",
							condition3Type,
							generateCampaignRuleParameters(
								condition3Operator,
								condition3Value,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleActions(
							"",
							action1Type,
							actionCampaignIds,
							[]string{},
							[]string{},
							[]string{},
							"",
							generateCampaignRuleParameters(
								"",
								"",
								"",
								"",
								abandonRate,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleActions(
							"",
							action2Type,
							actionCampaignIds,
							[]string{},
							[]string{},
							[]string{},
							"",
							generateCampaignRuleParameters(
								"",
								"",
								"",
								"",
								"",
								outboundLines,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "name", ruleName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "match_any_conditions", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "enabled", util.FalseValue),

					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign1ResourceLabel, "id"),

					// Condition 1
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.condition_type", condition1Type),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.operator", condition1Operator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.value", condition1Value),

					// Condition 2
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.1.condition_type", condition2Type),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.1.parameters.0.operator", condition2Operator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.1.parameters.0.value", condition2Value),

					// Condition 3
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.2.condition_type", condition3Type),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.2.parameters.0.operator", condition3Operator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.2.parameters.0.value", condition3Value),

					// Action 1
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.action_type", action1Type),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.abandon_rate", abandonRate),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.campaign_rule_action_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign2ResourceLabel, "id"),

					// Action 2
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.1.action_type", action2Type),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.1.parameters.0.outbound_line_count", outboundLines),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.1.campaign_rule_action_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign2ResourceLabel, "id"),
				),
			},
			// Update
			{
				Config: fmt.Sprintf(`
			data "genesyscloud_auth_division_home" "home" {}
			`) +
					routingQueue.GenerateRoutingQueueResourceBasic(queueLabel, queueNameAttr) +
					campaign1Resource +
					campaign2Resource +
					generateOutboundCampaignRule(
						resourceLabel,
						ruleName,
						util.TrueValue, // enabled
						util.TrueValue, // matchAnyConditions
						generateCampaignRuleEntity(
							campaignRuleEntityCampaignIds,
							[]string{},
							[]string{},
							[]string{},
						),
						generateCampaignRuleConditions(
							"",
							condition1TypeUpdate,
							generateCampaignRuleParameters(
								condition1OperatorUpdated,
								condition1ValueUpdated,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleConditions(
							"",
							condition2TypeUpdate,
							generateCampaignRuleParameters(
								condition2OperatorUpdated,
								condition2ValueUpdated,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleConditions(
							"",
							condition3TypeUpdate,
							generateCampaignRuleParameters(
								condition3OperatorUpdated,
								condition3ValueUpdated,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleActions(
							"",
							action1TypeUpdated,
							actionCampaignIds,
							[]string{},
							[]string{},
							[]string{},
							"",
							generateCampaignRuleParameters(
								"",
								"",
								"",
								"",
								"",
								"",
								relativeWeight,
								"",
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleActions(
							"",
							action2TypeUpdated,
							actionCampaignIds,
							[]string{},
							[]string{},
							[]string{},
							"",
							generateCampaignRuleParameters(
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								maxCpa,
								"",
								"",
								"",
								"",
								"",
								"",
							),
						),
						generateCampaignRuleActions(
							"",
							action3TypeUpdated,
							actionCampaignIds,
							[]string{},
							[]string{},
							[]string{},
							"",
							generateCampaignRuleParameters(
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								"",
								queueId,
								"",
								"",
								"",
								"",
								"",
							),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "name", ruleName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "match_any_conditions", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "enabled", util.TrueValue),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.condition_type", condition1TypeUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.operator", condition1OperatorUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.0.parameters.0.value", condition1ValueUpdated),

					// Condition 2
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.1.condition_type", condition2TypeUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.1.parameters.0.operator", condition2OperatorUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.1.parameters.0.value", condition2ValueUpdated),

					// Condition 3
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.2.condition_type", condition3TypeUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.2.parameters.0.operator", condition3OperatorUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_conditions.2.parameters.0.value", condition3ValueUpdated),

					// Action 1
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.action_type", action1TypeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.parameters.0.relative_weight", relativeWeight),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.0.campaign_rule_action_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign2ResourceLabel, "id"),

					// Action 2
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.1.action_type", action2TypeUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.1.parameters.0.max_calls_per_agent", maxCpa),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.1.campaign_rule_action_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign2ResourceLabel, "id"),

					resource.TestCheckResourceAttr("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.2.action_type", action3TypeUpdated),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.2.parameters.0.queue_id",
						routingQueue.ResourceType+"."+queueLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_campaignrule."+resourceLabel, "campaign_rule_actions.2.campaign_rule_action_entities.0.campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaign2ResourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_campaignrule." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyCampaignRuleDestroyed,
	})
}

func generateOutboundCampaignRule(resourceLabel string, name string, enabled string, matchAnyConditions string, nestedBlocks ...string) string {
	return fmt.Sprintf(`
resource "genesyscloud_outbound_campaignrule" "%s" {
	name                 = "%s"
	enabled              = %s
	match_any_conditions = %s
	%s
}`, resourceLabel, name, enabled, matchAnyConditions, strings.Join(nestedBlocks, "\n"))
}

func generateCampaignRuleEntity(campaignIds []string, sequenceIds []string, smsCampaignIds []string, emailCampaignIds []string) string {
	return fmt.Sprintf(`
	campaign_rule_entities {
		campaign_ids = [%s]
		sequence_ids = [%s]
		sms_campaign_ids = [%s]
		email_campaign_ids = [%s]
	}
`, strings.Join(campaignIds, ", "), strings.Join(sequenceIds, ", "), strings.Join(smsCampaignIds, ", "),
		strings.Join(emailCampaignIds, ", "))
}

func generateCampaignRuleActions(id string,
	actionType string,
	campaignIds []string,
	sequenceIds []string,
	smsCampaignIds []string,
	emailCampaignIds []string,
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
			sms_campaign_ids          = [%s]
			email_campaign_ids          = [%s]
			%s
		}
		%s
	}
`, id, actionType, strings.Join(campaignIds, ", "), strings.Join(sequenceIds, ", "), strings.Join(smsCampaignIds, ", "),
		strings.Join(emailCampaignIds, ", "), useTriggeringEntity, paramsBlock)
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

func generateCampaignRuleParameters(operator string,
	value string,
	dialingMode string,
	priority string,
	abandonRate string,
	outboundLineCount string,
	relativeWeight string,
	maxCallsPerAgent string,
	queueId string,
	messagesPerMinute string,
	smsMessagesPerMinute string,
	emailMessagesPerMinute string,
	smsContentTemplate string,
	emailContentTemplate string,
) string {
	var maxCallsPerAgentStr, messagesPerMinuteStr, smsMessagesPerMinuteStr, emailMessagesPerMinuteStr string
	if operator != "" {
		operator = fmt.Sprintf(`operator = "%s"`, operator)
	}
	if dialingMode != "" {
		dialingMode = fmt.Sprintf(`dialing_mode = "%s"`, dialingMode)
	}
	if priority != "" {
		priority = fmt.Sprintf(`priority = "%s"`, priority)
	}
	if abandonRate != "" {
		abandonRate = fmt.Sprintf(`abandon_rate = "%s"`, abandonRate)
	}
	if outboundLineCount != "" {
		outboundLineCount = fmt.Sprintf(`outbound_line_count = "%s"`, outboundLineCount)
	}
	if relativeWeight != "" {
		relativeWeight = fmt.Sprintf(`relative_weight = "%s"`, relativeWeight)
	}
	if maxCallsPerAgent != "" {
		maxCallsPerAgentStr = fmt.Sprintf(`max_calls_per_agent = "%s"`, maxCallsPerAgent)
	}
	if queueId != "" {
		queueId = fmt.Sprintf(`queue_id = %s`, queueId)
	}
	if messagesPerMinute != "" {
		messagesPerMinuteStr = fmt.Sprintf(`messages_per_minute = "%s"`, messagesPerMinute)
	}
	if smsMessagesPerMinute != "" {
		smsMessagesPerMinuteStr = fmt.Sprintf(`sms_messages_per_minute = "%s"`, smsMessagesPerMinute)
	}
	if emailMessagesPerMinute != "" {
		emailMessagesPerMinuteStr = fmt.Sprintf(`email_messages_per_minute = "%s"`, emailMessagesPerMinute)
	}
	if smsContentTemplate != "" {
		smsContentTemplate = fmt.Sprintf(`sms_content_template = %s`, smsContentTemplate)
	}
	if emailContentTemplate != "" {
		emailContentTemplate = fmt.Sprintf(`email_content_template = %s`, emailContentTemplate)
	}

	return fmt.Sprintf(`
		parameters {
			%s
			value        = "%s"
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
		}
`, operator, value, dialingMode, priority, abandonRate, outboundLineCount, relativeWeight, maxCallsPerAgentStr, queueId,
		messagesPerMinuteStr, smsMessagesPerMinuteStr, emailMessagesPerMinuteStr, smsContentTemplate, emailContentTemplate)
}

func testVerifyCampaignRuleDestroyed(state *terraform.State) error {
	outboundApi := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_campaignrule" {
			continue
		}
		campaignRule, resp, err := outboundApi.GetOutboundCampaignrule(rs.Primary.ID)
		if campaignRule != nil {
			return fmt.Errorf("campaign rule (%s) still exists", rs.Primary.ID)
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
	campaignResourceLabel,
	campaignName,
	campaignStatus,
	contactListResourceLabel,
	contactListName,
	locationResourceLabel,
	locationName,
	locationEmergencyNumber,
	siteResourceLabel,
	siteName,
	wrapupCodeResourceLabel,
	wrapupCodeName,
	flowResourceLabel,
	flowFilePath,
	flowName,
	flowDivisionName,
	carResourceLabel,
	carName string) string {

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
	`, campaignResourceLabel,
		campaignName,
		campaignStatus,
		contactListResourceLabel, // genesyscloud_outbound_campaign.contact_list_id
		siteResourceLabel,        // genesyscloud_outbound_campaign.site_id
		carResourceLabel,         // genesyscloud_outbound_campaign.call_analysis_response_set_id
		contactListResourceLabel,
		contactListName,
		locationResourceLabel,
		locationName,
		locationEmergencyNumber,
		siteResourceLabel,
		siteName,
		locationResourceLabel, // genesyscloud_telephony_providers_edges_site.location_id
		wrapupCodeResourceLabel,
		wrapupCodeName,
		flowResourceLabel,
		flowFilePath,
		flowName,
		flowDivisionName,
		contactListResourceLabel, // genesyscloud_flow
		wrapupCodeResourceLabel,  // genesyscloud_flow
		carResourceLabel,
		carName,
		flowName,          // genesyscloud_outbound_callanalysisresponseset.responses.callable_person.name
		flowResourceLabel, // genesyscloud_outbound_callanalysisresponseset.responses.callable_person.data
	)
}
