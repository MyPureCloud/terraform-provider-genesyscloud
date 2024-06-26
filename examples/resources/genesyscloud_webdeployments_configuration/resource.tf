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