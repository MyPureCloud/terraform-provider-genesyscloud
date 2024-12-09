package conversations_messaging_supportedcontent

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
The resource_genesyscloud_conversations_messaging_supportedcontent.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthSupportedContent retrieves all of the supported content via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthSupportedContents(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getSupportedContentProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	supportedContents, resp, err := proxy.getAllSupportedContent(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get supported content: %s", err), resp)
	}

	for _, supportedContent := range *supportedContents {
		resources[*supportedContent.Id] = &resourceExporter.ResourceMeta{BlockLabel: *supportedContent.Name}
	}

	return resources, nil
}

// createSupportedContent is used by the supported_content resource to create Genesys cloud supported content
func createSupportedContent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSupportedContentProxy(sdkConfig)

	supportedContentConfig := getSupportedContentFromResourceData(d)

	log.Printf("Creating supported content %s", *supportedContentConfig.Name)
	supportedContent, resp, err := proxy.createSupportedContent(ctx, &supportedContentConfig)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create supported content: %s", err), resp)
	}

	d.SetId(*supportedContent.Id)
	log.Printf("Created supported content %s", *supportedContent.Id)
	return readSupportedContent(ctx, d, meta)
}

// readSupportedContent is used by the supported_content resource to read an supported content from genesys cloud
func readSupportedContent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSupportedContentProxy(sdkConfig)

	log.Printf("Reading supported content %s", d.Id())

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSupportedContent(), constants.ConsistencyChecks(), ResourceType)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		supportedContent, resp, getErr := proxy.getSupportedContentById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read supported content %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read supported content %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", supportedContent.Name)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "media_types", supportedContent.MediaTypes, flattenMediaTypes)
		log.Printf("Read supported content %s %s", d.Id(), *supportedContent.Name)
		return cc.CheckState(d)
	})
}

// updateSupportedContent is used by the supported_content resource to update an supported content in Genesys Cloud
func updateSupportedContent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSupportedContentProxy(sdkConfig)

	supportedContentConfig := getSupportedContentFromResourceData(d)

	log.Printf("Updating supported content %s", *supportedContentConfig.Name)
	supportedContent, resp, err := proxy.updateSupportedContent(ctx, d.Id(), &supportedContentConfig)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update supported content: %s", err), resp)
	}

	log.Printf("Updated supported content %s", *supportedContent.Id)
	return readSupportedContent(ctx, d, meta)
}

// deleteSupportedContent is used by the supported_content resource to delete an supported content from Genesys cloud
func deleteSupportedContent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSupportedContentProxy(sdkConfig)

	_, err := proxy.deleteSupportedContent(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete supported content %s: %s", d.Id(), err)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getSupportedContentById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted supported content %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting supported content %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("supported content %s still exists: %s", d.Id(), err), resp))
	})
}
