package outbound_callabletimeset

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

/*
The resource_genesyscloud_outbound_callabletimeset.go contains all of the methods that perform the core logic for a resource.
*/

// getAllOutboundCallableTimesets retrieves all of the Outbound Callable Timesets via Terraform in the genesys cloud and is used for the exporter
func getAllOutboundCallableTimesets(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getOutboundCallabletimesetProxy(clientConfig)

	callabletimesets, resp, getErr := proxy.getAllOutboundCallableTimeset(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of callable timeset configs error: %s", getErr), resp)
	}
	for _, callabletimesets := range *callabletimesets {
		resources[*callabletimesets.Id] = &resourceExporter.ResourceMeta{BlockLabel: *callabletimesets.Name}
	}
	return resources, nil
}

// createOutboundCallabletimeset is used by the outbound_callabletimeset resource to create Outbound Callable Timesets
func createOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCallabletimesetProxy(sdkConfig)

	callableTimeset := getOutboundCallableTimesetFromResourceData(d)

	log.Printf("Creating Outbound Callabletimeset %s", name)
	outboundCallabletimeset, resp, err := proxy.createOutboundCallabletimeset(ctx, &callableTimeset)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create Outbound Callabletimeset %s error: %s", *callableTimeset.Name, err), resp)
	}

	d.SetId(*outboundCallabletimeset.Id)
	log.Printf("Created Outbound Callabletimeset %s %s", name, *outboundCallabletimeset.Id)
	return readOutboundCallabletimeset(ctx, d, meta)
}

// updateOutboundCallabletimeset is used by the outbound_callabletimeset resource to update an Outbound Callable Timeset
func updateOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCallabletimesetProxy(sdkConfig)

	callableTimeset := getOutboundCallableTimesetFromResourceData(d)

	log.Printf("Updating Outbound Callabletimeset %s", d.Id())
	outboundCallabletimeset, resp, err := proxy.updateOutboundCallabletimeset(ctx, d.Id(), &callableTimeset)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Outbound Callabletimeset %s error: %s", *callableTimeset.Name, err), resp)
	}
	log.Printf("Updated Outbound Callabletimeset %s", *outboundCallabletimeset.Id)
	return readOutboundCallabletimeset(ctx, d, meta)
}

// readOutboundCallabletimeset is used by the outbound_callabletimeset resource to read an Outbound Callable Timeset from the genesys cloud
func readOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCallabletimesetProxy(sdkConfig)

	log.Printf("Reading Outbound Callabletimeset %s", d.Id())

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundCallabletimeset(), constants.ConsistencyChecks(), ResourceType)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		callableTimeset, resp, getErr := proxy.getOutboundCallabletimesetById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Outbound Callabletimeset %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Outbound Callabletimeset %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", callableTimeset.Name)
		if callableTimeset.CallableTimes != nil {
			// Remove the milliseconds added to start_time and stop_time by the API
			trimTime(callableTimeset.CallableTimes)
			d.Set("callable_times", flattenCallableTimes(*callableTimeset.CallableTimes))
		}
		log.Printf("Read Outbound Callabletimeset %s %s", d.Id(), *callableTimeset.Name)
		return cc.CheckState(d)
	})
}

// deleteOutboundCallabletimeset is used by the outbound_callabletimeset resource to delete an existing Outbound Callable Timeset from the genesys cloud
func deleteOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCallabletimesetProxy(sdkConfig)

	log.Printf("Deleting Outbound Callabletimeset")
	resp, err := proxy.deleteOutboundCallabletimeset(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Outbound Callabletimeset %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		resp, err := proxy.deleteOutboundCallabletimeset(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Outbound Callabletimeset deleted
				log.Printf("Deleted Outbound Callabletimeset %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Outbound Callabletimeset %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Outbound Callabletimeset %s still exists", d.Id()), resp))
	})
}
