package outbound_contactlistfilter

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_outbound_contactlistfilter.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOutboundContactlistfilter retrieves all of the outbound contactlistfilter via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOutboundContactlistfilters(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getOutboundContactlistfilterProxy(clientConfig)

	contactListFilters, resp, err := proxy.getAllOutboundContactlistfilter(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get contact list filters error: %s", err), resp)
	}

	for _, contactListFilter := range *contactListFilters {
		resources[*contactListFilter.Id] = &resourceExporter.ResourceMeta{BlockLabel: *contactListFilter.Name}
	}

	return resources, nil
}

// createOutboundContactlistfilter is used by the outbound_contactlistfilter resource to create Genesys cloud outbound contactlistfilter
func createOutboundContactlistfilter(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundContactlistfilterProxy(sdkConfig)

	contactListFilter := getContactlistfilterFromResourceData(d)

	log.Printf("Creating Outbound Contact List Filter %s", *contactListFilter.Name)
	outboundContactListFilter, resp, err := proxy.createOutboundContactlistfilter(ctx, &contactListFilter)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create Outbound Contact List Filter %s error: %s", *contactListFilter.Name, err), resp)
	}

	d.SetId(*outboundContactListFilter.Id)

	log.Printf("Created Outbound Contact List Filter %s %s", *contactListFilter.Name, *outboundContactListFilter.Id)
	return readOutboundContactlistfilter(ctx, d, meta)
}

// readOutboundContactlistfilter is used by the outbound_contactlistfilter resource to read an outbound contactlistfilter from genesys cloud
func readOutboundContactlistfilter(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundContactlistfilterProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundContactlistfilter(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Outbound Contact List Filter %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkContactListFilter, resp, getErr := proxy.getOutboundContactlistfilterById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound Contact List Filter %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound Contact List Filter %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", sdkContactListFilter.Name)

		switch {
		case *sdkContactListFilter.SourceType == "ContactList":
			resourcedata.SetNillableReference(d, "contact_list_id", sdkContactListFilter.ContactList)
		case *sdkContactListFilter.SourceType == "ContactListTemplate":
			resourcedata.SetNillableReference(d, "contact_list_template_id", sdkContactListFilter.ContactListTemplate)
		default:
		}

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "clauses", sdkContactListFilter.Clauses, flattenContactListFilterClauses)
		resourcedata.SetNillableValue(d, "filter_type", sdkContactListFilter.FilterType)

		log.Printf("Read Outbound Contact List Filter %s %s", d.Id(), *sdkContactListFilter.Name)
		return cc.CheckState(d)
	})
}

// updateOutboundContactlistfilter is used by the outbound_contactlistfilter resource to update an outbound contactlistfilter in Genesys Cloud
func updateOutboundContactlistfilter(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundContactlistfilterProxy(sdkConfig)

	contactListFilter := getContactlistfilterFromResourceData(d)

	log.Printf("Updating Outbound Contact List Filter %s", *contactListFilter.Name)
	_, resp, err := proxy.updateOutboundContactlistfilter(ctx, d.Id(), &contactListFilter)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Outbound Contact List Filter %s error: %s", *contactListFilter.Name, err), resp)
	}

	log.Printf("Updated Outbound Contact List Filter %s", *contactListFilter.Name)
	return readOutboundContactlistfilter(ctx, d, meta)
}

// deleteOutboundContactlistfilter is used by the outbound_contactlistfilter resource to delete an outbound contactlistfilter from Genesys cloud
func deleteOutboundContactlistfilter(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundContactlistfilterProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Contact List Filter")
		resp, err := proxy.deleteOutboundContactlistfilter(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Outbound Contact List Filter %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getOutboundContactlistfilterById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Outbound Contact list filter deleted
				log.Printf("Deleted Outbound Contact List Filter %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting Outbound Contact List Filter %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Outbound Contact List Filter %s still exists", d.Id()), resp))
	})
}
