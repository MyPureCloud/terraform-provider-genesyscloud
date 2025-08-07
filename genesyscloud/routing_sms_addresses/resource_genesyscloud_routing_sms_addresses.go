package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
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
		var blockLabel string
		if entity.Name != nil {
			blockLabel = *entity.Name
		} else if entity.PostalCode != nil {
			blockLabel = *entity.PostalCode
		} else {
			blockLabel = *entity.Id
		}
		resources[*entity.Id] = &resourceExporter.ResourceMeta{BlockLabel: blockLabel}
	}
	return resources, nil
}

func createRoutingSmsAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSmsAddressProxy(sdkConfig)

	sdkSmsAddressProvision := platformclientv2.Smsaddressprovision{
		Name:               &name, // is optional but must be an empty string. Null here will return a 400 error
		Street:             platformclientv2.String(d.Get("street").(string)),
		City:               platformclientv2.String(d.Get("city").(string)),
		Region:             platformclientv2.String(d.Get("region").(string)),
		PostalCode:         platformclientv2.String(d.Get("postal_code").(string)),
		CountryCode:        platformclientv2.String(d.Get("country_code").(string)),
		AutoCorrectAddress: platformclientv2.Bool(d.Get("auto_correct_address").(bool)),
	}

	log.Printf("Creating Routing Sms Address %s", name)
	routingSmsAddress, resp, err := proxy.createSmsAddress(sdkSmsAddressProvision)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create sms address with name '%s'. Error: %s", name, err.Error()), resp)
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
			diagErr := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Routing Sms Address %s | error: %s", d.Id(), getErr), resp)
			if util.IsStatus404(resp) {
				return retry.RetryableError(diagErr)
			}
			return retry.NonRetryableError(diagErr)
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
