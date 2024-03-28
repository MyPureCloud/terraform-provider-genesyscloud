package routing_queue_outbound_email_address

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
	"log"
	consistencyChecker "terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
The resource_genesyscloud_routing_queue_outbound-email_address.go contains all the methods that perform the core logic for the resource.
*/

func getAllAuthRoutingQueueOutboundEmailAddress(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingQueueOutboundEmailAddressProxy(clientConfig)

	queues, _, err := proxy.getAllRoutingQueues(ctx)
	if err != nil {
		return nil, diag.Errorf("failed to get routing queues outbound email address: %s", err)
	}

	for _, queue := range *queues {
		if queue.OutboundEmailAddress != nil && *queue.OutboundEmailAddress != nil {
			resources[*queue.Id] = &resourceExporter.ResourceMeta{Name: *queue.Id + "-email-address"}
		}
	}

	return resources, nil
}

func createRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	queueId := d.Get("queue_id").(string)
	log.Printf("creating outbound email address for queue %s", queueId)
	d.SetId(queueId)

	return updateRoutingQueueOutboundEmailAddress(ctx, d, meta)
}

func readRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueOutboundEmailAddressProxy(sdkConfig)
	queueId := d.Id()

	log.Printf("Reading routing queue %s outbound email address", queueId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		queueEmailAddress, resp, getErr := proxy.getRoutingQueueOutboundEmailAddress(ctx, queueId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read outbound email address for queue %s: %s", queueId, getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read outbound email address for queue %s: %s", queueId, getErr))
		}

		cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueueOutboundEmailAddress())

		_ = d.Set("queue_id", queueId)
		resourcedata.SetNillableReference(d, "domain_id", queueEmailAddress.Domain)
		resourcedata.SetNillableValue()

		return cc.CheckState()
	})
}

func updateRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueOutboundEmailAddressProxy(sdkConfig)
	queueId := d.Id()

	log.Printf("updating outbound email address for queue %s", queueId)
	_, _, err := proxy.updateRoutingQueueOutboundEmailAddress(ctx, queueId, &sdkRules)
	if err != nil {
		return diag.Errorf("failed to update outbound email address for queue %s: %s", queueId, err)
	}
	log.Printf("updated outbound email address for queue %s", queueId)

	return readRoutingQueueOutboundEmailAddress(ctx, d, meta)
}

func deleteRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
