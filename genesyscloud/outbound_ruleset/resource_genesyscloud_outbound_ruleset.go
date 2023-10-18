package outbound_ruleset

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

/*
The resource_genesyscloud_outbound_ruleset.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOutboundRulesets retrieves all of the outbound rulesets via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOutboundRuleset(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getOutboundRulesetProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	rulesets, err := proxy.getAllOutboundRuleset(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get ruleset: %v", err)
	}

	for _, ruleset := range *rulesets {
		log.Printf("Dealing with ruleset id : %s", *ruleset.Id)
		resources[*ruleset.Id] = &resourceExporter.ResourceMeta{Name: *ruleset.Id}
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
