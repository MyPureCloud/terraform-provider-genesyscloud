package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func getAllDidPools(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	telephonyAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		didPools, _, getErr := telephonyAPI.GetTelephonyProvidersEdgesDidpools(100, pageNum, "", nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of DID pools: %v", getErr)
		}

		if didPools.Entities == nil || len(*didPools.Entities) == 0 {
			break
		}

		for _, didPool := range *didPools.Entities {
			if *didPool.State != "deleted" {
				resources[*didPool.Id] = &ResourceMeta{Name: *didPool.StartPhoneNumber}
			}
		}
	}

	return resources, nil
}

func telephonyDidPoolExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllDidPools),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceTelephonyDidPool() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud DID Pool",

		CreateContext: createWithPooledClient(createDidPool),
		ReadContext:   readWithPooledClient(readDidPool),
		UpdateContext: updateWithPooledClient(updateDidPool),
		DeleteContext: deleteWithPooledClient(deleteDidPool),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"start_phone_number": {
				Description:      "Starting phone number of the DID Pool range.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePhoneNumber,
			},
			"end_phone_number": {
				Description:      "Ending phone number of the DID Pool range.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePhoneNumber,
			},
			"description": {
				Description: "DID Pool description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"comments": {
				Description: "Comments for the DID Pool.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"pool_provider": {
				Description:  "Provider (PURE_CLOUD | PURE_CLOUD_VOICE).",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"PURE_CLOUD", "PURE_CLOUD_VOICE"}, false),
			},
		},
	}
}

func createDidPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startPhoneNumber := d.Get("start_phone_number").(string)
	endPhoneNumber := d.Get("end_phone_number").(string)
	description := d.Get("description").(string)
	comments := d.Get("comments").(string)
	poolProvider := d.Get("pool_provider").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Creating DID pool %s", startPhoneNumber)
	didPool, _, err := telephonyApi.PostTelephonyProvidersEdgesDidpools(platformclientv2.Didpool{
		StartPhoneNumber: &startPhoneNumber,
		EndPhoneNumber:   &endPhoneNumber,
		Description:      &description,
		Comments:         &comments,
		Provider:         &poolProvider,
	})
	if err != nil {
		return diag.Errorf("Failed to create DID pool %s: %s", startPhoneNumber, err)
	}

	d.SetId(*didPool.Id)

	log.Printf("Created DID pool %s %s", startPhoneNumber, *didPool.Id)
	return readDidPool(ctx, d, meta)
}

func readDidPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading DID pool %s", d.Id())
	return withRetriesForRead(ctx, 30*time.Second, d, func() *resource.RetryError {
		didPool, resp, getErr := telephonyApi.GetTelephonyProvidersEdgesDidpool(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read DID pool %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read DID pool %s: %s", d.Id(), getErr))
		}

		if *didPool.State == "deleted" {
			d.SetId("")
			return nil
		}

		d.Set("start_phone_number", *didPool.StartPhoneNumber)
		d.Set("end_phone_number", *didPool.EndPhoneNumber)

		if didPool.Description != nil {
			d.Set("description", *didPool.Description)
		} else {
			d.Set("description", nil)
		}

		if didPool.Comments != nil {
			d.Set("comments", *didPool.Comments)
		} else {
			d.Set("comments", nil)
		}

		if didPool.Provider != nil {
			d.Set("pool_provider", *didPool.Provider)
		} else {
			d.Set("pool_provider", nil)
		}

		log.Printf("Read DID pool %s %s", d.Id(), *didPool.StartPhoneNumber)
		return nil
	})
}

func updateDidPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startPhoneNumber := d.Get("start_phone_number").(string)
	endPhoneNumber := d.Get("end_phone_number").(string)
	description := d.Get("description").(string)
	comments := d.Get("comments").(string)
	poolProvider := d.Get("pool_provider").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	didPoolBody := platformclientv2.Didpool{
		StartPhoneNumber: &startPhoneNumber,
		EndPhoneNumber:   &endPhoneNumber,
		Description:      &description,
		Comments:         &comments,
		Provider:         &poolProvider,
	}

	log.Printf("Updating DID pool %s", d.Id())
	if _, _, err := telephonyApi.PutTelephonyProvidersEdgesDidpool(d.Id(), didPoolBody); err != nil {
		return diag.Errorf("Error updating DID pool %s: %s", startPhoneNumber, err)
	}

	log.Printf("Updated DID pool %s", d.Id())
	time.Sleep(5 * time.Second)
	return readDidPool(ctx, d, meta)
}

func deleteDidPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startPhoneNumber := d.Get("start_phone_number").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting DID pool with starting number %s", startPhoneNumber)
	if _, err := telephonyApi.DeleteTelephonyProvidersEdgesDidpool(d.Id()); err != nil {
		return diag.Errorf("Failed to delete DID pool with starting number %s: %s", startPhoneNumber, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		didPool, resp, err := telephonyApi.GetTelephonyProvidersEdgesDidpool(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// DID pool deleted
				log.Printf("Deleted DID pool %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting DID pool %s: %s", d.Id(), err))
		}

		if *didPool.State == "deleted" {
			// DID pool deleted
			log.Printf("Deleted DID pool %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("DID pool %s still exists", d.Id()))
	})
}
