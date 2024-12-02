package outbound_wrapupcode_mappings

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

// getOutboundWrapupCodeMappings is used by the exporter to return all wrapupcode mappings
func getOutboundWrapupCodeMappings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getOutboundWrapupCodeMappingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getAllOutboundWrapupCodeMappings(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s due to error: %s", ResourceType, err), resp)
	}

	_, resp, err = proxy.getAllWrapupCodes(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s due to error: %s", ResourceType, err), resp)
	}
	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "wrapupcodemappings"}
	return resources, nil
}

// createOutboundWrapUpCodeMappings is used to create the Terraform backing state associated with an outbound wrapup code mapping
func createOutboundWrapUpCodeMappings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Outbound Wrap-up Code Mappings")
	d.SetId("wrapupcodemappings")
	return updateOutboundWrapUpCodeMappings(ctx, d, meta)
}

// readOutboundWrapUpCodeMappings reads the current state of the outboundwrapupcode mapping object
func readOutboundWrapUpCodeMappings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundWrapupCodeMappingsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundWrapUpCodeMappings(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Outbound Wrap-up Code Mappings")

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkWrapupCodeMappings, resp, err := proxy.getAllOutboundWrapupCodeMappings(ctx)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound Wrap-up Code Mappings: %s", err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound Wrap-up Code Mappings: %s", err), resp))
		}

		wrapupCodes, resp, err := proxy.getAllWrapupCodes(ctx)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to get wrapup codes: %s", err), resp))
		}

		resourcedata.SetNillableValue(d, "default_set", sdkWrapupCodeMappings.DefaultSet)

		existingWrapupCodes := make([]string, 0)
		for _, wuc := range *wrapupCodes {
			existingWrapupCodes = append(existingWrapupCodes, *wuc.Id)
		}

		if sdkWrapupCodeMappings.Mapping != nil {
			_ = d.Set("mappings", flattenOutboundWrapupCodeMappings(d, sdkWrapupCodeMappings, &existingWrapupCodes))
		}

		log.Print("Read Outbound Wrap-up Code Mappings")
		return cc.CheckState(d)
	})
}

// updateOutboundWrapUpCodeMappings is sued to update the Terraform backing state associated with an outbound wrapup code mapping
func updateOutboundWrapUpCodeMappings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundWrapupCodeMappingsProxy(sdkConfig)

	log.Printf("Updating Outbound Wrap-up Code Mappings")
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		wrapupCodeMappings, resp, err := proxy.getAllOutboundWrapupCodeMappings(ctx)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get  wrap-up code mappings error: %s", err), resp)
		}

		wrapupCodeUpdate := platformclientv2.Wrapupcodemapping{
			DefaultSet: lists.BuildSdkStringList(d, "default_set"),
			Mapping:    buildWrapupCodeMappings(d),
			Version:    wrapupCodeMappings.Version,
		}
		_, resp, err = proxy.updateOutboundWrapUpCodeMappings(ctx, wrapupCodeUpdate)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update wrap-up code mappings %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}
	log.Print("Updated Outbound Wrap-up Code Mappings")
	return readOutboundWrapUpCodeMappings(ctx, d, meta)
}

// deleteOutboundWrapUpCodeMappings This a no up to satisfy the deletion of outbound wrapping resource
func deleteOutboundWrapUpCodeMappings(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Does not delete the wrap-up code mappings. This resource will just no longer manage them.
	return nil
}
