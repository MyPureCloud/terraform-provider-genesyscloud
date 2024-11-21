package outbound_digitalruleset

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

/*
resource_genesycloud_outbound_digitalruleset_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the outbound_digitalruleset resource.
3.  The datasource schema definitions for the outbound_digitalruleset datasource.
4.  The resource exporter configuration for the outbound_digitalruleset exporter.
*/
const ResourceType = "genesyscloud_outbound_digitalruleset"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceOutboundDigitalruleset())
	regInstance.RegisterDataSource(ResourceType, DataSourceOutboundDigitalruleset())
	regInstance.RegisterExporter(ResourceType, OutboundDigitalrulesetExporter())
}

var (
	contactColumnConditionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Description: `The name of the contact list column to evaluate.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`operator`: {
				Description: `The operator to use when comparing values.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`value`: {
				Description: `The value to compare against the contact's data.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`value_type`: {
				Description: `The data type the value should be treated as.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	contactAddressConditionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`operator`: {
				Description: `The operator to use when comparing address values.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`value`: {
				Description: `The value to compare against the contact's address.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	contactAddressTypeConditionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`operator`: {
				Description: `The operator to use when comparing the address types.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`value`: {
				Description: `The type value to compare against the contact column type.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	lastAttemptByColumnConditionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`email_column_name`: {
				Description: `The name of the contact column to evaluate for Email.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`sms_column_name`: {
				Description: `The name of the contact column to evaluate for SMS.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`operator`: {
				Description:  `The operator to use when comparing values.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"Before", "After"}, false),
			},
			`value`: {
				Description: `The period value to compare against the contact's data.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	lastAttemptOverallConditionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`media_types`: {
				Description: `A list of media types to evaluate.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`operator`: {
				Description:  `The operator to use when comparing values.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"Before", "After"}, false),
			},
			`value`: {
				Description: `The period value to compare against the contact's data.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	lastResultByColumnConditionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`email_column_name`: {
				Description: `The name of the contact column to evaluate for Email.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`email_wrapup_codes`: {
				Description: `A list of wrapup code identifiers to match for Email.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`sms_column_name`: {
				Description: `The name of the contact column to evaluate for SMS.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`sms_wrapup_codes`: {
				Description: `A list of wrapup code identifiers to match for SMS.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	lastResultOverallConditionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`email_wrapup_codes`: {
				Description: `A list of wrapup code identifiers to match for Email.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`sms_wrapup_codes`: {
				Description: `A list of wrapup code identifiers to match for SMS.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	digitalDataActionConditionPredicateResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`output_field`: {
				Description: `The name of an output field from the data action's output to use for this condition`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`output_operator`: {
				Description: `The operation with which to evaluate this condition`,
				Required:    true,
				Type:        schema.TypeString,
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

	dataActionContactColumnFieldMappingResource = &schema.Resource{
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

	dataActionConditionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`data_action_id`: {
				Description: `The Data Action Id to use for this condition.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`contact_id_field`: {
				Description: `The input field from the data action that the contactId will be passed into.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`data_not_found_resolution`: {
				Description: `The result of this condition if the data action returns a result indicating there was no data.`,
				Required:    true,
				Type:        schema.TypeBool,
			},
			`predicates`: {
				Description: `A list of predicates defining the comparisons to use for this condition.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        digitalDataActionConditionPredicateResource,
			},
			`contact_column_to_data_action_field_mappings`: {
				Description: `A list of mappings defining which contact data fields will be passed to which data action input fields.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        dataActionContactColumnFieldMappingResource,
			},
		},
	}

	digitalConditionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`inverted`: {
				Description: `If true, inverts the result of evaluating this condition. Default is false.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`contact_column_condition_settings`: {
				Description: `The settings for a 'contact list column' condition.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        contactColumnConditionSettingsResource,
			},
			`contact_address_condition_settings`: {
				Description: `The settings for a 'contact address' condition.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        contactAddressConditionSettingsResource,
			},
			`contact_address_type_condition_settings`: {
				Description: `The settings for a 'contact address type' condition.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        contactAddressTypeConditionSettingsResource,
			},
			`last_attempt_by_column_condition_settings`: {
				Description: `The settings for a 'last attempt by column' condition.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        lastAttemptByColumnConditionSettingsResource,
			},
			`last_attempt_overall_condition_settings`: {
				Description: `The settings for a 'last attempt overall' condition.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        lastAttemptOverallConditionSettingsResource,
			},
			`last_result_by_column_condition_settings`: {
				Description: `The settings for a 'last result by column' condition.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        lastResultByColumnConditionSettingsResource,
			},
			`last_result_overall_condition_settings`: {
				Description: `The settings for a 'last result overall' condition.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        lastResultOverallConditionSettingsResource,
			},
			`data_action_condition_settings`: {
				Description: `The settings for a 'data action' condition.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        dataActionConditionSettingsResource,
			},
		},
	}

	updateContactColumnActionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`properties`: {
				Description:      `A map of key-value pairs pertinent to the DialerAction. Different types of DialerActions require different properties. MODIFY_CONTACT_ATTRIBUTE with an updateOption of SET takes a contact column as the key and accepts any value. SCHEDULE_CALLBACK takes a key 'callbackOffset' that specifies how far in the future the callback should be scheduled, in minutes. SET_CALLER_ID takes two keys: 'callerAddress', which should be the caller id phone number, and 'callerName'. For either key, you can also specify a column on the contact to get the value from. To do this, specify 'contact.Column', where 'Column' is the name of the contact column from which to get the value. SET_SKILLS takes a key 'skills' with an array of skill ids wrapped into a string (Example: {'skills': '['skillIdHere']'} ).`,
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			`update_option`: {
				Description: `The type of update to make to the specified contact column(s).`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	appendToDncActionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`expire`: {
				Description: `Whether to expire the record appended to the DNC list.`,
				Required:    true,
				Type:        schema.TypeBool,
			},
			`expiration_duration`: {
				Description: `If 'expire' is set to true, how long to keep the record.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`list_type`: {
				Description: `The Dnc List Type to append entries to`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	markContactUncontactableActionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`media_types`: {
				Description: `A list of media types to evaluate.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	setContentTemplateActionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`sms_content_template_id`: {
				Description: `A string of sms contentTemplateId.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`email_content_template_id`: {
				Description: `A string of email contentTemplateId.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	setSmsPhoneNumberActionSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`sender_sms_phone_number`: {
				Description: `The string address for the sms phone number.`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	digitalActionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`update_contact_column_action_settings`: {
				Description: `The settings for an 'update contact column' action.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        updateContactColumnActionSettingsResource,
			},
			`do_not_send_action_settings`: {
				Description:      `The settings for a 'do not send' action.`,
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			`append_to_dnc_action_settings`: {
				Description: `The settings for an 'Append to DNC' action.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        appendToDncActionSettingsResource,
			},
			`mark_contact_uncontactable_action_settings`: {
				Description: `The settings for a 'mark contact uncontactable' action.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        markContactUncontactableActionSettingsResource,
			},
			`mark_contact_address_uncontactable_action_settings`: {
				Description:      `The settings for an 'mark contact address uncontactable' action.`,
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			`set_content_template_action_settings`: {
				Description: `The settings for a 'Set content template' action.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        setContentTemplateActionSettingsResource,
			},
			`set_sms_phone_number_action_settings`: {
				Description: `The settings for a 'set sms phone number' action.`,
				Optional:    true,
				Type:        schema.TypeSet,
				MaxItems:    1,
				Elem:        setSmsPhoneNumberActionSettingsResource,
			},
		},
	}

	digitalRuleResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the rule.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`order`: {
				Description: `The ranked order of the rule. Rules are processed from lowest number to highest.`,
				Required:    true,
				Type:        schema.TypeInt,
			},
			`category`: {
				Description:  `The category of the rule.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"PreContact", "PostContact"}, false),
			},
			`conditions`: {
				Description: `A list of conditions to evaluate. All of the Conditions must evaluate to true to trigger the actions.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        digitalConditionResource,
			},
			`actions`: {
				Description: `The list of actions to be taken if all conditions are true.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        digitalActionResource,
			},
		},
	}
)

// ResourceOutboundDigitalruleset registers the genesyscloud_outbound_digitalruleset resource with Terraform
func ResourceOutboundDigitalruleset() *schema.Resource {

	return &schema.Resource{
		Description: `Genesys Cloud outbound digitalruleset`,

		CreateContext: provider.CreateWithPooledClient(createOutboundDigitalruleset),
		ReadContext:   provider.ReadWithPooledClient(readOutboundDigitalruleset),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundDigitalruleset),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundDigitalruleset),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the digital rule set`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`contact_list_id`: {
				Description: `A ContactList to provide suggestions for contact columns on relevant conditions and actions.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`rules`: {
				Description: `The list of rules.`,
				Required:    true,
				Type:        schema.TypeList,
				Elem:        digitalRuleResource,
			},
		},
	}
}

// OutboundDigitalrulesetExporter returns the resourceExporter object used to hold the genesyscloud_outbound_digitalruleset exporter's config
func OutboundDigitalrulesetExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthOutboundDigitalrulesets),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"contact_list_id": {
				RefType: "genesyscloud_outbound_contact_list",
			},
		},
	}
}

// DataSourceOutboundDigitalruleset registers the genesyscloud_outbound_digitalruleset data source
func DataSourceOutboundDigitalruleset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound digitalruleset data source. Select an outbound digitalruleset by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundDigitalrulesetRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `outbound digitalruleset name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
