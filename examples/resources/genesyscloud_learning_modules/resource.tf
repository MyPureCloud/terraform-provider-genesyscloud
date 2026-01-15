resource "genesyscloud_learning_modules" "example_learning_module" {
  name                    = "Example name"
  description             = "Example description"
  completion_time_in_days = 10
  type                    = "Native"
  cover_art_id            = "05cde3c7-d48e-41af-a7a5-44f4f01acefb"
  length_in_minutes       = 15
  excluded_from_catalog   = false
  external_id             = ""
  enforce_content_order   = false
  is_published            = false
  inform_steps {
    type         = "Url"
    name         = "Example name"
    value        = "https://www.example.com"
    order        = 1
    display_name = "Example display name"
    description  = "Example description"
  }
  inform_steps {
    type         = "Richtext"
    name         = "Example name"
    value        = "<b>Example</b>"
    order        = 2
    display_name = "Example display name"
    description  = "Example description"
  }
  assessment_form {
    pass_percent = 80
    question_groups {
      name                       = "Question Group 1"
      type                       = "questionGroup"
      default_answers_to_highest = true
      default_answers_to_na      = true
      na_enabled                 = true
      weight                     = 1
      manual_weight              = true
      questions {
        text              = "Yes or no question."
        help_text         = "A simple question."
        type              = "multipleChoiceQuestion"
        na_enabled        = true
        comments_required = false
        is_kill           = true
        is_critical       = true
        answer_options {
          text  = "Yes"
          value = 1
        }
        answer_options {
          text  = "No"
          value = 0
        }
      }
      questions {
        text      = "Question with visibility condition"
        help_text = "A lot of fields are optional like this help text."
        type      = "multipleChoiceQuestion"
        answer_options {
          text  = "Yes"
          value = 1
        }
        answer_options {
          text  = "No"
          value = 0
        }
        visibility_condition {
          combining_operation = "AND"
          predicates          = ["/form/questionGroup/0/question/0/answer/0"]
        }
      }
    }
    question_groups {
      name   = "Question Group with Visibility Condition"
      type   = "questionGroup"
      weight = 1.5
      visibility_condition {
        combining_operation = "AND"
        predicates          = ["/form/questionGroup/0/question/0/answer/0"]
      }
      questions {
        text                    = "Free Text Question."
        type                    = "freeTextQuestion"
        max_response_characters = 100
      }
    }
  }
  review_assessment_results {
    by_assignees = true
    by_viewers   = true
  }
  auto_assign {
    enabled = true
    rule_id = "be53142d-3389-482e-aba5-92a13c671af1"
  }
}