package outbound_campaignrule

// @team: Outbound Rules
// @chat: #Outbound Rules
// @pm: Chad Mccormick
// @jira: OBR
// @description: Manages outbound campaign operations including automated voice dialing, SMS/email messaging campaigns, contact list management, and campaign rules for proactive customer outreach.

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesycloud_outbound_campaignrule_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the outbound_campaignrule resource.
3.  The datasource schema definitions for the outbound_campaignrule datasource.
4.  The resource exporter configuration for the outbound_campaignrule exporter.
*/
const ResourceType = "genesyscloud_outbound_campaignrule"

func getAllowedActions() []string {
	return []string{
		"turnOnCampaign",
		"turnOffCampaign",
		"turnOnSequence",
		"turnOffSequence",
		"setCampaignPriority",
		"recycleCampaign",
		"setCampaignDialingMode",
		"setCampaignAbandonRate",
		"setCampaignNumberOfLines",
		"setCampaignWeight",
		"setCampaignMaxCallsPerAgent",
		"changeCampaignQueue",
		"changeCampaignTemplate",
		"setCampaignMessagesPerMinute",
	}
}

func getAllowedConditions() []string {
	return []string{
		"campaignProgress",
		"campaignAgents",
		"campaignRecordsAttempted",
		"campaignContactsMessaged",
		"campaignBusinessSuccess",
		"campaignBusinessNeutral",
		"campaignBusinessFailure",
		"campaignValidAttempts",
		"campaignRightPartyContacts",
		"timeOfDay",
		"dayOfWeek",
		"dayOfMonth",
		"weekDayOfMonth",
		"specificDate",
		"campaignRunTime",
		"campaignWaitTime",
	}
}

var (
	outboundCampaignRuleEntities = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`campaign_ids`:       outboundCampaignRuleEntityCampaignRuleId,
			`sequence_ids`:       outboundCampaignRuleEntitySequenceRuleId,
			`sms_campaign_ids`:   outboundCampaignRuleEntitySmsCampaignRuleId,
			`email_campaign_ids`: outboundCampaignRuleEntityEmailCampaignRuleId,
		},
	}

	outboundCampaignRuleEntityCampaignRuleId = &schema.Schema{
		Description: `The list of campaigns for a CampaignRule to monitor. Required if the CampaignRule has any conditions that run on a campaign. Changing the outboundCampaignRuleEntityCampaignRuleId attribute will cause the outbound_campaignrule object to be dropped and recreated with a new ID.`,
		Optional:    true,
		ForceNew:    true,
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
	}

	outboundCampaignRuleEntitySequenceRuleId = &schema.Schema{
		Description: `The list of sequences for a CampaignRule to monitor. Required if the CampaignRule has any conditions that run on a sequence. Changing the outboundCampaignRuleEntitySequenceRuleId attribute will cause the outbound_campaignrule object to be dropped and recreated with a new ID.`,
		Optional:    true,
		ForceNew:    true,
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
	}

	outboundCampaignRuleEntitySmsCampaignRuleId = &schema.Schema{
		Description: `The list of SMS campaigns for a CampaignRule to monitor. Required if the CampaignRule has any conditions that run on an SMS campaign. Changing the outboundCampaignRuleEntityCampaignRuleId attribute will cause the outbound_campaignrule object to be dropped and recreated with a new ID.`,
		Optional:    true,
		ForceNew:    true,
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
	}

	outboundCampaignRuleEntityEmailCampaignRuleId = &schema.Schema{
		Description: `The list of Email campaigns for a CampaignRule to monitor. Required if the CampaignRule has any conditions that run on an Email campaign. Changing the outboundCampaignRuleEntityCampaignRuleId attribute will cause the outbound_campaignrule object to be dropped and recreated with a new ID.`,
		Optional:    true,
		ForceNew:    true,
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
	}

	outboundCampaignRuleActionEntities = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`campaign_ids`:       outboundCampaignRuleEntityCampaignRuleId,
			`sequence_ids`:       outboundCampaignRuleEntitySequenceRuleId,
			`sms_campaign_ids`:   outboundCampaignRuleEntitySmsCampaignRuleId,
			`email_campaign_ids`: outboundCampaignRuleEntityEmailCampaignRuleId,
			`use_triggering_entity`: {
				Description: `If true, the CampaignRuleAction will apply to the same entity that triggered the CampaignRuleCondition.`,
				Optional:    true,
				Type:        schema.TypeBool,
				Default:     false,
			},
		},
	}
)

var (
	outboundCampaignRuleWeekDayOfMonth = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`day_of_week`: {
				Description: `Day of week (1=Monday, 7=Sunday).`,
				Required:    true,
				Type:        schema.TypeInt,
			},
			`month`: {
				Description: `Month (1-12). Optional.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`occurrence`: {
				Description: `Occurrence (1-4, or -1 for last).`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
		},
	}

	outboundCampaignRuleDateTimeParameters = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`inverted`: {
				Description: `If true, inverts the result of evaluating this condition.`,
				Optional:    true,
				Default:     false,
				Type:        schema.TypeBool,
			},
			`time_of_day`: {
				Description: `Parameters for timeOfDay condition type.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`threshold_value`: {
							Description: `Time in HH:mm:ss.SSS format.`,
							Optional:    true,
							Type:        schema.TypeString,
						},
						`interval`: {
							Description: `Time interval for "between" operator.`,
							Optional:    true,
							MaxItems:    1,
							Type:        schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									`min`: {
										Description: `Minimum time in HH:mm:ss.SSS format.`,
										Required:    true,
										Type:        schema.TypeString,
									},
									`max`: {
										Description: `Maximum time in HH:mm:ss.SSS format.`,
										Required:    true,
										Type:        schema.TypeString,
									},
								},
							},
						},
					},
				},
			},
			`day_of_week`: {
				Description: `Parameters for dayOfWeek condition type.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`in_set`: {
							Description: `Days of week (1=Monday, 7=Sunday) for "equals" operator.`,
							Optional:    true,
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						`interval`: {
							Description: `Day interval for "between" operator.`,
							Optional:    true,
							MaxItems:    1,
							Type:        schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									`min`: {
										Description: `Minimum day (1-7).`,
										Required:    true,
										Type:        schema.TypeInt,
									},
									`max`: {
										Description: `Maximum day (1-7).`,
										Required:    true,
										Type:        schema.TypeInt,
									},
								},
							},
						},
					},
				},
			},
			`day_of_month`: {
				Description: `Parameters for dayOfMonth condition type.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`threshold_value`: {
							Description: `Day of month (1-31 or "LAST_DAY") for "before"/"after" operators.`,
							Optional:    true,
							Type:        schema.TypeString,
						},
						`in_set`: {
							Description: `Days of month (1-31, "LAST_DAY", "EVEN_DAY", "ODD_DAY") for "equals" operator.`,
							Optional:    true,
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						`interval`: {
							Description: `Day interval for "between" operator.`,
							Optional:    true,
							MaxItems:    1,
							Type:        schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									`min`: {
										Description: `Minimum day (1-31).`,
										Required:    true,
										Type:        schema.TypeString,
									},
									`max`: {
										Description: `Maximum day (1-31 or "LAST_DAY").`,
										Required:    true,
										Type:        schema.TypeString,
									},
								},
							},
						},
					},
				},
			},
			`specific_date`: {
				Description: `Parameters for specificDate condition type.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`include_year`: {
							Description: `If true, includes year in date comparison.`,
							Optional:    true,
							Default:     true,
							Type:        schema.TypeBool,
						},
						`threshold_value`: {
							Description: `Date in yyyy-MM-dd (with year) or MM-dd (without year) format.`,
							Optional:    true,
							Type:        schema.TypeString,
						},
						`interval`: {
							Description: `Date interval for "between" operator.`,
							Optional:    true,
							MaxItems:    1,
							Type:        schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									`min`: {
										Description: `Minimum date in yyyy-MM-dd or MM-dd format.`,
										Required:    true,
										Type:        schema.TypeString,
									},
									`max`: {
										Description: `Maximum date in yyyy-MM-dd or MM-dd format.`,
										Required:    true,
										Type:        schema.TypeString,
									},
								},
							},
						},
					},
				},
			},
			`week_day_of_month`: {
				Description: `Parameters for weekDayOfMonth condition type.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`threshold_value`: {
							Description: `The weekday-of-month value for "equals"/"before"/"after" operators.`,
							Optional:    true,
							MaxItems:    1,
							Type:        schema.TypeList,
							Elem:        outboundCampaignRuleWeekDayOfMonth,
						},
						`interval`: {
							Description: `Weekday-of-month interval for "between" operator.`,
							Optional:    true,
							MaxItems:    1,
							Type:        schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									`min`: {
										Description: `Minimum weekday-of-month.`,
										Required:    true,
										MaxItems:    1,
										Type:        schema.TypeList,
										Elem:        outboundCampaignRuleWeekDayOfMonth,
									},
									`max`: {
										Description: `Maximum weekday-of-month.`,
										Required:    true,
										MaxItems:    1,
										Type:        schema.TypeList,
										Elem:        outboundCampaignRuleWeekDayOfMonth,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	outboundCampaignRuleRunTimeSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`include_waiting_time`: {
				Description: `When true, counts all campaign running time. When false, only counts time when campaign is not waiting.`,
				Optional:    true,
				Default:     true,
				Type:        schema.TypeBool,
			},
		},
	}

	outboundCampaignRuleWaitTimeSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`wait_type`: {
				Description:  `Campaign wait type (Agents | Contacts | Lines).`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"Agents", "Contacts", "Lines"}, false),
			},
		},
	}
)

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceOutboundCampaignrule())
	regInstance.RegisterDataSource(ResourceType, DataSourceOutboundCampaignrule())
	regInstance.RegisterExporter(ResourceType, OutboundCampaignruleExporter())
}

// validateCampaignRuleConditions validates the relationship between campaign_rule_processing and condition fields
func validateCampaignRuleConditions(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	processing := d.Get("campaign_rule_processing").(string)
	hasConditionGroups := len(d.Get("condition_groups").([]interface{})) > 0
	hasRuleConditions := len(d.Get("campaign_rule_conditions").([]interface{})) > 0

	if processing == "v2" {
		if hasRuleConditions {
			return fmt.Errorf("when campaign_rule_processing is set to 'v2', use 'condition_groups' instead of 'campaign_rule_conditions'")
		}
		if !hasConditionGroups {
			return fmt.Errorf("when campaign_rule_processing is set to 'v2', 'condition_groups' is required")
		}
		// Validate time-based condition blocks match condition_type within condition_groups
		if err := validateConditionBlocksInGroups(d.Get("condition_groups").([]interface{})); err != nil {
			return err
		}
	} else {
		if hasConditionGroups {
			return fmt.Errorf("'condition_groups' can only be used when campaign_rule_processing is set to 'v2'")
		}
		if !hasRuleConditions {
			return fmt.Errorf("'campaign_rule_conditions' is required when not using campaign_rule_processing 'v2'")
		}
		// Validate that time-based condition types are not used in legacy mode
		if err := validateNoTimeBasedConditionsInLegacy(d.Get("campaign_rule_conditions").([]interface{})); err != nil {
			return err
		}
		// Also validate operator restrictions for legacy conditions (reject date/time operators on non-date types)
		if err := validateConditionBlocks(d.Get("campaign_rule_conditions").([]interface{})); err != nil {
			return err
		}
	}

	// Validate for_duration is not used in action parameters (handled by schema split - action uses campaignRuleActionParameters without for_duration)

	return nil
}

// validateConditionBlocksInGroups validates conditions within condition_groups
func validateConditionBlocksInGroups(groups []interface{}) error {
	for _, g := range groups {
		if g == nil {
			continue
		}
		groupMap := g.(map[string]interface{})
		conditions, ok := groupMap["conditions"].([]interface{})
		if !ok {
			continue
		}
		if err := validateConditionBlocks(conditions); err != nil {
			return err
		}
	}
	return nil
}

func validateNoTimeBasedConditionsInLegacy(conditions []interface{}) error {
	v2OnlyConditionTypes := map[string]bool{
		"timeOfDay":        true,
		"dayOfWeek":        true,
		"dayOfMonth":       true,
		"specificDate":     true,
		"weekDayOfMonth":   true,
		"campaignRunTime":  true,
		"campaignWaitTime": true,
	}

	for _, c := range conditions {
		if c == nil {
			continue
		}
		condMap := c.(map[string]interface{})
		condType, _ := condMap["condition_type"].(string)

		// Reject v2-only condition types
		if v2OnlyConditionTypes[condType] {
			return fmt.Errorf("condition_type %q requires campaign_rule_processing = \"v2\" with condition_groups; it cannot be used in legacy campaign_rule_conditions", condType)
		}

		// Reject v2-only nested blocks in legacy conditions
		if v, ok := condMap["date_time_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			return fmt.Errorf("date_time_parameters requires campaign_rule_processing = \"v2\" with condition_groups; it cannot be used in legacy campaign_rule_conditions")
		}
		if v, ok := condMap["campaign_run_time_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			return fmt.Errorf("campaign_run_time_settings requires campaign_rule_processing = \"v2\" with condition_groups; it cannot be used in legacy campaign_rule_conditions")
		}
		if v, ok := condMap["campaign_wait_time_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			return fmt.Errorf("campaign_wait_time_settings requires campaign_rule_processing = \"v2\" with condition_groups; it cannot be used in legacy campaign_rule_conditions")
		}

		// Reject for_duration in legacy condition parameters
		paramsRaw := condMap["parameters"]
		var paramsMap map[string]interface{}
		switch p := paramsRaw.(type) {
		case *schema.Set:
			if p != nil && p.Len() > 0 {
				paramsMap, _ = p.List()[0].(map[string]interface{})
			}
		case []interface{}:
			if len(p) > 0 && p[0] != nil {
				paramsMap, _ = p[0].(map[string]interface{})
			}
		}
		if paramsMap != nil {
			if v, ok := paramsMap["for_duration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				return fmt.Errorf("for_duration requires campaign_rule_processing = \"v2\" with condition_groups; it cannot be used in legacy campaign_rule_conditions")
			}
		}
	}
	return nil
}

// validateConditionBlocks validates that date_time_parameters, campaign_run_time_settings, and
// campaign_wait_time_settings are only used with their corresponding condition_type values,
// and that only the matching sub-block within date_time_parameters is populated.
func validateConditionBlocks(conditions []interface{}) error {
	dateTimeConditionTypes := map[string]string{
		"timeOfDay":      "time_of_day",
		"dayOfWeek":      "day_of_week",
		"dayOfMonth":     "day_of_month",
		"specificDate":   "specific_date",
		"weekDayOfMonth": "week_day_of_month",
	}

	for _, c := range conditions {
		if c == nil {
			continue
		}
		condMap := c.(map[string]interface{})
		condType, _ := condMap["condition_type"].(string)

		// Validate date_time_parameters only with date/time condition types
		if dtParams, ok := condMap["date_time_parameters"].([]interface{}); ok && len(dtParams) > 0 && dtParams[0] != nil {
			if _, isDateTimeType := dateTimeConditionTypes[condType]; !isDateTimeType {
				return fmt.Errorf("date_time_parameters can only be used with time-based condition types (timeOfDay, dayOfWeek, dayOfMonth, specificDate, weekDayOfMonth), got %q", condType)
			}
			// Validate only the matching sub-block is set
			dtMap := dtParams[0].(map[string]interface{})
			expectedSubBlock := dateTimeConditionTypes[condType]
			for subBlock := range dateTimeConditionTypes {
				tfName := dateTimeConditionTypes[subBlock]
				if tfName == expectedSubBlock {
					continue
				}
				if v, ok := dtMap[tfName].([]interface{}); ok && len(v) > 0 && v[0] != nil {
					return fmt.Errorf("for condition_type %q, only %q should be set in date_time_parameters, but %q is also set", condType, expectedSubBlock, tfName)
				}
			}
		}

		// Validate campaign_run_time_settings only with campaignRunTime
		if v, ok := condMap["campaign_run_time_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			if condType != "campaignRunTime" {
				return fmt.Errorf("campaign_run_time_settings can only be used with condition_type \"campaignRunTime\", got %q", condType)
			}
		}

		// Validate campaign_wait_time_settings only with campaignWaitTime
		if v, ok := condMap["campaign_wait_time_settings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			if condType != "campaignWaitTime" {
				return fmt.Errorf("campaign_wait_time_settings can only be used with condition_type \"campaignWaitTime\", got %q", condType)
			}
		}

		// Require matching blocks for v2-only condition types
		if _, isDateTimeType := dateTimeConditionTypes[condType]; isDateTimeType {
			dtParams, ok := condMap["date_time_parameters"].([]interface{})
			if !ok || len(dtParams) == 0 || dtParams[0] == nil {
				return fmt.Errorf("condition_type %q requires date_time_parameters with matching sub-block", condType)
			}
			// Verify matching sub-block exists
			dtMap := dtParams[0].(map[string]interface{})
			expectedSubBlock := dateTimeConditionTypes[condType]
			subBlock, ok := dtMap[expectedSubBlock].([]interface{})
			if !ok || len(subBlock) == 0 || subBlock[0] == nil {
				return fmt.Errorf("condition_type %q requires date_time_parameters.%s to be set", condType, expectedSubBlock)
			}
		}
		if condType == "campaignWaitTime" {
			v, ok := condMap["campaign_wait_time_settings"].([]interface{})
			if !ok || len(v) == 0 || v[0] == nil {
				return fmt.Errorf("condition_type \"campaignWaitTime\" requires campaign_wait_time_settings")
			}
		}

		// Validate date/time operators only allowed with date/time condition types
		if _, isDateTimeType := dateTimeConditionTypes[condType]; !isDateTimeType {
			paramsRaw := condMap["parameters"]
			var paramsMap map[string]interface{}
			switch p := paramsRaw.(type) {
			case *schema.Set:
				if p != nil && p.Len() > 0 {
					paramsMap, _ = p.List()[0].(map[string]interface{})
				}
			case []interface{}:
				if len(p) > 0 && p[0] != nil {
					paramsMap, _ = p[0].(map[string]interface{})
				}
			}
			if paramsMap != nil {
				operator, _ := paramsMap["operator"].(string)
				op := strings.ToLower(operator)
				if op == "before" || op == "after" || op == "between" {
					return fmt.Errorf("operator %q can only be used with date/time condition types, got condition_type %q", operator, condType)
				}
			}
		}

		// Validate non-date/time operators rejected for date/time condition types
		if _, isDateTimeType := dateTimeConditionTypes[condType]; isDateTimeType {
			paramsRaw := condMap["parameters"]
			var paramsMap map[string]interface{}
			switch p := paramsRaw.(type) {
			case *schema.Set:
				if p != nil && p.Len() > 0 {
					paramsMap, _ = p.List()[0].(map[string]interface{})
				}
			case []interface{}:
				if len(p) > 0 && p[0] != nil {
					paramsMap, _ = p[0].(map[string]interface{})
				}
			}
			if paramsMap != nil {
				operator, _ := paramsMap["operator"].(string)
				dateTimeOperators := map[string]bool{"before": true, "after": true, "between": true, "equals": true}
				if operator != "" && !dateTimeOperators[strings.ToLower(operator)] {
					return fmt.Errorf("operator %q cannot be used with date/time condition_type %q; valid operators are: equals, before, after, between", operator, condType)
				}
			}
		}

		// Validate operator↔field consistency for date/time conditions
		if _, isDateTimeType := dateTimeConditionTypes[condType]; isDateTimeType {
			if err := validateOperatorFields(condMap, condType, dateTimeConditionTypes); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateOperatorFields checks that the operator matches the fields provided in date_time_parameters
func validateOperatorFields(condMap map[string]interface{}, condType string, dateTimeConditionTypes map[string]string) error {
	// Get operator from parameters
	paramsRaw := condMap["parameters"]
	var paramsMap map[string]interface{}
	switch p := paramsRaw.(type) {
	case *schema.Set:
		if p != nil && p.Len() > 0 {
			paramsMap, _ = p.List()[0].(map[string]interface{})
		}
	case []interface{}:
		if len(p) > 0 && p[0] != nil {
			paramsMap, _ = p[0].(map[string]interface{})
		}
	}
	if paramsMap == nil {
		return nil
	}
	operator, _ := paramsMap["operator"].(string)
	if operator == "" {
		return nil
	}

	// Get the date_time_parameters sub-block
	dtParams, ok := condMap["date_time_parameters"].([]interface{})
	if !ok || len(dtParams) == 0 || dtParams[0] == nil {
		return nil
	}
	dtMap := dtParams[0].(map[string]interface{})
	subBlockName := dateTimeConditionTypes[condType]
	subBlock, ok := dtMap[subBlockName].([]interface{})
	if !ok || len(subBlock) == 0 || subBlock[0] == nil {
		return nil
	}
	subBlockMap := subBlock[0].(map[string]interface{})

	hasInterval := false
	if v, ok := subBlockMap["interval"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		hasInterval = true
	}
	hasThreshold := false
	if v, ok := subBlockMap["threshold_value"]; ok {
		switch tv := v.(type) {
		case string:
			hasThreshold = tv != ""
		case []interface{}:
			hasThreshold = len(tv) > 0 && tv[0] != nil
		}
	}
	hasInSet := false
	if v, ok := subBlockMap["in_set"].([]interface{}); ok && len(v) > 0 {
		hasInSet = true
	}

	switch strings.ToLower(operator) {
	case "between":
		if !hasInterval {
			return fmt.Errorf("operator \"between\" for condition_type %q requires an interval block in date_time_parameters.%s", condType, subBlockName)
		}
		if hasThreshold {
			return fmt.Errorf("operator \"between\" for condition_type %q should use interval, not threshold_value", condType)
		}
		if hasInSet {
			return fmt.Errorf("operator \"between\" for condition_type %q should use interval, not in_set", condType)
		}
	case "before", "after":
		if !hasThreshold {
			return fmt.Errorf("operator %q for condition_type %q requires threshold_value in date_time_parameters.%s", operator, condType, subBlockName)
		}
		if hasInterval {
			return fmt.Errorf("operator %q for condition_type %q should use threshold_value, not interval", operator, condType)
		}
		if hasInSet {
			return fmt.Errorf("operator %q for condition_type %q should use threshold_value, not in_set", operator, condType)
		}
	case "equals":
		if hasInterval {
			return fmt.Errorf("operator \"equals\" for condition_type %q should use threshold_value or in_set, not interval", condType)
		}
		if !hasThreshold && !hasInSet {
			return fmt.Errorf("operator \"equals\" for condition_type %q requires threshold_value or in_set in date_time_parameters.%s", condType, subBlockName)
		}
		if hasThreshold && hasInSet {
			return fmt.Errorf("operator \"equals\" for condition_type %q should use either threshold_value or in_set, not both", condType)
		}
	}

	return nil
}

// paramSchemaOptions controls which fields are included in the campaign rule parameter schema.
type paramSchemaOptions struct {
	includeForDuration       bool
	includeDateTimeOperators bool
}

// campaignRuleParameterSchema builds the shared parameter schema for campaign rule conditions and actions.
func campaignRuleParameterSchema(opts paramSchemaOptions) map[string]*schema.Schema {
	operators := []string{"equals", "greaterThan", "greaterThanEqualTo", "lessThan", "lessThanEqualTo"}
	if opts.includeDateTimeOperators {
		operators = append(operators, "before", "after", "between")
	}

	params := map[string]*schema.Schema{
		`operator`: {
			Description:  `The operator for comparison. Required for a CampaignRuleCondition.`,
			Optional:     true,
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(operators, true),
		},
		`value`: {
			Description: `The value for comparison. Required for a CampaignRuleCondition.`,
			Optional:    true,
			Type:        schema.TypeString,
		},
		`priority`: {
			Description:  `The priority to set a campaign to (1 | 2 | 3 | 4 | 5). Required for the 'setCampaignPriority' action.`,
			Optional:     true,
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"1", "2", "3", "4", "5"}, true),
		},
		`dialing_mode`: {
			Description:  `The dialing mode to set a campaign to. Required for the 'setCampaignDialingMode' action (agentless | preview | power | predictive | progressive | external).`,
			Optional:     true,
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"agentless", "preview", "power", "predictive", "progressive", "external"}, true),
		},
		`abandon_rate`: {
			Description: `Compliance Abandon Rate. Required for 'setCampaignAbandonRate' action`,
			Optional:    true,
			Type:        schema.TypeString,
			ValidateFunc: func(v interface{}, key string) (warns []string, errs []error) {
				f, err := strconv.ParseFloat(v.(string), 64)
				if err != nil || f <= 0.1 {
					errs = append(errs, fmt.Errorf("%q must be a float > 0.1", key))
				}
				return nil, nil
			},
		},
		`outbound_line_count`: {
			Description: `Number of Outbound lines. Required for 'setCampaignNumberOfLines' action`,
			Optional:    true,
			Type:        schema.TypeString,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				v := val.(string)
				if v != "" {
					if num, err := strconv.Atoi(v); err != nil || num < 0 {
						errs = append(errs, fmt.Errorf("%q must be a non-negative integer", key))
					}
				}
				return
			},
		},
		`relative_weight`: {
			Description: `Relative weight. Required for 'setCampaignWeight' action`,
			Optional:    true,
			Type:        schema.TypeString,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				if v := val.(string); v != "" {
					if num, err := strconv.Atoi(v); err != nil || num < 0 || num > 100 {
						errs = append(errs, fmt.Errorf("%q must be an integer between 0 and 100 inclusive", key))
					}
				}
				return
			},
		},
		`max_calls_per_agent`: {
			Description: `Max calls per agent. Optional parameter for 'setCampaignMaxCallsPerAgent' action`,
			Optional:    true,
			Type:        schema.TypeString,
			ValidateFunc: func(v interface{}, key string) (warns []string, errs []error) {
				f, err := strconv.ParseFloat(v.(string), 64)
				if err != nil || f <= 1.0 {
					errs = append(errs, fmt.Errorf("%q must be a float > 1.0", key))
				}
				return nil, nil
			},
		},
		`queue_id`: {
			Description: `The ID of the Queue. Required for 'changeCampaignQueue' action`,
			Optional:    true,
			Type:        schema.TypeString,
		},
		`messages_per_minute`: {
			Description: `The number of messages per minute to set a messaging campaign to.`,
			Optional:    true,
			Type:        schema.TypeString,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				if v := val.(string); v != "" {
					if num, err := strconv.Atoi(v); err != nil || num < 1 {
						errs = append(errs, fmt.Errorf("%q must be a positive integer", key))
					}
				}
				return
			},
		},
		`sms_messages_per_minute`: {
			Description: `The number of messages per minute to set a SMS messaging campaign to.`,
			Optional:    true,
			Type:        schema.TypeString,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				if v := val.(string); v != "" {
					if num, err := strconv.Atoi(v); err != nil || num < 1 {
						errs = append(errs, fmt.Errorf("%q must be a positive integer", key))
					}
				}
				return
			},
		},
		`email_messages_per_minute`: {
			Description: `The number of messages per minute to set an Email messaging campaign to.`,
			Optional:    true,
			Type:        schema.TypeString,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				if v := val.(string); v != "" {
					if num, err := strconv.Atoi(v); err != nil || num < 1 {
						errs = append(errs, fmt.Errorf("%q must be a positive integer", key))
					}
				}
				return
			},
		},
		`sms_content_template_id`: {
			Description: `The content template to set a SMS campaign to.`,
			Optional:    true,
			Type:        schema.TypeString,
		},
		`email_content_template_id`: {
			Description: `The content template to set an Email campaign to.`,
			Optional:    true,
			Type:        schema.TypeString,
		},
	}

	if opts.includeForDuration {
		params[`for_duration`] = &schema.Schema{
			Description: `Duration (in seconds) for which the condition must be continuously true before it is evaluated as true. Only valid in condition parameters with campaign_rule_processing = "v2".`,
			Optional:    true,
			MaxItems:    1,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					`seconds`: {
						Description:  `Duration in seconds.`,
						Required:     true,
						Type:         schema.TypeInt,
						ValidateFunc: validation.IntAtLeast(1),
					},
				},
			},
		}
	}

	return params
}

// ResourceOutboundCampaignrule registers the genesyscloud_outbound_campaignrule resource with Terraform
func ResourceOutboundCampaignrule() *schema.Resource {
	campaignRuleParameters := &schema.Resource{
		Schema: campaignRuleParameterSchema(paramSchemaOptions{includeForDuration: true, includeDateTimeOperators: true}),
	}

	// Action parameters: no for_duration, no date/time operators
	campaignRuleActionParameters := &schema.Resource{
		Schema: campaignRuleParameterSchema(paramSchemaOptions{includeForDuration: false, includeDateTimeOperators: false}),
	}

	outboundCampaignRuleCondition := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`id`: {
				Description: `The ID of the CampaignRuleCondition.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`parameters`: {
				Description: `The parameters for the CampaignRuleCondition.`,
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        campaignRuleParameters,
			},
			`condition_type`: {
				Description:  `The type of condition to evaluate (` + strings.Join(getAllowedConditions(), ` | `) + `)`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(getAllowedConditions(), true),
			},
			`date_time_parameters`: {
				Description: `Parameters for date/time conditions (timeOfDay, dayOfWeek, dayOfMonth, specificDate, weekDayOfMonth). Only valid with campaign_rule_processing = "v2" and condition_groups.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem:        outboundCampaignRuleDateTimeParameters,
			},
			`campaign_run_time_settings`: {
				Description: `Settings for campaignRunTime conditions. Only valid with campaign_rule_processing = "v2" and condition_groups.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem:        outboundCampaignRuleRunTimeSettings,
			},
			`campaign_wait_time_settings`: {
				Description: `Settings for campaignWaitTime conditions. Only valid with campaign_rule_processing = "v2" and condition_groups.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem:        outboundCampaignRuleWaitTimeSettings,
			},
		},
	}

	outboundCampaignRuleConditionGroup := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`match_any_conditions`: {
				Description: `Whether or not this condition group should be evaluated as true if any of sub conditions is matched.`,
				Required:    true,
				Type:        schema.TypeBool,
			},
			`conditions`: {
				Description: `The list of conditions in this group.`,
				Required:    true,
				MinItems:    1,
				Type:        schema.TypeList,
				Elem:        outboundCampaignRuleCondition,
			},
		},
	}

	outboundCampaignRuleExecutionSettings := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`frequency`: {
				Description:  `Execution control frequency. Valid values: onEachTrigger, oncePerDay.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"onEachTrigger", "oncePerDay"}, true),
			},
			`time_zone_id`: {
				Description: `The time zone for the execution control frequency="oncePerDay"; for example, Africa/Abidjan. This property is ignored when frequency is not "oncePerDay".`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	outboundCampaignRuleAction := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`id`: {
				Description: `The ID of the CampaignRuleAction.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`parameters`: {
				Description: `The parameters for the CampaignRuleAction. Required for certain actionTypes.`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        campaignRuleActionParameters,
			},
			`action_type`: {
				Description:  `The action to take on the campaignRuleActionEntities (` + strings.Join(getAllowedActions(), ` | `) + `)`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(getAllowedActions(), true),
			},
			`campaign_rule_action_entities`: {
				Description: `The list of entities that this action will apply to.`,
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        outboundCampaignRuleActionEntities,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud outbound campaign rule`,

		CreateContext: provider.CreateWithPooledClient(createOutboundCampaignRule),
		ReadContext:   provider.ReadWithPooledClient(readOutboundCampaignRule),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundCampaignRule),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundCampaignRule),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		CustomizeDiff: validateCampaignRuleConditions,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the campaign rule.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`campaign_rule_entities`: {
				Description: `The list of entities that this campaign rule monitors.`,
				Required:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        outboundCampaignRuleEntities,
			},
			`campaign_rule_conditions`: {
				Description:   `The list of conditions that are evaluated on the entities. Required when not using condition_groups (campaign_rule_processing "v2").`,
				Optional:      true,
				Type:          schema.TypeList,
				Elem:          outboundCampaignRuleCondition,
				ConflictsWith: []string{`condition_groups`},
			},
			`campaign_rule_actions`: {
				Description: `The list of actions that are executed if the conditions are satisfied.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        outboundCampaignRuleAction,
			},
			`match_any_conditions`: {
				Description: `Whether actions are executed if any condition is met, or only when all conditions are met.`,
				Optional:    true,
				Default:     false,
				Type:        schema.TypeBool,
			},
			`campaign_rule_processing`: {
				Description:  `Campaign rule processing algorithm. Use "v2" to enable condition groups.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"v2"}, true),
			},
			`condition_groups`: {
				Description:   `List of condition groups that are evaluated, used only with campaignRuleProcessing="v2".`,
				Optional:      true,
				Type:          schema.TypeList,
				Elem:          outboundCampaignRuleConditionGroup,
				ConflictsWith: []string{`campaign_rule_conditions`},
			},
			`execution_settings`: {
				Description: `Campaign rule execution settings.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Elem:        outboundCampaignRuleExecutionSettings,
			},
			`enabled`: {
				Description: `Whether or not this campaign rule is currently enabled.`,
				Optional:    true,
				Default:     false,
				Type:        schema.TypeBool,
			},
			`time_zone_id`: {
				Description: `Optional. Used for date/time conditions. If omitted, Genesys Cloud defaults to UTC.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

// OutboundCampaignruleExporter returns the resourceExporter object used to hold the genesyscloud_outbound_campaignrule exporter's config
func OutboundCampaignruleExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthCampaignRules),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`campaign_rule_actions.campaign_rule_action_entities.campaign_ids`: {
				RefType: "genesyscloud_outbound_campaign",
			},
			`campaign_rule_actions.campaign_rule_action_entities.sequence_ids`: {
				RefType: "genesyscloud_outbound_sequence",
			},
			`campaign_rule_actions.campaign_rule_action_entities.sms_campaign_ids`: {
				RefType: "genesyscloud_outbound_messagingcampaign",
			},
			`campaign_rule_actions.campaign_rule_action_entities.email_campaign_ids`: {
				RefType: "genesyscloud_outbound_messagingcampaign",
			},
			`campaign_rule_entities.campaign_ids`: {
				RefType: "genesyscloud_outbound_campaign",
			},
			`campaign_rule_entities.sequence_ids`: {
				RefType: "genesyscloud_outbound_sequence",
			},
			`campaign_rule_entities.sms_campaign_ids`: {
				RefType: "genesyscloud_outbound_messagingcampaign",
			},
			`campaign_rule_entities.email_campaign_ids`: {
				RefType: "genesyscloud_outbound_messagingcampaign",
			},
			`campaign_rule_actions.parameters.queue_id`: {
				RefType: "genesyscloud_routing_queue",
			},
			`campaign_rule_conditions.parameters.queue_id`: {
				RefType: "genesyscloud_routing_queue",
			},
			`campaign_rule_actions.parameters.sms_content_template_id`: {
				RefType: "genesyscloud_responsemanagement_response",
			},
			`campaign_rule_conditions.parameters.sms_content_template_id`: {
				RefType: "genesyscloud_responsemanagement_response",
			},
			`campaign_rule_actions.parameters.email_content_template_id`: {
				RefType: "genesyscloud_responsemanagement_response",
			},
			`campaign_rule_conditions.parameters.email_content_template_id`: {
				RefType: "genesyscloud_responsemanagement_response",
			},
			`condition_groups.conditions.parameters.queue_id`: {
				RefType: "genesyscloud_routing_queue",
			},
			`condition_groups.conditions.parameters.sms_content_template_id`: {
				RefType: "genesyscloud_responsemanagement_response",
			},
			`condition_groups.conditions.parameters.email_content_template_id`: {
				RefType: "genesyscloud_responsemanagement_response",
			},
		},
	}
}

// DataSourceOutboundCampaignrule registers the genesyscloud_outbound_campaignrule data source
func DataSourceOutboundCampaignrule() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound campaign rule data source. Select a campaign rule by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundCampaignruleRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Campaign Rule name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
