package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

const ResourceType = "genesyscloud_routing_sms_address"

func getAllRoutingSmsAddress(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingSmsAddressProxy(clientConfig)

	allSmsAddresses, resp, err := proxy.getAllSmsAddresses(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get sms addresses error: %s", err), resp)
	}

	for _, entity := range *allSmsAddresses {
		var name string
		if entity.Name != nil {
			name = *entity.Name
		} else {
			name = *entity.Id
		}
		resources[*entity.Id] = &resourceExporter.ResourceMeta{BlockLabel: name}
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

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
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
	routingSmsAddress, resp, err := proxy.createSmsAddress(sdkSmsAddressProvision)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create sms address %s error: %s", name, err), resp)
	}

	d.SetId(*routingSmsAddress.Id)

	log.Printf("Created Routing Sms Address %s %s", name, *routingSmsAddress.Id)
	return readRoutingSmsAddress(ctx, d, meta)
}

func readRoutingSmsAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSmsAddressProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingSmsAddress(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Routing Sms Address %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkSmsAddress, resp, getErr := proxy.getSmsAddressById(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Routing Sms Address %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Routing Sms Address %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", sdkSmsAddress.Name)
		resourcedata.SetNillableValue(d, "street", sdkSmsAddress.Street)
		resourcedata.SetNillableValue(d, "city", sdkSmsAddress.City)
		resourcedata.SetNillableValue(d, "region", sdkSmsAddress.Region)
		resourcedata.SetNillableValue(d, "postal_code", sdkSmsAddress.PostalCode)
		resourcedata.SetNillableValue(d, "country_code", sdkSmsAddress.CountryCode)

		log.Printf("Read Routing Sms Address %s", d.Id())
		return cc.CheckState(d)
	})
}

func deleteRoutingSmsAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSmsAddressProxy(sdkConfig)

	// AD-123 is the ID for a default address returned to all test orgs, it can't be deleted
	if d.Id() == "AD-123" {
		return nil
	}

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Routing Sms Address")
		resp, err := proxy.deleteSmsAddress(d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete routing sms address %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getSmsAddressById(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Routing Sms Address deleted
				log.Printf("Deleted Routing Sms Address %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Routing Sms Address %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Routing Sms Address %s still exists", d.Id()), resp))
	})
}
