package routing_queue_conditional_group_routing

import (
	"context"
	"fmt"
	consistencyChecker "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	featureToggles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The resource_genesyscloud_routing_queue_conditional_group_routing.go contains all the methods that perform the core logic for the resource.
*/

func getAllAuthRoutingQueueConditionalGroup(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingQueueConditionalGroupRoutingProxy(clientConfig)

	if exists := featureToggles.CSGToggleExists(); !exists {
		log.Printf("Environment variable %s not set, skipping exporter for %s", featureToggles.CSGToggleName(), ResourceType)
		return nil, nil
	}

	queues, resp, err := proxy.routingQueueProxy.GetAllRoutingQueues(ctx, "", false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get conditional group routing rules: %s", err), resp)
	}

	for _, queue := range *queues {
		if queue.ConditionalGroupRouting != nil && queue.ConditionalGroupRouting.Rules != nil {
			resources[*queue.Id+"/rules"] = &resourceExporter.ResourceMeta{BlockLabel: *queue.Name + "-rules"}
		}
	}

	return resources, nil
}

// createRoutingQueueConditionalRoutingGroup is used by the routing_queue_conditional_group_routing resource to create Conditional Group Routing Rules
func createRoutingQueueConditionalRoutingGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.CSGToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, "Environment variable ENABLE_STANDALONE_CGR not set", fmt.Errorf("environment variable %s not set", featureToggles.CSGToggleName()))
	}

	queueId := d.Get("queue_id").(string)
	log.Printf("creating conditional group routing rules for queue %s", queueId)
	d.SetId(queueId + "/rule") // Adding /rule to the id so the id doesn't conflict with the id of the routing queue these rules belong to

	return updateRoutingQueueConditionalRoutingGroup(ctx, d, meta)
}

// readRoutingQueueConditionalRoutingGroup is used by the routing_queue_conditional_group_routing resource to read Conditional Group Routing Rules
func readRoutingQueueConditionalRoutingGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.CSGToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, "Environment variable ENABLE_STANDALONE_CGR not set", fmt.Errorf("environment variable %s not set", featureToggles.CSGToggleName()))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueConditionalGroupRoutingProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueueConditionalGroupRouting(), constants.ConsistencyChecks(), ResourceType)
	queueId := strings.Split(d.Id(), "/")[0]

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		log.Printf("Reading routing queue %s conditional group routing rules", queueId)
		sdkRules, resp, getErr := proxy.getRoutingQueueConditionRouting(ctx, queueId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read conditional group routing for queue %s | error: %s", queueId, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read conditional group routing for queue %s | error: %s", queueId, getErr), resp))
		}
		log.Printf("Read routing queue %s conditional group routing rules", queueId)

		_ = d.Set("queue_id", queueId)
		_ = d.Set("rules", flattenConditionalGroupRouting(sdkRules))

		return cc.CheckState(d)
	})
}

// updateRoutingQueueConditionalRoutingGroup is used by the routing_queue_conditional_group_routing resource to update Conditional Group Routing Rules
func updateRoutingQueueConditionalRoutingGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.CSGToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, "Environment variable ENABLE_STANDALONE_CGR not set", fmt.Errorf("environment variable %s not set", featureToggles.CSGToggleName()))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueConditionalGroupRoutingProxy(sdkConfig)

	queueId := strings.Split(d.Id(), "/")[0]
	rules := d.Get("rules").([]interface{})

	sdkRules, err := buildConditionalGroupRouting(rules)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "Error building conditional group routing", err)
	}

	log.Printf("updating conditional group routing rules for queue %s", queueId)
	_, resp, err := proxy.updateRoutingQueueConditionRouting(ctx, queueId, &sdkRules)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Error updating routing queue conditional routing %s | error: %s", queueId, err), resp)
	}
	log.Printf("updated conditional group routing rules for queue %s", queueId)

	return readRoutingQueueConditionalRoutingGroup(ctx, d, meta)
}

// deleteRoutingQueueConditionalRoutingGroup is used by the routing_queue_conditional_group_routing resource to delete Conditional Group Routing Rules
func deleteRoutingQueueConditionalRoutingGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueConditionalGroupRoutingProxy(sdkConfig)
	queueId := strings.Split(d.Id(), "/")[0]

	log.Printf("Removing rules from queue %s", queueId)

	// check if routing queue still exists before trying to remove rules
	log.Printf("Reading queue '%s' to verify it exists before trying to remove its CGR rules", queueId)
	if _, resp, err := proxy.getRoutingQueueById(ctx, queueId); err != nil {
		if util.IsStatus404(resp) {
			log.Printf("conditional group routing rules parent queue %s already deleted", queueId)
			return nil
		}
		log.Printf("Failed to read routing queue '%s': %v", queueId, err)
	}

	// To delete conditional group routing, update the queue with no rules
	log.Printf("Updating routing queue '%s' to have no CGR rules", queueId)
	var newRules []platformclientv2.Conditionalgrouproutingrule
	if _, resp, err := proxy.updateRoutingQueueConditionRouting(ctx, queueId, &newRules); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to remove rules from queue %s: %s", queueId, err), resp)
	}

	// Verify there are no rules
	log.Printf("Reading queue '%s' CGR rules to verify that they have been removed", queueId)
	if rules, resp, err := proxy.getRoutingQueueConditionRouting(ctx, queueId); rules != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("conditional group routing rules still exist for queue %s: %s", queueId, err), resp)
	}

	log.Printf("Successfully removed rules from queue %s", queueId)
	return nil
}
