package journey_segment

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

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func getAllJourneySegments(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {

	proxy := getJourneySegmentProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	segments, proxyResponse, getErr := proxy.getAllJourneySegments(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of journey segments: %s", getErr), proxyResponse)
	}

	for _, segment := range *segments {
		resources[*segment.Id] = &resourceExporter.ResourceMeta{BlockLabel: *segment.DisplayName}
	}

	return resources, nil
}

func createJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneySegmentProxy(sdkConfig)
	segment := buildSdkJourneySegment(d)

	log.Printf("Creating journey segment %s", *segment.DisplayName)
	segmentResponse, proxyResponse, err := proxy.createJourneySegment(ctx, segment)
	if err != nil {
		input, _ := util.InterfaceToJson(*segment)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to create journey segment %s: %s\n(input: %+v)", *segmentResponse.DisplayName, err, input), proxyResponse)
	}

	d.SetId(*segmentResponse.Id)

	log.Printf("Created journey segment %s %s", *segmentResponse.DisplayName, *segmentResponse.Id)
	return readJourneySegment(ctx, d, meta)
}

func readJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneySegmentProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceJourneySegment(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading journey segment %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		segmentResponse, proxyResponse, getErr := proxy.getJourneySegmentById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey segment %s | error: %s", d.Id(), getErr), proxyResponse))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey segment %s | error: %s", d.Id(), getErr), proxyResponse))
		}

		flattenJourneySegment(d, segmentResponse)

		log.Printf("Read journey segment %s %s", d.Id(), *segmentResponse.DisplayName)
		return cc.CheckState(d)
	})
}

func updateJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneySegmentProxy(sdkConfig)
	patchSegment := buildSdkPatchSegment(d)

	log.Printf("Updating journey segment %s", d.Id())
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current journey segment version
		segmentResponse, proxyResponse, getErr := proxy.getJourneySegmentById(ctx, d.Id())
		if getErr != nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey segment %s error: %s", d.Id(), getErr), proxyResponse)
		}

		patchSegment.Version = segmentResponse.Version
		_, proxyResponse, patchErr := proxy.updateJourneySegment(ctx, d.Id(), patchSegment)
		if patchErr != nil {
			input, _ := util.InterfaceToJson(*patchSegment)
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Error updating journey segment %s: %s\n(input: %+v)", *patchSegment.DisplayName, patchErr, input), proxyResponse)
		}
		return proxyResponse, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated journey segment %s", d.Id())
	return readJourneySegment(ctx, d, meta)
}

func deleteJourneySegment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getJourneySegmentProxy(sdkConfig)

	displayName := d.Get("display_name").(string)
	log.Printf("Deleting journey segment with display name %s", displayName)
	if proxyResponse, err := proxy.deleteJourneySegment(ctx, d.Id()); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete journey segment with display name %s error: %s", displayName, err), proxyResponse)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, proxyResponse, err := proxy.getJourneySegmentById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyResponse) {
				// journey segment deleted
				log.Printf("Deleted journey segment %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting journey segment %s | error: %s", d.Id(), err), proxyResponse))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("journey journey %s still exists", d.Id()), proxyResponse))
	})
}
