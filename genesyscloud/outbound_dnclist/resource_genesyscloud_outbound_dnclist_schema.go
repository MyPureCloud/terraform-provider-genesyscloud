package outbound_dnclist

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"
)

const resourceName = "genesyscloud_outbound_dnclist"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceOutboundDncList())
	l.RegisterResource(resourceName, ResourceOutboundDncList())
	l.RegisterExporter(resourceName, OutboundDncListExporter())
}

func ResourceOutboundDncList() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound DNC List`,

		CreateContext: provider.CreateWithPooledClient(createOutboundDncList),
		ReadContext:   provider.ReadWithPooledClient(readOutboundDncList),
		UpdateContext: provider.UpdateWithPooledClient(updateOutboundDncList),
		DeleteContext: provider.DeleteWithPooledClient(deleteOutboundDncList),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the DncList.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`contact_method`: {
				Description:  `The contact method. Required if dncSourceType is rds.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`Email`, `Phone`}, false),
			},
			`login_id`: {
				Description: `A dnc.com loginId. Required if the dncSourceType is dnc.com.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`campaign_id`: {
				Description: `A dnc.com campaignId. Optional if the dncSourceType is dnc.com.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`dnc_codes`: {
				Description: `The list of dnc.com codes to be treated as DNC. Required if the dncSourceType is dnc.com.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{`B`, `C`, `D`, `E`, `F`, `G`, `H`, `I`, `L`, `M`, `O`, `P`, `R`, `S`, `T`, `V`, `W`, `X`, `Y`}, false),
				},
			},
			`license_id`: {
				Description: `A gryphon license number. Required if the dncSourceType is gryphon.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`division_id`: {
				Description: `The division this DNC List belongs to.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`dnc_source_type`: {
				Description:  `The type of the DNC List. Changing the dnc_source_attribute will cause the outbound_dnclist object to be dropped and recreated with new ID.`,
				Required:     true,
				ForceNew:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`rds`, `dnc.com`, `gryphon`}, false),
			},
			`entries`: {
				Description: `Rows to add to the DNC list. To emulate removing phone numbers, you can set expiration_date to a date in the past.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`expiration_date`: {
							Description:      `Expiration date for DNC phone numbers in yyyy-MM-ddTHH:mmZ format.`,
							Optional:         true,
							Type:             schema.TypeString,
							ValidateDiagFunc: validators.ValidateDateTime,
						},
						`phone_numbers`: {
							Description: `Phone numbers to add to a DNC list. Only possible if the dncSourceType is rds.  Phone numbers must be in an E.164 number format.`,
							Optional:    true,
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validators.ValidatePhoneNumber,
							},
						},
					},
				},
			},
		},
	}
}

func DataSourceOutboundDncList() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Outbound DNC Lists. Select a DNC list by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceOutboundDncListRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "DNC List name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func OutboundDncListExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllOutboundDncLists),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}
