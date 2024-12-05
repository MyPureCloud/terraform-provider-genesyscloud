package conversations_messaging_supportedcontent_default

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_conversations_messaging_supportedcontent_default.go contains all of the methods that perform the core logic for a resource.
*/

// getAuthConversationsMessagingSupportedcontentDefault retrieves all of the conversations messaging supportedcontent default via Terraform in the Genesys Cloud and is used for the exporter
func getAuthConversationsMessagingSupportedcontentDefaults(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getConversationsMessagingSupportedcontentDefaultProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getConversationsMessagingSupportedcontentDefault(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get conversations messaging supportedcontent default: %s", err), resp)
	}

	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "supported_content_default"}

	return resources, nil
}

// createConversationsMessagingSupportedcontentDefault is used by the conversations_messaging_supportedcontent_default resource to create Genesys cloud conversations messaging supportedcontent default
func createConversationsMessagingSupportedcontentDefault(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("supported_content_default")
	return updateConversationsMessagingSupportedcontentDefault(ctx, d, meta)
}

// readConversationsMessagingSupportedcontentDefault is used by the conversations_messaging_supportedcontent_default resource to read an conversations messaging supportedcontent default from genesys cloud
func readConversationsMessagingSupportedcontentDefault(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingSupportedcontentDefaultProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingSupportedcontentDefault(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading conversations supported content default %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		supportedContentDefault, resp, err := proxy.getConversationsMessagingSupportedcontentDefault(ctx)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read conversations supported content default %s: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read conversations supported content default %s: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "content_id", supportedContentDefault.Id)

		log.Printf("Read conversations supported content default %s %s", d.Id(), *supportedContentDefault.Id)
		return cc.CheckState(d)
	})
}

// updateConversationsMessagingSupportedcontentDefault is used by the conversations_messaging_supportedcontent_default resource to update an conversations messaging supportedcontent default in Genesys Cloud
func updateConversationsMessagingSupportedcontentDefault(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingSupportedcontentDefaultProxy(sdkConfig)
	supportedContentId := d.Get("content_id").(string)

	conversationsMessagingSupportedcontentDefault := platformclientv2.Supportedcontentreference{
		Id: &supportedContentId,
	}

	log.Printf("Updating conversations messaging supportedcontent default %s", supportedContentId)
	supportedContentReference, resp, err := proxy.updateConversationsMessagingSupportedcontentDefault(ctx, d.Id(), &conversationsMessagingSupportedcontentDefault)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update conversations messaging supportedcontent default: %s", err), resp)
	}

	log.Printf("Updated conversations messaging supportedcontent default %s", *supportedContentReference.Id)
	return readConversationsMessagingSupportedcontentDefault(ctx, d, meta)
}

// deleteConversationsMessagingSupportedcontentDefault is used by the conversations_messaging_supportedcontent_default resource to delete an conversations messaging supportedcontent default from Genesys cloud
func deleteConversationsMessagingSupportedcontentDefault(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
