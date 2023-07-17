package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

var (
	outboundrulesetdialerruleResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the rule.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`order`: {
				Description: `The ranked order of the rule. Rules are processed from lowest number to highest.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`category`: {
				Description:  `The category of the rule.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`DIALER_PRECALL`, `DIALER_WRAPUP`}, false),
			},
			`conditions`: {
				Description: `A list of Conditions. All of the Conditions must evaluate to true to trigger the actions.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        outboundrulesetconditionResource,
			},
			`actions`: {
				Description: `The list of actions to be taken if the conditions are true.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        outboundrulesetdialeractionResource,
			},
		},
	}
	outboundrulesetconditionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`type`: {
				Description:  `The type of the condition.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`wrapupCondition`, `systemDispositionCondition`, `contactAttributeCondition`, `phoneNumberCondition`, `phoneNumberTypeCondition`, `callAnalysisCondition`, `contactPropertyCondition`, `dataActionCondition`}, false),
			},
			`inverted`: {
				Description: `If true, inverts the result of evaluating this Condition. Default is false.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`attribute_name`: {
				Description: `An attribute name associated with this Condition. Required for a contactAttributeCondition.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`value`: {
				Description: `A value associated with this Condition. This could be text, a number, or a relative time. Not used for a DataActionCondition.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`value_type`: {
				Description:  `The type of the value associated with this Condition. Not used for a DataActionCondition.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`STRING`, `NUMERIC`, `DATETIME`, `PERIOD`}, false),
			},
			`operator`: {
				Description:  `An operation with which to evaluate the Condition. Not used for a DataActionCondition.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`EQUALS`, `LESS_THAN`, `LESS_THAN_EQUALS`, `GREATER_THAN`, `GREATER_THAN_EQUALS`, `CONTAINS`, `BEGINS_WITH`, `ENDS_WITH`, `BEFORE`, `AFTER`, `IN`}, false),
			},
			`codes`: {
				Description: `List of wrap-up code identifiers. Required for a wrapupCondition.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`property`: {
				Description: `A value associated with the property type of this Condition. Required for a contactPropertyCondition.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`property_type`: {
				Description:  `The type of the property associated with this Condition. Required for a contactPropertyCondition.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`LAST_ATTEMPT_BY_COLUMN`, `LAST_ATTEMPT_OVERALL`, `LAST_WRAPUP_BY_COLUMN`, `LAST_WRAPUP_OVERALL`}, false),
			},
			`data_action_id`: {
				Description: `The Data Action to use for this condition. Required for a dataActionCondition.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`data_not_found_resolution`: {
				Description: `The result of this condition if the data action returns a result indicating there was no data. Required for a DataActionCondition.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`contact_id_field`: {
				Description: `The input field from the data action that the contactId will be passed to for this condition. Valid for a dataActionCondition.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`call_analysis_result_field`: {
				Description: `The input field from the data action that the callAnalysisResult will be passed to for this condition. Valid for a wrapup dataActionCondition.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`agent_wrapup_field`: {
				Description: `The input field from the data action that the agentWrapup will be passed to for this condition. Valid for a wrapup dataActionCondition.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`contact_column_to_data_action_field_mappings`: {
				Description: `A list of mappings defining which contact data fields will be passed to which data action input fields for this condition. Valid for a dataActionCondition.`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        outboundrulesetcontactcolumntodataactionfieldmappingResource,
			},
			`predicates`: {
				Description: `A list of predicates defining the comparisons to use for this condition. Required for a dataActionCondition.`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        outboundrulesetdataactionconditionpredicateResource,
			},
		},
	}
	outboundrulesetcontactcolumntodataactionfieldmappingResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`contact_column_name`: {
				Description: `The name of a contact column whose data will be passed to the data action`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`data_action_field`: {
				Description: `The name of an input field from the data action that the contact column data will be passed to`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
	outboundrulesetdataactionconditionpredicateResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`output_field`: {
				Description: `The name of an output field from the data action's output to use for this condition`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`output_operator`: {
				Description:  `The operation with which to evaluate this condition`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`EQUALS`, `LESS_THAN`, `LESS_THAN_EQUALS`, `GREATER_THAN`, `GREATER_THAN_EQUALS`, `CONTAINS`, `BEGINS_WITH`, `ENDS_WITH`, `BEFORE`, `AFTER`}, false),
			},
			`comparison_value`: {
				Description: `The value to compare against for this condition`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`inverted`: {
				Description: `If true, inverts the result of evaluating this Predicate. Default is false.`,
				Required:    true,
				Type:        schema.TypeBool,
			},
			`output_field_missing_resolution`: {
				Description: `The result of this predicate if the requested output field is missing from the data action's result`,
				Required:    true,
				Type:        schema.TypeBool,
			},
		},
	}
	outboundrulesetdialeractionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`type`: {
				Description:  `The type of this DialerAction.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`Action`, `modifyContactAttribute`, `dataActionBehavior`}, false),
			},
			`action_type_name`: {
				Description:  `Additional type specification for this DialerAction.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`DO_NOT_DIAL`, `MODIFY_CONTACT_ATTRIBUTE`, `SWITCH_TO_PREVIEW`, `APPEND_NUMBER_TO_DNC_LIST`, `SCHEDULE_CALLBACK`, `CONTACT_UNCALLABLE`, `NUMBER_UNCALLABLE`, `SET_CALLER_ID`, `SET_SKILLS`, `DATA_ACTION`}, false),
			},
			`update_option`: {
				Description:  `Specifies how a contact attribute should be updated. Required for MODIFY_CONTACT_ATTRIBUTE.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`SET`, `INCREMENT`, `DECREMENT`, `CURRENT_TIME`}, false),
			},
			`properties`: {
				Description: `A map of key-value pairs pertinent to the DialerAction. Different types of DialerActions require different properties. MODIFY_CONTACT_ATTRIBUTE with an updateOption of SET takes a contact column as the key and accepts any value. SCHEDULE_CALLBACK takes a key 'callbackOffset' that specifies how far in the future the callback should be scheduled, in minutes. SET_CALLER_ID takes two keys: 'callerAddress', which should be the caller id phone number, and 'callerName'. For either key, you can also specify a column on the contact to get the value from. To do this, specify 'contact.Column', where 'Column' is the name of the contact column from which to get the value. SET_SKILLS takes a key 'skills' with an array of skill ids wrapped into a string (Example: {'skills': '['skillIdHere']'} ).`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`data_action_id`: {
				Description: `The Data Action to use for this action. Required for a dataActionBehavior.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`contact_column_to_data_action_field_mappings`: {
				Description: `A list of mappings defining which contact data fields will be passed to which data action input fields for this condition. Valid for a dataActionBehavior.`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        outboundrulesetcontactcolumntodataactionfieldmappingResource,
			},
			`contact_id_field`: {
				Description: `The input field from the data action that the contactId will be passed to for this condition. Valid for a dataActionBehavior.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`call_analysis_result_field`: {
				Description: `The input field from the data action that the callAnalysisResult will be passed to for this condition. Valid for a wrapup dataActionBehavior.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`agent_wrapup_field`: {
				Description: `The input field from the data action that the agentWrapup will be passed to for this condition. Valid for a wrapup dataActionBehavior.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
)

func resourceOutboundRuleset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound ruleset`,

		CreateContext: CreateWithPooledClient(createOutboundRuleset),
		ReadContext:   ReadWithPooledClient(readOutboundRuleset),
		UpdateContext: UpdateWithPooledClient(updateOutboundRuleset),
		DeleteContext: DeleteWithPooledClient(deleteOutboundRuleset),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the RuleSet.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`contact_list_id`: {
				Description: `A ContactList to provide user-interface suggestions for contact columns on relevant conditions and actions.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`queue_id`: {
				Description: `A Queue to provide user-interface suggestions for wrap-up codes on relevant conditions and actions.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`rules`: {
				Description: `The list of rules.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        outboundrulesetdialerruleResource,
			},
		},
	}
}

func getAllOutboundRuleset(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	outboundApi := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sdkrulesetentitylisting, _, getErr := outboundApi.GetOutboundRulesets(pageSize, pageNum, true, "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Error requesting page of Outbound Ruleset: %s", getErr)
		}

		if sdkrulesetentitylisting.Entities == nil || len(*sdkrulesetentitylisting.Entities) == 0 {
			break
		}

		for _, entity := range *sdkrulesetentitylisting.Entities {
			resources[*entity.Id] = &ResourceMeta{Name: *entity.Name}
		}
	}

	return resources, nil
}

func outboundRulesetExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllOutboundRuleset),
		RefAttrs: map[string]*RefAttrSettings{
			"contact_list_id": {
				RefType: "genesyscloud_outbound_contact_list",
			},
			"queue_id": {
				RefType: "genesyscloud_routing_queue",
			},
			"rules.conditions.codes": {
				RefType: "genesyscloud_routing_wrapupcode",
			},
			"rules.conditions.data_action_id": {
				RefType: "genesyscloud_integration_action",
			},
			"rules.actions.data_action_id": {
				RefType: "genesyscloud_integration_action",
			},
		},
	}
}

func createOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkruleset := platformclientv2.Ruleset{
		ContactList: buildSdkDomainEntityRef(d, "contact_list_id"),
		Queue:       buildSdkDomainEntityRef(d, "queue_id"),
		Rules:       buildSdkoutboundrulesetDialerruleSlice(d.Get("rules").([]interface{})),
	}

	if name != "" {
		sdkruleset.Name = &name
	}

	log.Printf("Creating Outbound Ruleset %s", name)
	outboundRuleset, _, err := outboundApi.PostOutboundRulesets(sdkruleset)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Ruleset %s: %s", name, err)
	}

	d.SetId(*outboundRuleset.Id)

	log.Printf("Created Outbound Ruleset %s %s", name, *outboundRuleset.Id)
	return readOutboundRuleset(ctx, d, meta)
}

func updateOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkruleset := platformclientv2.Ruleset{
		ContactList: buildSdkDomainEntityRef(d, "contact_list_id"),
		Queue:       buildSdkDomainEntityRef(d, "queue_id"),
		Rules:       buildSdkoutboundrulesetDialerruleSlice(d.Get("rules").([]interface{})),
	}

	if name != "" {
		sdkruleset.Name = &name
	}

	log.Printf("Updating Outbound Ruleset %s", name)
	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Ruleset version
		outboundRuleset, resp, getErr := outboundApi.GetOutboundRuleset(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Ruleset %s: %s", d.Id(), getErr)
		}
		sdkruleset.Version = outboundRuleset.Version
		outboundRuleset, _, updateErr := outboundApi.PutOutboundRuleset(d.Id(), sdkruleset)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound Ruleset %s: %s", name, updateErr)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Ruleset %s", name)
	return readOutboundRuleset(ctx, d, meta)
}

func readOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Ruleset %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *resource.RetryError {
		sdkruleset, resp, getErr := outboundApi.GetOutboundRuleset(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read Outbound Ruleset %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Outbound Ruleset %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceOutboundRuleset())

		if sdkruleset.Name != nil {
			d.Set("name", *sdkruleset.Name)
		}
		if sdkruleset.ContactList != nil && sdkruleset.ContactList.Id != nil {
			d.Set("contact_list_id", *sdkruleset.ContactList.Id)
		}
		if sdkruleset.Queue != nil && sdkruleset.Queue.Id != nil {
			d.Set("queue_id", *sdkruleset.Queue.Id)
		}
		if sdkruleset.Rules != nil {
			d.Set("rules", flattenSdkoutboundrulesetDialerruleSlice(*sdkruleset.Rules))
		}

		log.Printf("Read Outbound Ruleset %s %s", d.Id(), *sdkruleset.Name)

		return cc.CheckState()
	})
}

func deleteOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Ruleset")
		resp, err := outboundApi.DeleteOutboundRuleset(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Ruleset: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := outboundApi.GetOutboundRuleset(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Outbound Ruleset deleted
				log.Printf("Deleted Outbound Ruleset %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Outbound Ruleset %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Outbound Ruleset %s still exists", d.Id()))
	})
}

func buildSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(contactcolumntodataactionfieldmapping *schema.Set) *[]platformclientv2.Contactcolumntodataactionfieldmapping {
	if contactcolumntodataactionfieldmapping == nil {
		return nil
	}
	sdkContactcolumntodataactionfieldmappingSlice := make([]platformclientv2.Contactcolumntodataactionfieldmapping, 0)
	contactcolumntodataactionfieldmappingList := contactcolumntodataactionfieldmapping.List()
	for _, configcontactcolumntodataactionfieldmapping := range contactcolumntodataactionfieldmappingList {
		var sdkContactcolumntodataactionfieldmapping platformclientv2.Contactcolumntodataactionfieldmapping
		contactcolumntodataactionfieldmappingMap := configcontactcolumntodataactionfieldmapping.(map[string]interface{})
		if contactColumnName := contactcolumntodataactionfieldmappingMap["contact_column_name"].(string); contactColumnName != "" {
			sdkContactcolumntodataactionfieldmapping.ContactColumnName = &contactColumnName
		}
		if dataActionField := contactcolumntodataactionfieldmappingMap["data_action_field"].(string); dataActionField != "" {
			sdkContactcolumntodataactionfieldmapping.DataActionField = &dataActionField
		}

		sdkContactcolumntodataactionfieldmappingSlice = append(sdkContactcolumntodataactionfieldmappingSlice, sdkContactcolumntodataactionfieldmapping)
	}
	return &sdkContactcolumntodataactionfieldmappingSlice
}

func buildSdkoutboundrulesetDataactionconditionpredicateSlice(dataactionconditionpredicate *schema.Set) *[]platformclientv2.Dataactionconditionpredicate {
	if dataactionconditionpredicate == nil {
		return nil
	}
	sdkDataactionconditionpredicateSlice := make([]platformclientv2.Dataactionconditionpredicate, 0)
	dataactionconditionpredicateList := dataactionconditionpredicate.List()
	for _, configdataactionconditionpredicate := range dataactionconditionpredicateList {
		var sdkDataactionconditionpredicate platformclientv2.Dataactionconditionpredicate
		dataactionconditionpredicateMap := configdataactionconditionpredicate.(map[string]interface{})
		if outputField := dataactionconditionpredicateMap["output_field"].(string); outputField != "" {
			sdkDataactionconditionpredicate.OutputField = &outputField
		}
		if outputOperator := dataactionconditionpredicateMap["output_operator"].(string); outputOperator != "" {
			sdkDataactionconditionpredicate.OutputOperator = &outputOperator
		}
		if comparisonValue := dataactionconditionpredicateMap["comparison_value"].(string); comparisonValue != "" {
			sdkDataactionconditionpredicate.ComparisonValue = &comparisonValue
		}
		sdkDataactionconditionpredicate.Inverted = platformclientv2.Bool(dataactionconditionpredicateMap["inverted"].(bool))
		sdkDataactionconditionpredicate.OutputFieldMissingResolution = platformclientv2.Bool(dataactionconditionpredicateMap["output_field_missing_resolution"].(bool))

		sdkDataactionconditionpredicateSlice = append(sdkDataactionconditionpredicateSlice, sdkDataactionconditionpredicate)
	}
	return &sdkDataactionconditionpredicateSlice
}

func buildSdkoutboundrulesetConditionSlice(conditionList []interface{}) *[]platformclientv2.Condition {
	sdkConditionSlice := make([]platformclientv2.Condition, 0)
	for _, configcondition := range conditionList {
		var sdkCondition platformclientv2.Condition
		conditionMap := configcondition.(map[string]interface{})
		if varType := conditionMap["type"].(string); varType != "" {
			sdkCondition.VarType = &varType
		}
		sdkCondition.Inverted = platformclientv2.Bool(conditionMap["inverted"].(bool))
		if attributeName := conditionMap["attribute_name"].(string); attributeName != "" {
			sdkCondition.AttributeName = &attributeName
		}
		if value := conditionMap["value"].(string); value != "" {
			sdkCondition.Value = &value
		}
		if valueType := conditionMap["value_type"].(string); valueType != "" {
			sdkCondition.ValueType = &valueType
		}
		if operator := conditionMap["operator"].(string); operator != "" {
			sdkCondition.Operator = &operator
		}
		codes := make([]string, 0)
		for _, v := range conditionMap["codes"].([]interface{}) {
			codes = append(codes, v.(string))
		}
		sdkCondition.Codes = &codes
		if property := conditionMap["property"].(string); property != "" {
			sdkCondition.Property = &property
		}
		if propertyType := conditionMap["property_type"].(string); propertyType != "" {
			sdkCondition.PropertyType = &propertyType
		}
		sdkCondition.DataAction = &platformclientv2.Domainentityref{Id: platformclientv2.String(conditionMap["data_action_id"].(string))}
		sdkCondition.DataNotFoundResolution = platformclientv2.Bool(conditionMap["data_not_found_resolution"].(bool))
		if contactIdField := conditionMap["contact_id_field"].(string); contactIdField != "" {
			sdkCondition.ContactIdField = &contactIdField
		}
		if callAnalysisResultField := conditionMap["call_analysis_result_field"].(string); callAnalysisResultField != "" {
			sdkCondition.CallAnalysisResultField = &callAnalysisResultField
		}
		if agentWrapupField := conditionMap["agent_wrapup_field"].(string); agentWrapupField != "" {
			sdkCondition.AgentWrapupField = &agentWrapupField
		}
		if contactColumnToDataActionFieldMappings := conditionMap["contact_column_to_data_action_field_mappings"]; contactColumnToDataActionFieldMappings != nil {
			sdkCondition.ContactColumnToDataActionFieldMappings = buildSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(contactColumnToDataActionFieldMappings.(*schema.Set))
		}
		if predicates := conditionMap["predicates"]; predicates != nil {
			sdkCondition.Predicates = buildSdkoutboundrulesetDataactionconditionpredicateSlice(predicates.(*schema.Set))
		}

		sdkConditionSlice = append(sdkConditionSlice, sdkCondition)
	}
	return &sdkConditionSlice
}

func buildSdkoutboundrulesetDialeractionSlice(dialeractionList []interface{}) *[]platformclientv2.Dialeraction {
	sdkDialeractionSlice := make([]platformclientv2.Dialeraction, 0)
	for _, configdialeraction := range dialeractionList {
		var sdkDialeraction platformclientv2.Dialeraction
		dialeractionMap := configdialeraction.(map[string]interface{})
		if varType := dialeractionMap["type"].(string); varType != "" {
			sdkDialeraction.VarType = &varType
		}
		if actionTypeName := dialeractionMap["action_type_name"].(string); actionTypeName != "" {
			sdkDialeraction.ActionTypeName = &actionTypeName
		}
		if updateOption := dialeractionMap["update_option"].(string); updateOption != "" {
			sdkDialeraction.UpdateOption = &updateOption
		}
		if properties := dialeractionMap["properties"].(map[string]interface{}); properties != nil {
			sdkProperties := map[string]string{}
			for k, v := range properties {
				sdkProperties[k] = v.(string)
			}
			sdkDialeraction.Properties = &sdkProperties
		}
		sdkDialeraction.DataAction = &platformclientv2.Domainentityref{Id: platformclientv2.String(dialeractionMap["data_action_id"].(string))}
		if contactColumnToDataActionFieldMappings := dialeractionMap["contact_column_to_data_action_field_mappings"]; contactColumnToDataActionFieldMappings != nil {
			sdkDialeraction.ContactColumnToDataActionFieldMappings = buildSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(contactColumnToDataActionFieldMappings.(*schema.Set))
		}
		if contactIdField := dialeractionMap["contact_id_field"].(string); contactIdField != "" {
			sdkDialeraction.ContactIdField = &contactIdField
		}
		if callAnalysisResultField := dialeractionMap["call_analysis_result_field"].(string); callAnalysisResultField != "" {
			sdkDialeraction.CallAnalysisResultField = &callAnalysisResultField
		}
		if agentWrapupField := dialeractionMap["agent_wrapup_field"].(string); agentWrapupField != "" {
			sdkDialeraction.AgentWrapupField = &agentWrapupField
		}

		sdkDialeractionSlice = append(sdkDialeractionSlice, sdkDialeraction)
	}
	return &sdkDialeractionSlice
}

func buildSdkoutboundrulesetDialerruleSlice(dialerruleList []interface{}) *[]platformclientv2.Dialerrule {
	sdkDialerruleSlice := make([]platformclientv2.Dialerrule, 0)
	for _, configdialerrule := range dialerruleList {
		var sdkDialerrule platformclientv2.Dialerrule
		dialerruleMap := configdialerrule.(map[string]interface{})
		if name := dialerruleMap["name"].(string); name != "" {
			sdkDialerrule.Name = &name
		}
		sdkDialerrule.Order = platformclientv2.Int(dialerruleMap["order"].(int))
		if category := dialerruleMap["category"].(string); category != "" {
			sdkDialerrule.Category = &category
		}
		if conditions := dialerruleMap["conditions"]; conditions != nil {
			sdkDialerrule.Conditions = buildSdkoutboundrulesetConditionSlice(conditions.([]interface{}))
		}
		if actions := dialerruleMap["actions"]; actions != nil {
			sdkDialerrule.Actions = buildSdkoutboundrulesetDialeractionSlice(actions.([]interface{}))
		}

		sdkDialerruleSlice = append(sdkDialerruleSlice, sdkDialerrule)
	}
	return &sdkDialerruleSlice
}

func flattenSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(contactcolumntodataactionfieldmappings []platformclientv2.Contactcolumntodataactionfieldmapping) *schema.Set {
	if len(contactcolumntodataactionfieldmappings) == 0 {
		return nil
	}

	contactcolumntodataactionfieldmappingSet := schema.NewSet(schema.HashResource(outboundrulesetcontactcolumntodataactionfieldmappingResource), []interface{}{})
	for _, contactcolumntodataactionfieldmapping := range contactcolumntodataactionfieldmappings {
		contactcolumntodataactionfieldmappingMap := make(map[string]interface{})

		if contactcolumntodataactionfieldmapping.ContactColumnName != nil {
			contactcolumntodataactionfieldmappingMap["contact_column_name"] = *contactcolumntodataactionfieldmapping.ContactColumnName
		}
		if contactcolumntodataactionfieldmapping.DataActionField != nil {
			contactcolumntodataactionfieldmappingMap["data_action_field"] = *contactcolumntodataactionfieldmapping.DataActionField
		}

		contactcolumntodataactionfieldmappingSet.Add(contactcolumntodataactionfieldmappingMap)
	}

	return contactcolumntodataactionfieldmappingSet
}

func flattenSdkoutboundrulesetDataactionconditionpredicateSlice(dataactionconditionpredicates []platformclientv2.Dataactionconditionpredicate) *schema.Set {
	if len(dataactionconditionpredicates) == 0 {
		return nil
	}

	dataactionconditionpredicateSet := schema.NewSet(schema.HashResource(outboundrulesetdataactionconditionpredicateResource), []interface{}{})
	for _, dataactionconditionpredicate := range dataactionconditionpredicates {
		dataactionconditionpredicateMap := make(map[string]interface{})

		if dataactionconditionpredicate.OutputField != nil {
			dataactionconditionpredicateMap["output_field"] = *dataactionconditionpredicate.OutputField
		}
		if dataactionconditionpredicate.OutputOperator != nil {
			dataactionconditionpredicateMap["output_operator"] = *dataactionconditionpredicate.OutputOperator
		}
		if dataactionconditionpredicate.ComparisonValue != nil {
			dataactionconditionpredicateMap["comparison_value"] = *dataactionconditionpredicate.ComparisonValue
		}
		if dataactionconditionpredicate.Inverted != nil {
			dataactionconditionpredicateMap["inverted"] = *dataactionconditionpredicate.Inverted
		}
		if dataactionconditionpredicate.OutputFieldMissingResolution != nil {
			dataactionconditionpredicateMap["output_field_missing_resolution"] = *dataactionconditionpredicate.OutputFieldMissingResolution
		}

		dataactionconditionpredicateSet.Add(dataactionconditionpredicateMap)
	}

	return dataactionconditionpredicateSet
}

func flattenSdkoutboundrulesetConditionSlice(conditions []platformclientv2.Condition) []interface{} {
	if len(conditions) == 0 {
		return nil
	}

	var conditionList []interface{}
	for _, condition := range conditions {
		conditionMap := make(map[string]interface{})

		if condition.VarType != nil {
			conditionMap["type"] = *condition.VarType
		}
		if condition.Inverted != nil {
			conditionMap["inverted"] = *condition.Inverted
		}
		if condition.AttributeName != nil {
			conditionMap["attribute_name"] = *condition.AttributeName
		}
		if condition.Value != nil {
			conditionMap["value"] = *condition.Value
		}
		if condition.ValueType != nil {
			conditionMap["value_type"] = *condition.ValueType
		}
		if condition.Operator != nil {
			conditionMap["operator"] = *condition.Operator
		}
		if condition.Codes != nil {
			codes := make([]string, 0)
			for _, v := range *condition.Codes {
				codes = append(codes, v)
			}
			conditionMap["codes"] = codes
		}
		if condition.Property != nil {
			conditionMap["property"] = *condition.Property
		}
		if condition.PropertyType != nil {
			conditionMap["property_type"] = *condition.PropertyType
		}
		if condition.DataAction != nil {
			conditionMap["data_action_id"] = *condition.DataAction.Id
		}
		if condition.DataNotFoundResolution != nil {
			conditionMap["data_not_found_resolution"] = *condition.DataNotFoundResolution
		}
		if condition.ContactIdField != nil {
			conditionMap["contact_id_field"] = *condition.ContactIdField
		}
		if condition.CallAnalysisResultField != nil {
			conditionMap["call_analysis_result_field"] = *condition.CallAnalysisResultField
		}
		if condition.AgentWrapupField != nil {
			conditionMap["agent_wrapup_field"] = *condition.AgentWrapupField
		}
		if condition.ContactColumnToDataActionFieldMappings != nil {
			conditionMap["contact_column_to_data_action_field_mappings"] = flattenSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(*condition.ContactColumnToDataActionFieldMappings)
		}
		if condition.Predicates != nil {
			conditionMap["predicates"] = flattenSdkoutboundrulesetDataactionconditionpredicateSlice(*condition.Predicates)
		}

		conditionList = append(conditionList, conditionMap)
	}

	return conditionList
}

func flattenSdkoutboundrulesetDialeractionSlice(dialeractions []platformclientv2.Dialeraction) []interface{} {
	if len(dialeractions) == 0 {
		return nil
	}

	var dialeractionList []interface{}
	for _, dialeraction := range dialeractions {
		dialeractionMap := make(map[string]interface{})

		if dialeraction.VarType != nil {
			dialeractionMap["type"] = *dialeraction.VarType
		}
		if dialeraction.ActionTypeName != nil {
			dialeractionMap["action_type_name"] = *dialeraction.ActionTypeName
		}
		if dialeraction.UpdateOption != nil {
			dialeractionMap["update_option"] = *dialeraction.UpdateOption
		}
		if dialeraction.Properties != nil {
			results := make(map[string]interface{})
			for k, v := range *dialeraction.Properties {
				results[k] = v
			}
			dialeractionMap["properties"] = results
		}
		if dialeraction.DataAction != nil {
			dialeractionMap["data_action_id"] = *dialeraction.DataAction.Id
		}
		if dialeraction.ContactColumnToDataActionFieldMappings != nil {
			dialeractionMap["contact_column_to_data_action_field_mappings"] = flattenSdkoutboundrulesetContactcolumntodataactionfieldmappingSlice(*dialeraction.ContactColumnToDataActionFieldMappings)
		}
		if dialeraction.ContactIdField != nil {
			dialeractionMap["contact_id_field"] = *dialeraction.ContactIdField
		}
		if dialeraction.CallAnalysisResultField != nil {
			dialeractionMap["call_analysis_result_field"] = *dialeraction.CallAnalysisResultField
		}
		if dialeraction.AgentWrapupField != nil {
			dialeractionMap["agent_wrapup_field"] = *dialeraction.AgentWrapupField
		}

		dialeractionList = append(dialeractionList, dialeractionMap)
	}

	return dialeractionList
}

func flattenSdkoutboundrulesetDialerruleSlice(dialerrules []platformclientv2.Dialerrule) []interface{} {
	if len(dialerrules) == 0 {
		return nil
	}

	var dialerruleList []interface{}
	for _, dialerrule := range dialerrules {
		dialerruleMap := make(map[string]interface{})

		if dialerrule.Name != nil {
			dialerruleMap["name"] = *dialerrule.Name
		}
		if dialerrule.Order != nil {
			dialerruleMap["order"] = *dialerrule.Order
		}
		if dialerrule.Category != nil {
			dialerruleMap["category"] = *dialerrule.Category
		}
		if dialerrule.Conditions != nil {
			dialerruleMap["conditions"] = flattenSdkoutboundrulesetConditionSlice(*dialerrule.Conditions)
		}
		if dialerrule.Actions != nil {
			dialerruleMap["actions"] = flattenSdkoutboundrulesetDialeractionSlice(*dialerrule.Actions)
		}

		dialerruleList = append(dialerruleList, dialerruleMap)
	}

	return dialerruleList
}
