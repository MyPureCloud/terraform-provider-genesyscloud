package flow_loglevel

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_flow_loglevel.go contains all of the methods that perform the core logic for a resource.
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

	flowLogLevels, apiResponse, err := ep.getAllFlowLogLevels(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get flow log levels: %v", err), apiResponse)
	}

	for _, flowLogLevel := range *flowLogLevels {
		resources[*flowLogLevel.Id] = &resourceExporter.ResourceMeta{BlockLabel: *flowLogLevel.Id}
	}

	return resources, nil
}

// createFlowLogLevel is used by the flow_loglevel resource to create Genesyscloud flow_loglevel
func createFlowLogLevel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getFlowLogLevelProxy(sdkConfig)
	flowId := d.Get("flow_id").(string)
	log.Printf("Creating flow log level for flow  %s", flowId)

	flowLogLevelRequest := platformclientv2.Flowloglevelrequest{
		LogLevelCharacteristics: getFlowLogLevelFromResourceData(d),
	}

	flowLogLevel, apiResponse, err := ep.createFlowLogLevel(ctx, flowId, &flowLogLevelRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create flow log level: %s %s", err, d.Id()), apiResponse)
	}

	log.Printf("Sucessfully created flow log level for flow:  %s flowLogLevelId: %s", flowId, *flowLogLevel.Id)

	d.SetId(*flowLogLevel.Id)
	return readFlowLogLevel(ctx, d, meta)
}

// readFlowLogLevels is used by the flow_loglevel resource to read a flow log level from genesys cloud.
func readFlowLogLevel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getFlowLogLevelProxy(sdkConfig)
	flowId := d.Get("flow_id").(string)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceFlowLoglevel(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading readFlowLogLevel with flowId %s", flowId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		flowSettingsResponse, apiResponse, err := ep.getFlowLogLevelById(ctx, flowId)
		if err != nil {
			if util.IsStatus404ByInt(apiResponse.StatusCode) {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read flow log level %s | error: %s", flowId, err), apiResponse))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read flow log level %s | error: %s", flowId, err), apiResponse))
		}

		flowLogLevel := flowSettingsResponse.LogLevelCharacteristics

		resourcedata.SetNillableValue(d, "flow_log_level", flowLogLevel.Level)

		log.Printf("Read flow log level %s", flowId)
		return cc.CheckState(d)
	})
}

// updateFlowLogLevels is used by the flow_loglevel resource to update an flow log level in Genesys Cloud
func updateFlowLogLevel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getFlowLogLevelProxy(sdkConfig)
	flowId := d.Get("flow_id").(string)
	log.Printf("Updating flow log level for flow %s", flowId)

	flowLogLevelRequest := platformclientv2.Flowloglevelrequest{
		LogLevelCharacteristics: getFlowLogLevelFromResourceData(d),
	}

	updatedFlow, apiResponse, err := ep.updateFlowLogLevel(ctx, flowId, &flowLogLevelRequest)

	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update flow log level: %s %s", flowId, d.Id()), apiResponse)
	}

	log.Printf("Sucessfully updated flow log level for flow:  %s flowLogLevelId: %s", flowId, *updatedFlow.Id)

	return readFlowLogLevel(ctx, d, meta)
}

// deleteFlowLogLevels is used by the flow_loglevel resource to delete an flow log level from Genesys cloud.
func deleteFlowLogLevel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getFlowLogLevelProxy(sdkConfig)
	flowId := d.Get("flow_id").(string)
	log.Printf("Deleting flow log level for flow  %s", flowId)

	apiResponse, err := ep.deleteFlowLogLevelById(ctx, flowId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete flow log level %s: %s", flowId, d.Id()), apiResponse)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, apiResponse, err := ep.getFlowLogLevelById(ctx, flowId)

		if err == nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting flow log level %s %s | error: %s", flowId, d.Id(), err), apiResponse))
		}
		if util.IsStatus404ByInt(apiResponse.StatusCode) {
			log.Printf("Deleted flow log level %s", flowId)
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("flow log level %s still exists", flowId), apiResponse))
	})
}

// getFlowLogLevelFromResourceData maps data from schema ResourceData object to a platformclientv2.Flowloglevel
func getFlowLogLevelFromResourceData(d *schema.ResourceData) *platformclientv2.Flowloglevel {
	return &platformclientv2.Flowloglevel{
		Level: platformclientv2.String(d.Get("flow_log_level").(string)),
	}
}
