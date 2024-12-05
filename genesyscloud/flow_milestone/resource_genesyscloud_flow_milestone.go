package flow_milestone

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_flow_milestone.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthFlowMilestone retrieves all of the flow milestone via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthFlowMilestones(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newFlowMilestoneProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)
	flowMilestones, resp, err := proxy.getAllFlowMilestone(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get flow milestone error: %s", err), resp)
	}

	for _, flowMilestone := range *flowMilestones {
		resources[*flowMilestone.Id] = &resourceExporter.ResourceMeta{BlockLabel: *flowMilestone.Name}
	}
	return resources, nil
}

// createFlowMilestone is used by the flow_milestone resource to create Genesys cloud flow milestone
func createFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getFlowMilestoneProxy(sdkConfig)

	flowMilestone := getFlowMilestoneFromResourceData(d)

	log.Printf("Creating flow milestone %s", *flowMilestone.Name)
	flowMilestoneSdk, resp, err := proxy.createFlowMilestone(ctx, &flowMilestone)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create flow milestone %s error: %s", *flowMilestone.Name, err), resp)
	}

	d.SetId(*flowMilestoneSdk.Id)
	log.Printf("Created flow milestone %s", *flowMilestoneSdk.Id)
	return readFlowMilestone(ctx, d, meta)
}

// readFlowMilestone is used by the flow_milestone resource to read a flow milestone from genesys cloud
func readFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getFlowMilestoneProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceFlowMilestone(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading flow milestone %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		flowMilestone, resp, getErr := proxy.getFlowMilestoneById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read flow milestone %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read flow milestone %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", flowMilestone.Name)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", flowMilestone.Division)
		resourcedata.SetNillableValue(d, "description", flowMilestone.Description)

		log.Printf("Read flow milestone %s %s", d.Id(), *flowMilestone.Name)
		return cc.CheckState(d)
	})
}

// updateFlowMilestone is used by the flow_milestone resource to update an flow milestone in Genesys Cloud
func updateFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getFlowMilestoneProxy(sdkConfig)

	flowMilestone := getFlowMilestoneFromResourceData(d)

	log.Printf("Updating flow milestone %s", *flowMilestone.Name)
	flowMilestoneSdk, resp, err := proxy.updateFlowMilestone(ctx, d.Id(), &flowMilestone)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update flow milestone %s error: %s", *flowMilestone.Name, err), resp)
	}

	log.Printf("Updated flow milestone %s", *flowMilestoneSdk.Id)
	return readFlowMilestone(ctx, d, meta)
}

// deleteFlowMilestone is used by the flow_milestone resource to delete a flow milestone from Genesys cloud
func deleteFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getFlowMilestoneProxy(sdkConfig)

	resp, err := proxy.deleteFlowMilestone(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete flow milestone %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getFlowMilestoneById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted flow milestone %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting flow milestone %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("flow milestone %s still exists", d.Id()), resp))
	})
}
