package outbound_ruleset

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
The resource_genesyscloud_outbound_ruleset.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOutboundRulesets retrieves all of the outbound rulesets via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOutboundRuleset(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getOutboundRulesetProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	rulesets, rsErr := proxy.getAllOutboundRuleset(ctx)
	if rsErr != nil {
		return nil, diag.Errorf("Failed to get ruleset: %v", rsErr)
	}

	// DEVTOOLING-319: Filtering rule sets by removing the ones that are referencing skills that no longer exist
	skillExporter := gcloud.RoutingSkillExporter()
	skillMap, skillErr := skillExporter.GetResourcesFunc(ctx)
	if skillErr != nil {
		return nil, diag.Errorf("Failed to get skill resources: %v", skillErr)
	}
	filteredRuleSets, filterErr := filterOutboundRulesets(*rulesets, skillMap)
	if filterErr != nil {
		return nil, diag.Errorf("Failed to filter outbound rulesets: %v", filterErr)
	}

	for _, ruleset := range filteredRuleSets {
		log.Printf("Dealing with ruleset id : %s", *ruleset.Id)
		resources[*ruleset.Id] = &resourceExporter.ResourceMeta{Name: *ruleset.Name}
	}

	return resources, nil
}

// createOutboundRuleset is used by the outbound_ruleset resource to create Genesys cloud outbound_ruleset
func createOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newOutboundRulesetProxy(sdkConfig)

	outboundRuleset := getOutboundRulesetFromResourceData(d)

	ruleset, err := proxy.createOutboundRuleset(ctx, &outboundRuleset)
	if err != nil {
		return diag.Errorf("Failed to create ruleset: %s", err)
	}

	d.SetId(*ruleset.Id)
	log.Printf("Created Outbound Ruleset %s", *ruleset.Id)
	return readOutboundRuleset(ctx, d, meta)
}

// readOutboundRuleset is used by the outbound_ruleset resource to read an outbound ruleset from genesys cloud.
func readOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newOutboundRulesetProxy(sdkConfig)

	log.Printf("Reading Outbound Ruleset %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		ruleset, respCode, getErr := proxy.getOutboundRulesetById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read Outbound Ruleset %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Outbound Ruleset %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundRuleset())

		resourcedata.SetNillableValue(d, "name", ruleset.Name)
		resourcedata.SetNillableReference(d, "contact_list_id", ruleset.ContactList)
		resourcedata.SetNillableReference(d, "queue_id", ruleset.Queue)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "rules", ruleset.Rules, flattenDialerrules)

		log.Printf("Read Outbound Ruleset %s %s", d.Id(), *ruleset.Name)
		return cc.CheckState()
	})
}

// updateOutboundRuleset is used by the outbound_ruleset resource to update an outbound ruleset in Genesys Cloud
func updateOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newOutboundRulesetProxy(sdkConfig)

	outboundRuleset := getOutboundRulesetFromResourceData(d)

	ruleset, err := proxy.updateOutboundRuleset(ctx, d.Id(), &outboundRuleset)
	if err != nil {
		return diag.Errorf("Failed to update ruleset: %s", err)
	}

	log.Printf("Updated Outbound Ruleset %s", *ruleset.Id)
	return readOutboundRuleset(ctx, d, meta)
}

// deleteOutboundRuleset is used by the outbound_ruleset resource to delete an outbound ruleset from Genesys cloud.
func deleteOutboundRuleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newOutboundRulesetProxy(sdkConfig)

	_, err := proxy.deleteOutboundRuleset(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete ruleset %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 1800*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getOutboundRulesetById(ctx, d.Id())

		//Now that I am checking for th error string of API 404 and there is no error, I need to move the isStatus404
		//moved out of the code
		if gcloud.IsStatus404ByInt(respCode) {
			// Outbound Ruleset deleted
			log.Printf("Deleted Outbound Ruleset %s", d.Id())
			return nil
		}

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error deleting Outbound Ruleset %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Ruleset %s still exists", d.Id()))
	})
}

// filterOutboundRuleset filters rulesets by removing the ones that are referencing skills that no longer exist in GC
func filterOutboundRulesets(ruleSets []platformclientv2.Ruleset, skillMap resourceExporter.ResourceIDMetaMap) ([]platformclientv2.Ruleset, diag.Diagnostics) {
	var filteredRuleSets []platformclientv2.Ruleset
	log.Printf("Filtering outbound rule sets")

RuleSetLoop:
	for _, ruleSet := range ruleSets {
		for _, rule := range *ruleSet.Rules {
			// look through rule actions to check if the referenced skills exist in our skill map or not
			for _, action := range *rule.Actions {
				if action.ActionTypeName != nil && strings.ToLower(*action.ActionTypeName) == "set_skills" && action.Properties != nil {
					for id, value := range *action.Properties {
						if strings.ToLower(id) == "skills" {
							// the property value is a json string wrapping an array of skill ids, need to convert it back to a slice to check if each skill exists
							var skillIds []string
							err := json.Unmarshal([]byte(value), &skillIds)
							if err != nil {
								log.Printf("Error decoding JSON: %s", err)
								return nil, diag.Errorf("Failed to filter ruleset: %s", err)
							}
							for _, skillId := range skillIds {
								_, found := skillMap[skillId]
								if !found { // skill id referenced by the rule action is not found in the skill map, we filter the ruleset out and evaluate the next one.
									log.Printf("Removing ruleset %s, the skill %s used in action does not exist in GC anymore", *ruleSet.Id, skillId)
									continue RuleSetLoop
								}
							}
						}
					}
				}
			}
			// look through rule conditions to check if the referenced skills exist in our skill map or not
			for _, condition := range *rule.Conditions {
				if condition.AttributeName != nil && strings.ToLower(*condition.AttributeName) == "skill" {
					if condition.Value != nil {
						var found bool
						for _, value := range skillMap {
							if value.Name == *condition.Value {
								found = true
								break
							}
						}
						if !found { // skill name referenced by rule condition is not found in the skill map, we filter the rulset out and evaluate the next one
							log.Printf("Removing ruleset %s, the skill %s used in condition does not exist in GC anymore", *ruleSet.Id, *condition.Value)
							continue RuleSetLoop
						}
					}
				}
			}
		}
		filteredRuleSets = append(filteredRuleSets, ruleSet)
	}

	return filteredRuleSets, nil
}
