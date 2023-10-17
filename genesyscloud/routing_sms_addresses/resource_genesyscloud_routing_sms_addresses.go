package genesyscloud

import (
	"context"
	"fmt"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

const resourceName = "genesyscloud_routing_sms_address"

func getAllRoutingSmsAddress(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingSmsAddressProxy(clientConfig)

	allSmsAddresses, err := proxy.getAllSmsAddresses(ctx)
	if err != nil {
		return nil, diag.Errorf("failed to get sms addresses: %v", err)
	}

	for _, entity := range *allSmsAddresses {
		var name string
		if entity.Name != nil {
			name = *entity.Name
		} else {
			name = *entity.Id
		}
		resources[*entity.Id] = &resourceExporter.ResourceMeta{Name: name}
	}

	return resources, nil
}

func createRoutingSmsAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	street := d.Get("street").(string)
	city := d.Get("city").(string)
	region := d.Get("region").(string)
	postalCode := d.Get("postal_code").(string)
	countryCode := d.Get("country_code").(string)
	autoCorrectAddress := d.Get("auto_correct_address").(bool)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getRoutingSmsAddressProxy(sdkConfig)

	sdkSmsAddressProvision := platformclientv2.Smsaddressprovision{
		AutoCorrectAddress: &autoCorrectAddress,
	}

	if name != "" {
		sdkSmsAddressProvision.Name = &name
	}
	if street != "" {
		sdkSmsAddressProvision.Street = &street
	}
	if city != "" {
		sdkSmsAddressProvision.City = &city
	}
	if region != "" {
		sdkSmsAddressProvision.Region = &region
	}
	if postalCode != "" {
		sdkSmsAddressProvision.PostalCode = &postalCode
	}
	if countryCode != "" {
		sdkSmsAddressProvision.CountryCode = &countryCode
	}

	log.Printf("Creating Routing Sms Address %s", name)
	routingSmsAddress, _, err := proxy.createSmsAddress(sdkSmsAddressProvision)
	if err != nil {
		return diag.Errorf("Failed to create Routing Sms Addresse %s: %s", name, err)
	}

	d.SetId(*routingSmsAddress.Id)

	log.Printf("Created Routing Sms Address %s %s", name, *routingSmsAddress.Id)
	return readRoutingSmsAddress(ctx, d, meta)
}

func readRoutingSmsAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getRoutingSmsAddressProxy(sdkConfig)

	log.Printf("Reading Routing Sms Address %s", d.Id())
	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkSmsAddress, resp, getErr := proxy.getSmsAddressById(d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Routing Sms Address %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Routing Sms Address %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingSmsAddress())

		resourcedata.SetNillableValue(d, "name", sdkSmsAddress.Name)
		resourcedata.SetNillableValue(d, "street", sdkSmsAddress.Street)
		resourcedata.SetNillableValue(d, "city", sdkSmsAddress.City)
		resourcedata.SetNillableValue(d, "region", sdkSmsAddress.Region)
		resourcedata.SetNillableValue(d, "postal_code", sdkSmsAddress.PostalCode)
		resourcedata.SetNillableValue(d, "country_code", sdkSmsAddress.CountryCode)

		log.Printf("Read Routing Sms Address %s", d.Id())
		return cc.CheckState()
	})
}

func deleteRoutingSmsAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getRoutingSmsAddressProxy(sdkConfig)

	// AD-123 is the ID for a default address returned to all test orgs, it can't be deleted
	if d.Id() == "AD-123" {
		return nil
	}

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Routing Sms Address")
		resp, err := proxy.deleteSmsAddress(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Routing Sms Address: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getSmsAddressById(d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Routing Sms Address deleted
				log.Printf("Deleted Routing Sms Address %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Routing Sms Address %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Routing Sms Address %s still exists", d.Id()))
	})
}
