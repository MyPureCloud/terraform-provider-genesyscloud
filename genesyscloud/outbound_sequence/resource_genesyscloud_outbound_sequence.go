package outbound_sequence

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
The resource_genesyscloud_outbound_sequence.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOutboundSequence retrieves all of the outbound sequence via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOutboundSequences(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newOutboundSequenceProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	campaignSequences, err := proxy.getAllOutboundSequence(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get outbound sequence: %v", err)
	}

	for _, campaignSequence := range *campaignSequences {
		resources[*campaignSequence.Id] = &resourceExporter.ResourceMeta{Name: *campaignSequence.Name}
	}

	return resources, nil
}

// createOutboundSequence is used by the outbound_sequence resource to create Genesys cloud outbound sequence
func createOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundSequenceProxy(sdkConfig)
	status := d.Get("status").(string)

	outboundSequence := getOutboundSequenceFromResourceData(d)

	log.Printf("Creating outbound sequence %s", *outboundSequence.Name)
	campaignSequence, err := proxy.createOutboundSequence(ctx, &outboundSequence)
	if err != nil {
		return diag.Errorf("Failed to create outbound sequence: %s", err)
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
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundSequenceProxy(sdkConfig)

	log.Printf("Reading outbound sequence %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		campaignSequence, respCode, getErr := proxy.getOutboundSequenceById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read outbound sequence %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read outbound sequence %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundSequence())

		resourcedata.SetNillableValue(d, "name", campaignSequence.Name)
		if campaignSequence.Campaigns != nil {
			d.Set("campaign_ids", gcloud.SdkDomainEntityRefArrToList(*campaignSequence.Campaigns))
		}
		resourcedata.SetNillableValue(d, "status", campaignSequence.Status)
		resourcedata.SetNillableValue(d, "repeat", campaignSequence.Repeat)

		log.Printf("Read outbound sequence %s %s", d.Id(), *campaignSequence.Name)
		return cc.CheckState()
	})
}

// updateOutboundSequence is used by the outbound_sequence resource to update an outbound sequence in Genesys Cloud
func updateOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundSequenceProxy(sdkConfig)
	status := d.Get("status").(string)

	outboundSequence := getOutboundSequenceFromResourceData(d)
	if status != "off" {
		outboundSequence.Status = &status
	}

	log.Printf("Updating outbound sequence %s", *outboundSequence.Name)
	campaignSequence, err := proxy.updateOutboundSequence(ctx, d.Id(), &outboundSequence)
	if err != nil {
		return diag.Errorf("Failed to update outbound sequence: %s", err)
	}

	log.Printf("Updated outbound sequence %s", *campaignSequence.Id)
	return readOutboundSequence(ctx, d, meta)
}

// deleteOutboundSequence is used by the outbound_sequence resource to delete an outbound sequence from Genesys cloud
func deleteOutboundSequence(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundSequenceProxy(sdkConfig)

	// Sequence can't be deleted while running
	sequence, _, err := proxy.getOutboundSequenceById(ctx, d.Id())
	if *sequence.Status == "on" {
		if err != nil {
			return diag.Errorf("Failed to get outbound sequence %s: %s", d.Id(), err)
		}
		sequence.Status = platformclientv2.String("off")
		_, err = proxy.updateOutboundSequence(ctx, d.Id(), sequence)
		if err != nil {
			return diag.Errorf("Failed to turn off outbound sequence %s: %s", d.Id(), err)
		}
		time.Sleep(20 * time.Second) // Give the sequence a chance to turned off
	}

	_, err = proxy.deleteOutboundSequence(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete outbound sequence %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getOutboundSequenceById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted outbound sequence %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting outbound sequence %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("outbound sequence %s still exists", d.Id()))
	})
}
