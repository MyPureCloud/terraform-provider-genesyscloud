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
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

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

	digitalRuleSets, _, err := proxy.getAllOutboundDigitalruleset(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get outbound digitalruleset: %v", err)
	}

	for _, digitalRuleSet := range *digitalRuleSets {
		resources[*digitalRuleSet.Id] = &resourceExporter.ResourceMeta{Name: *digitalRuleSet.Name}
	}

	return resources, nil
}

// createOutboundDigitalruleset is used by the outbound_digitalruleset resource to create Genesys cloud outbound digitalruleset
func createOutboundDigitalruleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundDigitalrulesetProxy(sdkConfig)

	outboundDigitalruleset := getOutboundDigitalrulesetFromResourceData(d)

	log.Printf("Creating outbound digitalruleset %s", *outboundDigitalruleset.Name)
	digitalRuleSet, _, err := proxy.createOutboundDigitalruleset(ctx, &outboundDigitalruleset)
	if err != nil {
		return diag.Errorf("Failed to create outbound digitalruleset: %s", err)
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
				return retry.RetryableError(fmt.Errorf("Failed to read outbound digitalruleset %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read outbound digitalruleset %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundDigitalruleset(), constants.DefaultConsistencyChecks, resourceName)

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
	digitalRuleSet, _, err := proxy.updateOutboundDigitalruleset(ctx, d.Id(), &outboundDigitalruleset)
	if err != nil {
		return diag.Errorf("Failed to update outbound digitalruleset: %s", err)
	}

	log.Printf("Updated outbound digitalruleset %s", *digitalRuleSet.Id)
	return readOutboundDigitalruleset(ctx, d, meta)
}

// deleteOutboundDigitalruleset is used by the outbound_digitalruleset resource to delete an outbound digitalruleset from Genesys cloud
func deleteOutboundDigitalruleset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundDigitalrulesetProxy(sdkConfig)

	_, err := proxy.deleteOutboundDigitalruleset(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete outbound digitalruleset %s: %s", d.Id(), err)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getOutboundDigitalrulesetById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(respCode) {
				log.Printf("Deleted outbound digitalruleset %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting outbound digitalruleset %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("outbound digitalruleset %s still exists", d.Id()))
	})
}
