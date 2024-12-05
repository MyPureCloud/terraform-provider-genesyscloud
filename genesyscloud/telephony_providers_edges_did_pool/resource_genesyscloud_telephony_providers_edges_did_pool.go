package telephony_providers_edges_did_pool

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

// getAllDidPools retrieves all DID pools and is used for the exporter
func getAllDidPools(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getTelephonyDidPoolProxy(clientConfig)

	didPools, resp, err := proxy.getAllTelephonyDidPools(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get did pools error: %s", err), resp)
	}

	for _, didPool := range *didPools {
		if didPool.State != nil && *didPool.State != "deleted" {
			resources[*didPool.Id] = &resourceExporter.ResourceMeta{BlockLabel: *didPool.StartPhoneNumber}
		}
	}
	return resources, nil
}

// createDidPool is used by the resource to create a Genesys Cloud DID pool
func createDidPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startPhoneNumber := d.Get("start_phone_number").(string)
	endPhoneNumber := d.Get("end_phone_number").(string)
	description := d.Get("description").(string)
	comments := d.Get("comments").(string)
	poolProvider := d.Get("pool_provider").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTelephonyDidPoolProxy(sdkConfig)

	didPool := &platformclientv2.Didpool{
		StartPhoneNumber: &startPhoneNumber,
		EndPhoneNumber:   &endPhoneNumber,
		Description:      &description,
		Comments:         &comments,
		Provider:         &poolProvider,
	}
	log.Printf("Creating DID pool %s", startPhoneNumber)
	createdDidPool, resp, err := proxy.createTelephonyDidPool(ctx, didPool)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create DID pool %s error: %s", startPhoneNumber, err), resp)
	}

	d.SetId(*createdDidPool.Id)

	log.Printf("Created DID pool %s %s", startPhoneNumber, *createdDidPool.Id)
	return readDidPool(ctx, d, meta)
}

// readDidPool is used by the resource to read a Genesys Cloud DID pool
func readDidPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTelephonyDidPoolProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTelephonyDidPool(), constants.ConsistencyChecks(), ResourceType)
	utilE164 := util.NewUtilE164Service()

	log.Printf("Reading DID pool %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		didPool, resp, getErr := proxy.getTelephonyDidPoolById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read DID pool %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read DID pool %s | error: %s", d.Id(), getErr), resp))
		}

		if didPool.State != nil && *didPool.State == "deleted" {
			d.SetId("")
			return nil
		}

		_ = d.Set("start_phone_number", utilE164.FormatAsCalculatedE164Number(*didPool.StartPhoneNumber))
		_ = d.Set("end_phone_number", utilE164.FormatAsCalculatedE164Number(*didPool.EndPhoneNumber))

		resourcedata.SetNillableValue(d, "description", didPool.Description)
		resourcedata.SetNillableValue(d, "comments", didPool.Comments)
		resourcedata.SetNillableValue(d, "pool_provider", didPool.Provider)

		log.Printf("Read DID pool %s %s", d.Id(), *didPool.StartPhoneNumber)
		return cc.CheckState(d)
	})
}

// updateDidPool is used by the resource to update a Genesys Cloud DID pool
func updateDidPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startPhoneNumber := d.Get("start_phone_number").(string)
	endPhoneNumber := d.Get("end_phone_number").(string)
	description := d.Get("description").(string)
	comments := d.Get("comments").(string)
	poolProvider := d.Get("pool_provider").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTelephonyDidPoolProxy(sdkConfig)

	didPoolBody := &platformclientv2.Didpool{
		StartPhoneNumber: &startPhoneNumber,
		EndPhoneNumber:   &endPhoneNumber,
		Description:      &description,
		Comments:         &comments,
		Provider:         &poolProvider,
	}

	log.Printf("Updating DID pool %s", d.Id())
	if _, resp, err := proxy.updateTelephonyDidPool(ctx, d.Id(), didPoolBody); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update DID pool %s error: %s", startPhoneNumber, err), resp)
	}

	log.Printf("Updated DID pool %s", d.Id())
	return readDidPool(ctx, d, meta)
}

// deleteDidPool is used by the resource to delete a Genesys Cloud DID pool
func deleteDidPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startPhoneNumber := d.Get("start_phone_number").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTelephonyDidPoolProxy(sdkConfig)

	// DEVTOOLING-317: Unable to delete DID pool with a number assigned, retrying on HTTP 409
	diagErr := util.RetryWhen(util.IsStatus409, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting DID pool with starting number %s", startPhoneNumber)
		resp, err := proxy.deleteTelephonyDidPool(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete DID pool %s error: %s", startPhoneNumber, err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		didPool, resp, err := proxy.getTelephonyDidPoolById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// DID pool deleted
				log.Printf("Deleted DID pool %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting DID pool %s | error: %s", d.Id(), err), resp))
		}

		if didPool.State != nil && *didPool.State == "deleted" {
			// DID pool deleted
			log.Printf("Deleted DID pool %s", d.Id())
			return nil
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("DID pool %s still exists", d.Id()), resp))
	})
}
