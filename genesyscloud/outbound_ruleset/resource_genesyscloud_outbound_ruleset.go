package outbound_ruleset

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_outbound_ruleset.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOutboundRulesets retrieves all of the outbound rulesets via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOutboundRuleset(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getOutboundRulesetProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	rulesets, resp, rsErr := proxy.getAllOutboundRuleset(ctx)
	if rsErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get rulesets error: %s", rsErr), resp)
	}

	// DEVTOOLING-319: filters rule sets by removing the ones that reference skills that no longer exist in GC
	skillMap, skillErr := routingSkill.GetAllRoutingSkills(ctx, clientConfig)
	if skillErr != nil {
		return nil, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to get skill resources"), fmt.Errorf("%v", skillErr))
	}
	filteredRuleSets, filterErr := filterOutboundRulesets(*rulesets, skillMap)
	if filterErr != nil {
		return nil, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to filter outbound rulesets"), fmt.Errorf("%v", filterErr))
	}

	for _, ruleset := range filteredRuleSets {
		log.Printf("Dealing with ruleset id : %s", *ruleset.Id)
		resources[*ruleset.Id] = &resourceExporter.ResourceMeta{BlockLabel: *ruleset.Name}
	}
	return resources, nil
}

// filterOutboundRulesets filters rule sets by removing the ones that reference skills that no longer exist in GC
func filterOutboundRulesets(ruleSets []platformclientv2.Ruleset, skillMap resourceExporter.ResourceIDMetaMap) ([]platformclientv2.Ruleset, diag.Diagnostics) {
	var filteredRuleSets []platformclientv2.Ruleset
	log.Printf("Filtering outbound rule sets")

	for _, ruleSet := range ruleSets {
		var foundDeleted bool
		for _, rule := range *ruleSet.Rules {
			if doesRuleActionsRefDeletedSkill(rule, skillMap) || doesRuleConditionsRefDeletedSkill(rule, skillMap) {
				foundDeleted = true
				break
			}
		}
		if foundDeleted {
			log.Printf("Removing ruleset id '%s'", *ruleSet.Id)
		} else {
			// No references to a deleted skill in the ruleset, keep it
			filteredRuleSets = append(filteredRuleSets, ruleSet)
		}
	}
	return filteredRuleSets, nil
}

// createOutboundRuleset is used by the outbound_ruleset resource to create Genesys cloud outbound_ruleset
func createOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newOutboundRulesetProxy(sdkConfig)

	outboundRuleset := getOutboundRulesetFromResourceData(d)

	ruleset, resp, err := proxy.createOutboundRuleset(ctx, &outboundRuleset)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create ruleset %s error: %s", *outboundRuleset.Name, err), resp)
	}

	d.SetId(*ruleset.Id)
	log.Printf("Created Outbound Ruleset %s", *ruleset.Id)
	return readOutboundRuleset(ctx, d, meta)
}

// readOutboundRuleset is used by the outbound_ruleset resource to read an outbound ruleset from genesys cloud.
func readOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newOutboundRulesetProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundRuleset(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Outbound Ruleset %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		ruleset, resp, getErr := proxy.getOutboundRulesetById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Outbound Ruleset %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Outbound Ruleset %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", ruleset.Name)
		resourcedata.SetNillableReference(d, "contact_list_id", ruleset.ContactList)
		resourcedata.SetNillableReference(d, "queue_id", ruleset.Queue)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "rules", ruleset.Rules, flattenDialerrules)

		log.Printf("Read Outbound Ruleset %s %s", d.Id(), *ruleset.Name)
		return cc.CheckState(d)
	})
}

// updateOutboundRuleset is used by the outbound_ruleset resource to update an outbound ruleset in Genesys Cloud
func updateOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newOutboundRulesetProxy(sdkConfig)

	outboundRuleset := getOutboundRulesetFromResourceData(d)

	ruleset, resp, err := proxy.updateOutboundRuleset(ctx, d.Id(), &outboundRuleset)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update ruleset %s error: %s", *outboundRuleset.Name, err), resp)
	}

	log.Printf("Updated Outbound Ruleset %s", *ruleset.Id)
	return readOutboundRuleset(ctx, d, meta)
}

// deleteOutboundRuleset is used by the outbound_ruleset resource to delete an outbound ruleset from Genesys cloud.
func deleteOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := newOutboundRulesetProxy(sdkConfig)

	resp, err := proxy.deleteOutboundRuleset(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete ruleset %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 1800*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getOutboundRulesetById(ctx, d.Id())

		//Now that I am checking for th error string of API 404 and there is no error, I need to move the isStatus404
		//moved out of the code
		if util.IsStatus404(resp) {
			// Outbound Ruleset deleted
			log.Printf("Deleted Outbound Ruleset %s", d.Id())
			return nil
		}

		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Outbound Ruleset %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Outbound Ruleset %s still exists", d.Id()), resp))
	})
}
