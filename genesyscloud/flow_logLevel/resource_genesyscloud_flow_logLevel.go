package flow_logLevel

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

/*
The resource_genesyscloud_flow_logLevel.go contains all of the methods that perform the core logic for a resource.
In general a resource should have a approximately 5 methods in it:

1.  A getAll.... function that the CX as Code exporter will use during the process of exporting Genesys Cloud.
2.  A create.... function that the resource will use to create a Genesys Cloud object (e.g. genesycloud_flow_logLevel)
3.  A read.... function that looks up a single resource.
4.  An update... function that updates a single resource.
5.  A delete.... function that deletes a single resource.

Two things to note:

1.  All code in these methods should be focused on getting data in and out of Terraform.  All code that is used for interacting
    with a Genesys API should be encapsulated into a proxy class contained within the package.

2.  In general, to keep this file somewhat manageable, if you find yourself with a number of helper functions move them to a
utils function in the package.  This will keep the code manageable and easy to work through.
*/
// getAllFlowLogLevels retrieves all of the flow log levels via Terraform in the Genesys Cloud and is used for the exporter
func getAllFlowLogLevels(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	ep := getFlowLogLevelProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	flowLogLevels, err := ep.getAllFlowLogLevels(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get flow log levels: %v", err)
	}

	for _, flowLogLevel := range *flowLogLevels {
		log.Printf("Dealing with flow log level id : %s", *flowLogLevel.Id)
		resources[*flowLogLevel.Id] = &resourceExporter.ResourceMeta{Name: *flowLogLevel.Id}
	}

	return resources, nil
}

// createFlowLogLevel is used by the flow_logLevel resource to create Genesyscloud flow_logLevel
func createFlowLogLevel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getFlowLogLevelProxy(sdkConfig)
	flowId := d.Get("flow_id").(string)
	flowLogLevelRequest := getFlowLogLevelSettingsRequestFromResourceData(d)

	flowLogLevel, err := ep.createFlowLogLevel(ctx, flowId, &flowLogLevelRequest)
	if err != nil {
		return diag.Errorf("Failed to create flow log level: %s", err)
	}

	d.SetId(*flowLogLevel.Id)
	return readFlowLogLevel(ctx, d, meta)
}

// readFlowLogLevels is used by the flow_logLevel resource to read a flow log level from genesys cloud.
func readFlowLogLevel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getFlowLogLevelProxy(sdkConfig)
	flowId := d.Get("flow_id").(string)

	log.Printf("Reading readFlowLogLevel with flowId %s", flowId)
	if flowId == "" {
		log.Printf("flow log level with blank flowId %s", flowId)
		return diag.Errorf("flowId not found")
	}
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		flowSettingsResponse, respCode, err := ep.getFlowLogLevelById(ctx, flowId)
		if err != nil {
			log.Print(err)
			if util.IsStatus404ByInt(respCode) {
				return retry.NonRetryableError(fmt.Errorf("Failed to read flow log level %s: %s", flowId, err))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read flow log level %s: %s", flowId, err))
		}

		flowLogLevel := flowSettingsResponse.LogLevelCharacteristics

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceFlowLoglevel())
		resourcedata.SetNillableValue(d, "flow_log_level", flowLogLevel.Level)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "flow_characteristics", flowLogLevel.Characteristics, flattenFlowCharacteristics)

		log.Printf("Read flow log level %s", flowId)
		checkState := cc.CheckState()
		log.Printf("checkState result =  %v", checkState)
		return checkState
	})
}

// updateFlowLogLevels is used by the flow_logLevel resource to update an flow log level in Genesys Cloud
func updateFlowLogLevel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getFlowLogLevelProxy(sdkConfig)
	flowId := d.Get("flow_id").(string)

	_, _, err := ep.getFlowLogLevelById(ctx, flowId)
	flowLogLevelRequest := getFlowLogLevelSettingsRequestFromResourceData(d)
	_, err = ep.updateFlowLogLevel(ctx, flowId, &flowLogLevelRequest)

	if err != nil {
		return diag.Errorf("Failed to update flow log level: %s", err)
	}

	log.Printf("Updated flow log level")

	return readFlowLogLevel(ctx, d, meta)
}

// deleteFlowLogLevels is used by the flow_logLevel resource to delete an flow log level from Genesys cloud.
func deleteFlowLogLevel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getFlowLogLevelProxy(sdkConfig)
	flowId := d.Get("flow_id").(string)

	_, err := ep.deleteFlowLogLevelById(ctx, flowId)
	if err != nil {
		return diag.Errorf("Failed to delete flow log level %s: %s", flowId, err)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := ep.getFlowLogLevelById(ctx, flowId)

		if err == nil {
			return retry.NonRetryableError(fmt.Errorf("Error deleting flow log level %s: %s", flowId, err))
		}
		if util.IsStatus404ByInt(respCode) {
			// Success  : External contact deleted
			log.Printf("Deleted flow log level %s", flowId)
			return nil
		}

		return retry.RetryableError(fmt.Errorf("External contact %s still exists", flowId))
	})
}
