package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

func getAllExtensionPools(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	telephonyAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		extensionPools, _, getErr := telephonyAPI.GetTelephonyProvidersEdgesExtensionpools(pageSize, pageNum, "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of Extension pools: %v", getErr)
		}

		if extensionPools.Entities == nil || len(*extensionPools.Entities) == 0 {
			break
		}

		for _, extensionPool := range *extensionPools.Entities {
			if extensionPool.State != nil && *extensionPool.State != "deleted" {
				resources[*extensionPool.Id] = &ResourceMeta{Name: *extensionPool.StartNumber}
			}
		}
	}

	return resources, nil
}

func telephonyExtensionPoolExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllExtensionPools),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceTelephonyExtensionPool() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Extension Pool",

		CreateContext: createWithPooledClient(createExtensionPool),
		ReadContext:   readWithPooledClient(readExtensionPool),
		UpdateContext: updateWithPooledClient(updateExtensionPool),
		DeleteContext: deleteWithPooledClient(deleteExtensionPool),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"start_number": {
				Description:      "Starting phone number of the Extension Pool range. Changing the start_number attribute will cause the extension object to be dropped and recreated with a new ID.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePhoneNumber,
			},
			"end_number": {
				Description:      "Ending phone number of the Extension Pool range. Changing the end_number attribute will cause the extension object to be dropped and recreated with a new ID.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePhoneNumber,
			},
			"description": {
				Description: "Extension Pool description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func createExtensionPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startNumber := d.Get("start_number").(string)
	endNumber := d.Get("end_number").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Creating Extension pool %s", startNumber)
	extensionPool, _, err := telephonyApi.PostTelephonyProvidersEdgesExtensionpools(platformclientv2.Extensionpool{
		StartNumber: &startNumber,
		EndNumber:   &endNumber,
		Description: &description,
	})
	if err != nil {
		return diag.Errorf("Failed to create Extension pool %s: %s", startNumber, err)
	}

	d.SetId(*extensionPool.Id)

	log.Printf("Created Extension pool %s %s", startNumber, *extensionPool.Id)
	return readExtensionPool(ctx, d, meta)
}

func readExtensionPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading Extension pool %s", d.Id())
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		extensionPool, resp, getErr := telephonyApi.GetTelephonyProvidersEdgesExtensionpool(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read Extension pool %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Extension pool %s: %s", d.Id(), getErr))
		}

		if extensionPool.State != nil && *extensionPool.State == "deleted" {
			d.SetId("")
			return nil
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceTelephonyExtensionPool())
		d.Set("start_number", *extensionPool.StartNumber)
		d.Set("end_number", *extensionPool.EndNumber)

		if extensionPool.Description != nil {
			d.Set("description", *extensionPool.Description)
		} else {
			d.Set("description", nil)
		}

		log.Printf("Read Extension pool %s %s", d.Id(), *extensionPool.StartNumber)
		return cc.CheckState()
	})
}

func updateExtensionPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startNumber := d.Get("start_number").(string)
	endNumber := d.Get("end_number").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	extensionPoolBody := platformclientv2.Extensionpool{
		StartNumber: &startNumber,
		EndNumber:   &endNumber,
		Description: &description,
	}

	log.Printf("Updating Extension pool %s", d.Id())
	if _, _, err := telephonyApi.PutTelephonyProvidersEdgesExtensionpool(d.Id(), extensionPoolBody); err != nil {
		return diag.Errorf("Error updating Extension pool %s: %s", startNumber, err)
	}

	log.Printf("Updated Extension pool %s", d.Id())
	return readExtensionPool(ctx, d, meta)
}

func deleteExtensionPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	startNumber := d.Get("start_number").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting Extension pool with starting number %s", startNumber)
	if _, err := telephonyApi.DeleteTelephonyProvidersEdgesExtensionpool(d.Id()); err != nil {
		return diag.Errorf("Failed to delete Extension pool with starting number %s: %s", startNumber, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		extensionPool, resp, err := telephonyApi.GetTelephonyProvidersEdgesExtensionpool(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// Extension pool deleted
				log.Printf("Deleted Extension pool %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Extension pool %s: %s", d.Id(), err))
		}

		if extensionPool.State != nil && *extensionPool.State == "deleted" {
			// Extension pool deleted
			log.Printf("Deleted Extension pool %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Extension pool %s still exists", d.Id()))
	})
}
