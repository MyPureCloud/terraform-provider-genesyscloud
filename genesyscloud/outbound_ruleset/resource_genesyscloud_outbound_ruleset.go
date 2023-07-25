package outbound_ruleset

import (
	"context"
	"fmt"
	"log"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

// // Registering our resource provider for export
// func register_outbound_ruleset() {
// 	log.Printf("resource outbound_ruleset registered")
// 	regInstance.RegisterResource("genesyscloud_outbound_ruleset", ResourceOutboundRuleset())
// 	regInstance.RegisterDataSource("genesyscloud_outbound_ruleset", DataSourceOutboundRuleset())
// 	//gcloud.RegisterExporter("genesyscloud_outbound_ruleset", outboundRulesetExporter())
// }

// var regInstance registrar


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
				Required:    true,
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

var outboundAPIProxy *OutboundAPIProxy

func init() {
	outboundAPIProxy = NewOutboundAPIProxy()
}

func ResourceOutboundRuleset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound ruleset`,

		CreateContext: gcloud.CreateWithPooledClient(createOutboundRuleset),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundRuleset),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundRuleset),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundRuleset),
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

func getAllOutboundRuleset(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	outboundAPIProxy.ConfigureProxyApiInstance(clientConfig)
	return outboundAPIProxy.ReadAllOutboundRuleset(outboundAPIProxy)
}

func OutboundRulesetExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllOutboundRuleset),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
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
		JsonEncodeAttributes: []string{"rules.actions.properties.skills"},
	 	CustomAttributeResolver: map[string]*resourceExporter.RefAttrCustomResolver{
			"rules.actions.properties": {ResolverFunc: resourceExporter.RuleSetPropertyResolver},
	 		"rules.actions.properties.skills": {ResolverFunc: resourceExporter.RuleSetSkillPropertyResolver},
		},
	}
}

func createOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundAPIProxy.ConfigureProxyApiInstance(sdkConfig)

	sdkruleset := platformclientv2.Ruleset{
		ContactList: gcloud.BuildSdkDomainEntityRef(d, "contact_list_id"),
		Queue:       gcloud.BuildSdkDomainEntityRef(d, "queue_id"),
		Rules:       buildSdkoutboundrulesetDialerruleSlice(d.Get("rules").([]interface{})),
	}

	if name != "" {
		sdkruleset.Name = &name
	}

	log.Printf("Creating Outbound Ruleset %s", name)
	outboundRuleset, _, err := outboundAPIProxy.CreateOutboundRulesets(outboundAPIProxy, sdkruleset)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Ruleset %s: %s", name, err)
	}

	d.SetId(*outboundRuleset.Id)

	log.Printf("Created Outbound Ruleset %s %s", name, *outboundRuleset.Id)
	return readOutboundRuleset(ctx, d, meta)
}

func updateOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundAPIProxy.ConfigureProxyApiInstance(sdkConfig)

	sdkruleset := platformclientv2.Ruleset{
		ContactList: gcloud.BuildSdkDomainEntityRef(d, "contact_list_id"),
		Queue:       gcloud.BuildSdkDomainEntityRef(d, "queue_id"),
		Rules:       buildSdkoutboundrulesetDialerruleSlice(d.Get("rules").([]interface{})),
	}

	if name != "" {
		sdkruleset.Name = &name
	}

	log.Printf("Updating Outbound Ruleset %s", name)
	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Ruleset version
		outboundRuleset, resp, getErr := outboundAPIProxy.ReadOutboundRuleset(outboundAPIProxy, d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Ruleset %s: %s", d.Id(), getErr)
		}
		sdkruleset.Version = outboundRuleset.Version
		outboundRuleset, _, updateErr := outboundAPIProxy.UpdateOutboundRuleset(outboundAPIProxy, sdkruleset, d.Id())
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
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundAPIProxy.ConfigureProxyApiInstance(sdkConfig)

	log.Printf("Reading Outbound Ruleset %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *resource.RetryError {
		sdkruleset, resp, getErr := outboundAPIProxy.ReadOutboundRuleset(outboundAPIProxy,d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read Outbound Ruleset %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Outbound Ruleset %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundRuleset())

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
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundAPIProxy.ConfigureProxyApiInstance(sdkConfig)

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Ruleset")
		resp, err := outboundAPIProxy.DeleteOutboundRuleset(outboundAPIProxy,d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Ruleset: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := outboundAPIProxy.ReadOutboundRuleset(outboundAPIProxy,d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Outbound Ruleset deleted
				log.Printf("Deleted Outbound Ruleset %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Outbound Ruleset %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Outbound Ruleset %s still exists", d.Id()))
	})
}

