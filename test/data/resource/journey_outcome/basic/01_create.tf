resource "genesyscloud_journey_outcome" "terraform_test_-TEST-CASE-" {
  is_active    = true
  display_name = "terraform_test_-TEST-CASE-"
  description  = "test description of journey outcome"
  is_positive  = true
  journey {
    patterns {
      criteria {
        key                = "page.title"
        values             = ["Title"]
        operator           = "notEqual"
        should_ignore_case = true
      }
      count        = 1
      stream_type  = "Web"
      session_type = "web"
    }
  }
  #  Associated_value_field needs `eventtypes` to be created, which is a feature coming soon. More details available here:
  #  https://developer.genesys.cloud/commdigital/digital/webmessaging/journey/eventtypes
  #  https://all.docs.genesys.com/ATC/Current/AdminGuide/Custom_sessions

  #  associated_value_field {
  #    data_type = "Number"
  #    name      = "ItemNumber"
  #  }
}
