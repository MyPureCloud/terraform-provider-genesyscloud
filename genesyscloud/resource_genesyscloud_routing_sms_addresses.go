package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
	"log"
	"time"
)

func resourceRoutingSmsAddress() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud routing sms addresse`,

		CreateContext: createWithPooledClient(createRoutingSmsAddresse),
		ReadContext:   readWithPooledClient(readRoutingSmsAddresse),

		DeleteContext: deleteWithPooledClient(deleteRoutingSmsAddresse),
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
				Description: `The ISO country code of this address`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
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

func getAllRoutingSmsAddresse(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	routingApi := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sdksmsaddressentitylisting, _, getErr := routingApi.GetRoutingSmsAddresses(pageSize, pageNum)
		if getErr != nil {
			return nil, diag.Errorf("Error requesting page of Routing Sms Addresse: %s", getErr)
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

func routingSmsAddresseExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllRoutingSmsAddresse),
		RefAttrs:         map[string]*RefAttrSettings{
			// No RefAttrs
		},
		// TODO RemoveIfMissing, AllowZeroValues and UnResolvableAttributes may need to be added
	}
}

func createRoutingSmsAddresse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	street := d.Get("street").(string)
	city := d.Get("city").(string)
	region := d.Get("region").(string)
	postalCode := d.Get("postal_code").(string)
	countryCode := d.Get("country_code").(string)
	autoCorrectAddress := d.Get("auto_correct_address").(bool)

	sdkConfig := meta.(*providerMeta).ClientConfig
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

	log.Printf("Creating Routing Sms Addresse %s", name)
	routingSmsAddresse, _, err := routingApi.PostRoutingSmsAddresses(sdksmsaddressprovision)
	if err != nil {
		return diag.Errorf("Failed to create Routing Sms Addresse %s: %s", name, err)
	}

	d.SetId(*routingSmsAddresse.Id)

	log.Printf("Created Routing Sms Addresse %s %s", name, *routingSmsAddresse.Id)
	return readRoutingSmsAddresse(ctx, d, meta)
}

func readRoutingSmsAddresse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading Routing Sms Addresse %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		sdksmsaddress, resp, getErr := routingApi.GetRoutingSmsAddress(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read Routing Sms Addresse %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Routing Sms Addresse %s: %s", d.Id(), getErr))
		}

		// cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceRoutingSmsAddresse())

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

		log.Printf("Read Routing Sms Addresse %s %s", d.Id(), *sdksmsaddress.Name)
		return nil // TODO calling cc.CheckState() can cause some difficult to understand errors in development. When ready for a PR, remove this line and uncomment the consistency_checker initialization and the the below one
		// return cc.CheckState()
	})
}

func deleteRoutingSmsAddresse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	routingApi := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	diagErr := retryWhen(isStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Routing Sms Addresse")
		resp, err := routingApi.DeleteRoutingSmsAddress(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Routing Sms Addresse: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := routingApi.GetRoutingSmsAddress(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// Routing Sms Addresse deleted
				log.Printf("Deleted Routing Sms Addresse %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Routing Sms Addresse %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Routing Sms Addresse %s still exists", d.Id()))
	})
}
