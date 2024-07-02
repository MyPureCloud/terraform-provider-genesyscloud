package outbound_contactlistfilter

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesycloud_outbound_contactlistfilter_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the outbound_contactlistfilter resource.
3.  The datasource schema definitions for the outbound_contactlistfilter datasource.
4.  The resource exporter configuration for the outbound_contactlistfilter exporter.
*/
const resourceName = "genesyscloud_outbound_contactlistfilter"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceOutboundContactlistfilter())
	regInstance.RegisterDataSource(resourceName, DataSourceOutboundContactlistfilter())
	regInstance.RegisterExporter(resourceName, OutboundContactlistfilterExporter())
}

var (
	filterClauseResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`filter_type`: {
				Description:  `How to join predicates together.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`AND`, `OR`}, false),
			},
			`predicates`: {
				Description: `Conditions to filter the contacts by.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        predicateResource,
			},
		},
	}
	predicateResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column`: {
				Description: `Contact list column from the contact list filter's contact list.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`column_type`: {
				Description:  `The type of data in the contact column.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`numeric`, `alphabetic`}, false),
			},
			`operator`: {
				Description:  `The operator for this contact list filter predicate.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`EQUALS`, `LESS_THAN`, `LESS_THAN_EQUALS`, `GREATER_THAN`, `GREATER_THAN_EQUALS`, `CONTAINS`, `BEGINS_WITH`, `ENDS_WITH`, `BEFORE`, `AFTER`, `BETWEEN`, `IN`}, false),
			},
			`value`: {
				Description: `Value with which to compare the contact's data. This could be text, a number, or a relative time. A value for relative time should follow the format PxxDTyyHzzM, where xx, yy, and zz specify the days, hours and minutes. For example, a value of P01DT08H30M corresponds to 1 day, 8 hours, and 30 minutes from now. To specify a time in the past, include a negative sign before each numeric value. For example, a value of P-01DT-08H-30M corresponds to 1 day, 8 hours, and 30 minutes in the past. You can also do things like P01DT00H-30M, which would correspond to 23 hours and 30 minutes from now (1 day - 30 minutes).`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`var_range`: {
				Description: `A range of values. Required for operators BETWEEN and IN.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        rangeResource,
			},
			`inverted`: {
				Description: `Inverts the result of the predicate (i.e., if the predicate returns true, inverting it will return false).`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeBool,
			},
		},
	}
	rangeResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`min`: {
				Description: `The minimum value of the range. Required for the operator BETWEEN.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`max`: {
				Description: `The maximum value of the range. Required for the operator BETWEEN.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`min_inclusive`: {
				Description: `Whether or not to include the minimum in the range.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`max_inclusive`: {
				Description: `Whether or not to include the maximum in the range.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`in_set`: {
				Description: `A set of values that the contact data should be in. Required for the IN operator.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
)

// ResourceOutboundContactlistfilter registers the genesyscloud_outbound_contactlistfilter resource with Terraform
func ResourceOutboundContactlistfilter() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Contact List Filter`,

		CreateContext: provider.CreateWithPooledClient(createOutboundContactlistfilter),
		ReadContext:   provider.ReadWithPooledClient(readOutboundContactlistfilter),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundContactlistfilter),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundContactlistfilter),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the list.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`contact_list_id`: {
				Description:  `The contact list the filter is based on. Mutually exclusive to 'contact_list_template_id', however, one of the two must be specified`,
				Optional:     true,
				Type:         schema.TypeString,
				ExactlyOneOf: []string{"contact_list_id", "contact_list_template_id"},
			},
			`contact_list_template_id`: {
				Description:  `The contact list template the filter is based on. Mutually exclusive to 'contact_list_id', however, one of the two must be specified.`,
				Optional:     true,
				Type:         schema.TypeString,
				ExactlyOneOf: []string{"contact_list_id", "contact_list_template_id"},
			},
			`clauses`: {
				Description: `Groups of conditions to filter the contacts by.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        filterClauseResource,
			},
			`filter_type`: {
				Description:  `How to join clauses together.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`AND`, `OR`}, false),
			},
		},
	}
}

// OutboundContactlistfilterExporter returns the resourceExporter object used to hold the genesyscloud_outbound_contactlistfilter exporter's config
func OutboundContactlistfilterExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthOutboundContactlistfilters),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"contact_list_id":          {RefType: "genesyscloud_outbound_contact_list"},
			"contact_list_template_id": {RefType: "genesyscloud_outbound_contact_list_template"},
		},
	}
}

// DataSourceOutboundContactlistfilter registers the genesyscloud_outbound_contactlistfilter data source
func DataSourceOutboundContactlistfilter() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound Contact List Filters. Select a contact list filter by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundContactlistfilterRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Contact List Filter name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
