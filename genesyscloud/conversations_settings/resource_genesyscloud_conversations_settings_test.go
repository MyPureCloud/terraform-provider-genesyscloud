package conversations_settings

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceConversationsSettings(t *testing.T) {
	var (
		resourceLabel = "conversations_settings"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with specific settings
				Config: generateConversationsSettingsResource(
					resourceLabel,
					"true",  // communication_based_acw
					"true",  // complete_acw_when_agent_transitions_offline
					"false", // include_non_agent_conversation_summary
					"true",  // allow_callback_queue_selection
					"false", // callbacks_inherit_routing_from_inbound_call
					"true",  // total_active_callback
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "communication_based_acw", "true"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "complete_acw_when_agent_transitions_offline", "true"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "include_non_agent_conversation_summary", "false"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "allow_callback_queue_selection", "true"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "callbacks_inherit_routing_from_inbound_call", "false"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "total_active_callback", "true"),
				),
			},
			{
				// Update settings
				Config: generateConversationsSettingsResource(
					resourceLabel,
					"false", // communication_based_acw
					"false", // complete_acw_when_agent_transitions_offline
					"true",  // include_non_agent_conversation_summary
					"false", // allow_callback_queue_selection
					"true",  // callbacks_inherit_routing_from_inbound_call
					"false", // total_active_callback
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "communication_based_acw", "false"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "complete_acw_when_agent_transitions_offline", "false"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "include_non_agent_conversation_summary", "true"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "allow_callback_queue_selection", "false"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "callbacks_inherit_routing_from_inbound_call", "true"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "total_active_callback", "false"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_conversations_settings." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyConversationsSettingsDestroyed,
	})
}

func TestAccResourceConversationsSettingsMinimal(t *testing.T) {
	var (
		resourceLabel = "conversations_settings_minimal"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with only the two required fields for customer
				Config: generateConversationsSettingsResourceMinimal(
					resourceLabel,
					"true", // communication_based_acw
					"true", // complete_acw_when_agent_transitions_offline
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "communication_based_acw", "true"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "complete_acw_when_agent_transitions_offline", "true"),
				),
			},
			{
				// Update the two fields
				Config: generateConversationsSettingsResourceMinimal(
					resourceLabel,
					"false", // communication_based_acw
					"false", // complete_acw_when_agent_transitions_offline
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "communication_based_acw", "false"),
					resource.TestCheckResourceAttr("genesyscloud_conversations_settings."+resourceLabel, "complete_acw_when_agent_transitions_offline", "false"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_conversations_settings." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyConversationsSettingsDestroyed,
	})
}

func generateConversationsSettingsResource(
	resourceLabel string,
	communicationBasedAcw string,
	completeAcwWhenAgentTransitionsOffline string,
	includeNonAgentConversationSummary string,
	allowCallbackQueueSelection string,
	callbacksInheritRoutingFromInboundCall string,
	totalActiveCallback string,
) string {
	return fmt.Sprintf(`
resource "genesyscloud_conversations_settings" "%s" {
	communication_based_acw                        = %s
	complete_acw_when_agent_transitions_offline    = %s
	include_non_agent_conversation_summary         = %s
	allow_callback_queue_selection                 = %s
	callbacks_inherit_routing_from_inbound_call    = %s
	total_active_callback                          = %s
}
`, resourceLabel,
		communicationBasedAcw,
		completeAcwWhenAgentTransitionsOffline,
		includeNonAgentConversationSummary,
		allowCallbackQueueSelection,
		callbacksInheritRoutingFromInboundCall,
		totalActiveCallback,
	)
}

func generateConversationsSettingsResourceMinimal(
	resourceLabel string,
	communicationBasedAcw string,
	completeAcwWhenAgentTransitionsOffline string,
) string {
	return fmt.Sprintf(`
resource "genesyscloud_conversations_settings" "%s" {
	communication_based_acw                        = %s
	complete_acw_when_agent_transitions_offline    = %s
}
`, resourceLabel,
		communicationBasedAcw,
		completeAcwWhenAgentTransitionsOffline,
	)
}

func testVerifyConversationsSettingsDestroyed(state *terraform.State) error {
	// Singleton resources are not actually destroyed, they just remain with their current values
	// This function verifies that the resource is removed from Terraform state
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_conversations_settings" {
			continue
		}

		// If we get here, the resource still exists in state (which shouldn't happen after destroy)
		return fmt.Errorf("conversations_settings still exists in state")
	}

	// Resource successfully removed from state
	return nil
}
