package telephony_providers_edges_did_pool

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

// getAllDidPools retrieves all DID pools and is used for the exporter
func getAllDidPools(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getTelephonyDidPoolProxy(clientConfig)

	didPools, err := proxy.getAllTelephonyDidPools(ctx)
	if err != nil {
		return nil, diag.Errorf("failed to read did pools: %v", err)
	}

	for _, didPool := range *didPools {
		if didPool.State != nil && *didPool.State != "deleted" {
			resources[*didPool.Id] = &resourceExporter.ResourceMeta{Name: *didPool.StartPhoneNumber}
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

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTelephonyDidPoolProxy(sdkConfig)

	didPool := &platformclientv2.Didpool{
		StartPhoneNumber: &startPhoneNumber,
		EndPhoneNumber:   &endPhoneNumber,
		Description:      &description,
		Comments:         &comments,
		Provider:         &poolProvider,
	}
	log.Printf("Creating DID pool %s", startPhoneNumber)
	createdDidPool, err := proxy.createTelephonyDidPool(ctx, didPool)
	if err != nil {
		return diag.Errorf("Failed to create DID pool %s: %s", startPhoneNumber, err)
	}

	d.SetId(*createdDidPool.Id)

	log.Printf("Created DID pool %s %s", startPhoneNumber, *createdDidPool.Id)
	return readDidPool(ctx, d, meta)
}

// readDidPool is used by the resource to read a Genesys Cloud DID pool
func readDidPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTelephonyDidPoolProxy(sdkConfig)

	log.Printf("Reading DID pool %s", d.Id())
	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		didPool, respCode, getErr := proxy.getTelephonyDidPoolById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read DID pool %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read DID pool %s: %s", d.Id(), getErr))
		}

		if didPool.State != nil && *didPool.State == "deleted" {
			d.SetId("")
			return nil
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTelephonyDidPool())
		_ = d.Set("start_phone_number", *didPool.StartPhoneNumber)
		_ = d.Set("end_phone_number", *didPool.EndPhoneNumber)

		resourcedata.SetNillableValue(d, "description", didPool.Description)
		resourcedata.SetNillableValue(d, "comments", didPool.Comments)
		resourcedata.SetNillableValue(d, "pool_provider", didPool.Provider)

		log.Printf("Read DID pool %s %s", d.Id(), *didPool.StartPhoneNumber)
		return cc.CheckState()
	})
}

// updateDidPool is used by the resource to update a Genesys Cloud DID pool
func updateDidPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startPhoneNumber := d.Get("start_phone_number").(string)
	endPhoneNumber := d.Get("end_phone_number").(string)
	description := d.Get("description").(string)
	comments := d.Get("comments").(string)
	poolProvider := d.Get("pool_provider").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTelephonyDidPoolProxy(sdkConfig)

	didPoolBody := &platformclientv2.Didpool{
		StartPhoneNumber: &startPhoneNumber,
		EndPhoneNumber:   &endPhoneNumber,
		Description:      &description,
		Comments:         &comments,
		Provider:         &poolProvider,
	}

	log.Printf("Updating DID pool %s", d.Id())
	if _, err := proxy.updateTelephonyDidPool(ctx, d.Id(), didPoolBody); err != nil {
		return diag.Errorf("Error updating DID pool %s: %s", startPhoneNumber, err)
	}

	log.Printf("Updated DID pool %s", d.Id())
	return readDidPool(ctx, d, meta)
}

// deleteDidPool is used by the resource to delete a Genesys Cloud DID pool
func deleteDidPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startPhoneNumber := d.Get("start_phone_number").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTelephonyDidPoolProxy(sdkConfig)

	log.Printf("Deleting DID pool with starting number %s", startPhoneNumber)
	if err := proxy.deleteTelephonyDidPool(ctx, d.Id()); err != nil {
		return diag.Errorf("Failed to delete DID pool with starting number %s: %s", startPhoneNumber, err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		didPool, respCode, err := proxy.getTelephonyDidPoolById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				// DID pool deleted
				log.Printf("Deleted DID pool %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting DID pool %s: %s", d.Id(), err))
		}

		if didPool.State != nil && *didPool.State == "deleted" {
			// DID pool deleted
			log.Printf("Deleted DID pool %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("DID pool %s still exists", d.Id()))
	})
}
