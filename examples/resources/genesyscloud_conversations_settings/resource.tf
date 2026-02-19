resource "genesyscloud_conversations_settings" "example" {
  allow_callback_queue_selection              = true
  callbacks_inherit_routing_from_inbound_call = true
  communication_based_acw                     = true
  complete_acw_when_agent_transitions_offline = true
  include_non_agent_conversation_summary      = true
  total_active_callback                       = true
}
