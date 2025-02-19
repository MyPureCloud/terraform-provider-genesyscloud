package location

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceLocation())
	l.RegisterResource(ResourceType, ResourceLocation())
	l.RegisterExporter(ResourceType, LocationExporter())
}

const ResourceType = "genesyscloud_location"

func ResourceLocation() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Location",

		CreateContext: provider.CreateWithPooledClient(createLocation),
		ReadContext:   provider.ReadWithPooledClient(readLocation),
		UpdateContext: provider.UpdateWithPooledClient(updateLocation),
		DeleteContext: provider.DeleteWithPooledClient(deleteLocation),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Location name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"path": {
				Description: "A list of ancestor location IDs. This can be used to create sublocations.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"notes": {
				Description: "Notes for this location.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"emergency_number": {
				Description: "Emergency phone number for this location.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": {
							Description:      "Emergency phone number.  Must be in an E.164 number format.",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidatePhoneNumber,
							DiffSuppressFunc: comparePhoneNumbers,
						},
						"type": {
							Description:  "Type of emergency number (default | elin).",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "default",
							ValidateFunc: validation.StringInSlice([]string{"default", "elin"}, false),
						},
					},
				},
			},
			"address": {
				Description: "Address for this location. This cannot be changed while an emergency number is assigned.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"city": {
							Description: "Location city.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"country": {
							Description: "Country abbreviation.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"state": {
							Description: "Location state. Required for countries with states.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"street1": {
							Description: "Street address 1.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"street2": {
							Description: "Street address 2.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"zip_code": {
							Description: "Location zip code.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func DataSourceLocation() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Location. Select a location by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceLocationRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Location name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func LocationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllLocations),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"path": {RefType: "genesyscloud_location"},
		},
		CustomValidateExports: map[string][]string{
			"E164": {"emergency_number.number"},
		},
	}
}
