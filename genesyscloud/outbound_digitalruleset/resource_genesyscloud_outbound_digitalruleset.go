package outbound_digitalruleset

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_outbound_digitalruleset.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOutboundDigitalruleset retrieves all of the outbound digitalruleset via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOutboundDigitalrulesets(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newOutboundDigitalrulesetProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	digitalRuleSets, resp, err := proxy.getAllOutboundDigitalruleset(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get outbound digitalruleset: %v", err), resp)
	}

	for _, digitalRuleSet := range *digitalRuleSets {
		resources[*digitalRuleSet.Id] = &resourceExporter.ResourceMeta{BlockLabel: *digitalRuleSet.Name}
	}

	return resources, nil
}

// createOutboundDigitalruleset is used by the outbound_digitalruleset resource to create Genesys cloud outbound digitalruleset
func createOutboundDigitalruleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundDigitalrulesetProxy(sdkConfig)

	outboundDigitalruleset := getOutboundDigitalrulesetFromResourceData(d)

	log.Printf("Creating outbound digitalruleset %s", *outboundDigitalruleset.Name)
	digitalRuleSet, resp, err := proxy.createOutboundDigitalruleset(ctx, &outboundDigitalruleset)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create outbound digitalruleset: %s", err), resp)
	}

	d.SetId(*digitalRuleSet.Id)
	log.Printf("Created outbound digitalruleset %s", *digitalRuleSet.Id)
	return readOutboundDigitalruleset(ctx, d, meta)
}

// readOutboundDigitalruleset is used by the outbound_digitalruleset resource to read an outbound digitalruleset from genesys cloud
func readOutboundDigitalruleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundDigitalrulesetProxy(sdkConfig)

	log.Printf("Reading outbound digitalruleset %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		digitalRuleSet, resp, getErr := proxy.getOutboundDigitalrulesetById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read outbound digitalruleset %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read outbound digitalruleset %s: %s", d.Id(), getErr), resp))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundDigitalruleset(), constants.ConsistencyChecks(), ResourceType)

		resourcedata.SetNillableValue(d, "name", digitalRuleSet.Name)
		resourcedata.SetNillableReference(d, "contact_list_id", digitalRuleSet.ContactList)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "rules", digitalRuleSet.Rules, flattenDigitalRules)

		log.Printf("Read outbound digitalruleset %s %s", d.Id(), *digitalRuleSet.Name)
		return cc.CheckState(d)
	})
}

// updateOutboundDigitalruleset is used by the outbound_digitalruleset resource to update an outbound digitalruleset in Genesys Cloud
func updateOutboundDigitalruleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundDigitalrulesetProxy(sdkConfig)

	outboundDigitalruleset := getOutboundDigitalrulesetFromResourceData(d)

	log.Printf("Updating outbound digitalruleset %s", *outboundDigitalruleset.Name)
	digitalRuleSet, resp, err := proxy.updateOutboundDigitalruleset(ctx, d.Id(), &outboundDigitalruleset)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update outbound digitalruleset: %s", err), resp)
	}

	log.Printf("Updated outbound digitalruleset %s", *digitalRuleSet.Id)
	return readOutboundDigitalruleset(ctx, d, meta)
}

// deleteOutboundDigitalruleset is used by the outbound_digitalruleset resource to delete an outbound digitalruleset from Genesys cloud
func deleteOutboundDigitalruleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundDigitalrulesetProxy(sdkConfig)

	resp, err := proxy.deleteOutboundDigitalruleset(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete outbound digitalruleset %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getOutboundDigitalrulesetById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted outbound digitalruleset %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting outbound digitalruleset %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("outbound digitalruleset %s still exists", d.Id()), resp))
	})
}
