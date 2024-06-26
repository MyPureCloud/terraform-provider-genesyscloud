---
page_title: "genesyscloud_webdeployments_configuration Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Web Deployment Configuration
---
# genesyscloud_webdeployments_configuration (Resource)

Genesys Cloud Web Deployment Configuration

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/webdeployments/configurations](https://developer.dev-genesys.cloud/api/rest/v2/webdeployments/#get-api-v2-webdeployments-configurations)
* [POST /api/v2/webdeployments/configurations](https://developer.dev-genesys.cloud/api/rest/v2/webdeployments/#post-api-v2-webdeployments-configurations)
* [DELETE /api/v2/webdeployments/configurations/{configurationId}](https://developer.dev-genesys.cloud/api/rest/v2/webdeployments/#delete-api-v2-webdeployments-configurations--configurationId-)
* [GET /api/v2/webdeployments/configurations/{configurationId}/versions](https://developer.dev-genesys.cloud/api/rest/v2/webdeployments/#get-api-v2-webdeployments-configurations--configurationId--versions)
* [GET /api/v2/webdeployments/configurations/{configurationId}/versions/draft](https://developer.dev-genesys.cloud/api/rest/v2/webdeployments/#get-api-v2-webdeployments-configurations--configurationId--versions-draft)
* [PUT /api/v2/webdeployments/configurations/{configurationId}/versions/draft](https://developer.dev-genesys.cloud/api/rest/v2/webdeployments/#put-api-v2-webdeployments-configurations--configurationId--versions-draft)
* [POST /api/v2/webdeployments/configurations/{configurationId}/versions/draft/publish](https://developer.dev-genesys.cloud/api/rest/v2/webdeployments/#post-api-v2-webdeployments-configurations--configurationId--versions-draft-publish)
* [GET /api/v2/webdeployments/configurations/{configurationId}/versions/{versionId}](https://developer.dev-genesys.cloud/api/rest/v2/webdeployments/#get-api-v2-webdeployments-configurations--configurationId--versions--versionId-)

## Example Usage

```terraform
resource "genesyscloud_webdeployments_configuration" "exampleConfiguration" {
  name                  = "Example Web Deployment Configuration"
  description           = "This example configuration shows how to define a full web deployment configuration"
  languages             = ["en-us", "ja"]
  default_language      = "en-us"
  headless_mode_enabled = true
  custom_i18n_labels {
    language = "en-us"
    localized_labels {
      key   = "MessengerHomeHeaderTitle"
      value = "Custom Header Title"
    }
    localized_labels {
      key   = "MessengerHomeHeaderSubTitle"
      value = "Custom Header Subtitle"
    }
  }
  position {
    alignment    = "Auto"
    side_space   = 10
    bottom_space = 20
  }
  messenger {
    enabled = true
    launcher_button {
      visibility = "OnDemand"
    }
    home_screen {
      enabled  = true
      logo_url = "https://my-domain/images/my-logo.png"
    }
    styles {
      primary_color = "#B0B0B0"
    }
    file_upload {
      mode {
        file_types       = ["image/png"]
        max_file_size_kb = 256
      }
      mode {
        file_types       = ["image/jpeg"]
        max_file_size_kb = 128
      }
    }
    apps {
      conversations {
        enabled                     = true
        show_agent_typing_indicator = true
        show_user_typing_indicator  = true
        auto_start_enabled          = true
        markdown_enabled            = true
        conversation_disconnect {
          enabled = true
          type    = "Send"
        }
        conversation_clear_enabled = true
        humanize {
          enabled = true
          bot {
            name       = "Marvin"
            avatar_url = "https://my-domain-example.net/images/marvin.png"
          }
        }
      }
      knowledge {
        enabled           = true
        knowledge_base_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      }
    }
  }
  cobrowse {
    enabled                = true
    allow_agent_control    = true
    allow_agent_navigation = true
    channels               = ["Webmessaging", "Voice"]
    mask_selectors         = [".my-class", "#my-id"]
    readonly_selectors     = [".my-class", "#my-id"]
    pause_criteria = {
      url_fragment = "/sensitive"
      condition    = "includes"
    }
  }
  journey_events {
    enabled                   = true
    excluded_query_parameters = ["marketingCampaign"]

    pageview_config = "Auto"

    click_event {
      selector   = ".promo-button"
      event_name = "promo:interest"
    }
    click_event {
      selector   = ".cancel-button"
      event_name = "service:cancel"
    }

    form_track_event {
      selector                     = ".interest-submit"
      form_name                    = "interest"
      capture_data_on_form_abandon = true
      capture_data_on_form_submit  = false
    }

    form_track_event {
      selector                     = ".feedback-submit"
      form_name                    = "feedback"
      capture_data_on_form_abandon = false
      capture_data_on_form_submit  = true
    }

    idle_event {
      event_name         = "idle:short"
      idle_after_seconds = 30
    }

    idle_event {
      event_name         = "idle:long"
      idle_after_seconds = 120
    }

    in_viewport_event {
      selector   = ".promo-banner"
      event_name = "promo:visible"
    }

    in_viewport_event {
      selector   = ".call-to-action"
      event_name = "action:encouraged"
    }

    scroll_depth_event {
      event_name = "scroll:half"
      percentage = 50
    }

    scroll_depth_event {
      event_name = "scroll:footer"
      percentage = 90
    }
  }
  authentication_settings {
    enabled        = true
    integration_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `default_language` (String) The default language to use for the configuration.
- `languages` (List of String) A list of languages supported on the configuration.
- `name` (String) Deployment name

### Optional

- `authentication_settings` (Block List, Max: 1) Settings for authenticated webdeployments. (see [below for nested schema](#nestedblock--authentication_settings))
- `cobrowse` (Block List, Max: 1) Settings concerning cobrowse (see [below for nested schema](#nestedblock--cobrowse))
- `custom_i18n_labels` (Block List) The localization settings for homescreen app (see [below for nested schema](#nestedblock--custom_i18n_labels))
- `description` (String) Deployment description
- `headless_mode_enabled` (Boolean) Headless Mode Support which Controls UI components. When enabled, native UI components will be disabled and allows for custom-built UI.
- `journey_events` (Block List, Max: 1) Settings concerning journey events (see [below for nested schema](#nestedblock--journey_events))
- `messenger` (Block List, Max: 1) Settings concerning messenger (see [below for nested schema](#nestedblock--messenger))
- `position` (Block List, Max: 1) Settings concerning position (see [below for nested schema](#nestedblock--position))
- `status` (String) The current status of the deployment. Valid values: Pending, Active, Inactive, Error, Deleting.
- `support_center` (Block List, Max: 1) Settings concerning knowledge portal (previously support center) (see [below for nested schema](#nestedblock--support_center))

### Read-Only

- `id` (String) The ID of this resource.
- `version` (String) The version of the configuration.

<a id="nestedblock--authentication_settings"></a>
### Nested Schema for `authentication_settings`

Required:

- `enabled` (Boolean) Indicate if these auth is required for this deployment. If, for example, this flag is set to true then webmessaging sessions can not send messages unless the end-user is authenticated.
- `integration_id` (String) The integration identifier which contains the auth settings required on the deployment.


<a id="nestedblock--cobrowse"></a>
### Nested Schema for `cobrowse`

Optional:

- `allow_agent_control` (Boolean) Whether agent can take control over customer's screen or not
- `allow_agent_navigation` (Boolean) Whether agent can use navigation feature over customer's screen or not
- `channels` (List of String) List of channels through which cobrowse is available (for now only Webmessaging and Voice)
- `enabled` (Boolean) Whether or not cobrowse is enabled
- `mask_selectors` (List of String) List of CSS selectors which should be masked when screen sharing is active
- `pause_criteria` (Block List) Pause criteria that will pause cobrowse if some of them are met in the user's URL (see [below for nested schema](#nestedblock--cobrowse--pause_criteria))
- `readonly_selectors` (List of String) List of CSS selectors which should be read-only when screen sharing is active

<a id="nestedblock--cobrowse--pause_criteria"></a>
### Nested Schema for `cobrowse.pause_criteria`

Required:

- `condition` (String) The condition to be applied to the `url_fragment`. Conditions are 'includes', 'does_not_include', 'starts_with', 'ends_with', 'equals'
- `url_fragment` (String) A string representing a part of the URL that, when matched according to the specified condition, will trigger a pause in the cobrowse session



<a id="nestedblock--custom_i18n_labels"></a>
### Nested Schema for `custom_i18n_labels`

Optional:

- `language` (String) Language of localized labels in homescreen app (eg. en-us, de-de)
- `localized_labels` (Block List) Contains localized labels used in homescreen app (see [below for nested schema](#nestedblock--custom_i18n_labels--localized_labels))

<a id="nestedblock--custom_i18n_labels--localized_labels"></a>
### Nested Schema for `custom_i18n_labels.localized_labels`

Required:

- `key` (String) Contains localized label key used in messenger homescreen
- `value` (String) Contains localized label value used in messenger homescreen



<a id="nestedblock--journey_events"></a>
### Nested Schema for `journey_events`

Optional:

- `click_event` (Block List) Details about a selector event trigger (see [below for nested schema](#nestedblock--journey_events--click_event))
- `enabled` (Boolean) Whether or not journey event collection is enabled Defaults to `true`.
- `excluded_query_parameters` (List of String) List of parameters to be excluded from the query string
- `form_track_event` (Block List) Details about a forms tracking event trigger (see [below for nested schema](#nestedblock--journey_events--form_track_event))
- `idle_event` (Block List) Details about an idle event trigger (see [below for nested schema](#nestedblock--journey_events--idle_event))
- `in_viewport_event` (Block List) Details about a selector event trigger (see [below for nested schema](#nestedblock--journey_events--in_viewport_event))
- `pageview_config` (String) Controls how the pageview events are tracked.Valid values: Auto, Once, Off
- `scroll_depth_event` (Block List) Details about a scroll percentage event trigger (see [below for nested schema](#nestedblock--journey_events--scroll_depth_event))
- `search_query_parameters` (List of String) List of query parameters used for search (e.g. 'q')
- `should_keep_url_fragment` (Boolean) Whether or not to keep the URL fragment

<a id="nestedblock--journey_events--click_event"></a>
### Nested Schema for `journey_events.click_event`

Required:

- `event_name` (String) Name of event triggered when element matching selector is interacted with
- `selector` (String) Element that triggers event


<a id="nestedblock--journey_events--form_track_event"></a>
### Nested Schema for `journey_events.form_track_event`

Required:

- `capture_data_on_form_abandon` (Boolean) Whether to capture the form data in the form abandoned event
- `capture_data_on_form_submit` (Boolean) Whether to capture the form data in the form submitted event
- `form_name` (String) Prefix for the form submitted or abandoned event name
- `selector` (String) Form element that triggers the form submitted or abandoned event


<a id="nestedblock--journey_events--idle_event"></a>
### Nested Schema for `journey_events.idle_event`

Required:

- `event_name` (String) Name of event triggered after period of inactivity

Optional:

- `idle_after_seconds` (Number) Number of seconds of inactivity before an event is triggered


<a id="nestedblock--journey_events--in_viewport_event"></a>
### Nested Schema for `journey_events.in_viewport_event`

Required:

- `event_name` (String) Name of event triggered when element matching selector is interacted with
- `selector` (String) Element that triggers event


<a id="nestedblock--journey_events--scroll_depth_event"></a>
### Nested Schema for `journey_events.scroll_depth_event`

Required:

- `event_name` (String) Name of event triggered after scrolling to the specified percentage
- `percentage` (Number) Percentage of a webpage at which an event is triggered



<a id="nestedblock--messenger"></a>
### Nested Schema for `messenger`

Optional:

- `apps` (Block List, Max: 1) The apps embedded in the messenger (see [below for nested schema](#nestedblock--messenger--apps))
- `enabled` (Boolean) Whether or not messenger is enabled
- `file_upload` (Block List, Max: 1) File upload settings for messenger (see [below for nested schema](#nestedblock--messenger--file_upload))
- `home_screen` (Block List, Max: 1) The settings for the home screen (see [below for nested schema](#nestedblock--messenger--home_screen))
- `launcher_button` (Block List, Max: 1) The settings for the launcher button (see [below for nested schema](#nestedblock--messenger--launcher_button))
- `styles` (Block List, Max: 1) The style settings for messenger (see [below for nested schema](#nestedblock--messenger--styles))

<a id="nestedblock--messenger--apps"></a>
### Nested Schema for `messenger.apps`

Optional:

- `conversations` (Block List, Max: 1) Conversation settings that handles chats within the messenger (see [below for nested schema](#nestedblock--messenger--apps--conversations))
- `knowledge` (Block List, Max: 1) The knowledge base config for messenger (see [below for nested schema](#nestedblock--messenger--apps--knowledge))

<a id="nestedblock--messenger--apps--conversations"></a>
### Nested Schema for `messenger.apps.conversations`

Optional:

- `auto_start_enabled` (Boolean) The auto start for the messenger conversation
- `conversation_clear_enabled` (Boolean) The conversation clear settings for the messenger app
- `conversation_disconnect` (Block List, Max: 1) The conversation disconnect for the messenger app (see [below for nested schema](#nestedblock--messenger--apps--conversations--conversation_disconnect))
- `enabled` (Boolean) The toggle to enable or disable conversations
- `humanize` (Block List, Max: 1) The humanize conversations settings for the messenger app (see [below for nested schema](#nestedblock--messenger--apps--conversations--humanize))
- `markdown_enabled` (Boolean) The markdown for the messenger app
- `show_agent_typing_indicator` (Boolean) The toggle to enable or disable typing indicator for messenger
- `show_user_typing_indicator` (Boolean) The toggle to enable or disable typing indicator for messenger

<a id="nestedblock--messenger--apps--conversations--conversation_disconnect"></a>
### Nested Schema for `messenger.apps.conversations.conversation_disconnect`

Optional:

- `enabled` (Boolean) whether or not conversation disconnect setting is enabled
- `type` (String) Conversation disconnect type


<a id="nestedblock--messenger--apps--conversations--humanize"></a>
### Nested Schema for `messenger.apps.conversations.humanize`

Optional:

- `bot` (Block List, Max: 1) Bot messenger profile setting (see [below for nested schema](#nestedblock--messenger--apps--conversations--humanize--bot))
- `enabled` (Boolean) Whether or not humanize conversations setting is enabled

<a id="nestedblock--messenger--apps--conversations--humanize--bot"></a>
### Nested Schema for `messenger.apps.conversations.humanize.bot`

Optional:

- `avatar_url` (String) The avatar URL of the bot
- `name` (String) The name of the bot




<a id="nestedblock--messenger--apps--knowledge"></a>
### Nested Schema for `messenger.apps.knowledge`

Optional:

- `enabled` (Boolean) whether or not knowledge base is enabled
- `knowledge_base_id` (String) The knowledge base for messenger



<a id="nestedblock--messenger--file_upload"></a>
### Nested Schema for `messenger.file_upload`

Optional:

- `mode` (Block List) The list of supported file upload modes (see [below for nested schema](#nestedblock--messenger--file_upload--mode))

<a id="nestedblock--messenger--file_upload--mode"></a>
### Nested Schema for `messenger.file_upload.mode`

Optional:

- `file_types` (List of String) A list of supported content types for uploading files.Valid values: image/jpeg, image/gif, image/png
- `max_file_size_kb` (Number) The maximum file size for file uploads in kilobytes. Default is 10240 (10 MB)



<a id="nestedblock--messenger--home_screen"></a>
### Nested Schema for `messenger.home_screen`

Optional:

- `enabled` (Boolean) Whether or not home screen is enabled
- `logo_url` (String) URL for custom logo to appear in home screen


<a id="nestedblock--messenger--launcher_button"></a>
### Nested Schema for `messenger.launcher_button`

Optional:

- `visibility` (String) The visibility settings for the button.Valid values: On, Off, OnDemand


<a id="nestedblock--messenger--styles"></a>
### Nested Schema for `messenger.styles`

Optional:

- `primary_color` (String) The primary color of messenger in hexadecimal



<a id="nestedblock--position"></a>
### Nested Schema for `position`

Optional:

- `alignment` (String) The alignment for position
- `bottom_space` (Number) The bottomspace value for position
- `side_space` (Number) The sidespace value for position


<a id="nestedblock--support_center"></a>
### Nested Schema for `support_center`

Required:

- `enabled` (Boolean) Whether or not knowledge portal (previously support center) is enabled

Optional:

- `custom_messages` (Block List) Customizable display texts for knowledge portal (see [below for nested schema](#nestedblock--support_center--custom_messages))
- `enabled_categories` (Block List) Featured categories for knowledge portal (previously support center) home screen (see [below for nested schema](#nestedblock--support_center--enabled_categories))
- `feedback_enabled` (Boolean) Whether or not requesting customer feedback on article content and article search results is enabled
- `knowledge_base_id` (String) The knowledge base for knowledge portal (previously support center)
- `router_type` (String) Router type for knowledge portal
- `screens` (Block List) Available screens for the knowledge portal with its modules (see [below for nested schema](#nestedblock--support_center--screens))
- `style_setting` (Block List, Max: 1) Style attributes for knowledge portal (previously support center) (see [below for nested schema](#nestedblock--support_center--style_setting))

<a id="nestedblock--support_center--custom_messages"></a>
### Nested Schema for `support_center.custom_messages`

Required:

- `default_value` (String) Default value for the custom message
- `type` (String) The custom message type. (Welcome or Fallback)


<a id="nestedblock--support_center--enabled_categories"></a>
### Nested Schema for `support_center.enabled_categories`

Required:

- `category_id` (String) The knowledge base category id

Optional:

- `image_uri` (String) Source URL for the featured category


<a id="nestedblock--support_center--screens"></a>
### Nested Schema for `support_center.screens`

Required:

- `module_settings` (Block List, Min: 1) Module settings for the screen, valid modules for each screenType: Home: Search, Categories, TopViewedArticles; Category: Search, Categories; SearchResults: Search, Results; Article: Search, Article; (see [below for nested schema](#nestedblock--support_center--screens--module_settings))
- `type` (String) The type of the screen

<a id="nestedblock--support_center--screens--module_settings"></a>
### Nested Schema for `support_center.screens.module_settings`

Required:

- `enabled` (Boolean) Whether or not knowledge portal (previously support center) screen module is enabled
- `type` (String) Screen module type

Optional:

- `compact_category_module_template_active` (Boolean) Whether the Support Center Compact Category Module Template is active or not
- `detailed_category_module_template` (Block List, Max: 1) Detailed category module template settings (see [below for nested schema](#nestedblock--support_center--screens--module_settings--detailed_category_module_template))

<a id="nestedblock--support_center--screens--module_settings--detailed_category_module_template"></a>
### Nested Schema for `support_center.screens.module_settings.detailed_category_module_template`

Required:

- `active` (Boolean) Whether the Support Center Detailed Category Module Template is active or not
- `sidebar_enabled` (Boolean) Whether the Support Center Detailed Category Module Sidebar is active or not




<a id="nestedblock--support_center--style_setting"></a>
### Nested Schema for `support_center.style_setting`

Optional:

- `global_style_setting` (Block List, Max: 1) Knowledge portal (previously support center) global customizations (see [below for nested schema](#nestedblock--support_center--style_setting--global_style_setting))
- `hero_style_setting` (Block List, Max: 1) Knowledge portal (previously support center) hero customizations (see [below for nested schema](#nestedblock--support_center--style_setting--hero_style_setting))

<a id="nestedblock--support_center--style_setting--global_style_setting"></a>
### Nested Schema for `support_center.style_setting.global_style_setting`

Required:

- `background_color` (String) Global background color, in hexadecimal format, eg #ffffff
- `font_family` (String) Global font family
- `primary_color` (String) Global primary color, in hexadecimal format, eg #ffffff
- `primary_color_dark` (String) Global dark primary color, in hexadecimal format, eg #ffffff
- `primary_color_light` (String) Global light primary color, in hexadecimal format, eg #ffffff
- `text_color` (String) Global text color, in hexadecimal format, eg #ffffff


<a id="nestedblock--support_center--style_setting--hero_style_setting"></a>
### Nested Schema for `support_center.style_setting.hero_style_setting`

Required:

- `background_color` (String) Background color for hero section, in hexadecimal format, eg #ffffff
- `image_uri` (String) Background image for hero section
- `text_color` (String) Text color for hero section, in hexadecimal format, eg #ffffff

