package outbound_callabletimeset

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

/*
The resource_genesyscloud_outbound_callabletimeset.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOutboundCallabletimeset retrieves all of the Outbound Callabletimeset via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOutboundCallabletimesets(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getOutboundCallabletimesetProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	callableTimeSets, err := proxy.getAllOutboundCallabletimeset(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get ruleset: %v", err)
	}

	for _, callableTimeSet := range *callableTimeSets {
		log.Printf("Dealing with Outbound Callabletimeset id : %s", *callableTimeSet.Id)
		resources[*callableTimeSet.Id] = &resourceExporter.ResourceMeta{Name: *callableTimeSet.Id}
	}

	return resources, nil
}

// createOutboundCallabletimeset is used by the outbound_callabletimeset resource to create Genesys cloud Outbound Callabletimeset
func createOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newOutboundCallabletimesetProxy(sdkConfig)

	outboundCallabletimeset := getOutboundCallabletimesetFromResourceData(d)
	callableTimeSet, err := proxy.createOutboundCallabletimeset(ctx, &outboundCallabletimeset)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Callabletimeset: %s", err)
	}

	d.SetId(*callableTimeSet.Id)
	log.Printf("Created Outbound Callabletimeset %s", *callableTimeSet.Id)
	return readOutboundCallabletimeset(ctx, d, meta)
}

// readOutboundCallabletimeset is used by the outbound_callabletimeset resource to read an Outbound Callabletimeset from genesys cloud
func readOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newOutboundCallabletimesetProxy(sdkConfig)

	log.Printf("Reading Outbound Callabletimeset %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		callableTimeSet, respCode, getErr := proxy.getOutboundCallabletimesetById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read Outbound Callabletimeset %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Outbound Callabletimeset %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundCallabletimeset())

		resourcedata.SetNillableValue(d, "name", callableTimeSet.Name)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "callable_times", callableTimeSet.CallableTimes, flattenCallableTimes)

		log.Printf("Read Outbound Callabletimeset %s %s", d.Id(), *callableTimeSet.Name)
		return cc.CheckState()
	})
}

// updateOutboundCallabletimeset is used by the outbound_callabletimeset resource to update an Outbound Callabletimeset in Genesys Cloud
func updateOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newOutboundCallabletimesetProxy(sdkConfig)
	outboundCallabletimeset := getOutboundCallabletimesetFromResourceData(d)

	callableTimeSet, err := proxy.updateOutboundCallabletimeset(ctx, d.Id(), &outboundCallabletimeset)
	if err != nil {
		return diag.Errorf("Failed to update Outbound Callabletimeset: %s", err)
	}

	log.Printf("Updated Outbound Callabletimeset %s", *callableTimeSet.Id)
	return readOutboundCallabletimeset(ctx, d, meta)
}

// deleteOutboundCallabletimeset is used by the outbound_callabletimeset resource to delete an Outbound Callabletimeset from Genesys cloud
func deleteOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newOutboundCallabletimesetProxy(sdkConfig)

	_, err := proxy.deleteOutboundCallabletimeset(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete Outbound Callabletimeset %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getOutboundCallabletimesetById(ctx, d.Id())

		if err == nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted Outbound Callabletimeset %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Outbound Callabletimeset %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Callabletimeset %s still exists", d.Id()))
	})
}
