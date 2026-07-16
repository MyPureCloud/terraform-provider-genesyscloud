package routing_queue_identity_resolution

/*
The resource_genesyscloud_routing_queue_identity_resolution.go file contains all the methods that perform the core logic for the resource.
*/

import (
	"context"
	"fmt"
	"log"

	consistencyChecker "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
)

func getAllRoutingQueueIdentityResolution(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingQueueIdentityResolutionProxy(clientConfig)

	queues, resp, err := proxy.routingQueueProxy.GetAllRoutingQueues(ctx, "", false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, "failed to list routing queues for identity resolution export", resp)
	}

	for _, queue := range *queues {
		if queue.Id == nil || queue.Name == nil {
			continue
		}

		config, getResp, getErr := proxy.getRoutingQueueIdentityResolution(ctx, *queue.Id)
		if getErr != nil {
			if util.IsStatus404(getResp) {
				continue
			}
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get identity resolution for queue %s", *queue.Id), getResp)
		}

		if isDefaultIdentityResolutionConfig(config) {
			continue
		}

		resources[*queue.Id] = &resourceExporter.ResourceMeta{BlockLabel: *queue.Name + "-identity-resolution"}
	}

	return resources, nil
}

func createRoutingQueueIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	queueId := d.Get("queue_id").(string)
	log.Printf("creating identity resolution for queue %s", queueId)
	d.SetId(queueId)

	return updateRoutingQueueIdentityResolution(ctx, d, meta)
}

func readRoutingQueueIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueIdentityResolutionProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueueIdentityResolution(), constants.ConsistencyChecks(), ResourceType)

	queueId := d.Id()
	log.Printf("reading identity resolution for queue %s", queueId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		config, resp, getErr := proxy.getRoutingQueueIdentityResolution(ctx, queueId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				log.Printf("parent queue %s not found, removing identity resolution from state", queueId)
				d.SetId("")
				return nil
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read identity resolution for queue %s: %s", queueId, getErr), resp))
		}

		_ = d.Set("queue_id", queueId)
		_ = d.Set("call_on_behalf_of_queue", flattenCallOnBehalfOfQueue(config.CallOnBehalfOfQueue))

		log.Printf("read identity resolution for queue %s", queueId)
		return cc.CheckState(d)
	})
}

func updateRoutingQueueIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueIdentityResolutionProxy(sdkConfig)
	queueId := d.Id()

	config, err := buildIdentityResolutionQueueConfig(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to build identity resolution config", err)
	}

	log.Printf("updating identity resolution for queue %s", queueId)
	_, resp, putErr := proxy.putRoutingQueueIdentityResolution(ctx, queueId, config)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update identity resolution for queue %s", queueId), resp)
	}

	log.Printf("updated identity resolution for queue %s", queueId)
	return readRoutingQueueIdentityResolution(ctx, d, meta)
}

func deleteRoutingQueueIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueIdentityResolutionProxy(sdkConfig)
	queueId := d.Id()

	log.Printf("resetting identity resolution for queue %s to default", queueId)

	_, resp, getErr := proxy.getRoutingQueueById(ctx, queueId)
	if getErr != nil {
		if util.IsStatus404(resp) {
			log.Printf("parent queue %s already deleted", queueId)
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to verify queue %s before resetting identity resolution", queueId), resp)
	}

	defaultConfig := buildDefaultIdentityResolutionQueueConfig()
	_, putResp, putErr := proxy.putRoutingQueueIdentityResolution(ctx, queueId, defaultConfig)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to reset identity resolution for queue %s", queueId), putResp)
	}

	log.Printf("reset identity resolution for queue %s to default", queueId)
	return nil
}
