package outbound_campaignrule

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

func getAllAuthCampaignRules(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getOutboundCampaignruleProxy(clientConfig)

	campaignRules, resp, err := proxy.getAllOutboundCampaignrule(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get outbound campaign rules error: %s", err), resp)
	}

	for _, campaignRule := range *campaignRules {
		resources[*campaignRule.Id] = &resourceExporter.ResourceMeta{BlockLabel: *campaignRule.Name}
	}
	return resources, nil
}

func createOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCampaignruleProxy(sdkConfig)

	rule := getCampaignruleFromResourceData(d)
	// API requires create in "off" state; enable via update if desired
	rule.Enabled = platformclientv2.Bool(false)

	log.Printf("Creating Outbound Campaign Rule %s", *rule.Name)
	outboundCampaignRule, resp, err := proxy.createOutboundCampaignrule(ctx, &rule)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create Outbound Campaign Rule %s error: %s", *rule.Name, err), resp)
	}

	d.SetId(*outboundCampaignRule.Id)
	log.Printf("Created Outbound Campaign Rule %s %s", *outboundCampaignRule.Name, *outboundCampaignRule.Id)

	if d.Get("enabled").(bool) {
		return updateOutboundCampaignRule(ctx, d, meta)
	}
	return readOutboundCampaignRule(ctx, d, meta)
}

func updateOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCampaignruleProxy(sdkConfig)
	enabled := d.Get("enabled").(bool)

	rule := getCampaignruleFromResourceData(d)
	if enabled {
		rule.Enabled = platformclientv2.Bool(true)
	}

	log.Printf("Updating Outbound Campaign Rule %s", *rule.Name)
	_, resp, err := proxy.updateOutboundCampaignrule(ctx, d.Id(), &rule)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update campaign rule %s error: %s", *rule.Name, err), resp)
	}

	log.Printf("Updated Outbound Campaign Rule %s", *rule.Name)
	return readOutboundCampaignRule(ctx, d, meta)
}

func readOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCampaignruleProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundCampaignrule(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Outbound Campaign Rule %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		campaignRule, resp, getErr := proxy.getOutboundCampaignruleById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound Campaign Rule %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound Campaign Rule %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", campaignRule.Name)
		resourcedata.SetNillableValueWithSchemaSetWithFunc(d, "campaign_rule_entities", campaignRule.CampaignRuleEntities, flattenCampaignRuleEntities)
		resourcedata.SetNillableValue(d, "campaign_rule_processing", campaignRule.CampaignRuleProcessing)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "condition_groups", campaignRule.ConditionGroups, flattenCampaignRuleConditionGroups)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "campaign_rule_conditions", campaignRule.CampaignRuleConditions, flattenCampaignRuleConditions)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "campaign_rule_actions", campaignRule.CampaignRuleActions, func(actions *[]platformclientv2.Campaignruleaction) []interface{} {
			return flattenCampaignRuleAction(actions, flattenCampaignRuleActionEntities)
		})
		resourcedata.SetNillableValue(d, "match_any_conditions", campaignRule.MatchAnyConditions)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "execution_settings", campaignRule.ExecutionSettings, flattenExecutionSettings)
		resourcedata.SetNillableValue(d, "enabled", campaignRule.Enabled)

		log.Printf("Read Outbound Campaign Rule %s %s", d.Id(), *campaignRule.Name)
		return cc.CheckState(d)
	})
}

func deleteOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundCampaignruleProxy(sdkConfig)

	ruleEnabled := d.Get("enabled").(bool)
	if ruleEnabled {
		// Have to disable rule before we can delete
		log.Printf("Disabling Outbound Campaign Rule")
		d.Set("enabled", false)
		rule, resp, err := proxy.getOutboundCampaignruleById(ctx, d.Id())
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get Outbound campaign rule %s error: %s", d.Id(), err), resp)
		}
		rule.Enabled = platformclientv2.Bool(false)
		_, resp, err = proxy.updateOutboundCampaignrule(ctx, d.Id(), rule)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to disable outbound campagin rule %s error: %s", d.Id(), err), resp)
		}
	}

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Campaign Rule")
		resp, err := proxy.deleteOutboundCampaignrule(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete outbound campaign rule %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getOutboundCampaignruleById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Outbound Campaign Rule deleted
				log.Printf("Deleted Outbound Campaign Rule %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting Outbound Campaign Rule %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Outbound Campaign Rule %s still exists", d.Id()), resp))
	})
}
