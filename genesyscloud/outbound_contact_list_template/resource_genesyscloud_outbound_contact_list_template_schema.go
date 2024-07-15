package outbound_contact_list_template

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesycloud_outbound_contact_list_template_schema.go holds three functions within it:

1.  The resource schema definitions for the outbound_contact_list_template resource.
2.  The datasource schema definitions for the outbound_contact_list_template datasource.
3.  The resource exporter configuration for the outbound_contact_list_template exporter.
*/

var (
	outboundContactListTemplateContactPhoneNumberColumnResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Description: `The name of the phone column.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`type`: {
				Description: `Indicates the type of the phone column. For example, 'cell' or 'home'.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`callable_time_column`: {
				Description: `A column that indicates the timezone to use for a given contact when checking callable times. Not allowed if 'automaticTimeZoneMapping' is set to true.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	outboundContactListTemplateEmailColumnResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Description: `The name of the email column.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`type`: {
				Description: `Indicates the type of the email column. For example, 'work' or 'personal'.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`contactable_time_column`: {
				Description: `A column that indicates the timezone to use for a given contact when checking contactable times.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	outboundContactListTemplateColumnDataTypeSpecification = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Description: `The column name of a column selected for dynamic queueing.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`column_data_type`: {
				Description:  `The data type of the column selected for dynamic queueing (TEXT, NUMERIC or TIMESTAMP)`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"TEXT", "NUMERIC", "TIMESTAMP"}, false),
			},
			`min`: {
				Description: `The minimum length of the numeric column selected for dynamic queueing.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`max`: {
				Description: `The maximum length of the numeric column selected for dynamic queueing.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`max_length`: {
				Description: `The maximum length of the text column selected for dynamic queueing.`,
				Required:    true,
				Type:        schema.TypeInt,
			},
		},
	}
)

func ResourceOutboundContactListTemplate() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Contact List Template`,

		CreateContext: provider.CreateWithPooledClient(createOutboundContactListTemplate),
		ReadContext:   provider.ReadWithPooledClient(readOutboundContactListTemplate),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundContactListTemplate),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundContactListTemplate),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name for the contact list template.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`column_names`: {
				Description: `The names of the contact template data columns. Changing the column_names attribute will cause the outbound_contact_list_template object to be dropped and recreated with a new ID`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`phone_columns`: {
				Description: `Indicates which columns are phone numbers. Changing the phone_columns attribute will cause the outbound_contact_list_template object to be dropped and recreated with a new ID. Required if email_columns is empty`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeSet,
				Elem:        outboundContactListTemplateContactPhoneNumberColumnResource,
			},
			`email_columns`: {
				Description: `Indicates which columns are email addresses. Changing the email_columns attribute will cause the outbound_contact_list_template object to be dropped and recreated with a new ID. Required if phone_columns is empty`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeSet,
				Elem:        outboundContactListTemplateEmailColumnResource,
			},
			`preview_mode_column_name`: {
				Description: `A column to check if a contact should always be dialed in preview mode.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`preview_mode_accepted_values`: {
				Description: `The values in the preview_mode_column_name column that indicate a contact should always be dialed in preview mode.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`attempt_limit_id`: {
				Description: `Attempt Limit for this Contact List Template.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`automatic_time_zone_mapping`: {
				Description: `Indicates if automatic time zone mapping is to be used for this Contact List Template. Changing the automatic_time_zone_mappings attribute will cause the outbound_contact_list_template object to be dropped and recreated with a new ID`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeBool,
			},
			`zip_code_column_name`: {
				Description: `The name of contact list column containing the zip code for use with automatic time zone mapping. Only allowed if 'automatic_time_zone_mapping' is set to true. Changing the zip_code_column_name attribute will cause the outbound_contact_list_template object to be dropped and recreated with a new ID`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			`column_data_type_specifications`: {
				Description: `The settings of the columns selected for dynamic queueing. If updated, the contact list template is dropped and recreated with a new ID`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				Elem:        outboundContactListTemplateColumnDataTypeSpecification,
			},
		},
	}
}

func OutboundContactListTemplateExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllOutboundContactListTemplates),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"attempt_limit_id": {RefType: "genesyscloud_outbound_attempt_limit"},
			"division_id":      {RefType: "genesyscloud_auth_division"},
		},
	}
}

func DataSourceOutboundContactListTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound Contact Lists Templates. Select a contact list template by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundContactListTemplateRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Contact List Template name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
