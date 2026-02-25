package routing_queue_conditional_group_activation

import (
	"context"
	"fmt"
	"log"
	"strings"

	consistencyChecker "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	featureToggles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

func getAllRoutingQueueConditionalGroupActivation(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingQueueConditionalGroupActivationProxy(clientConfig)

	if exists := featureToggles.CGAToggleExists(); !exists {
		log.Printf("Environment variable %s not set, skipping exporter for %s", featureToggles.CGAToggleName(), ResourceType)
		return nil, nil
	}

	queues, resp, err := proxy.routingQueueProxy.GetAllRoutingQueues(ctx, "", false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get conditional group activation rules: %s", err), resp)
	}

	for _, queue := range *queues {
		if queue.ConditionalGroupActivation != nil && queue.ConditionalGroupActivation.Rules != nil {
			resources[*queue.Id+"/cga"] = &resourceExporter.ResourceMeta{BlockLabel: *queue.Name + "-cga"}
		}
	}

	return resources, nil
}

func createRoutingQueueConditionalGroupActivation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.CGAToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, "Environment variable ENABLE_STANDALONE_CGA not set", fmt.Errorf("environment variable %s not set", featureToggles.CGAToggleName()))
	}

	queueId := d.Get("queue_id").(string)
	log.Printf("creating conditional group activation rules for queue %s", queueId)
	d.SetId(queueId + "/cga")

	return updateRoutingQueueConditionalGroupActivation(ctx, d, meta)
}

func readRoutingQueueConditionalGroupActivation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.CGAToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, "Environment variable ENABLE_STANDALONE_CGA not set", fmt.Errorf("environment variable %s not set", featureToggles.CGAToggleName()))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueConditionalGroupActivationProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingQueueConditionalGroupActivation(), constants.ConsistencyChecks(), ResourceType)
	queueId := strings.Split(d.Id(), "/")[0]

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		log.Printf("Reading routing queue %s conditional group activation rules", queueId)
		sdkCga, resp, getErr := proxy.getRoutingQueueConditionActivation(ctx, queueId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read conditional group activation for queue %s | error: %s", queueId, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read conditional group activation for queue %s | error: %s", queueId, getErr), resp))
		}
		log.Printf("Read routing queue %s conditional group activation rules", queueId)

		_ = d.Set("queue_id", queueId)

		if sdkCga != nil {
			flattened := flattenConditionalGroupActivation(sdkCga)
			if pilotRule, ok := flattened["pilot_rule"]; ok {
				_ = d.Set("pilot_rule", pilotRule)
			}
			if rules, ok := flattened["rules"]; ok {
				_ = d.Set("rules", rules)
			}
		}

		return cc.CheckState(d)
	})
}

func updateRoutingQueueConditionalGroupActivation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if exists := featureToggles.CGAToggleExists(); !exists {
		return util.BuildDiagnosticError(ResourceType, "Environment variable ENABLE_STANDALONE_CGA not set", fmt.Errorf("environment variable %s not set", featureToggles.CGAToggleName()))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueConditionalGroupActivationProxy(sdkConfig)

	queueId := strings.Split(d.Id(), "/")[0]

	cgaConfig := make(map[string]interface{})
	if v, ok := d.GetOk("pilot_rule"); ok {
		cgaConfig["pilot_rule"] = v
	}
	if v, ok := d.GetOk("rules"); ok {
		cgaConfig["rules"] = v
	}

	sdkCga := buildConditionalGroupActivation(cgaConfig)

	log.Printf("updating conditional group activation rules for queue %s", queueId)
	_, resp, err := proxy.updateRoutingQueueConditionActivation(ctx, queueId, &sdkCga)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Error updating routing queue conditional group activation %s | error: %s", queueId, err), resp)
	}
	log.Printf("updated conditional group activation rules for queue %s", queueId)

	return readRoutingQueueConditionalGroupActivation(ctx, d, meta)
}

func deleteRoutingQueueConditionalGroupActivation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingQueueConditionalGroupActivationProxy(sdkConfig)
	queueId := strings.Split(d.Id(), "/")[0]

	log.Printf("Removing conditional group activation rules from queue %s", queueId)

	log.Printf("Reading queue '%s' to verify it exists before trying to remove its CGA rules", queueId)
	if _, resp, err := proxy.getRoutingQueueById(ctx, queueId); err != nil {
		if util.IsStatus404(resp) {
			log.Printf("conditional group activation rules parent queue %s already deleted", queueId)
			return nil
		}
		log.Printf("Failed to read routing queue '%s': %v", queueId, err)
	}

	log.Printf("Updating routing queue '%s' to have no CGA rules", queueId)
	emptyCga := platformclientv2.Conditionalgroupactivation{}
	if _, resp, err := proxy.updateRoutingQueueConditionActivation(ctx, queueId, &emptyCga); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to remove conditional group activation rules from queue %s: %s", queueId, err), resp)
	}

	log.Printf("Reading queue '%s' CGA rules to verify that they have been removed", queueId)
	if cga, resp, err := proxy.getRoutingQueueConditionActivation(ctx, queueId); cga != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("conditional group activation rules still exist for queue %s: %s", queueId, err), resp)
	}

	log.Printf("Successfully removed conditional group activation rules from queue %s", queueId)
	return nil
}
