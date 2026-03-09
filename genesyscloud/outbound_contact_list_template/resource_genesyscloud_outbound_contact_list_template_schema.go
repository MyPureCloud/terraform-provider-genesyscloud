package outbound_contact_list_template

// @team: Outbound Digital
// @chat: #genesys-cloud-digital-campaigns
// @description: Manages outbound campaign operations including automated voice dialing, SMS/email messaging campaigns, contact list management, and campaign rules for proactive customer outreach.

import (
	"context"
	"fmt"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func normalizeOutboundContactListTemplateTimeColumnFields(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	// Normalize plan-time values so the deprecated `*_time_column` and new `*_time_column_name`
	// attributes don't cause TypeSet element mismatches/diffs during migration.
	if v := diff.Get("phone_columns"); v != nil {
		if s, ok := v.(*schema.Set); ok && s.Len() > 0 {
			newSet := schema.NewSet(hashOutboundContactListTemplatePhoneColumn, []interface{}{})
			for _, item := range s.List() {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				if newName, _ := m["callable_time_column_name"].(string); newName == "" {
					if oldName, _ := m["callable_time_column"].(string); oldName != "" {
						m["callable_time_column_name"] = oldName
					}
				}
				newSet.Add(m)
			}
			_ = diff.SetNew("phone_columns", newSet)
		}
	}

	if v := diff.Get("email_columns"); v != nil {
		if s, ok := v.(*schema.Set); ok && s.Len() > 0 {
			newSet := schema.NewSet(hashOutboundContactListTemplateEmailColumn, []interface{}{})
			for _, item := range s.List() {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				if newName, _ := m["contactable_time_column_name"].(string); newName == "" {
					if oldName, _ := m["contactable_time_column"].(string); oldName != "" {
						m["contactable_time_column_name"] = oldName
					}
				}
				newSet.Add(m)
			}
			_ = diff.SetNew("email_columns", newSet)
		}
	}

	return nil
}

func hashOutboundContactListTemplatePhoneColumn(v interface{}) int {
	m, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}
	columnName, _ := m["column_name"].(string)
	colType, _ := m["type"].(string)
	timeColName, _ := m["callable_time_column_name"].(string)
	if timeColName == "" {
		timeColName, _ = m["callable_time_column"].(string)
	}
	return schema.HashString(fmt.Sprintf("%s|%s|%s", columnName, colType, timeColName))
}

func hashOutboundContactListTemplateEmailColumn(v interface{}) int {
	m, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}
	columnName, _ := m["column_name"].(string)
	colType, _ := m["type"].(string)
	timeColName, _ := m["contactable_time_column_name"].(string)
	if timeColName == "" {
		timeColName, _ = m["contactable_time_column"].(string)
	}
	return schema.HashString(fmt.Sprintf("%s|%s|%s", columnName, colType, timeColName))
}

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
				Deprecated:  "Use `callable_time_column_name` instead.",
				Optional:    true,
				Type:        schema.TypeString,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// When automatic timezone mapping is enabled, the API may drop callable time columns.
					// Suppress diffs to prevent perpetual drift.
					return d.Get("automatic_time_zone_mapping").(bool)
				},
			},
			`callable_time_column_name`: {
				Description: `A column name that indicates the timezone to use for a given contact when checking callable times.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("automatic_time_zone_mapping").(bool)
				},
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
				Deprecated:  "Use `contactable_time_column_name` instead.",
				Optional:    true,
				Type:        schema.TypeString,
			},
			`contactable_time_column_name`: {
				Description: `A column name that indicates the timezone to use for a given contact when checking contactable times.`,
				Optional:    true,
				Computed:    true,
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
		CustomizeDiff: customdiff.Sequence(normalizeOutboundContactListTemplateTimeColumnFields),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 2,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 1,
				Type:    resourceOutboundContactListTemplateV1().CoreConfigSchema().ImpliedType(),
				Upgrade: stateUpgraderOutboundContactListTemplateV1ToV2,
			},
		},
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
				Set:         hashOutboundContactListTemplatePhoneColumn,
				Elem:        outboundContactListTemplateContactPhoneNumberColumnResource,
			},
			`email_columns`: {
				Description: `Indicates which columns are email addresses. Changing the email_columns attribute will cause the outbound_contact_list_template object to be dropped and recreated with a new ID. Required if phone_columns is empty`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeSet,
				Set:         hashOutboundContactListTemplateEmailColumn,
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

func resourceOutboundContactListTemplateV1() *schema.Resource {
	outboundContactListTemplateContactPhoneNumberColumnResourceV1 := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Required: true,
				Type:     schema.TypeString,
			},
			`type`: {
				Required: true,
				Type:     schema.TypeString,
			},
			`callable_time_column`: {
				Optional: true,
				Type:     schema.TypeString,
			},
		},
	}

	outboundContactListTemplateEmailColumnResourceV1 := &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Required: true,
				Type:     schema.TypeString,
			},
			`type`: {
				Required: true,
				Type:     schema.TypeString,
			},
			`contactable_time_column`: {
				Optional: true,
				Type:     schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Required: true,
				Type:     schema.TypeString,
			},
			`column_names`: {
				Required: true,
				ForceNew: true,
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			`phone_columns`: {
				Optional: true,
				ForceNew: true,
				Type:     schema.TypeSet,
				Elem:     outboundContactListTemplateContactPhoneNumberColumnResourceV1,
			},
			`email_columns`: {
				Optional: true,
				ForceNew: true,
				Type:     schema.TypeSet,
				Elem:     outboundContactListTemplateEmailColumnResourceV1,
			},
			`preview_mode_column_name`: {
				Optional: true,
				Type:     schema.TypeString,
			},
			`preview_mode_accepted_values`: {
				Optional: true,
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			`attempt_limit_id`: {
				Optional: true,
				Type:     schema.TypeString,
			},
			`automatic_time_zone_mapping`: {
				Optional: true,
				ForceNew: true,
				Type:     schema.TypeBool,
			},
			`zip_code_column_name`: {
				Optional: true,
				ForceNew: true,
				Type:     schema.TypeString,
			},
			`column_data_type_specifications`: {
				Optional: true,
				ForceNew: true,
				Type:     schema.TypeList,
				Elem:     outboundContactListTemplateColumnDataTypeSpecification,
			},
		},
	}
}

func stateUpgraderOutboundContactListTemplateV1ToV2(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	migrateSet := func(v interface{}, legacyKey, nameKey string) {
		list, ok := v.([]interface{})
		if !ok {
			return
		}
		for _, item := range list {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			legacy, _ := m[legacyKey].(string)
			name, _ := m[nameKey].(string)
			if name == "" && legacy != "" {
				m[nameKey] = legacy
			}
			if legacy == "" && name != "" {
				m[legacyKey] = name
			}
		}
	}

	if v, ok := rawState["phone_columns"]; ok {
		migrateSet(v, "callable_time_column", "callable_time_column_name")
	}
	if v, ok := rawState["email_columns"]; ok {
		migrateSet(v, "contactable_time_column", "contactable_time_column_name")
	}

	return rawState, nil
}
