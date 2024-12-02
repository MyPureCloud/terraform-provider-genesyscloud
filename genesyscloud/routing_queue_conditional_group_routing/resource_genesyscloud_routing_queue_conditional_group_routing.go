package routing_queue_conditional_group_routing

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
The resource_genesyscloud_routing_queue_conditional_group_routing.go contains all the methods that perform the core logic for the resource.
*/

func getAllAuthRoutingQueueConditionalGroup(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingQueueConditionalGroupRoutingProxy(clientConfig)

	if exists := featureToggles.CSGToggleExists(); !exists {
		log.Printf("Environment variable %s not set, skipping exporter for %s", featureToggles.CSGToggleName(), ResourceType)
		return nil, nil
	}

	queues, resp, err := proxy.routingQueueProxy.GetAllRoutingQueues(ctx, "")
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

func buildConditionalGroupRouting(rules []interface{}) ([]platformclientv2.Conditionalgrouproutingrule, error) {
	var sdkRules []platformclientv2.Conditionalgrouproutingrule
	for i, rule := range rules {
		configRule := rule.(map[string]interface{})
		sdkRule := platformclientv2.Conditionalgrouproutingrule{
			Operator:       platformclientv2.String(configRule["operator"].(string)),
			ConditionValue: platformclientv2.Float64(configRule["condition_value"].(float64)),
		}

		if evaluatedQueue, ok := configRule["evaluated_queue_id"].(string); ok && evaluatedQueue != "" {
			if i == 0 {
				return nil, fmt.Errorf("for rule 1, the current queue is used so evaluated_queue_id should not be specified")
			}
			sdkRule.Queue = &platformclientv2.Domainentityref{Id: &evaluatedQueue}
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkRule.Metric, configRule, "metric")
		if waitSeconds, ok := configRule["wait_seconds"].(int); ok {
			sdkRule.WaitSeconds = &waitSeconds
		}

		if memberGroupList, ok := configRule["groups"].([]interface{}); ok {
			var sdkMemberGroups []platformclientv2.Membergroup
			for _, memberGroup := range memberGroupList {
				memberGroupMap, ok := memberGroup.(map[string]interface{})
				if !ok {
					continue
				}

				sdkMemberGroup := platformclientv2.Membergroup{
					Id:      platformclientv2.String(memberGroupMap["member_group_id"].(string)),
					VarType: platformclientv2.String(memberGroupMap["member_group_type"].(string)),
				}
				sdkMemberGroups = append(sdkMemberGroups, sdkMemberGroup)
			}
			sdkRule.Groups = &sdkMemberGroups
		}

		sdkRules = append(sdkRules, sdkRule)
	}

	return sdkRules, nil
}

func flattenConditionalGroupRouting(sdkRules *[]platformclientv2.Conditionalgrouproutingrule) []interface{} {
	if sdkRules == nil {
		return nil
	}

	var rules []interface{}
	for i, sdkRule := range *sdkRules {
		rule := make(map[string]interface{})

		// The first rule is assumed to apply to this queue, so evaluated_queue_id should be omitted
		if i > 0 {
			resourcedata.SetMapReferenceValueIfNotNil(rule, "evaluated_queue_id", sdkRule.Queue)
		}
		resourcedata.SetMapValueIfNotNil(rule, "wait_seconds", sdkRule.WaitSeconds)
		resourcedata.SetMapValueIfNotNil(rule, "operator", sdkRule.Operator)
		resourcedata.SetMapValueIfNotNil(rule, "condition_value", sdkRule.ConditionValue)
		resourcedata.SetMapValueIfNotNil(rule, "metric", sdkRule.Metric)

		if sdkRule.Groups != nil {
			memberGroups := make([]interface{}, 0)
			for _, group := range *sdkRule.Groups {
				memberGroupMap := make(map[string]interface{})

				resourcedata.SetMapValueIfNotNil(memberGroupMap, "member_group_id", group.Id)
				resourcedata.SetMapValueIfNotNil(memberGroupMap, "member_group_type", group.VarType)

				memberGroups = append(memberGroups, memberGroupMap)
			}
			rule["groups"] = memberGroups
		}

		rules = append(rules, rule)
	}
	return rules
}
