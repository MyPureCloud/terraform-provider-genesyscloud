---
page_title: "genesyscloud_routing_queue Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Routing Queue
---
# genesyscloud_routing_queue (Resource)

Genesys Cloud Routing Queue

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

- [POST /api/v2/routing/queues](https://developer.mypurecloud.com/api/rest/v2/routing/#post-api-v2-routing-queues)
- [GET /api/v2/routing/queues/{queueId}/members](https://developer.mypurecloud.com/api/rest/v2/routing/#get-api-v2-routing-queues--queueId--members)
- [GET /api/v2/routing/queues/{queueId}](https://developer.mypurecloud.com/api/rest/v2/routing/#get-api-v2-routing-queues--queueId-)
- [POST /api/v2/routing/queues/{queueId}/members](https://developer.mypurecloud.com/api/rest/v2/routing/#post-api-v2-routing-queues--queueId--members)
- [PATCH /api/v2/routing/queues/{queueId}/members/{memberId}](https://developer.mypurecloud.com/api/rest/v2/routing/#patch-api-v2-routing-queues--queueId--members--memberId-)
- [DELETE /api/v2/routing/queues/{queueId}](https://developer.mypurecloud.com/api/rest/v2/routing/#delete-api-v2-routing-queues--queueId-)
- [GET /api/v2/routing/queues/{queueId}/wrapupcodes](https://developer.mypurecloud.com/api/rest/v2/routing/#get-api-v2-routing-queues--queueId--wrapupcodes)
- [POST /api/v2/routing/queues/{queueId}/wrapupcodes](https://developer.mypurecloud.com/api/rest/v2/routing/#post-api-v2-routing-queues--queueId--wrapupcodes)
- [DELETE /api/v2/routing/queues/{queueId}/wrapupcodes/{codeId}](https://developer.mypurecloud.com/api/rest/v2/routing/#delete-api-v2-routing-queues--queueId--wrapupcodes--codeId-)

## Schema Migration: Routing Queue V1 to V2

### Migration Details

As of v1.61.0 of the provider, the Genesys Cloud Routing Queue resource type includes a schema migration that removes several vestigial attributes from the media settings blocks.

#### Removed Attributes

The following attributes have been removed from the following media settings blocks: `media_settings_call`, `media_settings_email`, `media_settings_chat`, and `media_settings_message`:

- `mode`
- `enable_auto_dial_and_end`
- `auto_dial_delay_seconds`
- `auto_end_delay_seconds`

#### Migration Process

The migration of the state will automatically occur when running terraform init with version 1.61.0 or later of the provider. The migration process:

- Preserves all other existing attributes and their values
- Removes the deprecated attributes listed above from the state
- Maintains the functionality of the queue resource
-

#### Example State Changes

Before migration:

```hcl
resource "genesyscloud_routing_queue" "example" {
  name = "Example Queue"
  media_settings_callback {
    enable_auto_answer        = false
    mode                      = "AgentFirst"
    alerting_timeout_sec      = 30
    auto_end_delay_seconds    = 300
    enable_auto_dial_and_end  = false
    service_level_duration_ms = 20000
    service_level_percentage  = 0.8
    auto_dial_delay_seconds   = 300
  }
  media_settings_chat {
    enable_auto_answer        = false
    enable_auto_dial_and_end  = false
    service_level_duration_ms = 20000
    service_level_percentage  = 0.8
    alerting_timeout_sec      = 30
  }
  media_settings_message {
    service_level_duration_ms = 20000
    service_level_percentage  = 0.8
    alerting_timeout_sec      = 30
    enable_auto_answer        = false
    enable_auto_dial_and_end  = false
  }
  media_settings_call {
    service_level_percentage  = 0.8
    alerting_timeout_sec      = 8
    enable_auto_answer        = false
    enable_auto_dial_and_end  = false
    service_level_duration_ms = 20000
  }
  media_settings_email {
    enable_auto_dial_and_end  = false
    service_level_duration_ms = 86400000
    service_level_percentage  = 0.8
    alerting_timeout_sec      = 300
    enable_auto_answer        = false
  }
  ...
}
```

After migration:

```hcl
resource "genesyscloud_routing_queue" "example" {
  name = "Example Queue"
  media_settings_callback {
    enable_auto_answer        = false
    mode                      = "AgentFirst"
    alerting_timeout_sec      = 30
    auto_end_delay_seconds    = 300
    enable_auto_dial_and_end  = false
    service_level_duration_ms = 20000
    service_level_percentage  = 0.8
    auto_dial_delay_seconds   = 300
  }
  media_settings_chat {
    service_level_duration_ms = 20000
    service_level_percentage  = 0.8
    alerting_timeout_sec      = 30
  }
  media_settings_message {
    service_level_duration_ms = 20000
    service_level_percentage  = 0.8
    alerting_timeout_sec      = 30
  }
  media_settings_call {
    service_level_percentage  = 0.8
    alerting_timeout_sec      = 8
    service_level_duration_ms = 20000
  }
  media_settings_email {
    service_level_duration_ms = 86400000
    service_level_percentage  = 0.8
    alerting_timeout_sec      = 300
  }
}
```

#### Action Required

The state will be automatically upgraded when you run terraform init with version 1.60.0 or later of the provider. After this, you will have to update your config to remove these attributes from the `media_settings_call`, `media_settings_email`, `media_settings_chat`, and `media_settings_message` config blocks as they are no longer supported.


## Example Usage

```terraform
resource "genesyscloud_routing_queue" "example_queue" {
  name                     = "Example Queue"
  division_id              = data.genesyscloud_auth_division_home.home.id
  description              = "This is an example description"
  acw_wrapup_prompt        = "MANDATORY_TIMEOUT"
  acw_timeout_ms           = 300000
  skill_evaluation_method  = "BEST"
  queue_flow_id            = genesyscloud_flow.inqueue_flow.id
  whisper_prompt_id        = genesyscloud_architect_user_prompt.welcome_greeting.id
  auto_answer_only         = true
  enable_transcription     = true
  enable_audio_monitoring  = true
  enable_manual_assignment = true
  calling_party_name       = "Example Inc."
  groups                   = [genesyscloud_group.example_group.id, genesyscloud_group.example_group2.id]

  media_settings_call {
    alerting_timeout_sec      = 30
    service_level_percentage  = 0.7
    service_level_duration_ms = 10000
  }
  routing_rules {
    operator     = "MEETS_THRESHOLD"
    threshold    = 9
    wait_seconds = 300
  }

  default_script_ids = {
    EMAIL = genesyscloud_script.example_script.id
    # CHAT  = data.genesyscloud_script.chat.id
  }

  wrapup_codes = [genesyscloud_routing_wrapupcode.win.id]
}

resource "genesyscloud_routing_queue" "example_queue2" {
  name                     = "Example Queue 2"
  division_id              = data.genesyscloud_auth_division_home.home.id
  description              = "This is an example description 2"
  acw_wrapup_prompt        = "MANDATORY_TIMEOUT"
  acw_timeout_ms           = 300000
  skill_evaluation_method  = "BEST"
  queue_flow_id            = genesyscloud_flow.inqueue_flow.id
  whisper_prompt_id        = genesyscloud_architect_user_prompt.welcome_greeting.id
  auto_answer_only         = true
  enable_transcription     = true
  enable_audio_monitoring  = true
  enable_manual_assignment = true
  calling_party_name       = "Example Inc."
  groups                   = [genesyscloud_group.example_group.id, genesyscloud_group.example_group2.id]

  media_settings_call {
    alerting_timeout_sec      = 30
    service_level_percentage  = 0.7
    service_level_duration_ms = 10000
  }
  routing_rules {
    operator     = "MEETS_THRESHOLD"
    threshold    = 9
    wait_seconds = 300
  }

  default_script_ids = {
    EMAIL = genesyscloud_script.example_script.id
    # CHAT  = data.genesyscloud_script.chat.id
  }

  wrapup_codes = [genesyscloud_routing_wrapupcode.win.id]
}

resource "genesyscloud_routing_queue" "example_queue_with_bullseye_ring" {
  name                     = "Example Queue Bullseye"
  division_id              = data.genesyscloud_auth_division_home.home.id
  description              = "This is an example description"
  acw_wrapup_prompt        = "MANDATORY_TIMEOUT"
  acw_timeout_ms           = 300000
  skill_evaluation_method  = "BEST"
  queue_flow_id            = genesyscloud_flow.inqueue_flow.id
  whisper_prompt_id        = genesyscloud_architect_user_prompt.welcome_greeting.id
  auto_answer_only         = true
  enable_transcription     = true
  enable_audio_monitoring  = true
  enable_manual_assignment = true
  calling_party_name       = "Example Inc."

  # outbound_messaging_sms_address_id = "+13179821000"
  outbound_email_address {
    domain_id = genesyscloud_routing_email_domain.example_domain_com.id
    route_id  = genesyscloud_routing_email_route.example_route.id
  }
  media_settings_call {
    alerting_timeout_sec      = 30
    service_level_percentage  = 0.7
    service_level_duration_ms = 10000
  }
  routing_rules {
    operator     = "MEETS_THRESHOLD"
    threshold    = 9
    wait_seconds = 300
  }
  bullseye_rings {
    expansion_timeout_seconds = 15.1
    skills_to_remove          = [genesyscloud_routing_skill.example_skill.id]

    member_groups {
      member_group_id   = genesyscloud_group.bullseye_rings_group1.id
      member_group_type = "GROUP"
    }
    member_groups {
      member_group_id   = genesyscloud_group.bullseye_rings_group2.id
      member_group_type = "GROUP"
    }
  }
  default_script_ids = {
    EMAIL = genesyscloud_script.example_script.id
    # CHAT  = data.genesyscloud_script.chat.id
  }
  members {
    user_id  = genesyscloud_user.queue_user3.id
    ring_num = 2
  }

  wrapup_codes = [genesyscloud_routing_wrapupcode.win.id]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Queue name.

### Optional

- `acw_timeout_ms` (Number) The amount of time the agent can stay in ACW. Only set when ACW is MANDATORY_TIMEOUT, MANDATORY_FORCED_TIMEOUT or AGENT_REQUESTED.
- `acw_wrapup_prompt` (String) This field controls how the UI prompts the agent for a wrapup (MANDATORY | OPTIONAL | MANDATORY_TIMEOUT | MANDATORY_FORCED_TIMEOUT | AGENT_REQUESTED). Defaults to `MANDATORY_TIMEOUT`.
- `agent_owned_routing` (Block List, Max: 1) Agent Owned Routing. (see [below for nested schema](#nestedblock--agent_owned_routing))
- `auto_answer_only` (Boolean) Specifies whether the configured whisper should play for all ACD calls, or only for those which are auto-answered. Defaults to `true`.
- `bullseye_rings` (Block List, Max: 5) The bullseye ring settings for the queue. (see [below for nested schema](#nestedblock--bullseye_rings))
- `calling_party_name` (String) The name to use for caller identification for outbound calls from this queue.
- `calling_party_number` (String) The phone number to use for caller identification for outbound calls from this queue.
- `canned_response_libraries` (Block List, Max: 1) Agent Owned Routing. (see [below for nested schema](#nestedblock--canned_response_libraries))
- `conditional_group_routing_rules` (Block List, Max: 5) The Conditional Group Routing settings for the queue. **Note**: conditional_group_routing_rules is deprecated in genesyscloud_routing_queue. CGR is now a standalone resource, please set ENABLE_STANDALONE_CGR in your environment variables to enable and use genesyscloud_routing_queue_conditional_group_routing (see [below for nested schema](#nestedblock--conditional_group_routing_rules))
- `default_script_ids` (Map of String) The default script IDs for each communication type. Communication types: (CALL | CALLBACK | CHAT | COBROWSE | EMAIL | MESSAGE | SOCIAL_EXPRESSION | VIDEO | SCREENSHARE)
- `description` (String) Queue description.
- `direct_routing` (Block List, Max: 1) Used by the System to set Direct Routing settings for a system Direct Routing queue. (see [below for nested schema](#nestedblock--direct_routing))
- `division_id` (String) The division to which this queue will belong. If not set, the home division will be used.
- `email_in_queue_flow_id` (String) The in-queue flow ID to use for email conversations waiting in queue.
- `enable_audio_monitoring` (Boolean) Indicates whether audio monitoring is enabled for this queue.
- `enable_manual_assignment` (Boolean) Indicates whether manual assignment is enabled for this queue. Defaults to `false`.
- `enable_transcription` (Boolean) Indicates whether voice transcription is enabled for this queue. Defaults to `false`.
- `groups` (Set of String) List of group ids assigned to the queue
- `ignore_members` (Boolean) If true, queue members will not be managed through Terraform state or API updates. This provides backwards compatibility for configurations where queue members are managed outside of Terraform.
- `last_agent_routing_mode` (String) The Last Agent Routing Mode for the queue.
- `media_settings_call` (Block List, Max: 1) Call media settings. (see [below for nested schema](#nestedblock--media_settings_call))
- `media_settings_callback` (Block List, Max: 1) Callback media settings. (see [below for nested schema](#nestedblock--media_settings_callback))
- `media_settings_chat` (Block List, Max: 1) Chat media settings. (see [below for nested schema](#nestedblock--media_settings_chat))
- `media_settings_email` (Block List, Max: 1) Email media settings. (see [below for nested schema](#nestedblock--media_settings_email))
- `media_settings_message` (Block List, Max: 1) Message media settings. (see [below for nested schema](#nestedblock--media_settings_message))
- `members` (Block Set) Users in the queue. If not set, this resource will not manage members. If a user is already assigned to this queue via a group, attempting to assign them using this field will cause an error to be thrown. (see [below for nested schema](#nestedblock--members))
- `message_in_queue_flow_id` (String) The in-queue flow ID to use for message conversations waiting in queue.
- `on_hold_prompt_id` (String) The audio to be played when calls on this queue are on hold. If not configured, the default on-hold music will play.
- `outbound_email_address` (Block List, Max: 1) The outbound email address settings for this queue. **Note**: outbound_email_address is deprecated in genesyscloud_routing_queue. OEA is now a standalone resource, please set ENABLE_STANDALONE_EMAIL_ADDRESS in your environment variables to enable and use genesyscloud_routing_queue_outbound_email_address (see [below for nested schema](#nestedblock--outbound_email_address))
- `outbound_messaging_open_messaging_recipient_id` (String) The unique ID of the outbound messaging open messaging recipient for the queue.
- `outbound_messaging_sms_address_id` (String) The unique ID of the outbound messaging SMS address for the queue.
- `outbound_messaging_whatsapp_recipient_id` (String) The unique ID of the outbound messaging whatsapp recipient for the queue.
- `peer_id` (String) The ID of an associated external queue
- `queue_flow_id` (String) The in-queue flow ID to use for call conversations waiting in queue.
- `routing_rules` (Block List, Max: 6) The routing rules for the queue, used for routing to known or preferred agents. (see [below for nested schema](#nestedblock--routing_rules))
- `scoring_method` (String) The Scoring Method for the queue. Defaults to TimestampAndPriority. Defaults to `TimestampAndPriority`.
- `skill_evaluation_method` (String) The skill evaluation method to use when routing conversations (NONE | BEST | ALL). Defaults to `ALL`.
- `skill_groups` (Set of String) List of skill group ids assigned to the queue.
- `source_queue_id` (String) The id of an existing queue to copy the settings (does not include GPR settings) from when creating a new queue.
- `suppress_in_queue_call_recording` (Boolean) Indicates whether recording in-queue calls is suppressed for this queue. Defaults to `true`.
- `teams` (Set of String) List of ids assigned to the queue
- `whisper_prompt_id` (String) The prompt ID used for whisper on the queue, if configured.
- `wrapup_codes` (Set of String) IDs of wrapup codes assigned to this queue. If not set, this resource will not manage wrapup codes.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--agent_owned_routing"></a>
### Nested Schema for `agent_owned_routing`

Optional:

- `enable_agent_owned_callbacks` (Boolean) Enable Agent Owned Callbacks
- `max_owned_callback_delay_hours` (Number) Max Owned Call Back Delay Hours >= 7
- `max_owned_callback_hours` (Number) Auto End Delay Seconds Must be >= 7


<a id="nestedblock--bullseye_rings"></a>
### Nested Schema for `bullseye_rings`

Required:

- `expansion_timeout_seconds` (Number) Seconds to wait in this ring before moving to the next.

Optional:

- `member_groups` (Block Set) (see [below for nested schema](#nestedblock--bullseye_rings--member_groups))
- `skills_to_remove` (Set of String) Skill IDs to remove on ring exit.

<a id="nestedblock--bullseye_rings--member_groups"></a>
### Nested Schema for `bullseye_rings.member_groups`

Required:

- `member_group_id` (String) ID (GUID) for Group, SkillGroup, Team
- `member_group_type` (String) The type of the member group. Accepted values: TEAM, GROUP, SKILLGROUP



<a id="nestedblock--canned_response_libraries"></a>
### Nested Schema for `canned_response_libraries`

Optional:

- `library_ids` (Set of String) Set of canned response library IDs associated with the queue. Populate this field only when the mode is set to SelectedOnly.
- `mode` (String) The association mode of canned response libraries to queue.Valid values: All, SelectedOnly, None.


<a id="nestedblock--conditional_group_routing_rules"></a>
### Nested Schema for `conditional_group_routing_rules`

Required:

- `groups` (Block Set, Min: 1) The group(s) to activate if the rule evaluates as true. (see [below for nested schema](#nestedblock--conditional_group_routing_rules--groups))

Optional:

- `condition_value` (Number) The limit value, beyond which a rule evaluates as true.
- `metric` (String) The queue metric being evaluated. Valid values: EstimatedWaitTime, ServiceLevel Defaults to `EstimatedWaitTime`.
- `operator` (String) The operator that compares the actual value against the condition value. Valid values: GreaterThan, GreaterThanOrEqualTo, LessThan, LessThanOrEqualTo.
- `queue_id` (String) The ID of the queue being evaluated for this rule. For rule 1, this is always be the current queue, so no queue id should be specified for the first rule.
- `wait_seconds` (Number) The number of seconds to wait in this rule, if it evaluates as true, before evaluating the next rule. For the final rule, this is ignored, so need not be specified. Defaults to `2`.

<a id="nestedblock--conditional_group_routing_rules--groups"></a>
### Nested Schema for `conditional_group_routing_rules.groups`

Required:

- `member_group_id` (String) ID (GUID) for Group, SkillGroup, Team
- `member_group_type` (String) The type of the member group. Accepted values: TEAM, GROUP, SKILLGROUP



<a id="nestedblock--direct_routing"></a>
### Nested Schema for `direct_routing`

Optional:

- `agent_wait_seconds` (Number) The queue default time a Direct Routing interaction will wait for an agent before it goes to configured backup. Defaults to `60`.
- `backup_queue_id` (String) Direct Routing default backup queue id (if none supplied this queue will be used as backup).
- `call_use_agent_address_outbound` (Boolean) Boolean indicating if user Direct Routing addresses should be used outbound on behalf of queue in place of Queue address for calls. Defaults to `true`.
- `email_use_agent_address_outbound` (Boolean) Boolean indicating if user Direct Routing addresses should be used outbound on behalf of queue in place of Queue address for emails. Defaults to `true`.
- `message_use_agent_address_outbound` (Boolean) Boolean indicating if user Direct Routing addresses should be used outbound on behalf of queue in place of Queue address for messages. Defaults to `true`.
- `wait_for_agent` (Boolean) Boolean indicating if Direct Routing interactions should wait for the targeted agent by default. Defaults to `false`.


<a id="nestedblock--media_settings_call"></a>
### Nested Schema for `media_settings_call`

Optional:

- `alerting_timeout_sec` (Number) Alerting timeout in seconds. Must be >= 7
- `enable_auto_answer` (Boolean) Auto-Answer for digital channels(Email, Message) Defaults to `false`.
- `service_level_duration_ms` (Number) Service Level target in milliseconds. Must be >= 1000
- `service_level_percentage` (Number) The desired Service Level. A float value between 0 and 1.
- `sub_type_settings` (Block List) Auto-Answer for digital channels(Email, Message) (see [below for nested schema](#nestedblock--media_settings_call--sub_type_settings))

<a id="nestedblock--media_settings_call--sub_type_settings"></a>
### Nested Schema for `media_settings_call.sub_type_settings`

Required:

- `enable_auto_answer` (Boolean) Indicates if auto-answer is enabled for the given media type or subtype (default is false). Subtype settings take precedence over media type settings.
- `media_type` (String) The name of the social media company



<a id="nestedblock--media_settings_callback"></a>
### Nested Schema for `media_settings_callback`

Optional:

- `alerting_timeout_sec` (Number) Alerting timeout in seconds. Must be >= 7
- `answering_machine_flow_id` (String) The inbound flow to transfer to if an answering machine is detected during the outbound call of a customer first callback when answeringMachineReactionType is set to TransferToFlow.
- `answering_machine_reaction_type` (String) The action to take if an answering machine is detected during the outbound call of a customer first callback. Valid values include: HangUp, TransferToQueue, TransferToFlow
- `auto_answer_alert_tone_seconds` (Number) How long to play the alerting tone for an auto-answer interaction.
- `auto_dial_delay_seconds` (Number) Auto Dial Delay Seconds.
- `auto_end_delay_seconds` (Number) Auto End Delay Seconds.
- `enable_auto_answer` (Boolean) Auto-Answer for digital channels(Email, Message) Defaults to `false`.
- `enable_auto_dial_and_end` (Boolean) Auto Dial and End Defaults to `false`.
- `live_voice_flow_id` (String) The inbound flow to transfer to if a live voice is detected during the outbound call of a customer first callback.
- `live_voice_reaction_type` (String) The action to take if a live voice is detected during the outbound call of a customer first callback. Valid values include: HangUp, TransferToQueue, TransferToFlow
- `manual_answer_alert_tone_seconds` (Number) How long to play the alerting tone for a manual-answer interaction.
- `mode` (String) The mode callbacks will use on this queue.
- `pacing_modifier` (Number) Controls the maximum number of outbound calls at one time when mode is CustomerFirst.
- `service_level_duration_ms` (Number) Service Level target in milliseconds. Must be >= 1000
- `service_level_percentage` (Number) The desired Service Level. A float value between 0 and 1.
- `sub_type_settings` (Block List) Auto-Answer for digital channels(Email, Message) (see [below for nested schema](#nestedblock--media_settings_callback--sub_type_settings))

<a id="nestedblock--media_settings_callback--sub_type_settings"></a>
### Nested Schema for `media_settings_callback.sub_type_settings`

Required:

- `enable_auto_answer` (Boolean) Indicates if auto-answer is enabled for the given media type or subtype (default is false). Subtype settings take precedence over media type settings.
- `media_type` (String) The name of the social media company



<a id="nestedblock--media_settings_chat"></a>
### Nested Schema for `media_settings_chat`

Optional:

- `alerting_timeout_sec` (Number) Alerting timeout in seconds. Must be >= 7
- `enable_auto_answer` (Boolean) Auto-Answer for digital channels(Email, Message) Defaults to `false`.
- `service_level_duration_ms` (Number) Service Level target in milliseconds. Must be >= 1000
- `service_level_percentage` (Number) The desired Service Level. A float value between 0 and 1.
- `sub_type_settings` (Block List) Auto-Answer for digital channels(Email, Message) (see [below for nested schema](#nestedblock--media_settings_chat--sub_type_settings))

<a id="nestedblock--media_settings_chat--sub_type_settings"></a>
### Nested Schema for `media_settings_chat.sub_type_settings`

Required:

- `enable_auto_answer` (Boolean) Indicates if auto-answer is enabled for the given media type or subtype (default is false). Subtype settings take precedence over media type settings.
- `media_type` (String) The name of the social media company



<a id="nestedblock--media_settings_email"></a>
### Nested Schema for `media_settings_email`

Optional:

- `alerting_timeout_sec` (Number) Alerting timeout in seconds. Must be >= 7
- `enable_auto_answer` (Boolean) Auto-Answer for digital channels(Email, Message) Defaults to `false`.
- `service_level_duration_ms` (Number) Service Level target in milliseconds. Must be >= 1000
- `service_level_percentage` (Number) The desired Service Level. A float value between 0 and 1.
- `sub_type_settings` (Block List) Auto-Answer for digital channels(Email, Message) (see [below for nested schema](#nestedblock--media_settings_email--sub_type_settings))

<a id="nestedblock--media_settings_email--sub_type_settings"></a>
### Nested Schema for `media_settings_email.sub_type_settings`

Required:

- `enable_auto_answer` (Boolean) Indicates if auto-answer is enabled for the given media type or subtype (default is false). Subtype settings take precedence over media type settings.
- `media_type` (String) The name of the social media company



<a id="nestedblock--media_settings_message"></a>
### Nested Schema for `media_settings_message`

Optional:

- `alerting_timeout_sec` (Number) Alerting timeout in seconds. Must be >= 7
- `enable_auto_answer` (Boolean) Auto-Answer for digital channels(Email, Message) Defaults to `false`.
- `service_level_duration_ms` (Number) Service Level target in milliseconds. Must be >= 1000
- `service_level_percentage` (Number) The desired Service Level. A float value between 0 and 1.
- `sub_type_settings` (Block List) Auto-Answer for digital channels(Email, Message) (see [below for nested schema](#nestedblock--media_settings_message--sub_type_settings))

<a id="nestedblock--media_settings_message--sub_type_settings"></a>
### Nested Schema for `media_settings_message.sub_type_settings`

Required:

- `enable_auto_answer` (Boolean) Indicates if auto-answer is enabled for the given media type or subtype (default is false). Subtype settings take precedence over media type settings.
- `media_type` (String) The name of the social media company



<a id="nestedblock--members"></a>
### Nested Schema for `members`

Required:

- `user_id` (String) User ID

Optional:

- `ring_num` (Number) Ring number between 1 and 6 for this user in the queue. Defaults to `1`.


<a id="nestedblock--outbound_email_address"></a>
### Nested Schema for `outbound_email_address`

Required:

- `domain_id` (String) Unique ID of the email domain. e.g. "test.example.com"
- `route_id` (String) Unique ID of the email route.


<a id="nestedblock--routing_rules"></a>
### Nested Schema for `routing_rules`

Optional:

- `operator` (String) Matching operator (MEETS_THRESHOLD | ANY). MEETS_THRESHOLD matches any agent with a score at or above the rule's threshold. ANY matches all specified agents, regardless of score. Defaults to `MEETS_THRESHOLD`.
- `threshold` (Number) Threshold required for routing attempt (generally an agent score). Ignored for operator ANY.
- `wait_seconds` (Number) Seconds to wait in this rule before moving to the next. Defaults to `5`.

