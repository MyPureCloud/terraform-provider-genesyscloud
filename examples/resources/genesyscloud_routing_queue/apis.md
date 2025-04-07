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
