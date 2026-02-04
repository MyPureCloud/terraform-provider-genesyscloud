package conversations_settings

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

// getAllConversationsSettings retrieves all conversations settings for export
func getAllConversationsSettings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getConversationsSettingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getConversationsSettings(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get %s due to error: %s", ResourceType, err), resp)
	}

	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "conversations_settings"}
	return resources, nil
}

// createConversationsSettings creates the conversations settings resource
func createConversationsSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating Conversations Settings")
	d.SetId("settings")
	return updateConversationsSettings(ctx, d, meta)
}

// readConversationsSettings reads the conversations settings from the API
func readConversationsSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsSettingsProxy(sdkConfig)

	log.Printf("Reading conversations settings")

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		settings, resp, getErr := proxy.getConversationsSettings(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Conversations Settings %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Conversations Settings %s | error: %s", d.Id(), getErr), resp))
		}

		log.Printf("DEBUG: API returned: CommunicationBasedACW=%v, IncludeNonAgentConversationSummary=%v", 
			*settings.CommunicationBasedACW, *settings.IncludeNonAgentConversationSummary)

		// Map API response to Terraform state
		resourcedata.SetNillableValue(d, "communication_based_acw", settings.CommunicationBasedACW)
		resourcedata.SetNillableValue(d, "include_non_agent_conversation_summary", settings.IncludeNonAgentConversationSummary)
		resourcedata.SetNillableValue(d, "allow_callback_queue_selection", settings.AllowCallbackQueueSelection)
		resourcedata.SetNillableValue(d, "callbacks_inherit_routing_from_inbound_call", settings.CallbacksInheritRoutingFromInboundCall)
		resourcedata.SetNillableValue(d, "complete_acw_when_agent_transitions_offline", settings.CompleteAcwWhenAgentTransitionsOffline)
		resourcedata.SetNillableValue(d, "total_active_callback", settings.TotalActiveCallback)

		log.Printf("Read Conversations Settings")
		return nil
	})
}

// updateConversationsSettings updates the conversations settings via the API
func updateConversationsSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsSettingsProxy(sdkConfig)

	log.Printf("Updating Conversations Settings")

	// Build update request from Terraform config
	// Note: Use d.Get() instead of d.GetOk() for booleans because GetOk returns false
	// for both "not set" and "set to false", making it impossible to set false values
	communicationBasedACW := d.Get("communication_based_acw").(bool)
	includeNonAgentConversationSummary := d.Get("include_non_agent_conversation_summary").(bool)
	allowCallbackQueueSelection := d.Get("allow_callback_queue_selection").(bool)
	callbacksInheritRoutingFromInboundCall := d.Get("callbacks_inherit_routing_from_inbound_call").(bool)
	completeAcwWhenAgentTransitionsOffline := d.Get("complete_acw_when_agent_transitions_offline").(bool)
	totalActiveCallback := d.Get("total_active_callback").(bool)

	update := platformclientv2.Settings{
		CommunicationBasedACW:                  &communicationBasedACW,
		IncludeNonAgentConversationSummary:     &includeNonAgentConversationSummary,
		AllowCallbackQueueSelection:            &allowCallbackQueueSelection,
		CallbacksInheritRoutingFromInboundCall: &callbacksInheritRoutingFromInboundCall,
		CompleteAcwWhenAgentTransitionsOffline: &completeAcwWhenAgentTransitionsOffline,
		TotalActiveCallback:                    &totalActiveCallback,
	}

	log.Printf("DEBUG: Sending update with values: CommunicationBasedACW=%v, IncludeNonAgentConversationSummary=%v", 
		*update.CommunicationBasedACW, *update.IncludeNonAgentConversationSummary)

	// PATCH returns no body, so we just check for errors
	resp, err := proxy.updateConversationsSettings(ctx, &update)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update conversations settings %s error: %s", d.Id(), err), resp)
	}

	log.Printf("DEBUG: PATCH successful, response status: %d", resp.StatusCode)

	// Wait for API to propagate changes (same as routing_settings)
	time.Sleep(5 * time.Second)

	log.Printf("Updated Conversations Settings")

	// Read back the settings to refresh state
	return readConversationsSettings(ctx, d, meta)
}

// deleteConversationsSettings handles deletion (no-op for singleton resources)
func deleteConversationsSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Singleton settings resources typically don't support deletion
	// The settings remain in the organization with their current values
	log.Printf("Deleting (no-op) Conversations Settings")
	return nil
}
