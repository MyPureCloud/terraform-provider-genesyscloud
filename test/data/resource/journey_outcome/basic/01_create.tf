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
  #  Currently beta feature, can't tested
  #  associated_value_field {
  #    data_type = "Number"
  #    name      = "ItemNumber"
  #  }
}
