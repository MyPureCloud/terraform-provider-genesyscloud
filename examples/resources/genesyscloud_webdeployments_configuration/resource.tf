resource "genesyscloud_webdeployments_configuration" "exampleConfiguration" {
  name             = "Example Web Deployment Configuration"
  description      = "This example configuration shows how to define a full web deployment configuration"
  languages        = ["en-us", "ja"]
  default_language = "en-us"
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
  }
  cobrowse {
    enabled             = true
    allow_agent_control = true
    channels            = ["Webmessaging", "Voice"]
    mask_selectors      = [".my-class", "#my-id"]
    readonly_selectors  = [".my-class", "#my-id"]
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
}