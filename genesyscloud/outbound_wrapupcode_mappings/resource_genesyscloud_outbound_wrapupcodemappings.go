package outbound_wrapupcode_mappings

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

// getOutboundWrapupCodeMappings is used by the exporter to return all wrapupcode mappings
func getOutboundWrapupCodeMappings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getOutboundWrapupCodeMappingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getAllOutboundWrapupCodeMappings(ctx)
	if err != nil {
		if gcloud.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, diag.Errorf("Failed to get wrap-up code mappings: %v", err)
	}

	resources["0"] = &resourceExporter.ResourceMeta{Name: "wrapupcodemappings"}
	return resources, nil
}

// createOutboundWrapUpCodeMappings is used to create the Terraform backing state associated with an outbound wrapup code mapping
func createOutboundWrapUpCodeMappings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Outbound Wrap-up Code Mappings")
	d.SetId("wrapupcodemappings")
	return updateOutboundWrapUpCodeMappings(ctx, d, meta)
}

// updateOutboundWrapUpCodeMappings is sued to update the Terraform backing state associated with an outbound wrapup code mapping
func updateOutboundWrapUpCodeMappings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundWrapupCodeMappingsProxy(sdkConfig)

	log.Printf("Updating Outbound Wrap-up Code Mappings")
	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		wrapupCodeMappings, resp, err := proxy.getAllOutboundWrapupCodeMappings(ctx)
		if err != nil {
			return resp, diag.Errorf("failed to read wrap-up code mappings: %s", err)
		}

		wrapupCodeUpdate := platformclientv2.Wrapupcodemapping{
			DefaultSet: lists.BuildSdkStringListFromInterfaceArray(d, "default_set"),
			Mapping:    buildWrapupCodeMappings(d),
			Version:    wrapupCodeMappings.Version,
		}
		_, _, err = proxy.updateOutboundWrapUpCodeMappings(ctx, wrapupCodeUpdate)
		if err != nil {
			return resp, diag.Errorf("failed to update wrap-up code mappings: %s", err)
		}
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Print("Updated Outbound Wrap-up Code Mappings")
	return readOutboundWrapUpCodeMappings(ctx, d, meta)
}

// readOutboundWrapUpCodeMappings reads the current state of the outboundwrapupcode mapping object
func readOutboundWrapUpCodeMappings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundWrapupCodeMappingsProxy(sdkConfig)

	log.Printf("Reading Outbound Wrap-up Code Mappings")

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkWrapupCodeMappings, resp, err := proxy.getAllOutboundWrapupCodeMappings(ctx)
		if err != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Outbound Wrap-up Code Mappings: %s", err))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Outbound Wrap-up Code Mappings: %s", err))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundWrapUpCodeMappings())

		// Match new random ordering of list returned from API
		if sdkWrapupCodeMappings.DefaultSet != nil {
			defaultSet := make([]string, 0)
			schemaDefaultSet := d.Get("default_set").([]interface{})
			for _, v := range schemaDefaultSet {
				defaultSet = append(defaultSet, v.(string))
			}
			if lists.AreEquivalent(defaultSet, *sdkWrapupCodeMappings.DefaultSet) {
				d.Set("default_set", defaultSet)
			} else {
				d.Set("default_set", lists.StringListToInterfaceList(*sdkWrapupCodeMappings.DefaultSet))
			}
		}

		if sdkWrapupCodeMappings.Mapping != nil {
			d.Set("mappings", flattenOutboundWrapupCodeMappings(d, sdkWrapupCodeMappings))
		}

		log.Print("Read Outbound Wrap-up Code Mappings")
		return cc.CheckState()
	})
}

// deleteOutboundWrapUpCodeMappings This a no up to satisfy the deletion of outbound wrapping resource
func deleteOutboundWrapUpCodeMappings(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Does not delete the wrap-up code mappings. This resource will just no longer manage them.
	return nil
}
