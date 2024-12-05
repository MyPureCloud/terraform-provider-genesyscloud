package routing_queue_outbound_email_address

import (
	"context"
	"fmt"
	"log"
	"strings"
	consistencyChecker "terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_routing_queue_outbound-email_address.go contains all the methods that perform the core logic for the resource.
*/

func getAllAuthRoutingQueueOutboundEmailAddress(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	if exists := featureToggles.OEAToggleExists(); !exists {
		log.Printf("Environment variable %s not set, skipping exporter for %s", featureToggles.OEAToggleName(), ResourceType)
		return nil, nil
	}

	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingQueueOutboundEmailAddressProxy(clientConfig)

	queues, resp, err := proxy.routingQueueProxy.GetAllRoutingQueues(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, "failed to get outbound email addresses for routing queues", resp)
	}

	for _, queue := range *queues {
		if queue.OutboundEmailAddress != nil && !isQueueEmailAddressEmpty(*queue.OutboundEmailAddress) {
			resources[*queue.Id] = &resourceExporter.ResourceMeta{BlockLabel: *queue.Name + "-email-address"}
		}
	}

	return resources, nil
}

func createRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.OEAToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, "Environment variable ENABLE_STANDALONE_EMAIL_ADDRESS not set", fmt.Errorf("environment variable %s not set", featureToggles.OEAToggleName()))
	}

	queueId := d.Get("queue_id").(string)
	log.Printf("creating outbound email address for queue %s", queueId)
	d.SetId(queueId)

	return updateRoutingQueueOutboundEmailAddress(ctx, d, meta)
}

func readRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.OEAToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, "Environment variable ENABLE_STANDALONE_EMAIL_ADDRESS not set", fmt.Errorf("environment variable %s not set", featureToggles.OEAToggleName()))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueOutboundEmailAddressProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueueOutboundEmailAddress(), constants.ConsistencyChecks(), ResourceType)

	queueId := d.Id()

	log.Printf("Reading outbound email address for queue %s", queueId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		queueEmailAddress, resp, getErr := proxy.getRoutingQueueOutboundEmailAddress(ctx, queueId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read outbound email address for queue %s: %s", queueId, getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read outbound email address for queue %s: %s", queueId, getErr))
		}

		_ = d.Set("queue_id", queueId)
		resourcedata.SetNillableReference(d, "domain_id", queueEmailAddress.Domain)

		// The route property is a **Inboundroute hence all the checks
		if queueEmailAddress.Route != nil && *queueEmailAddress.Route != nil && (*queueEmailAddress.Route).Id != nil {
			_ = d.Set("route_id", *(*queueEmailAddress.Route).Id)
		}

		log.Printf("Reading outbound email address for queue %s", queueId)
		return cc.CheckState(d)
	})
}

func updateRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.OEAToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, "Environment variable ENABLE_STANDALONE_EMAIL_ADDRESS not set", fmt.Errorf("environment variable %s not set", featureToggles.OEAToggleName()))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueOutboundEmailAddressProxy(sdkConfig)
	queueId := d.Id()

	inboundRoute := &platformclientv2.Inboundroute{
		Id: platformclientv2.String(d.Get("route_id").(string)),
	}

	emailAddress := platformclientv2.Queueemailaddress{
		Domain: &platformclientv2.Domainentityref{
			Id: platformclientv2.String(d.Get("domain_id").(string)),
		},
		Route: &inboundRoute,
	}

	log.Printf("updating outbound email address for queue %s", queueId)
	_, resp, err := proxy.updateRoutingQueueOutboundEmailAddress(ctx, queueId, &emailAddress)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update outbound email address for queue %s", queueId), resp)
	}
	log.Printf("updated outbound email address for queue %s", queueId)

	return readRoutingQueueOutboundEmailAddress(ctx, d, meta)
}

func deleteRoutingQueueOutboundEmailAddress(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.OEAToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, "Environment variable ENABLE_STANDALONE_EMAIL_ADDRESS not set", fmt.Errorf("environment variable %s not set", featureToggles.OEAToggleName()))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueOutboundEmailAddressProxy(sdkConfig)
	queueId := d.Id()

	log.Printf("Removing email address from queue %s", queueId)

	// check if routing queue still exists before trying to remove outbound email address
	_, resp, err := proxy.getRoutingQueueOutboundEmailAddress(ctx, queueId)
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("outbound email address's parent queue %s already deleted", queueId)
			return nil
		}

		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to remove outbound email address for queue %s", queueId), resp)
	}

	// To delete, update the queue with an empty email address
	var emptyAddress platformclientv2.Queueemailaddress
	_, _, err = proxy.updateRoutingQueueOutboundEmailAddress(ctx, queueId, &emptyAddress)
	if err != nil && !strings.Contains(err.Error(), "error updating outbound email address for routing queue") {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to remove outbound email address from queue %s", queueId), resp)
	}

	// Verify there is no email address
	rules, _, _ := proxy.getRoutingQueueOutboundEmailAddress(ctx, queueId)
	if rules != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("outbound email address still exist for queue %s", queueId), resp)
	}

	log.Printf("Removed outbound email address from queue %s", queueId)
	return nil
}
