resource "genesyscloud_quality_forms_evaluation" "example_evaluation_form" {
  name      = "Example Evaluation Form"
  published = true
  question_groups {
    name                       = "Question Group 1"
    default_answers_to_highest = true
    default_answers_to_na      = true
    na_enabled                 = true
    weight                     = 1
    manual_weight              = true
    questions {
      text              = "Yes or no question."
      help_text         = "A simple question."
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
    weight = 1.5
    visibility_condition {
      combining_operation = "AND"
      predicates          = ["/form/questionGroup/0/question/0/answer/0"]
    }
    questions {
      text = "Multiple Choice Question."
      answer_options {
        text  = "1"
        value = 1
      }
      answer_options {
        text  = "2"
        value = 2
      }
      answer_options {
        text  = "3"
        value = 3
      }
    }
  }
}
