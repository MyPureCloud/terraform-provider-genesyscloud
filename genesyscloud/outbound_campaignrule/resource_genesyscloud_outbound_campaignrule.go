package outbound_campaignrule

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

func getAllAuthCampaignRules(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getOutboundCampaignruleProxy(clientConfig)

	campaignRules, err := proxy.getAllOutboundCampaignrule(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get outbound campaign rules: %v", err)
	}

	for _, campaignRule := range *campaignRules {
		resources[*campaignRule.Id] = &resourceExporter.ResourceMeta{Name: *campaignRule.Name}
	}

	return resources, nil
}

func createOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundCampaignruleProxy(sdkConfig)

	rule := getCampaignruleFromResourceData(d)

	log.Printf("Creating Outbound Campaign Rule %s", *rule.Name)
	outboundCampaignRule, err := proxy.createOutboundCampaignrule(ctx, &rule)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Campaign Rule %s: %s", *rule.Name, err)
	}

	d.SetId(*outboundCampaignRule.Id)
	log.Printf("Created Outbound Campaign Rule %s %s", *outboundCampaignRule.Name, *outboundCampaignRule.Id)

	enabled := d.Get("enabled").(bool)
	// Campaign rules can be enabled after creation
	if enabled {
		d.Set("enabled", enabled)
		diag := updateOutboundCampaignRule(ctx, d, meta)
		if diag != nil {
			return diag
		}
	}

	return readOutboundCampaignRule(ctx, d, meta)
}

func updateOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundCampaignruleProxy(sdkConfig)
	enabled := d.Get("enabled").(bool)

	rule := getCampaignruleFromResourceData(d)
	if enabled {
		rule.Enabled = platformclientv2.Bool(true)
	}

	log.Printf("Updating Outbound Campaign Rule %s", *rule.Name)
	_, err := proxy.updateOutboundCampaignrule(ctx, d.Id(), &rule)
	if err != nil {
		return diag.Errorf("Failed to update campaign rule %s: %s", d.Id(), err)
	}

	log.Printf("Updated Outbound Campaign Rule %s", *rule.Name)
	return readOutboundCampaignRule(ctx, d, meta)
}

func readOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundCampaignruleProxy(sdkConfig)

	log.Printf("Reading Outbound Campaign Rule %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		campaignRule, resp, getErr := proxy.getOutboundCampaignruleById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read Outbound Campaign Rule %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read Outbound Campaign Rule %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundCampaignrule())

		resourcedata.SetNillableValue(d, "name", campaignRule.Name)
		if campaignRule.CampaignRuleEntities != nil {
			d.Set("campaign_rule_entities", flattenCampaignRuleEntities(campaignRule.CampaignRuleEntities))
		}
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "campaign_rule_conditions", campaignRule.CampaignRuleConditions, flattenCampaignRuleConditions)
		if campaignRule.CampaignRuleActions != nil {
			d.Set("campaign_rule_actions", flattenCampaignRuleAction(campaignRule.CampaignRuleActions, flattenCampaignRuleActionEntities))
		} else {
			d.Set("campaign_rule_actions", nil)
		}
		resourcedata.SetNillableValue(d, "match_any_conditions", campaignRule.MatchAnyConditions)
		resourcedata.SetNillableValue(d, "enabled", campaignRule.Enabled)

		log.Printf("Read Outbound Campaign Rule %s %s", d.Id(), *campaignRule.Name)
		return cc.CheckState()
	})
}

func deleteOutboundCampaignRule(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundCampaignruleProxy(sdkConfig)

	ruleEnabled := d.Get("enabled").(bool)
	if ruleEnabled {
		// Have to disable rule before we can delete
		log.Printf("Disabling Outbound Campaign Rule")
		d.Set("enabled", false)
		rule, _, err := proxy.getOutboundCampaignruleById(ctx, d.Id())
		if err != nil {
			return diag.Errorf("Failed to find Outbound campaign rule %s for delete: %s", d.Id(), err)
		}
		rule.Enabled = platformclientv2.Bool(false)
		_, err = proxy.updateOutboundCampaignrule(ctx, d.Id(), rule)
		if err != nil {
			return diag.Errorf("Failed to disable outbound campaign rule %s: %s", d.Id(), err)
		}
	}

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Campaign Rule")
		resp, err := proxy.deleteOutboundCampaignrule(ctx, d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Campaign Rule: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getOutboundCampaignruleById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404ByInt(resp) {
				// Outbound Campaign Rule deleted
				log.Printf("Deleted Outbound Campaign Rule %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting Outbound Campaign Rule %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Campaign Rule %s still exists", d.Id()))
	})
}
