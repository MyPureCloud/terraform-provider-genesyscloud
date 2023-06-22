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
	"github.com/mypurecloud/platform-client-sdk-go/v102/platformclientv2"
)

func resourceRoutingSmsAddress() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud routing sms address`,

		CreateContext: CreateWithPooledClient(createRoutingSmsAddress),
		ReadContext:   ReadWithPooledClient(readRoutingSmsAddress),
		DeleteContext: DeleteWithPooledClient(deleteRoutingSmsAddress),
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
				ValidateDiagFunc: validateCountryCode,
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

func getAllRoutingSmsAddress(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	routingApi := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sdksmsaddressentitylisting, _, getErr := routingApi.GetRoutingSmsAddresses(pageSize, pageNum)
		if getErr != nil {
			return nil, diag.Errorf("Error requesting page of Routing Sms Address: %s", getErr)
		}

		if sdksmsaddressentitylisting.Entities == nil || len(*sdksmsaddressentitylisting.Entities) == 0 {
			break
		}

		for _, entity := range *sdksmsaddressentitylisting.Entities {
			resources[*entity.Id] = &ResourceMeta{Name: *entity.Name}
		}
	}

	return resources, nil
}

func routingSmsAddressExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllRoutingSmsAddress),
	}
}

func createRoutingSmsAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	street := d.Get("street").(string)
	city := d.Get("city").(string)
	region := d.Get("region").(string)
	postalCode := d.Get("postal_code").(string)
	countryCode := d.Get("country_code").(string)
	autoCorrectAddress := d.Get("auto_correct_address").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	sdksmsaddressprovision := platformclientv2.Smsaddressprovision{
		AutoCorrectAddress: &autoCorrectAddress,
	}

	if name != "" {
		sdksmsaddressprovision.Name = &name
	}
	if street != "" {
		sdksmsaddressprovision.Street = &street
	}
	if city != "" {
		sdksmsaddressprovision.City = &city
	}
	if region != "" {
		sdksmsaddressprovision.Region = &region
	}
	if postalCode != "" {
		sdksmsaddressprovision.PostalCode = &postalCode
	}
	if countryCode != "" {
		sdksmsaddressprovision.CountryCode = &countryCode
	}

	log.Printf("Creating Routing Sms Address %s", name)
	routingSmsAddress, _, err := routingApi.PostRoutingSmsAddresses(sdksmsaddressprovision)
	if err != nil {
		return diag.Errorf("Failed to create Routing Sms Addresse %s: %s", name, err)
	}

	d.SetId(*routingSmsAddress.Id)

	log.Printf("Created Routing Sms Address %s %s", name, *routingSmsAddress.Id)
	return readRoutingSmsAddress(ctx, d, meta)
}

func readRoutingSmsAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading Routing Sms Address %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *resource.RetryError {
		sdksmsaddress, resp, getErr := routingApi.GetRoutingSmsAddress(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read Routing Sms Address %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Routing Sms Address %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceRoutingSmsAddress())

		if sdksmsaddress.Name != nil {
			d.Set("name", *sdksmsaddress.Name)
		}
		if sdksmsaddress.Street != nil {
			d.Set("street", *sdksmsaddress.Street)
		}
		if sdksmsaddress.City != nil {
			d.Set("city", *sdksmsaddress.City)
		}
		if sdksmsaddress.Region != nil {
			d.Set("region", *sdksmsaddress.Region)
		}
		if sdksmsaddress.PostalCode != nil {
			d.Set("postal_code", *sdksmsaddress.PostalCode)
		}
		if sdksmsaddress.CountryCode != nil {
			d.Set("country_code", *sdksmsaddress.CountryCode)
		}

		log.Printf("Read Routing Sms Address %s %s", d.Id(), *sdksmsaddress.Name)
		return cc.CheckState()
	})
}

func deleteRoutingSmsAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	// AD-123 is the ID for a default address returned to all test orgs, it can't be deleted
	if d.Id() == "AD-123" {
		return nil
	}

	diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Routing Sms Address")
		resp, err := routingApi.DeleteRoutingSmsAddress(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Routing Sms Address: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := routingApi.GetRoutingSmsAddress(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Routing Sms Address deleted
				log.Printf("Deleted Routing Sms Address %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Routing Sms Address %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Routing Sms Address %s still exists", d.Id()))
	})
}
