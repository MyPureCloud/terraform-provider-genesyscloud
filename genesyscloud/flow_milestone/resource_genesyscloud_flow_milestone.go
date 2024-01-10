package flow_milestone

import (
	"context"
	"fmt"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

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

	flowMilestones, err := proxy.getAllFlowMilestone(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get flow milestone: %v", err)
	}

	for _, flowMilestone := range *flowMilestones {
		resources[*flowMilestone.Id] = &resourceExporter.ResourceMeta{Name: *flowMilestone.Name}
	}

	return resources, nil
}

// createFlowMilestone is used by the flow_milestone resource to create Genesys cloud flow milestone
func createFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getFlowMilestoneProxy(sdkConfig)

	flowMilestone := getFlowMilestoneFromResourceData(d)

	log.Printf("Creating flow milestone %s", *flowMilestone.Name)
	flowMilestoneSdk, err := proxy.createFlowMilestone(ctx, &flowMilestone)
	if err != nil {
		return diag.Errorf("Failed to create flow milestone: %s", err)
	}

	d.SetId(*flowMilestoneSdk.Id)
	log.Printf("Created flow milestone %s", *flowMilestoneSdk.Id)
	return readFlowMilestone(ctx, d, meta)
}

// readFlowMilestone is used by the flow_milestone resource to read a flow milestone from genesys cloud
func readFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getFlowMilestoneProxy(sdkConfig)

	log.Printf("Reading flow milestone %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		flowMilestone, respCode, getErr := proxy.getFlowMilestoneById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read flow milestone %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read flow milestone %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceFlowMilestone())

		resourcedata.SetNillableValue(d, "name", flowMilestone.Name)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", flowMilestone.Division)
		resourcedata.SetNillableValue(d, "description", flowMilestone.Description)

		log.Printf("Read flow milestone %s %s", d.Id(), *flowMilestone.Name)
		return cc.CheckState()
	})
}

// updateFlowMilestone is used by the flow_milestone resource to update an flow milestone in Genesys Cloud
func updateFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getFlowMilestoneProxy(sdkConfig)

	flowMilestone := getFlowMilestoneFromResourceData(d)

	log.Printf("Updating flow milestone %s", *flowMilestone.Name)
	flowMilestoneSdk, err := proxy.updateFlowMilestone(ctx, d.Id(), &flowMilestone)
	if err != nil {
		return diag.Errorf("Failed to update flow milestone: %s", err)
	}

	log.Printf("Updated flow milestone %s", *flowMilestoneSdk.Id)
	return readFlowMilestone(ctx, d, meta)
}

// deleteFlowMilestone is used by the flow_milestone resource to delete a flow milestone from Genesys cloud
func deleteFlowMilestone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getFlowMilestoneProxy(sdkConfig)

	_, err := proxy.deleteFlowMilestone(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete flow milestone %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getFlowMilestoneById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted flow milestone %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting flow milestone %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("flow milestone %s still exists", d.Id()))
	})
}
