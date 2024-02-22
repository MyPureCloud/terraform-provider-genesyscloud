package outbound_callanalysisresponseset

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	gcloud "terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"
)

/*
The resource_genesyscloud_outbound_callanalysisresponseset.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOutboundCallanalysisresponseset retrieves all of the outbound callanalysisresponseset via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOutboundCallanalysisresponsesets(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getOutboundCallanalysisresponsesetProxy(clientConfig)

	responseSets, getErr := proxy.getAllOutboundCallanalysisresponseset(ctx)
	if getErr != nil {
		return nil, diag.Errorf("Failed to get page of call analysis response set configs: %v", getErr)
	}
	for _, responseSet := range *responseSets {
		resources[*responseSet.Id] = &resourceExporter.ResourceMeta{Name: *responseSet.Name}
	}

	return resources, nil
}

// createOutboundCallanalysisresponseset is used by the outbound_callanalysisresponseset resource to create Genesys cloud outbound callanalysisresponseset
func createOutboundCallanalysisresponseset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCallanalysisresponsesetProxy(sdkConfig)

	responseSet := getResponseSetFromResourceData(d)

	log.Printf("Creating Outbound Call Analysis Response Set %s", *responseSet.Name)
	outboundCallanalysisresponseset, err := proxy.createOutboundCallanalysisresponseset(ctx, &responseSet)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Call Analysis Response Set %s: %s", *responseSet.Name, err)
	}

	d.SetId(*outboundCallanalysisresponseset.Id)

	log.Printf("Created Outbound Call Analysis Response Set %s %s", *responseSet.Name, *outboundCallanalysisresponseset.Id)
	return readOutboundCallanalysisresponseset(ctx, d, meta)
}

// readOutboundCallanalysisresponseset is used by the outbound_callanalysisresponseset resource to read an outbound callanalysisresponseset from genesys cloud
func readOutboundCallanalysisresponseset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCallanalysisresponsesetProxy(sdkConfig)

	log.Printf("Reading Outbound Call Analysis Response Set %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		responseSet, resp, getErr := proxy.getOutboundCallanalysisresponsesetById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read Outbound Call Analysis Response Set %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read Outbound Call Analysis Response Set %s: %s", d.Id(), getErr))
		}
		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundCallanalysisresponseset())

		resourcedata.SetNillableValue(d, "name", responseSet.Name)
		resourcedata.SetNillableValue(d, "beep_detection_enabled", responseSet.BeepDetectionEnabled)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "responses", responseSet.Responses, flattenSdkOutboundCallAnalysisResponseSetReaction)

		log.Printf("Read Outbound Call Analysis Response Set %s %s", d.Id(), *responseSet.Name)
		return cc.CheckState()
	})
}

// updateOutboundCallanalysisresponseset is used by the outbound_callanalysisresponseset resource to update an outbound callanalysisresponseset in Genesys Cloud
func updateOutboundCallanalysisresponseset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCallanalysisresponsesetProxy(sdkConfig)

	responseSet := getResponseSetFromResourceData(d)

	log.Printf("Updating Outbound Call Analysis Response Set %s %s", *responseSet.Name, d.Id())
	_, err := proxy.updateOutboundCallanalysisresponseset(ctx, d.Id(), &responseSet)
	if err != nil {
		return diag.Errorf("Failed to update Outbound Call Analysis Response Set %s: %s", *responseSet.Name, err)
	}

	log.Printf("Updated Outbound Call Analysis Response Set %s %s", *responseSet.Name, d.Id())
	return readOutboundCallanalysisresponseset(ctx, d, meta)
}

// deleteOutboundCallanalysisresponseset is used by the outbound_callanalysisresponseset resource to delete an outbound callanalysisresponseset from Genesys cloud
func deleteOutboundCallanalysisresponseset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCallanalysisresponsesetProxy(sdkConfig)

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Call Analysis Response Set")
		resp, err := proxy.deleteOutboundCallanalysisresponseset(ctx, d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Call Analysis Response Set: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getOutboundCallanalysisresponsesetById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404ByInt(resp) {
				// Outbound Call Analysis Response Set deleted
				log.Printf("Deleted Outbound Call Analysis Response Set %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting Outbound Call Analysis Response Set %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Call Analysis Response Set %s still exists", d.Id()))
	})
}
