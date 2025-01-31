variable "element_first_id" {
  type        = string
  default     = "ac6c61b5-1cd4-4c6e-a8a5-edb74d9117eb"
  description = "ID of the first element"
}

variable "element_second_id" {
  type        = string
  default     = "7e78450e-1fae-4e15-84b1-e7ffc74fb961"
  description = "ID of the second element"
}

resource "genesyscloud_journey_views" "journey_view" {
  name     = "Cx as Code sample"
  duration = "P1Y"
  elements {
    id   = var.element_first_id
    name = "Bot Start"
    attributes {
      type   = "Event"
      id     = "94677135-0e97-1a87-44f0-ee7f5506bd09"
      source = "Voice"
    }
    display_attributes {
      x   = 556
      y   = 246
      col = 0
    }
    filter {
      type = "And"
      predicates {

        dimension = "channels"
        values    = ["CALLBACK", "CALL"]
        operator  = "Matches"
        no_value  = false
      }
    }
    followed_by {
      id = var.element_second_id
      constraint_within {
        unit  = "Minutes"
        value = 60
      }
    }
  }
  elements {
    id   = var.element_second_id
    name = "Bot End"
    attributes {
      type   = "Event"
      id     = "b56bd9cc-74a9-1f0e-9ecf-494b4024a3be"
      source = "Voice"
    }
    display_attributes {
      x   = 956
      y   = 246
      col = 1
    }
    filter {
      type = "And"
      predicates {
        dimension = "channels"
        values    = ["CALLBACK", "CALL"]
        operator  = "Matches"
        no_value  = false
      }
      number_predicates {
        dimension = "turnCount"
        operator  = "Matches"
        no_value  = false
        range {
          gt {
            number = 1.0
          }
        }
      }
    }
  }
  charts {
    name          = "New Chart"
    version       = 1
    group_by_time = "Day"
    metrics {
      id         = "36b3f717-f309-425d-9d5e-3af2071ad0a4"
      element_id = var.element_second_id
      aggregate  = "CustomerCount"
    }
    display_attributes {
      var_type    = "Column"
      show_legend = true
    }
    group_by_max = 28
  }
  charts {
    name    = "New Chart B"
    version = 1
    group_by_attributes {
      element_id = var.element_second_id
      attribute  = "vendor"
    }
    metrics {
      id         = "723981e9-2810-42d1-94ef-42bbb4bbc134"
      element_id = var.element_second_id
      aggregate  = "EventCount"
    }
    display_attributes {
      var_type    = "Column"
      show_legend = true
    }
    group_by_max = 10
  }

}