package outbound_sequence

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
The resource_genesyscloud_outbound_sequence.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOutboundSequence retrieves all of the outbound sequence via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOutboundSequences(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newOutboundSequenceProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	campaignSequences, resp, err := proxy.getAllOutboundSequence(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get outbound sequences error: %s", err), resp)
	}

	for _, campaignSequence := range *campaignSequences {
		resources[*campaignSequence.Id] = &resourceExporter.ResourceMeta{BlockLabel: *campaignSequence.Name}
	}
	return resources, nil
}

// createOutboundSequence is used by the outbound_sequence resource to create Genesys cloud outbound sequence
func createOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundSequenceProxy(sdkConfig)
	status := d.Get("status").(string)

	outboundSequence := getOutboundSequenceFromResourceData(d)

	log.Printf("Creating outbound sequence %s", *outboundSequence.Name)
	campaignSequence, resp, err := proxy.createOutboundSequence(ctx, &outboundSequence)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create outbound sequence %s error: %s", *outboundSequence.Name, err), resp)
	}

	d.SetId(*campaignSequence.Id)
	// Campaigns sequences can be enabled after creation
	if status == "on" {
		d.Set("status", status)
		diag := updateOutboundSequence(ctx, d, meta)
		if diag != nil {
			return diag
		}
	}

	log.Printf("Created outbound sequence %s", *campaignSequence.Id)
	return readOutboundSequence(ctx, d, meta)
}

// readOutboundSequence is used by the outbound_sequence resource to read an outbound sequence from genesys cloud
func readOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundSequenceProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundSequence(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading outbound sequence %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		campaignSequence, resp, getErr := proxy.getOutboundSequenceById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read outbound sequence %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read outbound sequence %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", campaignSequence.Name)
		if campaignSequence.Campaigns != nil {
			d.Set("campaign_ids", util.SdkDomainEntityRefArrToList(*campaignSequence.Campaigns))
		}
		resourcedata.SetNillableValue(d, "status", campaignSequence.Status)
		resourcedata.SetNillableValue(d, "repeat", campaignSequence.Repeat)

		log.Printf("Read outbound sequence %s %s", d.Id(), *campaignSequence.Name)
		return cc.CheckState(d)
	})
}

// updateOutboundSequence is used by the outbound_sequence resource to update an outbound sequence in Genesys Cloud
func updateOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundSequenceProxy(sdkConfig)
	status := d.Get("status").(string)

	outboundSequence := getOutboundSequenceFromResourceData(d)
	if status != "off" {
		outboundSequence.Status = &status
	}

	log.Printf("Updating outbound sequence %s", *outboundSequence.Name)
	campaignSequence, resp, err := proxy.updateOutboundSequence(ctx, d.Id(), &outboundSequence)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update outbound sequence %s error: %s", *outboundSequence.Name, err), resp)
	}

	log.Printf("Updated outbound sequence %s", *campaignSequence.Id)
	return readOutboundSequence(ctx, d, meta)
}

// deleteOutboundSequence is used by the outbound_sequence resource to delete an outbound sequence from Genesys cloud
func deleteOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundSequenceProxy(sdkConfig)

	// Sequence can't be deleted while running
	sequence, resp, err := proxy.getOutboundSequenceById(ctx, d.Id())
	if *sequence.Status == "on" {
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get outbound sequence %s error: %s", d.Id(), err), resp)
		}
		sequence.Status = platformclientv2.String("off")
		_, resp, err = proxy.updateOutboundSequence(ctx, d.Id(), sequence)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to turn off outbound sequence %s error: %s", d.Id(), err), resp)
		}
		time.Sleep(20 * time.Second) // Give the sequence a chance to turned off
	}

	resp, err = proxy.deleteOutboundSequence(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete outbound sequence %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getOutboundSequenceById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted outbound sequence %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting outbound sequence %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("outbound sequence %s still exists", d.Id()), resp))
	})
}
