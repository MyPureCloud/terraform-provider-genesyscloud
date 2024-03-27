package routing_queue_outbound_email_address

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
	consistencyChecker "terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

func createRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	queueId := d.Get("queue_id").(string)
	log.Printf("creating conditional group routing rules for queue %s", queueId)
	d.SetId(queueId)

	return updateRoutingQueueOutboundEmailAddress(ctx, d, meta)
}

func readRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueOutboundEmailAddressProxy(sdkConfig)
	queueId := strings.Split(d.Id(), "/")[0]

	log.Printf("Reading routing queue %s outbound email address", queueId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {

		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read conditional group routing for queue %s: %s", queueId, getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read conditional group routing for queue %s: %s", queueId, getErr))
		}

		cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueueOutboundEmailAddress())

		_ = d.Set("queue_id", queueId)

		return cc.CheckState()
	})
}

func updateRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func deleteRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
