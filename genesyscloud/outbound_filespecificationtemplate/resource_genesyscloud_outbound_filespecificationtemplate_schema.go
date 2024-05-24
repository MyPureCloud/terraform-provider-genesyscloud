package outbound_filespecificationtemplate

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_outbound_filespecificationtemplate"

var (
	outboundFileSpecificationTemplateColumnInformationResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column_name`: {
				Description: `Column name. Mandatory for Fixed position/length file format.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`column_number`: {
				Description: `0 based column number in delimited file format.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`start_position`: {
				Description: `Zero-based position of the first column's character. Mandatory for Fixed position/length file format.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`length`: {
				Description: `Column width. Mandatory for Fixed position/length file format.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
		},
	}

	outboundFileSpecificationTemplatePreprocessingRuleResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`find`: {
				Description: `The regular expression to which file lines are to be matched`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`replace_with`: {
				Description: `The string to be substituted for each match.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`global`: {
				Description: `Replaces all matching substrings in every line.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`ignore_case`: {
				Description: `Enables case-insensitive matching.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
		},
	}
)

func ResourceOutboundFileSpecificationTemplate() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound File Specification Template`,

		CreateContext: provider.CreateWithPooledClient(createOutboundFileSpecificationTemplate),
		ReadContext:   provider.ReadWithPooledClient(readOutboundFileSpecificationTemplate),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundFileSpecificationTemplate),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundFileSpecificationTemplate),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the File Specification template.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`description`: {
				Description: `Description of the file specification template`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`format`: {
				Description:  `File format`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"FixedLength", "Delimited"}, false),
			},
			`number_of_header_lines_skipped`: {
				Description: `Number of heading lines to be skipped`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`number_of_trailer_lines_skipped`: {
				Description: `Number of trailing lines to be skipped`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`header`: {
				Description: `If true indicates that delimited file has a header row, which can provide column names`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`delimiter`: {
				Description:  `Kind of delimiter`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"Comma", "Pipe", "Colon", "Tab", "Semicolon", "Custom"}, false),
				Default:      "Comma",
			},
			`delimiter_value`: {
				Description: `Delimiter character, used only when delimiter="Custom"`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`column_information`: {
				Description: `Columns specification`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        outboundFileSpecificationTemplateColumnInformationResource,
			},
			`preprocessing_rule`: {
				Description: `Preprocessing rule`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        outboundFileSpecificationTemplatePreprocessingRuleResource,
			},
		},
	}
}

func dataSourceOutboundFileSpecificationTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound File Specification Template. Select a file specification template by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundFileSpecificationTemplateRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "File Specification Template name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, dataSourceOutboundFileSpecificationTemplate())
	l.RegisterResource(resourceName, ResourceOutboundFileSpecificationTemplate())
	l.RegisterExporter(resourceName, OutboundFileSpecificationTemplateExporter())
}

func OutboundFileSpecificationTemplateExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllFileSpecificationTemplates),
	}
}
