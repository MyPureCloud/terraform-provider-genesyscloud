package genesyscloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"
)

// SetRegistrar registers all the resources, data sources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceRoutingSmsAddress())
	l.RegisterResource(resourceName, ResourceRoutingSmsAddress())
	l.RegisterExporter(resourceName, RoutingSmsAddressExporter())
}

func RoutingSmsAddressExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingSmsAddress),
	}
}

func ResourceRoutingSmsAddress() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud routing sms address`,

		CreateContext: provider.CreateWithPooledClient(createRoutingSmsAddress),
		ReadContext:   provider.ReadWithPooledClient(readRoutingSmsAddress),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingSmsAddress),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `Name associated with this address`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			`street`: {
				Description: `The number and street address where this address is located.`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			`city`: {
				Description: `The city in which this address is in`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			`region`: {
				Description: `The state or region this address is in`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			`postal_code`: {
				Description: `The postal code this address is in`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			`country_code`: {
				Description:      `The ISO country code of this address`,
				Required:         true,
				ForceNew:         true,
				Type:             schema.TypeString,
				ValidateDiagFunc: validators.ValidateCountryCode,
			},
			`auto_correct_address`: {
				Description: `This is used when the address is created. If the value is not set or true, then the system will, if necessary, auto-correct the address you provide. Set this value to false if the system should not auto-correct the address.`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeBool,
			},
		},
	}
}

func DataSourceRoutingSmsAddress() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Routing Sms Address. Select a Routing Sms Address by name.`,

		ReadContext: provider.ReadWithPooledClient(dataSourceRoutingSmsAddressRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Routing Sms Address name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
