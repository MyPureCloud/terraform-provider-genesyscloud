package outbound

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
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

	outboundCampaignRuleEntities = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`campaign_ids`: outboundCampaignRuleEntityCampaignRuleId,
			`sequence_ids`: outboundCampaignRuleEntitySequenceRuleId,
		},
	}

	outboundCampaignRuleActionEntities = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`campaign_ids`: outboundCampaignRuleEntityCampaignRuleId,
			`sequence_ids`: outboundCampaignRuleEntitySequenceRuleId,
			`use_triggering_entity`: {
				Description: `If true, the CampaignRuleAction will apply to the same entity that triggered the CampaignRuleCondition.`,
				Optional:    true,
				Type:        schema.TypeBool,
				Default:     false,
			},
		},
	}

	campaignRuleParameters = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`operator`: {
				Description:  `The operator for comparison. Required for a CampaignRuleCondition.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"equals", "greaterThan", "greaterThanEqualTo", "lessThan", "lessThanEqualTo"}, true),
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
		},
	}

	outboundCampaignRuleCondition = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`id`: {
				Description: `The ID of the CampaignRuleCondition.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`parameters`: {
				Description: `The parameters for the CampaignRuleCondition.`,
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        campaignRuleParameters,
			},
			`condition_type`: {
				Description:  `The type of condition to evaluate (campaignProgress | campaignAgents).`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"campaignProgress", "campaignAgents"}, true),
			},
		},
	}

	outboundCampaignRuleAction = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`id`: {
				Description: `The ID of the CampaignRuleAction.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`parameters`: {
				Description: `The parameters for the CampaignRuleAction. Required for certain actionTypes.`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        campaignRuleParameters,
			},
			`action_type`: {
				Description: `The action to take on the campaignRuleActionEntities
(turnOnCampaign | turnOffCampaign | turnOnSequence | turnOffSequence | setCampaignPriority | recycleCampaign | setCampaignDialingMode).`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"turnOnCampaign", "turnOffCampaign", "turnOnSequence", "turnOffSequence", "setCampaignPriority", "recycleCampaign", "setCampaignDialingMode"}, true),
			},
			`campaign_rule_action_entities`: {
				Description: `The list of entities that this action will apply to.`,
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        outboundCampaignRuleActionEntities,
			},
		},
	}
)

func getAllCampaignRules(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		campaignRuleConfigs, _, getErr := outboundAPI.GetOutboundCampaignrules(pageSize, pageNum, true, "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of campaign rule configs: %v", getErr)
		}

		if campaignRuleConfigs.Entities == nil || len(*campaignRuleConfigs.Entities) == 0 {
			break
		}

		for _, campaignRuleConfig := range *campaignRuleConfigs.Entities {
			resources[*campaignRuleConfig.Id] = &resourceExporter.ResourceMeta{Name: *campaignRuleConfig.Name}
		}
	}

	return resources, nil
}

func OutboundCampaignRuleExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllCampaignRules),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`campaign_rule_actions.campaign_rule_action_entities.campaign_ids`: {
				RefType: "genesyscloud_outbound_campaign",
			},
			`campaign_rule_actions.campaign_rule_action_entities.sequence_ids`: {
				RefType: "genesyscloud_outbound_sequence",
			},
			`campaign_rule_entities.campaign_ids`: {
				RefType: "genesyscloud_outbound_campaign",
			},
			`campaign_rule_entities.sequence_ids`: {
				RefType: "genesyscloud_outbound_sequence",
			},
		},
	}
}

func ResourceOutboundCampaignRule() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound campaign rule`,

		CreateContext: gcloud.CreateWithPooledClient(createOutboundCampaignRule),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundCampaignRule),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundCampaignRule),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundCampaignRule),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
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
				Description: `The list of conditions that are evaluated on the entities.`,
				Required:    true,
				MinItems:    1,
				Type:        schema.TypeList,
				Elem:        outboundCampaignRuleCondition,
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
			`enabled`: {
				Description: `Whether or not this campaign rule is currently enabled.`,
				Optional:    true,
				Default:     false,
				Type:        schema.TypeBool,
			},
		},
	}
}

func createOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	matchAnyConditions := d.Get("match_any_conditions").(bool)
	enabled := d.Get("enabled").(bool)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	campaignRuleEntities := d.Get("campaign_rule_entities").(*schema.Set)
	campaignRuleConditions := d.Get("campaign_rule_conditions").([]interface{})
	campaignRuleActions := d.Get("campaign_rule_actions").([]interface{})

	sdkCampaignRule := platformclientv2.Campaignrule{
		CampaignRuleEntities:   buildOutboundCampaignRuleEntities(campaignRuleEntities),
		CampaignRuleConditions: buildOutboundCampaignRuleConditionSlice(campaignRuleConditions),
		CampaignRuleActions:    buildOutboundCampaignRuleActionSlice(campaignRuleActions),
		MatchAnyConditions:     &matchAnyConditions,
	}

	if name != "" {
		sdkCampaignRule.Name = &name
	}

	// All campaign rules have to be created in an "off" state to start out with
	defaultStatus := false
	sdkCampaignRule.Enabled = &defaultStatus

	log.Printf("Creating Outbound Campaign Rule %s", name)
	outboundCampaignRule, _, err := outboundApi.PostOutboundCampaignrules(sdkCampaignRule)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Campaign Rule %s: %s", name, err)
	}

	d.SetId(*outboundCampaignRule.Id)
	log.Printf("Created Outbound Campaign Rule %s %s", name, *outboundCampaignRule.Id)

	// Campaign rules can be enabled after creation
	if enabled {
		d.Set("enabled", enabled)
		diag := updateOutboundCampaignRule(ctx, d, meta)
		if diag != nil {
			return diag
		}
	}

	return readOutboundCampaignRule(ctx, d, meta)
}

func updateOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	matchAnyConditions := d.Get("match_any_conditions").(bool)

	// Required on updates
	enabled := d.Get("enabled").(bool)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkCampaignRule := platformclientv2.Campaignrule{
		CampaignRuleEntities:   buildOutboundCampaignRuleEntities(d.Get("campaign_rule_entities").(*schema.Set)),
		CampaignRuleConditions: buildOutboundCampaignRuleConditionSlice(d.Get("campaign_rule_conditions").([]interface{})),
		CampaignRuleActions:    buildOutboundCampaignRuleActionSlice(d.Get("campaign_rule_actions").([]interface{})),
		MatchAnyConditions:     &matchAnyConditions,
		Enabled:                &enabled,
	}

	if name != "" {
		sdkCampaignRule.Name = &name
	}

	log.Printf("Updating Outbound Campaign Rule %s", name)
	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Campaign Rule version
		outboundCampaignRule, resp, getErr := outboundApi.GetOutboundCampaignrule(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Campaign Rule %s: %s", d.Id(), getErr)
		}
		sdkCampaignRule.Version = outboundCampaignRule.Version
		outboundCampaignRule, _, updateErr := outboundApi.PutOutboundCampaignrule(d.Id(), sdkCampaignRule)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound Campaign Rule %s: %s", name, updateErr)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Campaign Rule %s", name)
	return readOutboundCampaignRule(ctx, d, meta)
}

func readOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Campaign Rule %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkCampaignRule, resp, getErr := outboundApi.GetOutboundCampaignrule(d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read Outbound Campaign Rule %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read Outbound Campaign Rule %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundCampaignRule())

		if sdkCampaignRule.Name != nil {
			d.Set("name", *sdkCampaignRule.Name)
		}
		if sdkCampaignRule.CampaignRuleEntities != nil {
			d.Set("campaign_rule_entities", flattenOutboundCampaignRuleEntities(sdkCampaignRule.CampaignRuleEntities))
		}
		if sdkCampaignRule.CampaignRuleConditions != nil {
			d.Set("campaign_rule_conditions", flattenOutboundCampaignRuleConditionSlice(*sdkCampaignRule.CampaignRuleConditions))
		}
		if sdkCampaignRule.CampaignRuleActions != nil {
			d.Set("campaign_rule_actions", flattenOutboundCampaignRuleActionSlice(sdkCampaignRule.CampaignRuleActions))
		}
		if sdkCampaignRule.MatchAnyConditions != nil {
			d.Set("match_any_conditions", *sdkCampaignRule.MatchAnyConditions)
		}
		if sdkCampaignRule.Enabled != nil {
			d.Set("enabled", *sdkCampaignRule.Enabled)
		}

		log.Printf("Read Outbound Campaign Rule %s %s", d.Id(), *sdkCampaignRule.Name)
		return cc.CheckState()
	})
}

func deleteOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	ruleEnabled := d.Get("enabled").(bool)
	if ruleEnabled {
		// Have to disable rule before we can delete
		log.Printf("Disabling Outbound Campaign Rule")
		d.Set("enabled", false)
		diagErr := updateOutboundCampaignRule(ctx, d, meta)
		if diagErr != nil {
			return diagErr
		}
	}

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Campaign Rule")
		resp, err := outboundApi.DeleteOutboundCampaignrule(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Campaign Rule: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := outboundApi.GetOutboundCampaignrule(d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Outbound Campaign Rule deleted
				log.Printf("Deleted Outbound Campaign Rule %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting Outbound Campaign Rule %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Campaign Rule %s still exists", d.Id()))
	})
}

func buildOutboundCampaignRuleEntities(entities *schema.Set) *platformclientv2.Campaignruleentities {
	if entities == nil {
		return nil
	}
	var campaignRuleEntities platformclientv2.Campaignruleentities

	campaignRuleEntitiesList := entities.List()

	if len(campaignRuleEntitiesList) <= 0 {
		return &campaignRuleEntities
	}

	campaignRuleEntitiesMap := campaignRuleEntitiesList[0].(map[string]interface{})
	if campaigns := campaignRuleEntitiesMap["campaign_ids"].([]interface{}); campaigns != nil {
		campaignRuleEntities.Campaigns = gcloud.BuildSdkDomainEntityRefArrFromArr(campaigns)
	}
	if sequences := campaignRuleEntitiesMap["sequence_ids"].([]interface{}); sequences != nil {
		campaignRuleEntities.Sequences = gcloud.BuildSdkDomainEntityRefArrFromArr(sequences)
	}

	return &campaignRuleEntities
}

func buildOutboundCampaignRuleConditionSlice(campaignRuleConditions []interface{}) *[]platformclientv2.Campaignrulecondition {
	var campaignRuleConditionRefs []platformclientv2.Campaignrulecondition

	for _, c := range campaignRuleConditions {
		sdkCondition := platformclientv2.Campaignrulecondition{}
		currentCondition := c.(map[string]interface{})
		id := currentCondition["id"].(string)
		conditionType := currentCondition["condition_type"].(string)

		sdkCondition.Parameters = buildOutboundCampaignRuleParameters(currentCondition["parameters"].(*schema.Set))

		if id != "" {
			sdkCondition.Id = &id
		}

		if conditionType != "" {
			sdkCondition.ConditionType = &conditionType
		}

		campaignRuleConditionRefs = append(campaignRuleConditionRefs, sdkCondition)
	}

	return &campaignRuleConditionRefs
}

func buildOutboundCampaignRuleActionSlice(campaignRuleActions []interface{}) *[]platformclientv2.Campaignruleaction {
	var sdkCampaignRuleActions []platformclientv2.Campaignruleaction

	for _, c := range campaignRuleActions {
		var sdkCampaignRuleAction platformclientv2.Campaignruleaction
		currentAction := c.(map[string]interface{})

		id := currentAction["id"].(string)
		actionType := currentAction["action_type"].(string)

		sdkCampaignRuleAction.Id = &id
		sdkCampaignRuleAction.ActionType = &actionType

		sdkCampaignRuleAction.Parameters = buildOutboundCampaignRuleParameters(currentAction["parameters"].(*schema.Set))

		sdkCampaignRuleAction.CampaignRuleActionEntities = buildOutboundCampaignRuleActionEntities(currentAction["campaign_rule_action_entities"].(*schema.Set))

		sdkCampaignRuleActions = append(sdkCampaignRuleActions, sdkCampaignRuleAction)
	}

	return &sdkCampaignRuleActions
}

func buildOutboundCampaignRuleParameters(set *schema.Set) *platformclientv2.Campaignruleparameters {
	var sdkCampaignRuleParameters platformclientv2.Campaignruleparameters

	paramsList := set.List()

	if len(paramsList) <= 0 {
		return &sdkCampaignRuleParameters
	}

	paramsMap := paramsList[0].(map[string]interface{})

	operator := paramsMap["operator"].(string)
	paramValue := paramsMap["value"].(string)
	priority := paramsMap["priority"].(string)
	dialingMode := paramsMap["dialing_mode"].(string)

	if paramValue != "" {
		sdkCampaignRuleParameters.Value = &paramValue
	}

	if priority != "" {
		sdkCampaignRuleParameters.Priority = &priority
	}

	if dialingMode != "" {
		sdkCampaignRuleParameters.DialingMode = &dialingMode
	}

	if operator != "" {
		sdkCampaignRuleParameters.Operator = &operator
	}

	return &sdkCampaignRuleParameters
}

func buildOutboundCampaignRuleActionEntities(set *schema.Set) *platformclientv2.Campaignruleactionentities {
	var (
		sdkCampaignRuleActionEntities platformclientv2.Campaignruleactionentities
		entities                      = set.List()
	)

	if len(entities) <= 0 {
		return &sdkCampaignRuleActionEntities
	}

	entitiesMap := entities[0].(map[string]interface{})
	useTriggeringEntity := entitiesMap["use_triggering_entity"].(bool)

	sdkCampaignRuleActionEntities.UseTriggeringEntity = &useTriggeringEntity

	if campaignIds := entitiesMap["campaign_ids"].([]interface{}); campaignIds != nil {
		sdkCampaignRuleActionEntities.Campaigns = gcloud.BuildSdkDomainEntityRefArrFromArr(campaignIds)
	}

	if sequenceIds := entitiesMap["sequence_ids"].([]interface{}); sequenceIds != nil {
		sdkCampaignRuleActionEntities.Sequences = gcloud.BuildSdkDomainEntityRefArrFromArr(sequenceIds)
	}

	return &sdkCampaignRuleActionEntities
}

func flattenOutboundCampaignRuleEntities(campaignRuleEntities *platformclientv2.Campaignruleentities) *schema.Set {
	var (
		campaignRuleEntitiesSet = schema.NewSet(schema.HashResource(outboundCampaignRuleEntities), []interface{}{})
		campaignRuleEntitiesMap = make(map[string]interface{})

		// had to change from []string to []interface{}
		campaigns []interface{}
		sequences []interface{}
	)

	if campaignRuleEntities == nil {
		return nil
	}

	if campaignRuleEntities.Campaigns != nil {
		for _, v := range *campaignRuleEntities.Campaigns {
			campaigns = append(campaigns, *v.Id)
		}
	}

	if campaignRuleEntities.Sequences != nil {
		for _, v := range *campaignRuleEntities.Sequences {
			sequences = append(sequences, *v.Id)
		}
	}

	campaignRuleEntitiesMap["campaign_ids"] = campaigns
	campaignRuleEntitiesMap["sequence_ids"] = sequences

	campaignRuleEntitiesSet.Add(campaignRuleEntitiesMap)
	return campaignRuleEntitiesSet
}

func flattenOutboundCampaignRuleConditionSlice(campaignRuleConditions []platformclientv2.Campaignrulecondition) []interface{} {
	if campaignRuleConditions == nil {
		return nil
	}

	var ruleConditionList []interface{}

	for _, currentSdkCondition := range campaignRuleConditions {
		campaignRuleConditionsMap := make(map[string]interface{})

		if currentSdkCondition.ConditionType != nil {
			campaignRuleConditionsMap["condition_type"] = *currentSdkCondition.ConditionType
		}

		if currentSdkCondition.Parameters != nil {
			campaignRuleConditionsMap["parameters"] = flattenRuleParameters(*currentSdkCondition.Parameters)
		}

		ruleConditionList = append(ruleConditionList, campaignRuleConditionsMap)
	}
	return ruleConditionList
}

func flattenOutboundCampaignRuleActionSlice(campaignRuleActions *[]platformclientv2.Campaignruleaction) []interface{} {
	if campaignRuleActions == nil {
		return nil
	}

	var ruleActionsList []interface{}

	for _, currentAction := range *campaignRuleActions {
		actionMap := make(map[string]interface{})

		if currentAction.Id != nil {
			actionMap["id"] = *currentAction.Id
		}

		if currentAction.ActionType != nil {
			actionMap["action_type"] = *currentAction.ActionType
		}

		if currentAction.Parameters != nil {
			actionMap["parameters"] = flattenRuleParameters(*currentAction.Parameters)
		}

		if currentAction.CampaignRuleActionEntities != nil {
			actionMap["campaign_rule_action_entities"] = flattenCampaignRuleActionEntities(currentAction.CampaignRuleActionEntities)
		}

		ruleActionsList = append(ruleActionsList, actionMap)
	}

	return ruleActionsList
}

func flattenCampaignRuleActionEntities(sdkActionEntity *platformclientv2.Campaignruleactionentities) *schema.Set {
	var (
		campaigns   []interface{}
		sequences   []interface{}
		entitiesSet = schema.NewSet(schema.HashResource(outboundCampaignRuleActionEntities), []interface{}{})
		entitiesMap = make(map[string]interface{})
	)

	if sdkActionEntity == nil {
		return nil
	}

	if sdkActionEntity.Campaigns != nil {
		for _, campaign := range *sdkActionEntity.Campaigns {
			campaigns = append(campaigns, *campaign.Id)
		}
	}

	if sdkActionEntity.Sequences != nil {
		for _, sequence := range *sdkActionEntity.Sequences {
			sequences = append(sequences, *sequence.Id)
		}
	}

	entitiesMap["campaign_ids"] = campaigns
	entitiesMap["sequence_ids"] = sequences
	entitiesMap["use_triggering_entity"] = *sdkActionEntity.UseTriggeringEntity

	entitiesSet.Add(entitiesMap)
	return entitiesSet
}

func flattenRuleParameters(sdkParams platformclientv2.Campaignruleparameters) []interface{} {
	paramsMap := make(map[string]interface{})

	if sdkParams.Operator != nil {
		paramsMap["operator"] = *sdkParams.Operator
	}

	if sdkParams.Value != nil {
		paramsMap["value"] = *sdkParams.Value
	}

	if sdkParams.Priority != nil {
		paramsMap["priority"] = *sdkParams.Priority
	}

	if sdkParams.DialingMode != nil {
		paramsMap["dialing_mode"] = *sdkParams.DialingMode
	}

	return []interface{}{paramsMap}
}
